package cli

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthReportsMissingToken(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd := NewRootCommand(&out, &errOut)
	cmd.SetArgs([]string{"auth"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want error")
	}

	got := errOut.String()
	if !strings.Contains(got, "token is required") {
		t.Fatalf("stderr missing token error: %s", got)
	}
	if !strings.Contains(got, "pass --token or set ZINIAO_TOKEN") {
		t.Fatalf("stderr missing token hint: %s", got)
	}
}

func TestAPIListsModules(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd := NewRootCommand(&out, &errOut)
	cmd.SetArgs([]string{"api"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, stderr = %s", err, errOut.String())
	}
	if got := out.String(); !strings.Contains(got, "ziniao") || !strings.Contains(got, "紫鸟业务接口") {
		t.Fatalf("api output missing module: %s", got)
	}
}

func TestAPIShowsDocument(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd := NewRootCommand(&out, &errOut)
	cmd.SetArgs([]string{"api", "ziniao", "user", "list"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, stderr = %s", err, errOut.String())
	}
	got := out.String()
	if !strings.Contains(got, `"name": "list"`) || !strings.Contains(got, `"/api/user/list"`) {
		t.Fatalf("api document output = %s", got)
	}
}

func TestAPIRejectsFullWithAPIName(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd := NewRootCommand(&out, &errOut)
	cmd.SetArgs([]string{"api", "ziniao", "user", "list", "--full"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "--full cannot be used") {
		t.Fatalf("error = %v", err)
	}
}

func TestHTTPCommandSendsRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		if r.URL.Path != "/api/user/create" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("page") != "1" {
			t.Fatalf("query page = %s", r.URL.Query().Get("page"))
		}
		if r.Header.Get("X-Test") != "yes" {
			t.Fatalf("X-Test = %s", r.Header.Get("X-Test"))
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("Authorization = %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	t.Setenv("ZINIAO_BASE_URL", server.URL)
	t.Setenv("ZINIAO_TOKEN", "test-token")

	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd := NewRootCommand(&out, &errOut)
	cmd.SetArgs([]string{"http", "POST", "/api/user/create", "--query", "page=1", "--header", "X-Test=yes", "--body", `{"name":"test"}`})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, stderr = %s", err, errOut.String())
	}
	if got := out.String(); !strings.Contains(got, `"ok": true`) {
		t.Fatalf("http output = %s", got)
	}
}

func TestVersionCommand(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd := NewRootCommand(&out, &errOut)
	cmd.SetArgs([]string{"version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, stderr = %s", err, errOut.String())
	}
	if got := out.String(); !strings.Contains(got, "zn-cli") {
		t.Fatalf("version output = %s", got)
	}
}
