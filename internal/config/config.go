package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"

	"ziniao/internal/apperr"
	"ziniao/internal/variant"
)

const (
	OutputText    = "text"
	OutputJSON    = "json"
	DefaultOutput = OutputText

	EnvProxyBaseURL = "CLI_PROXY_BASE"
	EnvProxyToken   = "CLI_PROXY_TOKEN"
	EnvDebug        = "ZINIAO_DEBUG"
)

const DefaultTimeout = 10 * time.Second

type Config struct {
	ProxyBaseURL  string
	AuthKey       string
	CLIType       string
	CLITypeHeader string
}

func AppName() string {
	return variant.Current().AppName
}

func Version() string {
	return variant.Current().VersionString()
}

func CLIType() string {
	return variant.Current().CLIType
}

func CLITypeHeader() string {
	return variant.Current().CLITypeHeaderName()
}

func EnvPrefix() string {
	return variant.Current().EnvPrefixName()
}

func Configure(v *viper.Viper) error {
	v.SetEnvPrefix(EnvPrefix())
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()
	return nil
}

func Load(v *viper.Viper) (Config, error) {
	if err := Configure(v); err != nil {
		return Config{}, err
	}

	vr := variant.Current()
	return Config{
		ProxyBaseURL:  strings.TrimSpace(os.Getenv(EnvProxyBaseURL)),
		AuthKey:       firstNonEmptyEnv(EnvProxyToken, "VENDOR_PROXY_TOKEN"),
		CLIType:       strings.TrimSpace(vr.CLIType),
		CLITypeHeader: vr.CLITypeHeaderName(),
	}, nil
}

func (c Config) UseRemoteBackend() bool {
	return strings.TrimSpace(c.ProxyBaseURL) != ""
}

func (c Config) RequireAuthKey() error {
	if strings.TrimSpace(c.AuthKey) == "" {
		return apperr.New(apperr.KindConfig, "auth key is required", "set CLI_PROXY_TOKEN.")
	}
	return nil
}

func APIInspectHint() string {
	return fmt.Sprintf("run %s api to inspect available modules and APIs.", AppName())
}

func ModuleRequiredHint() string {
	return fmt.Sprintf("run %s config module set <name> or pass --module.", AppName())
}

func ConfigModuleClearHint() string {
	return fmt.Sprintf("check state file or run %s config module clear.", AppName())
}

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}
