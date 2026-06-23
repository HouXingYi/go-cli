package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"ziniao/internal/apperr"
)

const (
	EnvModule    = "ZINIAO_MODULE"
	EnvConfigDir = "ZINIAO_CONFIG_DIR"
	stateFile    = "state.yaml"
)

type State struct {
	Module string `yaml:"module"`
}

func NormalizeModule(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func StatePath() (string, error) {
	dir := strings.TrimSpace(os.Getenv(EnvConfigDir))
	if dir == "" {
		var err error
		dir, err = os.UserConfigDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(dir, "zn-cli")
	}
	return filepath.Join(dir, stateFile), nil
}

func LoadState() (State, error) {
	path, err := StatePath()
	if err != nil {
		return State{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return State{}, nil
		}
		return State{}, err
	}

	var state State
	if err := yaml.Unmarshal(data, &state); err != nil {
		return State{}, apperr.Wrap(apperr.KindConfig, "failed to read CLI state", "check state file or run zn-cli config module clear.", err)
	}
	state.Module = NormalizeModule(state.Module)
	return state, nil
}

func SaveState(state State) error {
	path, err := StatePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return apperr.Wrap(apperr.KindConfig, "failed to save CLI state", "check config directory permissions.", err)
	}

	v := viper.New()
	v.Set("module", NormalizeModule(state.Module))
	if err := v.WriteConfigAs(path); err != nil {
		return apperr.Wrap(apperr.KindConfig, "failed to save CLI state", "check config directory permissions.", err)
	}
	return os.Chmod(path, 0o600)
}

func ClearState() error {
	path, err := StatePath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return apperr.Wrap(apperr.KindConfig, "failed to clear CLI state", "check config directory permissions.", err)
	}
	return nil
}

func ResolveModule(flagValue string) (string, error) {
	if module := NormalizeModule(flagValue); module != "" {
		return module, nil
	}
	if module := NormalizeModule(os.Getenv(EnvModule)); module != "" {
		return module, nil
	}
	state, err := LoadState()
	if err != nil {
		return "", err
	}
	if module := state.Module; module != "" {
		return module, nil
	}
	return "", apperr.New(apperr.KindConfig, "module is required", "run zn-cli config module set <name> or pass --module.")
}
