package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/timothy/gc-cli/internal/auth"
	"github.com/timothy/gc-cli/internal/config"
)

func newAuthCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with Google OAuth2",
	}
	cmd.AddCommand(newAuthLoginCmd(app), newAuthStatusCmd(app), newAuthLogoutCmd(app))
	return cmd
}

func newAuthLoginCmd(app *App) *cobra.Command {
	var profile string
	var scopesRaw string
	var openBrowser bool
	var clientID string
	var clientSecret string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate a profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			if clientID != "" || clientSecret != "" {
				_, err := app.ConfigStore.Update(func(cfg *config.Config) error {
					if clientID != "" {
						cfg.OAuthClientID = clientID
					}
					if clientSecret != "" {
						cfg.OAuthClientSecret = clientSecret
					}
					return nil
				})
				if err != nil {
					return err
				}
			}
			scopes := parseCSV(scopesRaw)
			if len(scopes) == 0 {
				scopes = auth.DefaultReadScopes
			}
			result, err := app.Auth.Login(ctx, auth.LoginRequest{
				Profile:     profile,
				Scopes:      scopes,
				OpenBrowser: openBrowser,
				OnAuthURL:   app.onAuthURL,
			})
			if err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(result)
			}
			_, _ = fmt.Fprintf(app.Out, "Authenticated profile %s\n", result.Profile)
			if result.Email != "" {
				_, _ = fmt.Fprintf(app.Out, "Email: %s\n", result.Email)
			}
			_, _ = fmt.Fprintf(app.Out, "Granted scopes: %d\n", len(result.ScopesGranted))
			return nil
		},
	}
	cmd.Flags().StringVar(&profile, "profile", "", "Profile name (defaults to active profile)")
	cmd.Flags().StringVar(&scopesRaw, "scopes", "", "Comma-separated OAuth scopes")
	cmd.Flags().BoolVar(&openBrowser, "open-browser", true, "Automatically open the authorization URL in your browser")
	cmd.Flags().StringVar(&clientID, "client-id", "", "OAuth client ID to persist in config")
	cmd.Flags().StringVar(&clientSecret, "client-secret", "", "OAuth client secret to persist in config")
	return cmd
}

func newAuthStatusCmd(app *App) *cobra.Command {
	var profile string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			status, err := app.Auth.Status(profile)
			if err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(status)
			}
			_, _ = fmt.Fprintf(app.Out, "Profile: %s\nAuthenticated: %t\n", status.Profile, status.Authenticated)
			if status.Email != "" {
				_, _ = fmt.Fprintf(app.Out, "Email: %s\n", status.Email)
			}
			if !status.TokenExpiresAt.IsZero() {
				_, _ = fmt.Fprintf(app.Out, "Token expires: %s\n", status.TokenExpiresAt.Format("2006-01-02 15:04:05 MST"))
			}
			_, _ = fmt.Fprintf(app.Out, "Scopes granted: %d\n", len(status.ScopesGranted))
			return nil
		},
	}
	cmd.Flags().StringVar(&profile, "profile", "", "Profile name (defaults to active profile)")
	return cmd
}

func newAuthLogoutCmd(app *App) *cobra.Command {
	var profile string
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Remove the saved OAuth token",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := app.Auth.Logout(profile); err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(map[string]string{"status": "logged_out", "profile": profile})
			}
			if profile == "" {
				profile = "active"
			}
			_, _ = fmt.Fprintf(app.Out, "Logged out profile %s\n", profile)
			return nil
		},
	}
	cmd.Flags().StringVar(&profile, "profile", "", "Profile name (defaults to active profile)")
	return cmd
}

func parseCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}
