package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"ziniao/internal/config"
)

type runtime struct {
	out    io.Writer
	errOut io.Writer
	viper  *viper.Viper
}

func Execute() int {
	root := NewRootCommand(os.Stdout, os.Stderr)
	if err := root.Execute(); err != nil {
		return 1
	}
	return 0
}

func NewRootCommand(out, errOut io.Writer) *cobra.Command {
	rt := &runtime{
		out:    out,
		errOut: errOut,
		viper:  viper.New(),
	}

	root := &cobra.Command{
		Use:           config.AppName,
		Short:         "Ziniao HTTP API CLI.",
		Long:          "zn-cli sends authenticated HTTP requests and inspects the API catalog.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	root.AddCommand(newAgentCommand(rt))
	root.AddCommand(newAuthCommand(rt))
	root.AddCommand(newConfigCommand(rt))
	root.AddCommand(newHTTPCommand(rt))
	root.AddCommand(newAPICommand(rt))
	root.AddCommand(newVersionCommand(rt))

	return root
}

func mustBindFlag(v *viper.Viper, key string, flag *pflag.Flag) {
	if flag == nil {
		panic(fmt.Sprintf("missing flag for key %s", key))
	}
	if err := v.BindPFlag(key, flag); err != nil {
		panic(err)
	}
}

func (rt *runtime) loadConfig() (config.Config, error) {
	return config.Load(rt.viper)
}
