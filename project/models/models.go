package models

import (
	"context"
)

type Response struct {
	Date        string `json:"date"`
	Title       string `json:"title"`
	Url         string `json:"url"`
	Explanation string `json:"explanation"`
	Img         []byte `json:"img"`
}

type ApiResponse struct {
	Date            string `json:"date"`
	Explanation     string `json:"explanation"`
	HDurl           string `json:"hdurl"`
	Media_type      string `json:"media_type"`
	Service_version string `json:"service_version"`
	Title           string `json:"title"`
	Url             string `json:"url"`
}

type Repository interface {
	Get(ctx context.Context) ([]*Response, error)
	GetWithDate(ctx context.Context, date string) (*Response, error)
	Write(string, string, string, string, []byte) error
}
