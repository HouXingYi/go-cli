// Package debug provides developer-only diagnostics controlled by ZINIAO_DEBUG.
// It is intentionally not exposed via CLI flags or user documentation.
package debug

import (
	"fmt"
	"io"
	"os"
	"strings"
)

const EnvDebug = "ZINIAO_DEBUG"

// Output is the destination for debug messages. Tests may replace it.
var Output io.Writer = os.Stderr

// Enabled reports whether ZINIAO_DEBUG is set to a truthy value.
func Enabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(EnvDebug)))
	return v == "1" || v == "true" || v == "yes"
}

// Printf writes a formatted line to Output when debug mode is enabled.
func Printf(format string, args ...any) {
	if !Enabled() {
		return
	}
	_, _ = fmt.Fprintf(Output, format, args...)
}

// RedactHeader masks sensitive header values before logging.
func RedactHeader(key, value string) string {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "authorization":
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(value)), "bearer ") {
			return "Bearer ***"
		}
		return "***"
	default:
		return value
	}
}
