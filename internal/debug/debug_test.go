package debug

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestEnabled(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "empty", value: "", want: false},
		{name: "zero", value: "0", want: false},
		{name: "one", value: "1", want: true},
		{name: "true", value: "true", want: true},
		{name: "yes", value: "YES", want: true},
		{name: "false", value: "false", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(EnvDebug, tt.value)
			if got := Enabled(); got != tt.want {
				t.Fatalf("Enabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrintfDisabled(t *testing.T) {
	t.Setenv(EnvDebug, "")

	var buf bytes.Buffer
	oldOutput := Output
	Output = &buf
	t.Cleanup(func() {
		Output = oldOutput
	})

	Printf("[debug] should not appear\n")
	if buf.Len() != 0 {
		t.Fatalf("Printf() wrote %q when debug disabled", buf.String())
	}
}

func TestPrintfEnabled(t *testing.T) {
	t.Setenv(EnvDebug, "1")

	var buf bytes.Buffer
	oldOutput := Output
	Output = &buf
	t.Cleanup(func() {
		Output = oldOutput
	})

	Printf("[debug] hello %s\n", "world")
	if got := buf.String(); got != "[debug] hello world\n" {
		t.Fatalf("Printf() = %q", got)
	}
}

func TestRedactHeader(t *testing.T) {
	tests := []struct {
		key   string
		value string
		want  string
	}{
		{key: "Authorization", value: "Bearer secret-key", want: "Bearer ***"},
		{key: "authorization", value: "bearer secret-key", want: "Bearer ***"},
		{key: "Authorization", value: "Basic abc", want: "***"},
		{key: "Content-Type", value: "application/json", want: "application/json"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if got := RedactHeader(tt.key, tt.value); got != tt.want {
				t.Fatalf("RedactHeader() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOutputDefaultsToStderr(t *testing.T) {
	if Output != os.Stderr {
		t.Fatalf("Output = %v, want os.Stderr", Output)
	}
	if !strings.Contains(EnvDebug, "ZINIAO") {
		t.Fatalf("EnvDebug = %q", EnvDebug)
	}
}
