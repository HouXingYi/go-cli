package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestLoadReadsEnvironment(t *testing.T) {
	t.Setenv("ZINIAO_BASE_URL", "https://api.example.com")
	t.Setenv("ZINIAO_TOKEN", "test-token")
	t.Setenv("ZINIAO_TIMEOUT", "5s")
	t.Setenv("ZINIAO_OUTPUT", "json")

	cfg, err := Load(viper.New(), "")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.BaseURL != "https://api.example.com" {
		t.Fatalf("BaseURL = %q", cfg.BaseURL)
	}
	if cfg.Token != "test-token" {
		t.Fatalf("Token = %q", cfg.Token)
	}
	if cfg.Timeout.String() != "5s" {
		t.Fatalf("Timeout = %s", cfg.Timeout)
	}
	if cfg.Output != OutputJSON {
		t.Fatalf("Output = %q", cfg.Output)
	}
}

func TestLoadRejectsInvalidOutput(t *testing.T) {
	v := viper.New()
	v.Set("output", "xml")

	_, err := Load(v, "")
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}
}

func TestRequireToken(t *testing.T) {
	err := Config{}.RequireToken()
	if err == nil {
		t.Fatal("RequireToken() error = nil, want error")
	}
}

func TestRequireBaseURL(t *testing.T) {
	err := Config{BaseURL: "not-a-url"}.RequireBaseURL()
	if err == nil {
		t.Fatal("RequireBaseURL() error = nil, want error")
	}
}
