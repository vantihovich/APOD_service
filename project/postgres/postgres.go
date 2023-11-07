package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	log "github.com/sirupsen/logrus"

	config "github.com/vantihovich/APOD_service/project/configuration"
	models "github.com/vantihovich/APOD_service/project/models"
)

type DB struct {
	pool *pgxpool.Pool
	cfg  string
}

func New(cfg config.App) (db DB) {
	db.cfg = fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s",
		cfg.Db.User,
		cfg.Db.Password,
		cfg.Db.Host,
		cfg.Db.Port,
		cfg.Db.Name)
	return db
}

func (db *DB) Open() error {
	pool, err := pgxpool.Connect(context.Background(), db.cfg)
	if err != nil {
		log.WithError(err).Error("Unable to connect to database")
		return err
	}
	log.Info("Successfully connected to DB")
	db.pool = pool
	return nil
}

func (db *DB) Get(ctx context.Context) ([]*models.Response, error) {
	var pods []*models.Response
	var pod *models.Response = &models.Response{}

	stmnt := `SELECT date, title, url, explanation, img FROM pods`

	r, err := db.pool.Query(ctx, stmnt)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.WithField("POD not found", err).Debug(
				"Valid error when worker hasn't processed data yet")
			return nil, errors.New("No rows in return")
		}
		log.WithError(err).Error(
			"err executing or parsing the request of POD by date to DB")
		return nil, err
	}
	defer r.Close()

	for r.Next() {

		err = r.Scan(
			&pod.Date,
			&pod.Title,
			&pod.Url,
			&pod.Explanation,
			&pod.Img)
		if err != nil {
			log.WithField("Pod not parsed", err).Debug("err scanning pod fields")
			return nil, errors.New("error in scanning response")
		}

		pods = append(pods, pod)
	}

	// check for errors from iterating over rows.Next()
	if err = r.Err(); err != nil {
		log.WithField("rows not parsed", err).Debug("err scanning rows")
		return nil, errors.New("error in iterating for scanning response")
	}

	return pods, nil
}

func (db *DB) GetWithDate(ctx context.Context, date string) (
	*models.Response,
	error) {
	var pod *models.Response = &models.Response{}

	stmnt := `SELECT date, title, url, explanation FROM pods WHERE date=$1`

	err := db.pool.QueryRow(ctx, stmnt, date).Scan(
		&pod.Date,
		&pod.Title,
		&pod.Url,
		&pod.Explanation,
		&pod.Img)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.WithField("POD not found", err).Debug(
				"Valid error when worker hasn't processed data yet")
			return nil, errors.New("No rows in return")
		}
		log.WithError(err).Error(
			"err executing or parsing the request of POD by date to DB")
		return nil, err
	}

	return pod, nil
}

func (db *DB) Write(date, title, url, explanation string, data []byte) error {
	ctx := context.Background()

	stmnt := `INSERT INTO pods (date, title, url, explanation, img) 
			  VALUES ($1, $2, $3, $4, $5)`

	_, err := db.pool.Exec(
		ctx,
		stmnt,
		date,
		title,
		url,
		explanation,
		data)

	if err != nil {
		log.WithError(err).Error("err executing the DB request to add new entry")
		return err
	}

	return nil
}
