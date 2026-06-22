package backend

import (
	"context"
	"encoding/json"
	"net/url"

	"ziniao/internal/config"
)

type Backend interface {
	Proxy(ctx context.Context, provider string, method, path string, query url.Values, body json.RawMessage) (json.RawMessage, error)
	Catalog(ctx context.Context, module, business, api string, full bool) (json.RawMessage, error)
}

func New(cfg config.Config) Backend {
	if cfg.UseRemoteBackend() {
		return NewHTTPBackend(cfg)
	}
	return NewMockBackend()
}
