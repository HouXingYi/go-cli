package zncliguide

import (
	"strings"

	_ "embed"
)

//go:embed SKILL.md
var content string

// Content returns the embedded agent guide (SKILL.md) with normalized LF line endings.
func Content() string {
	return strings.ReplaceAll(content, "\r\n", "\n")
}
