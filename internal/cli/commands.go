package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"ziniao/internal/apperr"
	"ziniao/internal/backend"
	"ziniao/internal/catalog"
	"ziniao/internal/config"
	"ziniao/internal/output"
	"ziniao/internal/variant"
)

func (rt *runtime) renderer() output.Renderer {
	return output.New(config.DefaultOutput, rt.out, rt.errOut)
}

func (rt *runtime) writeCommandError(err error) error {
	if writeErr := rt.renderer().Error(err); writeErr != nil {
		return writeErr
	}
	return err
}

func newAuthCommand(rt *runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "auth",
		Short: "Configure CLI authentication",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := rt.loadConfig()
			if err != nil {
				return rt.writeCommandError(err)
			}

			if err := cfg.RequireAuthKey(); err != nil {
				return rt.writeCommandError(err)
			}

			data := map[string]string{
				"status": "configured",
			}
			return rt.renderer().Success("Authentication configured successfully.", data)
		},
	}
}

func newHTTPCommand(rt *runtime) *cobra.Command {
	var module string
	var query string
	var body string

	httpCmd := &cobra.Command{
		Use:   "http <method> <path>",
		Short: "Send an authenticated HTTP request",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedModule, err := config.ResolveModule(module)
			if err != nil {
				return rt.writeCommandError(err)
			}

			cfg, err := rt.loadConfig()
			if err != nil {
				return rt.writeCommandError(err)
			}
			if err := cfg.RequireAuthKey(); err != nil {
				return rt.writeCommandError(err)
			}

			rawQuery, err := parseOptionalJSONObject(query, "query")
			if err != nil {
				return rt.writeCommandError(err)
			}

			rawBody, err := parseOptionalJSONValue(body, "body")
			if err != nil {
				return rt.writeCommandError(err)
			}

			be := backend.New(cfg)
			data, err := be.Proxy(cmd.Context(), resolvedModule, args[0], args[1], rawQuery, rawBody)
			if err != nil {
				return rt.writeCommandError(err)
			}

			return rt.renderer().Success(formatJSONRaw(data), data)
		},
	}

	httpCmd.Flags().StringVar(&module, "module", "", "module identifier (overrides default)")
	httpCmd.Flags().StringVar(&query, "query", "", "JSON query parameters object")
	httpCmd.Flags().StringVar(&body, "body", "", "JSON request body")
	return httpCmd
}

func newConfigCommand(rt *runtime) *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
	}
	configCmd.AddCommand(newConfigModuleCommand(rt))
	return configCmd
}

func newConfigModuleCommand(rt *runtime) *cobra.Command {
	moduleCmd := &cobra.Command{
		Use:   "module",
		Short: "Manage default module",
	}

	setCmd := &cobra.Command{
		Use:   "set <module>",
		Short: "Set default module",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			module := config.NormalizeModule(args[0])
			if module == "" {
				return rt.writeCommandError(apperr.New(apperr.KindConfig, "module name is required", "pass a module name, e.g. ziniao."))
			}
			if err := config.SaveState(config.State{Module: module}); err != nil {
				return rt.writeCommandError(err)
			}
			data := map[string]string{"module": module}
			return rt.renderer().Success(fmt.Sprintf("Default module set to %s.", module), data)
		},
	}

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Show default module",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			module := config.NormalizeModule(os.Getenv(config.EnvModule))
			source := "environment"
			if module == "" {
				state, err := config.LoadState()
				if err != nil {
					return rt.writeCommandError(err)
				}
				module = state.Module
				source = "config"
			}
			if module == "" {
				return rt.writeCommandError(apperr.New(apperr.KindConfig, "no default module configured", fmt.Sprintf("run %s config module set <name> or %s api to list modules.", config.AppName(), config.AppName())))
			}
			data := map[string]string{"module": module, "source": source}
			return rt.renderer().Success(module, data)
		},
	}

	clearCmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear persisted default module",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.ClearState(); err != nil {
				return rt.writeCommandError(err)
			}
			return rt.renderer().Success("Default module cleared.", map[string]string{"status": "cleared"})
		},
	}

	moduleCmd.AddCommand(setCmd, getCmd, clearCmd)
	return moduleCmd
}

func newAPICommand(rt *runtime) *cobra.Command {
	var full bool

	apiCmd := &cobra.Command{
		Use:   "api [module] [business] [api]",
		Short: "Inspect backend HTTP API catalog",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 3 {
				return fmt.Errorf("accepts at most 3 arg(s), received %d", len(args))
			}
			if full && len(args) == 3 {
				return fmt.Errorf("--full cannot be used when [api] is specified")
			}
			if full && len(args) != 2 {
				return fmt.Errorf("--full requires [module] and [business]")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := rt.loadConfig()
			if err != nil {
				return rt.writeCommandError(err)
			}

			be := backend.New(cfg)
			message, data, err := runAPIQuery(cmd.Context(), be, args, full)
			if err != nil {
				return rt.writeCommandError(err)
			}
			return rt.renderer().Success(message, data)
		},
	}

	apiCmd.Flags().BoolVar(&full, "full", false, "return full API documents for a business module")
	apiCmd.ValidArgsFunction = completeAPIArgs(rt)
	return apiCmd
}

