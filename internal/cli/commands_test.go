package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestAuthVerifyReportsMissingToken(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd := NewRootCommand(&out, &errOut)
	cmd.SetArgs([]string{"auth", "verify"})

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
