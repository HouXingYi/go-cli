package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"ziniao/internal/apperr"
	"ziniao/internal/config"
)

type HTTPBackend struct {
	baseURL    *url.URL
	authKey    string
	httpClient *http.Client
}

func NewHTTPBackend(cfg config.Config) *HTTPBackend {
	parsed, _ := url.Parse(strings.TrimSpace(cfg.ProxyBaseURL))
	return &HTTPBackend{
		baseURL: parsed,
		authKey: strings.TrimSpace(cfg.AuthKey),
		httpClient: &http.Client{
			Timeout: config.DefaultTimeout,
		},
	}
}

func (b *HTTPBackend) Proxy(ctx context.Context, provider string, method, path string, query url.Values, body json.RawMessage) (json.RawMessage, error) {
	proxyBody := map[string]interface{}{
		"method": strings.ToUpper(strings.TrimSpace(method)),
	}
	if queryMap := valuesToMap(query); len(queryMap) > 0 {
		proxyBody["query"] = queryMap
	}
	if bodyMap := rawMessageToMap(body); len(bodyMap) > 0 {
		proxyBody["body"] = bodyMap
	}

	requestPath := "/cli/" + strings.Trim(strings.TrimSpace(provider), "/") + "/" + trimPath(path)
	return b.do(ctx, http.MethodPost, requestPath, nil, proxyBody)
}

func (b *HTTPBackend) Catalog(ctx context.Context, module, business, api string, full bool) (json.RawMessage, error) {
	requestPath := "/cli-api"
	if module != "" {
		requestPath += "/" + strings.Trim(strings.TrimSpace(module), "/")
	}

	query := url.Values{}
	if business != "" {
		query.Set("business", business)
	}
	if api != "" {
		query.Set("api", api)
	}
	if full {
		query.Set("full", "true")
	}

	return b.do(ctx, http.MethodGet, requestPath, query, nil)
}

func (b *HTTPBackend) do(ctx context.Context, method, path string, query url.Values, body interface{}) (json.RawMessage, error) {
	if b.baseURL == nil || b.baseURL.Scheme == "" || b.baseURL.Host == "" {
		return nil, apperr.New(apperr.KindConfig, "proxy base url is invalid", "set VENDOR_PROXY_BASE.")
	}

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, apperr.Wrap(apperr.KindAPI, "failed to encode request body", "check request parameters.", err)
		}
		reader = bytes.NewReader(payload)
	}

	ref := &url.URL{Path: path}
	resolved := b.baseURL.ResolveReference(ref)
	if len(query) > 0 {
		resolved.RawQuery = query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, resolved.String(), reader)
	if err != nil {
		return nil, apperr.Wrap(apperr.KindAPI, "failed to create request", "check VENDOR_PROXY_BASE and request path.", err)
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if b.authKey != "" {
		req.Header.Set("Authorization", "Bearer "+b.authKey)
	}

	resp, err := b.httpClient.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
			return nil, apperr.Wrap(apperr.KindNetwork, "request timed out", "check network connectivity.", err)
		}
		return nil, apperr.Wrap(apperr.KindNetwork, "request failed", "check VENDOR_PROXY_BASE and network connectivity.", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apperr.Wrap(apperr.KindAPI, "failed to read response", "check whether the API response is available.", err)
	}

	return parseEnvelope(resp.StatusCode, responseBody)
}
