package output

import (
	"bytes"
	"strings"
	"testing"

	"ziniao/internal/apperr"
	"ziniao/internal/config"
)

func TestSuccessText(t *testing.T) {
	var out bytes.Buffer
	renderer := New(config.OutputText, &out, &bytes.Buffer{})

	if err := renderer.Success("ok", map[string]string{"message": "ok"}); err != nil {
		t.Fatalf("Success() error = %v", err)
	}

	if got := strings.TrimSpace(out.String()); got != "ok" {
		t.Fatalf("text output = %q", got)
	}
}

func TestSuccessJSON(t *testing.T) {
	var out bytes.Buffer
	renderer := New(config.OutputJSON, &out, &bytes.Buffer{})

	if err := renderer.Success("ok", map[string]string{"message": "ok"}); err != nil {
		t.Fatalf("Success() error = %v", err)
	}

	got := out.String()
	if !strings.Contains(got, `"success": true`) {
		t.Fatalf("json output missing success: %s", got)
	}
	if !strings.Contains(got, `"message": "ok"`) {
		t.Fatalf("json output missing message: %s", got)
	}
}

func TestErrorTextWithHint(t *testing.T) {
	var errOut bytes.Buffer
	renderer := New(config.OutputText, &bytes.Buffer{}, &errOut)

	err := apperr.New(apperr.KindConfig, "token is required", "set ZINIAO_TOKEN.")
	if writeErr := renderer.Error(err); writeErr != nil {
		t.Fatalf("Error() error = %v", writeErr)
	}

	got := errOut.String()
	if !strings.Contains(got, "Error: token is required") {
		t.Fatalf("error output missing message: %s", got)
	}
	if !strings.Contains(got, "Hint: set ZINIAO_TOKEN.") {
		t.Fatalf("error output missing hint: %s", got)
	}
}
