package config

import (
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"

	"ziniao/internal/apperr"
)

const (
	AppName       = "zn-cli"
	Version       = "0.1.0"
	EnvPrefix     = "ZINIAO"
	OutputText    = "text"
	OutputJSON    = "json"
	DefaultOutput = OutputText

	EnvProxyBaseURL = "VENDOR_PROXY_BASE"
	EnvAuthKey      = "CLI_AUTH_KEY"
)

const DefaultTimeout = 10 * time.Second

type Config struct {
	ProxyBaseURL string
	AuthKey      string
}

func Configure(v *viper.Viper) error {
	v.SetEnvPrefix(EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()
	return nil
}

func Load(v *viper.Viper) (Config, error) {
	if err := Configure(v); err != nil {
		return Config{}, err
	}

	return Config{
		ProxyBaseURL: strings.TrimSpace(os.Getenv(EnvProxyBaseURL)),
		AuthKey:      strings.TrimSpace(os.Getenv(EnvAuthKey)),
	}, nil
}

func (c Config) UseRemoteBackend() bool {
	return strings.TrimSpace(c.ProxyBaseURL) != ""
}

func (c Config) RequireAuthKey() error {
	if strings.TrimSpace(c.AuthKey) == "" {
		return apperr.New(apperr.KindConfig, "auth key is required", "set CLI_AUTH_KEY.")
	}
	return nil
}
