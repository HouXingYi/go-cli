package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"ziniao/internal/config"
	"ziniao/internal/variant"
)

type runtime struct {
	out    io.Writer
	errOut io.Writer
	viper  *viper.Viper
}

func Execute(v variant.Variant) int {
	variant.SetCurrent(v)
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

	v := variant.Current()
	root := &cobra.Command{
		Use:           v.AppName,
		Short:         fmt.Sprintf("%s HTTP API CLI.", v.AppName),
		Long:          fmt.Sprintf("%s sends authenticated HTTP requests and inspects the API catalog.", v.AppName),
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

	if v.ExtendRoot != nil {
		v.ExtendRoot(root)
	}

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
