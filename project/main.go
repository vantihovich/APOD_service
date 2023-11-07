package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"

	cnfg "github.com/vantihovich/APOD_service/project/configuration"
	"github.com/vantihovich/APOD_service/project/handlers"
	postgr "github.com/vantihovich/APOD_service/project/postgres"
)

func main() {
	srv := &http.Server{ //TODO implement configs for project
		Addr:    ":3000",
		Handler: service(),
	}

	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sig
		shutDownCtx, shutDownCnclFunc := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			<-shutDownCtx.Done()
			if shutDownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		if err := srv.Shutdown(shutDownCtx); err != nil {
			log.WithError(err).Fatal("Could not shutdown the server")
		}

		serverStopCtx()
		shutDownCnclFunc()
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.WithError(err).Fatal("An error starting server")
	}

	<-serverCtx.Done()
}

func service() http.Handler {
	log.Info("Configs loading")

	cfgDB, err := cnfg.LoadDB()
	if err != nil {
		log.WithError(err).Fatal("Failed to load DB config")
	}

	log.Info("сonnecting to DB")
	db := postgr.New(cfgDB)
	if err := db.Open(); err != nil {
		log.WithError(err).Fatal("Failed to establish connection with DB")
	}

	repoHandler := handlers.NewRepoHandler(&db)

	requestURL := fmt.Sprintf("https://api.nasa.gov/planetary/apod?api_key=DEMO_KEY")
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		log.WithError(err).Fatal("Failed to create API request")
	}

	go worker(req, repoHandler)

	r := chi.NewRouter()

	r.Route("/album", func(r chi.Router) {
		r.Get("/all", repoHandler.GetAll)
		r.Get("/date", repoHandler.GetByDate)
	})

	return r
}

func worker(req *http.Request, db *handlers.RepoHandler) {
	for {
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.WithError(err).Info("Failed to make the request to API")
		}

		respBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.WithError(err).Info("Failed to parse API response")
		}

		err = db.WriteNew(respBody)
		if err != nil {
			log.WithError(err).Info("Failed to write API response to DB")
		}
		log.Info("wrote API response to db")

		timer := time.NewTimer(24 * time.Hour)
		<-timer.C
	}
}
