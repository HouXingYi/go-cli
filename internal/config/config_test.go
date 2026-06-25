package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestLoadReadsEnvironment(t *testing.T) {
	t.Setenv(EnvProxyToken, "test-key")
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
	if cfg.CLIType != "zn-ent" {
		t.Fatalf("CLIType = %q, want zn-ent", cfg.CLIType)
	}
	if cfg.CLITypeHeader != "Cli-Type" {
		t.Fatalf("CLITypeHeader = %q, want Cli-Type", cfg.CLITypeHeader)
	}
}

func TestLoadReadsVendorProxyTokenFallback(t *testing.T) {
	t.Setenv(EnvProxyToken, "")
	t.Setenv("VENDOR_PROXY_TOKEN", "fallback-token")

	cfg, err := Load(viper.New())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.AuthKey != "fallback-token" {
		t.Fatalf("AuthKey = %q, want fallback-token", cfg.AuthKey)
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
