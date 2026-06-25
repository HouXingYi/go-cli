package variant

import (
	"strings"

	sharedguide "ziniao/skills/shared"
)

// RenderGuide returns the agent guide for v with AppName placeholders resolved.
func RenderGuide(v Variant) string {
	return strings.ReplaceAll(sharedguide.Template(), "{{AppName}}", v.AppName)
}