func newVersionCommand(rt *runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print CLI version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := variant.Current()
			line := fmt.Sprintf("%s %s (cli-type: %s)", v.AppName, v.VersionString(), v.CLIType)
			data := map[string]string{
				"name":     v.AppName,
				"version":  v.VersionString(),
				"cli_type": v.CLIType,
			}
			return rt.renderer().Success(line, data)
		},
	}
}

func runAPIQuery(ctx context.Context, be backend.Backend, args []string, full bool) (string, interface{}, error) {
	module, business, api := "", "", ""
	switch len(args) {
	case 1:
		module = args[0]
	case 2:
		module, business = args[0], args[1]
	case 3:
		module, business, api = args[0], args[1], args[2]
	}

	raw, err := be.Catalog(ctx, module, business, api, full)
	if err != nil {
		return "", nil, err
	}

	switch len(args) {
	case 0:
		items, err := catalog.ParseModules(raw)
		return catalog.FormatModules(items), items, err
	case 1:
		items, err := catalog.ParseBusinesses(raw)
		return catalog.FormatBusinesses(items), items, err
	case 2:
		if full {
			return formatJSONRaw(raw), raw, nil
		}
		items, err := catalog.ParseAPISummaries(raw)
		return catalog.FormatAPISummaries(items), items, err
	default:
		return formatJSONRaw(raw), raw, nil
	}
}

func completeAPIArgs(rt *runtime) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		cfg, err := rt.loadConfig()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		be := backend.New(cfg)
		mock, _ := be.(*backend.MockBackend)

		var candidates []string
		switch len(args) {
		case 0:
			if mock != nil {
				for _, name := range mock.ListModuleNames() {
					candidates = appendCompletion(candidates, name, "", toComplete)
				}
				return candidates, cobra.ShellCompDirectiveNoFileComp
			}
			raw, err := be.Catalog(cmd.Context(), "", "", "", false)
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			items, err := catalog.ParseModules(raw)
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			for _, item := range items {
				candidates = appendCompletion(candidates, item.Name, item.Title, toComplete)
			}
		case 1:
			if mock != nil {
				for _, name := range mock.ListBusinessNames(args[0]) {
					candidates = appendCompletion(candidates, name, "", toComplete)
				}
				return candidates, cobra.ShellCompDirectiveNoFileComp
			}
			raw, err := be.Catalog(cmd.Context(), args[0], "", "", false)
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			items, err := catalog.ParseBusinesses(raw)
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			for _, item := range items {
				candidates = appendCompletion(candidates, item.Name, item.Title, toComplete)
			}
		case 2:
			if mock != nil {
				for _, name := range mock.ListAPINames(args[0], args[1]) {
					candidates = appendCompletion(candidates, name, "", toComplete)
				}
				return candidates, cobra.ShellCompDirectiveNoFileComp
			}
			raw, err := be.Catalog(cmd.Context(), args[0], args[1], "", false)
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			items, err := catalog.ParseAPISummaries(raw)
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			for _, item := range items {
				candidates = appendCompletion(candidates, item.Name, item.Title, toComplete)
			}
		}
		return candidates, cobra.ShellCompDirectiveNoFileComp
	}
}

func appendCompletion(candidates []string, name, title, prefix string) []string {
	if prefix != "" && !strings.HasPrefix(name, prefix) {
		return candidates
	}
	if title == "" {
		return append(candidates, name)
	}
	return append(candidates, fmt.Sprintf("%s\t%s", name, title))
}

func parseOptionalJSONObject(raw, flag string) (json.RawMessage, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}
	msg := json.RawMessage(trimmed)
	if !json.Valid(msg) {
		return nil, apperr.New(apperr.KindConfig, flag+" is invalid json", "pass --"+flag+" with a valid JSON object.")
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(msg, &obj); err != nil || obj == nil {
		return nil, apperr.New(apperr.KindConfig, flag+" is invalid json", "pass --"+flag+" with a valid JSON object.")
	}
	return msg, nil
}

func parseOptionalJSONValue(raw, flag string) (json.RawMessage, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}
	msg := json.RawMessage(trimmed)
	if !json.Valid(msg) {
		return nil, apperr.New(apperr.KindConfig, flag+" is invalid json", "pass --"+flag+" with a valid JSON object or array.")
	}
	return msg, nil
}

func formatJSON(value interface{}) string {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Sprint(value)
	}
	return string(payload)
}

func formatJSONRaw(value json.RawMessage) string {
	if len(value) == 0 {
		return ""
	}
	if json.Valid(value) {
		var parsed interface{}
		if err := json.Unmarshal(value, &parsed); err == nil {
			return formatJSON(parsed)
		}
	}
	return string(value)
}
