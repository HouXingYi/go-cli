package cli

import (
	"github.com/spf13/cobra"

	"ziniao/internal/app"
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
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication commands",
	}

	verifyCmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify whether the token is valid",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := rt.loadConfig()
			if err != nil {
				return rt.writeCommandError(config.DefaultOutput, err)
			}

			renderer := rt.renderer(cfg.Output)
			service, err := rt.buildService(cfg)
			if err != nil {
				return rt.writeCommandError(cfg.Output, err)
			}

			result, err := service.VerifyAuth(cmd.Context())
			if err != nil {
				return rt.writeCommandError(cfg.Output, err)
			}
			return renderer.Success(result.Message, result)
		},
	}

	authCmd.AddCommand(verifyCmd)
	return authCmd
}

func newRequestCommand(rt *runtime) *cobra.Command {
	requestCmd := &cobra.Command{
		Use:   "request",
		Short: "HTTP request commands",
	}

	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run the configured HTTP request",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := rt.loadConfig()
			if err != nil {
				return rt.writeCommandError(config.DefaultOutput, err)
			}

			renderer := rt.renderer(cfg.Output)
			service, err := rt.buildService(cfg)
			if err != nil {
				return rt.writeCommandError(cfg.Output, err)
			}

			result, err := service.RunRequest(cmd.Context())
			if err != nil {
				return rt.writeCommandError(cfg.Output, err)
			}
			return renderer.Success(result.Message, result)
		},
	}

	requestCmd.AddCommand(runCmd)
	return requestCmd
}
