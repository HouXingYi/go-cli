package backend

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"ziniao/internal/apperr"
	"ziniao/internal/config"
)

func TestMockBackendCatalogModules(t *testing.T) {
	be := NewMockBackend()
	raw, err := be.Catalog(context.Background(), "", "", "", false)
	if err != nil {
		t.Fatalf("Catalog() error = %v", err)
	}

	var payload struct {
		Items []struct {
			Name string `json:"name"`
		} `json:"items"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal error = %v", err)
	}
	if len(payload.Items) == 0 || payload.Items[0].Name != "ziniao" {
		t.Fatalf("items = %#v", payload.Items)
	}
}

func TestMockBackendProxyUserList(t *testing.T) {
	be := NewMockBackend()
	raw, err := be.Proxy(context.Background(), "ziniao", "GET", "/api/user/list", url.Values{"page": []string{"1"}}, nil)
	if err != nil {
		t.Fatalf("Proxy() error = %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal error = %v", err)
	}
	if payload["api"] != "list" {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestMockBackendReportsUnknownAPI(t *testing.T) {
	be := NewMockBackend()
	_, err := be.Catalog(context.Background(), "ziniao", "user", "missing", false)
	if err == nil {
		t.Fatal("Catalog() error = nil, want error")
	}
	if got := apperr.Kind(err); got != apperr.KindAPI {
		t.Fatalf("error kind = %q", got)
	}
}

func TestHTTPBackendProxy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		if r.URL.Path != "/cli/ziniao/user/list" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("Authorization = %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ret":0,"status":"success","msg":"ok","data":{"ok":true}}`))
	}))
	defer server.Close()

	be := NewHTTPBackend(config.Config{
		ProxyBaseURL: server.URL,
		AuthKey:      "test-key",
	})

	raw, err := be.Proxy(context.Background(), "ziniao", "GET", "/user/list", url.Values{"page": []string{"1"}}, nil)
	if err != nil {
		t.Fatalf("Proxy() error = %v", err)
	}

	var payload map[string]bool
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal error = %v", err)
	}
	if !payload["ok"] {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestHTTPBackendCatalog(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s", r.Method)
		}
		if r.URL.Path != "/cli-api/ziniao" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("business") != "user" {
			t.Fatalf("business = %s", r.URL.Query().Get("business"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ret":0,"status":"success","data":{"module":"ziniao","business":"user","items":[{"name":"list","method":"GET","url":"/api/user/list","title":"查询用户列表"}]}}`))
	}))
	defer server.Close()

	be := NewHTTPBackend(config.Config{
		ProxyBaseURL: server.URL,
		AuthKey:      "test-key",
	})

	raw, err := be.Catalog(context.Background(), "ziniao", "user", "", false)
	if err != nil {
		t.Fatalf("Catalog() error = %v", err)
	}

	var payload struct {
		Items []struct {
			Name string `json:"name"`
		} `json:"items"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal error = %v", err)
	}
	if len(payload.Items) != 1 || payload.Items[0].Name != "list" {
		t.Fatalf("items = %#v", payload.Items)
	}
}

func TestParseEnvelopeRetError(t *testing.T) {
	_, err := parseEnvelope(http.StatusOK, []byte(`{"ret":30002,"status":"failed","msg":"no credentials"}`))
	if err == nil {
		t.Fatal("parseEnvelope() error = nil, want error")
	}
	if got := apperr.Kind(err); got != apperr.KindAPI {
		t.Fatalf("error kind = %q", got)
	}
}
