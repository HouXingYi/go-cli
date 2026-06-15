package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ziniao/internal/apperr"
)

const (
	authVerifyPath = "/auth/verify"
	requestRunPath = "/request/run"
)

type Options struct {
	BaseURL    string
	Token      string
	Timeout    time.Duration
	HTTPClient *http.Client
}

type Client struct {
	baseURL    *url.URL
	token      string
	httpClient *http.Client
}

type AuthVerifyResponse struct {
	Message string `json:"message"`
	Status  string `json:"status,omitempty"`
}

type RunResponse struct {
	Message string      `json:"message"`
	Result  interface{} `json:"result,omitempty"`
}

func New(options Options) (*Client, error) {
	baseURL, err := url.Parse(options.BaseURL)
	if err != nil || baseURL.Scheme == "" || baseURL.Host == "" {
		return nil, apperr.New(apperr.KindConfig, "base_url is invalid", "use an absolute URL like https://api.example.com.")
	}

	httpClient := options.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: options.Timeout,
		}
	}
	if httpClient.Timeout == 0 && options.Timeout > 0 {
		copyClient := *httpClient
		copyClient.Timeout = options.Timeout
		httpClient = &copyClient
	}

	return &Client{
		baseURL:    baseURL,
		token:      strings.TrimSpace(options.Token),
		httpClient: httpClient,
	}, nil
}

func (c *Client) VerifyAuth(ctx context.Context) (AuthVerifyResponse, error) {
	var response AuthVerifyResponse
	if err := c.do(ctx, http.MethodGet, authVerifyPath, nil, &response); err != nil {
		return AuthVerifyResponse{}, err
	}
	if response.Message == "" {
		response.Message = "Auth verified successfully."
	}
	return response, nil
}

func (c *Client) RunRequest(ctx context.Context) (RunResponse, error) {
	var response RunResponse
	if err := c.do(ctx, http.MethodGet, requestRunPath, nil, &response); err != nil {
		return RunResponse{}, err
	}
	if response.Message == "" {
		response.Message = "Request completed successfully."
	}
	return response, nil
}

func (c *Client) do(ctx context.Context, method, path string, body interface{}, dest interface{}) error {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return apperr.Wrap(apperr.KindAPI, "failed to encode request body", "check request parameters.", err)
		}
		reader = bytes.NewReader(payload)
	}

	requestURL := c.resolve(path)
	req, err := http.NewRequestWithContext(ctx, method, requestURL, reader)
	if err != nil {
		return apperr.Wrap(apperr.KindAPI, "failed to create request", "check base_url and request path.", err)
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
			return apperr.Wrap(apperr.KindNetwork, "request timed out", "increase --timeout or check network connectivity.", err)
		}
		return apperr.Wrap(apperr.KindNetwork, "request failed", "check base_url and network connectivity.", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return classifyStatus(resp)
	}

	if dest == nil || resp.Body == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(dest); err != nil && !errors.Is(err, io.EOF) {
		return apperr.Wrap(apperr.KindAPI, "failed to parse response", "check whether the API returns valid JSON.", err)
	}

	return nil
}

func (c *Client) resolve(path string) string {
	ref := &url.URL{Path: path}
	return c.baseURL.ResolveReference(ref).String()
}

func classifyStatus(resp *http.Response) error {
	message := responseMessage(resp)
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return apperr.New(apperr.KindAuth, message, "check whether the token is valid or expired.")
	case http.StatusForbidden:
		return apperr.New(apperr.KindAuth, message, "check whether the token has permission for this API.")
	default:
		return apperr.New(apperr.KindAPI, message, "check the API response and request parameters.")
	}
}

func responseMessage(resp *http.Response) string {
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil || len(body) == 0 {
		return fmt.Sprintf("api returned %s", resp.Status)
	}

	var payload struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &payload); err == nil {
		if payload.Message != "" {
			return payload.Message
		}
		if payload.Error != "" {
			return payload.Error
		}
	}

	return fmt.Sprintf("api returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
}
