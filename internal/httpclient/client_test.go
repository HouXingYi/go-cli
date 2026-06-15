package httpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ziniao/internal/apperr"
)

func TestVerifyAuthSendsBearerToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != authVerifyPath {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("Authorization = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message":"verified","status":"ok"}`))
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

	result, err := client.VerifyAuth(context.Background())
	if err != nil {
		t.Fatalf("VerifyAuth() error = %v", err)
	}
	if result.Message != "verified" {
		t.Fatalf("Message = %q", result.Message)
	}
}

func TestRunRequestHandlesServerError(t *testing.T) {
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

	_, err = client.RunRequest(context.Background())
	if err == nil {
		t.Fatal("RunRequest() error = nil, want error")
	}
	if got := apperr.Kind(err); got != apperr.KindAPI {
		t.Fatalf("error kind = %q", got)
	}
}

func TestVerifyAuthHandlesUnauthorized(t *testing.T) {
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

	_, err = client.VerifyAuth(context.Background())
	if err == nil {
		t.Fatal("VerifyAuth() error = nil, want error")
	}
	if got := apperr.Kind(err); got != apperr.KindAuth {
		t.Fatalf("error kind = %q", got)
	}
}
