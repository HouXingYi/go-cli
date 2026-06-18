package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"ziniao/internal/app"
	"ziniao/internal/catalog"
	"ziniao/internal/config"
	"ziniao/internal/httpclient"
	"ziniao/internal/output"
)

func (rt *runtime) renderer(format string) output.Renderer {
	return output.New(format, rt.out, rt.errOut)
}

func (rt *runtime) buildService(cfg config.Config) (app.Service, error) {
	if err := cfg.RequireToken(); err != nil {
		return app.Service{}, err
	}
	if err := cfg.RequireBaseURL(); err != nil {
		return app.Service{}, err
	}

	api, err := httpclient.New(httpclient.Options{
		BaseURL: cfg.BaseURL,
		Token:   cfg.Token,
		Timeout: cfg.Timeout,
	})
	if err != nil {
		return app.Service{}, err
	}

	return app.NewService(api, cfg.Token), nil
}

func (rt *runtime) writeCommandError(format string, err error) error {
	if writeErr := rt.renderer(format).Error(err); writeErr != nil {
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
				return rt.writeCommandError(config.DefaultOutput, err)
			}

			if err := cfg.RequireToken(); err != nil {
				return rt.writeCommandError(cfg.Output, err)
			}
			if err := cfg.RequireBaseURL(); err != nil {
				return rt.writeCommandError(cfg.Output, err)
			}

			data := map[string]string{
				"baseUrl": cfg.BaseURL,
				"status":  "configured",
			}
			return rt.renderer(cfg.Output).Success("Authentication configured successfully.", data)
		},
	}
}

func newHTTPCommand(rt *runtime) *cobra.Command {
	var queryPairs []string
	var headerPairs []string
	var body string

	httpCmd := &cobra.Command{
		Use:   "http <method> <path>",
		Short: "Send an authenticated HTTP request",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := rt.loadConfig()
			if err != nil {
				return rt.writeCommandError(config.DefaultOutput, err)
			}

			service, err := rt.buildService(cfg)
			if err != nil {
				return rt.writeCommandError(cfg.Output, err)
			}

			query, err := parseQueryPairs(queryPairs)
			if err != nil {
				return rt.writeCommandError(cfg.Output, err)
			}
			headers, err := parseHeaderPairs(headerPairs)
			if err != nil {
				return rt.writeCommandError(cfg.Output, err)
			}

			var rawBody json.RawMessage
			if strings.TrimSpace(body) != "" {
				rawBody = json.RawMessage(strings.TrimSpace(body))
			}

			response, err := service.HTTPRequest(cmd.Context(), app.HTTPRequest{
				Method:  args[0],
				Path:    args[1],
				Query:   query,
				Headers: headers,
				Body:    rawBody,
			})
			if err != nil {
				return rt.writeCommandError(cfg.Output, err)
			}
			return rt.renderer(cfg.Output).Success(formatHTTPResponse(response), response)
		},
	}

	httpCmd.Flags().StringArrayVar(&queryPairs, "query", nil, "append URL query parameter as key=value")
	httpCmd.Flags().StringArrayVar(&headerPairs, "header", nil, "append request header as key=value")
	httpCmd.Flags().StringVar(&body, "body", "", "JSON request body")
	return httpCmd
}

func newAPICommand(rt *runtime) *cobra.Command {
	provider := catalog.NewMockProvider()
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
				return rt.writeCommandError(config.DefaultOutput, err)
			}

			message, data, err := runAPIQuery(cmd, provider, args, full)
			if err != nil {
				return rt.writeCommandError(cfg.Output, err)
			}
			return rt.renderer(cfg.Output).Success(message, data)
		},
	}

	apiCmd.Flags().BoolVar(&full, "full", false, "return full API documents for a business module")
	apiCmd.ValidArgsFunction = completeAPIArgs(provider)
	return apiCmd
}

