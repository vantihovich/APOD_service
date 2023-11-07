package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/vantihovich/APOD_service/project/models"
)

type RepoHandler struct {
	repo models.Repository
}

func NewRepoHandler(rep models.Repository) *RepoHandler {
	return &RepoHandler{
		repo: rep,
	}
}

type DateRequest struct {
	Date string `json:"date"`
}

func (h *RepoHandler) WriteNew(data []byte) error {
	//an attempt to add new entry

	apiResp := models.ApiResponse{}
	json.Unmarshal(data, &apiResp)

	picRequestURL := fmt.Sprintf(apiResp.Url)
	picReq, err := http.NewRequest(http.MethodGet, picRequestURL, nil)
	if err != nil {
		log.WithError(err).Fatal("Failed to create API request")
	}

	picResp, err := http.DefaultClient.Do(picReq)
	if err != nil {
		log.WithError(err).Info("Failed to make the request to get picture")
	}

	pic, err := ioutil.ReadAll(picResp.Body)
	if err != nil {
		log.WithError(err).Info("Failed to parse picture response")
	}

	err = h.repo.Write(
		apiResp.Date,
		apiResp.Title,
		apiResp.Url,
		apiResp.Explanation,
		pic)
	if err != nil {
		log.WithError(err).Info("error occurred when adding POD to DB")
	}

	log.Debug("POD added succesfully")
	return err
}

func (h *RepoHandler) GetAll(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	response, err := h.repo.Get(ctx)
	if err != nil {
		log.WithError(err).Info("DB request of all PODs returned error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Debug("Retrieved all PODs succesfully")

	b, err := json.Marshal(response)
	if err != nil {
		log.WithError(err).Info("error marshalling the response")
	}

	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func (h *RepoHandler) GetByDate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	date := r.URL.Query().Get("date")

	response, err := h.repo.GetWithDate(ctx, date)
	if err != nil {
		log.WithError(err).Info("DB request of POD by date returned error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Debug("Retrieved POD by date succesfully")

	b, err := json.Marshal(response)
	if err != nil {
		log.WithError(err).Info("error marshalling the response")
	}

	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
