package config

import (
	"errors"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"

	"ziniao/internal/apperr"
)

const (
	AppName           = "ziniao"
	EnvPrefix         = "ZINIAO"
	DefaultConfigName = ".ziniao"
	DefaultOutput     = OutputText
	OutputText        = "text"
	OutputJSON        = "json"
)

const DefaultTimeout = 10 * time.Second

type Config struct {
	BaseURL string
	Token   string
	Timeout time.Duration
	Output  string
	Verbose bool
}

func Configure(v *viper.Viper, configFile string) error {
	v.SetEnvPrefix(EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	v.SetDefault("timeout", DefaultTimeout.String())
	v.SetDefault("output", DefaultOutput)

	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		v.SetConfigName(DefaultConfigName)
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		if home, err := os.UserHomeDir(); err == nil {
			v.AddConfigPath(home)
		}
	}

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if configFile != "" || !errors.As(err, &notFound) {
			return apperr.Wrap(apperr.KindConfig, "failed to read config file", "check --config path and file format.", err)
		}
	}

	return nil
}

func Load(v *viper.Viper, configFile string) (Config, error) {
	if err := Configure(v, configFile); err != nil {
		return Config{}, err
	}

	timeout, err := time.ParseDuration(v.GetString("timeout"))
	if err != nil {
		return Config{}, apperr.Wrap(apperr.KindConfig, "timeout is invalid", "use a duration like 10s, 1m, or 500ms.", err)
	}

	cfg := Config{
		BaseURL: strings.TrimSpace(v.GetString("base_url")),
		Token:   strings.TrimSpace(v.GetString("token")),
		Timeout: timeout,
		Output:  strings.ToLower(strings.TrimSpace(v.GetString("output"))),
		Verbose: v.GetBool("verbose"),
	}

	if cfg.Output == "" {
		cfg.Output = DefaultOutput
	}
	if cfg.Output != OutputText && cfg.Output != OutputJSON {
		return Config{}, apperr.New(apperr.KindConfig, "output format is invalid", "use --output text or --output json.")
	}

	return cfg, nil
}

func (c Config) RequireToken() error {
	if strings.TrimSpace(c.Token) == "" {
		return apperr.New(apperr.KindConfig, "token is required", "pass --token or set ZINIAO_TOKEN.")
	}
	return nil
}

func (c Config) RequireBaseURL() error {
	if strings.TrimSpace(c.BaseURL) == "" {
		return apperr.New(apperr.KindConfig, "base_url is required", "pass --base-url or set ZINIAO_BASE_URL.")
	}

	parsed, err := url.ParseRequestURI(c.BaseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return apperr.New(apperr.KindConfig, "base_url is invalid", "use an absolute URL like https://api.example.com.")
	}

	return nil
}
