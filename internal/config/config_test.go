package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestLoadReadsEnvironment(t *testing.T) {
	t.Setenv(EnvAuthKey, "test-key")
	t.Setenv(EnvProxyBaseURL, "https://api.example.com/vendor-proxy")

	cfg, err := Load(viper.New())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.AuthKey != "test-key" {
		t.Fatalf("AuthKey = %q", cfg.AuthKey)
	}
	if cfg.ProxyBaseURL != "https://api.example.com/vendor-proxy" {
		t.Fatalf("ProxyBaseURL = %q", cfg.ProxyBaseURL)
	}
}

func TestUseRemoteBackend(t *testing.T) {
	if (Config{}).UseRemoteBackend() {
		t.Fatal("empty config should not use remote backend")
	}
	if !(Config{ProxyBaseURL: "https://api.example.com"}).UseRemoteBackend() {
		t.Fatal("config with base URL should use remote backend")
	}
}

func TestRequireAuthKey(t *testing.T) {
	err := Config{}.RequireAuthKey()
	if err == nil {
		t.Fatal("RequireAuthKey() error = nil, want error")
	}
}
