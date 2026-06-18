package app

import (
	"context"

	"ziniao/internal/httpclient"
)

type HTTPRequest = httpclient.RequestOptions
type HTTPResponse = httpclient.Response

func (s Service) HTTPRequest(ctx context.Context, request HTTPRequest) (HTTPResponse, error) {
	if err := s.requireToken(); err != nil {
		return HTTPResponse{}, err
	}
	return s.api.Request(ctx, request)
}