func newVersionCommand(rt *runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print CLI version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := rt.loadConfig()
			if err != nil {
				return rt.writeCommandError(config.DefaultOutput, err)
			}
			data := map[string]string{
				"name":    config.AppName,
				"version": config.Version,
			}
			return rt.renderer(cfg.Output).Success(fmt.Sprintf("%s %s", config.AppName, config.Version), data)
		},
	}
}

func runAPIQuery(cmd *cobra.Command, provider catalog.Provider, args []string, full bool) (string, interface{}, error) {
	switch len(args) {
	case 0:
		items, err := provider.ListModules(cmd.Context())
		return formatModules(items), items, err
	case 1:
		items, err := provider.ListBusinesses(cmd.Context(), args[0])
		return formatBusinesses(items), items, err
	case 2:
		if full {
			items, err := provider.ListFullAPIs(cmd.Context(), args[0], args[1])
			return formatJSON(items), items, err
		}
		items, err := provider.ListAPIs(cmd.Context(), args[0], args[1])
		return formatAPISummaries(items), items, err
	default:
		doc, err := provider.GetAPI(cmd.Context(), args[0], args[1], args[2])
		return formatJSON(doc), doc, err
	}
}

func completeAPIArgs(provider catalog.Provider) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var candidates []string
		switch len(args) {
		case 0:
			items, err := provider.ListModules(cmd.Context())
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			for _, item := range items {
				candidates = appendCompletion(candidates, item.Name, item.Title, toComplete)
			}
		case 1:
			items, err := provider.ListBusinesses(cmd.Context(), args[0])
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			for _, item := range items {
				candidates = appendCompletion(candidates, item.Name, item.Title, toComplete)
			}
		case 2:
			items, err := provider.ListAPIs(cmd.Context(), args[0], args[1])
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

func parseQueryPairs(pairs []string) (url.Values, error) {
	values := url.Values{}
	for _, pair := range pairs {
		key, value, err := splitKeyValue(pair)
		if err != nil {
			return nil, err
		}
		values.Add(key, value)
	}
	return values, nil
}

func parseHeaderPairs(pairs []string) (http.Header, error) {
	headers := http.Header{}
	for _, pair := range pairs {
		key, value, err := splitKeyValue(pair)
		if err != nil {
			return nil, err
		}
		headers.Add(key, value)
	}
	return headers, nil
}

func splitKeyValue(pair string) (string, string, error) {
	key, value, ok := strings.Cut(pair, "=")
	key = strings.TrimSpace(key)
	if !ok || key == "" {
		return "", "", fmt.Errorf("invalid key=value pair %q", pair)
	}
	return key, strings.TrimSpace(value), nil
}

func formatModules(items []catalog.ModuleSummary) string {
	lines := make([]string, 0, len(items))
	for _, item := range items {
		lines = append(lines, fmt.Sprintf("%-10s %s", item.Name, item.Title))
	}
	return strings.Join(lines, "\n")
}

func formatBusinesses(items []catalog.BusinessSummary) string {
	lines := make([]string, 0, len(items))
	for _, item := range items {
		lines = append(lines, fmt.Sprintf("%-10s %s", item.Name, item.Title))
	}
	return strings.Join(lines, "\n")
}

func formatAPISummaries(items []catalog.APISummary) string {
	lines := make([]string, 0, len(items))
	for _, item := range items {
		lines = append(lines, fmt.Sprintf("%-10s %-6s %-20s %s", item.Name, item.Method, item.URL, item.Title))
	}
	return strings.Join(lines, "\n")
}

func formatJSON(value interface{}) string {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Sprint(value)
	}
	return string(payload)
}

func formatHTTPResponse(response httpclient.Response) string {
	if len(response.Body) == 0 {
		return response.Status
	}
	if json.Valid(response.Body) {
		var value interface{}
		if err := json.Unmarshal(response.Body, &value); err == nil {
			return formatJSON(value)
		}
	}
	return string(response.Body)
}
