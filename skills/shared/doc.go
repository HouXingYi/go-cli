package sharedguide

import (
	"strings"

	_ "embed"
)

//go:embed SKILL.md
var content string

// Template returns the shared agent guide template with {{AppName}} placeholders.
func Template() string {
	return strings.ReplaceAll(content, "\r\n", "\n")
}
