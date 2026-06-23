package cli

import (
	"io"

	"github.com/spf13/cobra"

	"ziniao/internal/agent"
)

func newAgentCommand(rt *runtime) *cobra.Command {
	agentCmd := &cobra.Command{
		Use:   "agent",
		Short: "Agent-oriented CLI guidance",
	}

	guideCmd := &cobra.Command{
		Use:   "guide",
		Short: "Print agent usage guide (embedded SKILL.md)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := io.WriteString(rt.out, agent.Guide())
			return err
		},
	}

	agentCmd.AddCommand(guideCmd)
	return agentCmd
}
