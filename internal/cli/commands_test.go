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
	if !strings.Contains(got, "set CLI_PROXY_TOKEN") {
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

func TestAPIShowsRawDocumentFields(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd := NewRootCommand(&out, &errOut)
	cmd.SetArgs([]string{"api", "ziniao", "user", "list"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, stderr = %s", err, errOut.String())
	}
	got := out.String()
	if !strings.Contains(got, `"requestBody"`) {
		t.Fatalf("api document output missing requestBody: %s", got)
	}
	if !strings.Contains(got, `"example"`) {
		t.Fatalf("api document output missing example: %s", got)
	}
}

func TestAPIFullShowsRawDocumentFields(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd := NewRootCommand(&out, &errOut)
	cmd.SetArgs([]string{"api", "ziniao", "user", "--full"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, stderr = %s", err, errOut.String())
	}
	got := out.String()
	if !strings.Contains(got, `"requestBody"`) {
		t.Fatalf("api --full output missing requestBody: %s", got)
	}
	if !strings.Contains(got, `"example"`) {
		t.Fatalf("api --full output missing example: %s", got)
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
	t.Setenv(config.EnvProxyToken, "test-key")
	t.Setenv(config.EnvConfigDir, t.TempDir())
	t.Setenv(config.EnvModule, "")

	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd := NewRootCommand(&out, &errOut)
	cmd.SetArgs([]string{"http", "GET", "/api/user/list", "--module", "ziniao", "--query", `{"page":1}`})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, stderr = %s", err, errOut.String())
	}
	if got := out.String(); !strings.Contains(got, `"api": "list"`) {
		t.Fatalf("http output = %s", got)
	}
}

func TestHTTPCommandUsesDefaultModuleFromState(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(config.EnvProxyToken, "test-key")
	t.Setenv(config.EnvConfigDir, dir)
	t.Setenv(config.EnvModule, "")

	if err := config.SaveState(config.State{Module: "ziniao"}); err != nil {
		t.Fatalf("SaveState() error = %v", err)
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd := NewRootCommand(&out, &errOut)
	cmd.SetArgs([]string{"http", "GET", "/api/user/list", "--query", `{"page":1}`})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, stderr = %s", err, errOut.String())
	}
	if got := out.String(); !strings.Contains(got, `"api": "list"`) {
		t.Fatalf("http output = %s", got)
	}
}

func TestHTTPCommandUsesDefaultModuleFromEnv(t *testing.T) {
	t.Setenv(config.EnvProxyToken, "test-key")
	t.Setenv(config.EnvConfigDir, t.TempDir())
	t.Setenv(config.EnvModule, "ziniao")

	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd := NewRootCommand(&out, &errOut)
	cmd.SetArgs([]string{"http", "GET", "/api/user/list"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, stderr = %s", err, errOut.String())
	}
	if got := out.String(); !strings.Contains(got, `"api": "list"`) {
		t.Fatalf("http output = %s", got)
	}
}

func TestHTTPCommandFlagOverridesDefaultModule(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(config.EnvProxyToken, "test-key")
	t.Setenv(config.EnvConfigDir, dir)
	t.Setenv(config.EnvModule, "erp")

	if err := config.SaveState(config.State{Module: "ziniao"}); err != nil {
		t.Fatalf("SaveState() error = %v", err)
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd := NewRootCommand(&out, &errOut)
	cmd.SetArgs([]string{"http", "GET", "/api/user/list", "--module", "ziniao"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, stderr = %s", err, errOut.String())
	}
	if got := out.String(); !strings.Contains(got, `"api": "list"`) {
		t.Fatalf("http output = %s", got)
	}
}

func TestHTTPCommandRejectsInvalidQueryJSON(t *testing.T) {
	t.Setenv(config.EnvProxyToken, "test-key")
	t.Setenv(config.EnvModule, "ziniao")

	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd := NewRootCommand(&out, &errOut)
	cmd.SetArgs([]string{"http", "GET", "/api/user/list", "--query", "not-json"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want error")
	}
	if !strings.Contains(errOut.String(), "query is invalid json") {
		t.Fatalf("stderr = %s", errOut.String())
	}
}

func TestHTTPCommandRejectsQueryArray(t *testing.T) {
	t.Setenv(config.EnvProxyToken, "test-key")
	t.Setenv(config.EnvModule, "ziniao")

	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd := NewRootCommand(&out, &errOut)
	cmd.SetArgs([]string{"http", "GET", "/api/user/list", "--query", "[]"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want error")
	}
	if !strings.Contains(errOut.String(), "query is invalid json") {
		t.Fatalf("stderr = %s", errOut.String())
	}
}

func TestHTTPCommandRequiresModule(t *testing.T) {
	t.Setenv(config.EnvProxyToken, "test-key")
	t.Setenv(config.EnvConfigDir, t.TempDir())
	t.Setenv(config.EnvModule, "")

	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd := NewRootCommand(&out, &errOut)
	cmd.SetArgs([]string{"http", "GET", "/api/user/list"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "module") && !strings.Contains(errOut.String(), "module is required") {
		t.Fatalf("error = %v, stderr = %s", err, errOut.String())
	}
}

func TestConfigModuleSetAndGet(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(config.EnvConfigDir, dir)
	t.Setenv(config.EnvModule, "")

	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd := NewRootCommand(&out, &errOut)
	cmd.SetArgs([]string{"config", "module", "set", "ziniao"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("set Execute() error = %v, stderr = %s", err, errOut.String())
	}
	if !strings.Contains(out.String(), "ziniao") {
		t.Fatalf("set output = %s", out.String())
	}

	out.Reset()
	errOut.Reset()
	cmd.SetArgs([]string{"config", "module", "get"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("get Execute() error = %v, stderr = %s", err, errOut.String())
	}
	if got := strings.TrimSpace(out.String()); got != "ziniao" {
		t.Fatalf("get output = %q, want ziniao", got)
	}
}

func TestConfigModuleClear(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(config.EnvConfigDir, dir)
	t.Setenv(config.EnvModule, "")

	if err := config.SaveState(config.State{Module: "ziniao"}); err != nil {
		t.Fatalf("SaveState() error = %v", err)
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd := NewRootCommand(&out, &errOut)
	cmd.SetArgs([]string{"config", "module", "clear"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("clear Execute() error = %v, stderr = %s", err, errOut.String())
	}

	out.Reset()
	errOut.Reset()
	cmd.SetArgs([]string{"config", "module", "get"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("get after clear error = nil, want error")
	}
	if !strings.Contains(errOut.String(), "no default module configured") {
		t.Fatalf("stderr = %s", errOut.String())
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
	if got := out.String(); !strings.Contains(got, "zn-ent") || !strings.Contains(got, "cli-type: zn-ent") {
		t.Fatalf("version output = %s", got)
	}
}

func TestAgentGuidePrintsContent(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd := NewRootCommand(&out, &errOut)
	cmd.SetArgs([]string{"agent", "guide"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, stderr = %s", err, errOut.String())
	}

	got := out.String()
	if !strings.HasPrefix(got, "---\n") {
		t.Fatalf("agent guide output missing frontmatter: %.80q", got)
	}
	if !strings.Contains(got, "name: zn-ent") {
		t.Fatalf("agent guide output missing name: zn-ent: %s", got)
	}
	if !strings.Contains(got, "## 推荐工作流") {
		t.Fatalf("agent guide output missing workflow section: %s", got)
	}
}

func TestAgentGuideNoAuthRequired(t *testing.T) {
	t.Setenv(config.EnvProxyToken, "")

	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd := NewRootCommand(&out, &errOut)
	cmd.SetArgs([]string{"agent", "guide"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, stderr = %s", err, errOut.String())
	}
	if errOut.Len() != 0 {
		t.Fatalf("stderr = %s, want empty", errOut.String())
	}
	if strings.Contains(out.String(), "Authentication configured") {
		t.Fatalf("agent guide should not use Renderer wrapper: %s", out.String())
	}
}
