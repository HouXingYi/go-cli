package agent

import "ziniao/skills/zn-cli"

// Guide returns the embedded agent guide (SKILL.md).
func Guide() string {
	return zncliguide.Content()
}
