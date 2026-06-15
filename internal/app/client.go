package app

import (
	"context"
	"strings"

	"ziniao/internal/apperr"
)

type APIClient interface {
	VerifyAuth(ctx context.Context) (AuthVerification, error)
	RunRequest(ctx context.Context) (RequestResult, error)
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
		return apperr.New(apperr.KindConfig, "token is required", "pass --token or set ZINIAO_TOKEN.")
	}
	return nil
}
