package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestLoadReadsEnvironment(t *testing.T) {
	t.Setenv("ZINIAO_TOKEN", "test-token")

	cfg, err := Load(viper.New())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Token != "test-token" {
		t.Fatalf("Token = %q", cfg.Token)
	}
}

func TestRequireToken(t *testing.T) {
	err := Config{}.RequireToken()
	if err == nil {
		t.Fatal("RequireToken() error = nil, want error")
	}
}
