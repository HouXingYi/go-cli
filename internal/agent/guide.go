package agent

import "ziniao/internal/variant"

// Guide returns the agent guide for the current CLI variant.
func Guide() string {
	return variant.RenderGuide(variant.Current())
}
