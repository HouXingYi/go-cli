package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestStatePathUsesConfigDirEnv(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(EnvConfigDir, dir)

	path, err := StatePath()
	if err != nil {
		t.Fatalf("StatePath() error = %v", err)
	}
	if want := filepath.Join(dir, stateFile); path != want {
		t.Fatalf("StatePath() = %q, want %q", path, want)
	}
}

func TestSaveLoadAndClearState(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(EnvConfigDir, dir)

	if err := SaveState(State{Module: "Ziniao"}); err != nil {
		t.Fatalf("SaveState() error = %v", err)
	}

	state, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState() error = %v", err)
	}
	if state.Module != "ziniao" {
		t.Fatalf("Module = %q, want ziniao", state.Module)
	}

	info, err := os.Stat(filepath.Join(dir, stateFile))
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Fatalf("state file mode = %o, want 0600", info.Mode().Perm())
	}

	if err := ClearState(); err != nil {
		t.Fatalf("ClearState() error = %v", err)
	}
	state, err = LoadState()
	if err != nil {
		t.Fatalf("LoadState() after clear error = %v", err)
	}
	if state.Module != "" {
		t.Fatalf("Module after clear = %q, want empty", state.Module)
	}
}

func TestLoadStateMissingFile(t *testing.T) {
	t.Setenv(EnvConfigDir, t.TempDir())

	state, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState() error = %v", err)
	}
	if state.Module != "" {
		t.Fatalf("Module = %q, want empty", state.Module)
	}
}

func TestResolveModulePriority(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(EnvConfigDir, dir)
	t.Setenv(EnvModule, "")

	if err := SaveState(State{Module: "ziniao"}); err != nil {
		t.Fatalf("SaveState() error = %v", err)
	}

	module, err := ResolveModule("")
	if err != nil || module != "ziniao" {
		t.Fatalf("ResolveModule() = %q, %v; want ziniao", module, err)
	}

	t.Setenv(EnvModule, "erp")
	module, err = ResolveModule("")
	if err != nil || module != "erp" {
		t.Fatalf("ResolveModule() with env = %q, %v; want erp", module, err)
	}

	module, err = ResolveModule("ziniao")
	if err != nil || module != "ziniao" {
		t.Fatalf("ResolveModule(flag) = %q, %v; want ziniao", module, err)
	}
}

func TestResolveModuleRequiresValue(t *testing.T) {
	t.Setenv(EnvConfigDir, t.TempDir())
	t.Setenv(EnvModule, "")

	_, err := ResolveModule("")
	if err == nil {
		t.Fatal("ResolveModule() error = nil, want error")
	}
}
