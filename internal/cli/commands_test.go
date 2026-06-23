package cli

import (
	"bytes"
	"strings"
	"testing"

	"ziniao/internal/config"
)

func TestAuthReportsMissingAuthKey(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd := NewRootCommand(&out, &errOut)
	cmd.SetArgs([]string{"auth"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want error")
	}

	got := errOut.String()
	if !strings.Contains(got, "auth key is required") {
		t.Fatalf("stderr missing auth key error: %s", got)
	}
	if !strings.Contains(got, "set CLI_AUTH_KEY") {
		t.Fatalf("stderr missing auth key hint: %s", got)
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

func TestHTTPCommandUsesMockBackend(t *testing.T) {
	t.Setenv(config.EnvAuthKey, "test-key")

	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd := NewRootCommand(&out, &errOut)
	cmd.SetArgs([]string{"http", "GET", "/api/user/list", "--provider", "ziniao", "--query", `{"page":1}`})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, stderr = %s", err, errOut.String())
	}
	if got := out.String(); !strings.Contains(got, `"api": "list"`) {
		t.Fatalf("http output = %s", got)
	}
}

func TestHTTPCommandRejectsInvalidQueryJSON(t *testing.T) {
	t.Setenv(config.EnvAuthKey, "test-key")

	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd := NewRootCommand(&out, &errOut)
	cmd.SetArgs([]string{"http", "GET", "/api/user/list", "--provider", "ziniao", "--query", "not-json"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want error")
	}
	if !strings.Contains(errOut.String(), "query is invalid json") {
		t.Fatalf("stderr = %s", errOut.String())
	}
}

func TestHTTPCommandRejectsQueryArray(t *testing.T) {
	t.Setenv(config.EnvAuthKey, "test-key")

	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd := NewRootCommand(&out, &errOut)
	cmd.SetArgs([]string{"http", "GET", "/api/user/list", "--provider", "ziniao", "--query", "[]"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want error")
	}
	if !strings.Contains(errOut.String(), "query is invalid json") {
		t.Fatalf("stderr = %s", errOut.String())
	}
}

func TestHTTPCommandRequiresProvider(t *testing.T) {
	t.Setenv(config.EnvAuthKey, "test-key")

	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd := NewRootCommand(&out, &errOut)
	cmd.SetArgs([]string{"http", "GET", "/api/user/list"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "provider") && !strings.Contains(errOut.String(), "provider is required") {
		t.Fatalf("error = %v, stderr = %s", err, errOut.String())
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
