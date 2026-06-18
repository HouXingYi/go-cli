package httpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"ziniao/internal/apperr"
)

func TestRequestSendsBearerTokenAndOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %q", r.Method)
		}
		if r.URL.Path != "/api/user/create" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("page") != "1" {
			t.Fatalf("query page = %q", r.URL.Query().Get("page"))
		}
		if r.Header.Get("X-Test") != "yes" {
			t.Fatalf("X-Test = %q", r.Header.Get("X-Test"))
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("Authorization = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client, err := New(Options{
		BaseURL: server.URL,
		Token:   "test-token",
		Timeout: time.Second,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result, err := client.Request(context.Background(), RequestOptions{
		Method: "POST",
		Path:   "/api/user/create",
		Query:  url.Values{"page": []string{"1"}},
		Headers: http.Header{
			"X-Test": []string{"yes"},
		},
		Body: json.RawMessage(`{"name":"test"}`),
	})
	if err != nil {
		t.Fatalf("Request() error = %v", err)
	}
	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d", result.StatusCode)
	}
	if string(result.Body) != `{"ok":true}` {
		t.Fatalf("Body = %s", string(result.Body))
	}
}

func TestRequestHandlesServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"server failed"}`))
	}))
	defer server.Close()

	client, err := New(Options{
		BaseURL: server.URL,
		Token:   "test-token",
		Timeout: time.Second,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = client.Request(context.Background(), RequestOptions{Method: "GET", Path: "/api/user/list"})
	if err == nil {
		t.Fatal("Request() error = nil, want error")
	}
	if got := apperr.Kind(err); got != apperr.KindAPI {
		t.Fatalf("error kind = %q", got)
	}
}

func TestRequestHandlesUnauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer server.Close()

	client, err := New(Options{
		BaseURL: server.URL,
		Token:   "bad-token",
		Timeout: time.Second,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = client.Request(context.Background(), RequestOptions{Method: "GET", Path: "/api/user/list"})
	if err == nil {
		t.Fatal("Request() error = nil, want error")
	}
	if got := apperr.Kind(err); got != apperr.KindAuth {
		t.Fatalf("error kind = %q", got)
	}
}

func TestRequestRejectsInvalidJSONBody(t *testing.T) {
	client, err := New(Options{
		BaseURL: "https://api.example.com",
		Token:   "test-token",
		Timeout: time.Second,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = client.Request(context.Background(), RequestOptions{
		Method: "POST",
		Path:   "/api/user/create",
		Body:   json.RawMessage(`{"name"`),
	})
	if err == nil {
		t.Fatal("Request() error = nil, want error")
	}
	if got := apperr.Kind(err); got != apperr.KindConfig {
		t.Fatalf("error kind = %q", got)
	}
}
