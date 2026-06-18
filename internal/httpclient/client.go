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

type RequestOptions struct {
	Method  string
	Path    string
	Query   url.Values
	Headers http.Header
	Body    json.RawMessage
}

type Response struct {
	StatusCode int             `json:"statusCode"`
	Status     string          `json:"status"`
	Headers    http.Header     `json:"headers,omitempty"`
	Body       json.RawMessage `json:"body,omitempty"`
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

func (c *Client) Request(ctx context.Context, options RequestOptions) (Response, error) {
	method := strings.ToUpper(strings.TrimSpace(options.Method))
	if method == "" {
		return Response{}, apperr.New(apperr.KindConfig, "method is required", "use a method like GET, POST, PUT, or DELETE.")
	}
	path := strings.TrimSpace(options.Path)
	if path == "" || !strings.HasPrefix(path, "/") {
		return Response{}, apperr.New(apperr.KindConfig, "path is invalid", "use an absolute backend path like /api/user/list.")
	}

	var reader io.Reader
	if len(options.Body) > 0 {
		if !json.Valid(options.Body) {
			return Response{}, apperr.New(apperr.KindConfig, "body is invalid json", "pass --body with a valid JSON object or array.")
		}
		reader = bytes.NewReader(options.Body)
	}

	requestURL := c.resolveWithQuery(path, options.Query)
	req, err := http.NewRequestWithContext(ctx, method, requestURL, reader)
	if err != nil {
		return Response{}, apperr.Wrap(apperr.KindAPI, "failed to create request", "check base_url and request path.", err)
	}

	req.Header.Set("Accept", "application/json")
	for key, values := range options.Headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	if reader != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
			return Response{}, apperr.Wrap(apperr.KindNetwork, "request timed out", "increase --timeout or check network connectivity.", err)
		}
		return Response{}, apperr.Wrap(apperr.KindNetwork, "request failed", "check base_url and network connectivity.", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, apperr.Wrap(apperr.KindAPI, "failed to read response", "check whether the API response is available.", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return Response{}, classifyStatusBody(resp, body)
	}

	return Response{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Headers:    resp.Header.Clone(),
		Body:       json.RawMessage(body),
	}, nil
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

func (c *Client) resolveWithQuery(path string, query url.Values) string {
	ref := &url.URL{Path: path}
	resolved := c.baseURL.ResolveReference(ref)
	if len(query) > 0 {
		resolved.RawQuery = query.Encode()
	}
	return resolved.String()
}

func classifyStatus(resp *http.Response) error {
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		body = nil
	}
	return classifyStatusBody(resp, body)
}

func classifyStatusBody(resp *http.Response, body []byte) error {
	message := responseMessage(resp, body)
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return apperr.New(apperr.KindAuth, message, "check whether the token is valid or expired.")
	case http.StatusForbidden:
		return apperr.New(apperr.KindAuth, message, "check whether the token has permission for this API.")
	default:
		return apperr.New(apperr.KindAPI, message, "check the API response and request parameters.")
	}
}

func responseMessage(resp *http.Response, body []byte) string {
	if len(body) == 0 {
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
