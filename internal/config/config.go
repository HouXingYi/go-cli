package config

import (
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
)

var DefaultBaseURL = "https://gateway.ziniao.com"

const DefaultTimeout = 10 * time.Second

type Config struct {
	Token string
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
		Token: strings.TrimSpace(v.GetString("token")),
	}, nil
}

func (c Config) RequireToken() error {
	if strings.TrimSpace(c.Token) == "" {
		return apperr.New(apperr.KindConfig, "token is required", "set ZINIAO_TOKEN.")
	}
	return nil
}
