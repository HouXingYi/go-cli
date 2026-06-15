package app

import (
	"context"

	"ziniao/internal/httpclient"
)

type AuthVerification = httpclient.AuthVerifyResponse
type RequestResult = httpclient.RunResponse

func (s Service) VerifyAuth(ctx context.Context) (AuthVerification, error) {
	if err := s.requireToken(); err != nil {
		return AuthVerification{}, err
	}
	return s.api.VerifyAuth(ctx)
}

func (s Service) RunRequest(ctx context.Context) (RequestResult, error) {
	if err := s.requireToken(); err != nil {
		return RequestResult{}, err
	}
	return s.api.RunRequest(ctx)
}
