package variant

import "github.com/spf13/cobra"

const DefaultVersion = "0.1.0"

const defaultCLITypeHeader = "Cli-Type"

// Variant describes a build-time CLI product (binary name, identity headers, paths).
type Variant struct {
	ID            string
	AppName       string
	CLIType       string
	CLITypeHeader string
	StateDir      string
	EnvPrefix     string
	Version       string
	ExtendRoot    func(*cobra.Command)
}

var current = Ent

// Current returns the active CLI variant for this process.
func Current() Variant {
	return current
}

// SetCurrent selects the CLI variant for this process (used by cmd entrypoints and tests).
func SetCurrent(v Variant) {
	current = v
}

// All returns registered variants for build tooling and tests.
func All() []Variant {
	return []Variant{Ent, Eco}
}

func (v Variant) CLITypeHeaderName() string {
	if h := v.CLITypeHeader; h != "" {
		return h
	}
	return defaultCLITypeHeader
}

func (v Variant) StateDirName() string {
	if d := v.StateDir; d != "" {
		return d
	}
	return v.AppName
}

func (v Variant) VersionString() string {
	if v.Version != "" {
		return v.Version
	}
	return DefaultVersion
}

func (v Variant) EnvPrefixName() string {
	if p := v.EnvPrefix; p != "" {
		return p
	}
	return "ZINIAO"
}
