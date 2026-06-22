package app

import (
	"context"
	"strings"

	"ziniao/internal/apperr"
	"ziniao/internal/httpclient"
)

type APIClient interface {
	Request(ctx context.Context, options httpclient.RequestOptions) (httpclient.Response, error)
}

type Service struct {
	api   APIClient
	token string
}

func NewService(api APIClient, token string) Service {
	return Service{
		api:   api,
		token: strings.TrimSpace(token),
	}
}

func (s Service) requireToken() error {
	if s.token == "" {
		return apperr.New(apperr.KindConfig, "token is required", "set ZINIAO_TOKEN.")
	}
	return nil
}
