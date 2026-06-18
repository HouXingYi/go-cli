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
	out        io.Writer
	errOut     io.Writer
	viper      *viper.Viper
	configFile string
}

func Execute() int {
	root := NewRootCommand(os.Stdout, os.Stderr)
	if err := root.Execute(); err != nil {
		return 1
	}
	return 0
}

func NewRootCommand(out, errOut io.Writer) *cobra.Command {
	// 初始化 CLI 运行时，集中保存输出流、错误流和配置读取器。
	rt := &runtime{
		out:    out,
		errOut: errOut,
		viper:  viper.New(),
	}

	// 定义根命令；未指定子命令时展示帮助信息。
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

	// 注册所有子命令共享的全局参数。
	flags := root.PersistentFlags()
	flags.StringVar(&rt.configFile, "config", "", fmt.Sprintf("config file path (default $HOME/%s.yaml)", config.DefaultConfigName))
	flags.String("token", "", "access token")
	flags.String("base-url", "", "HTTP API base URL")
	flags.String("timeout", config.DefaultTimeout.String(), "HTTP request timeout")
	flags.String("output", config.DefaultOutput, "output format: text or json")
	flags.Bool("verbose", false, "enable verbose logs")

	// 将命令行参数绑定到 viper，后续统一从配置层读取最终值。
	mustBindFlag(rt.viper, "token", flags.Lookup("token"))
	mustBindFlag(rt.viper, "base_url", flags.Lookup("base-url"))
	mustBindFlag(rt.viper, "timeout", flags.Lookup("timeout"))
	mustBindFlag(rt.viper, "output", flags.Lookup("output"))
	mustBindFlag(rt.viper, "verbose", flags.Lookup("verbose"))

	// 挂载业务子命令。
	root.AddCommand(newAuthCommand(rt))
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
	return config.Load(rt.viper, rt.configFile)
}
