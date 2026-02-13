package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/timboy697/gc-cli/internal/auth"
	"github.com/timboy697/gc-cli/internal/config"
	"github.com/timboy697/gc-cli/internal/tui"

	"github.com/urfave/cli/v2"
)

var Version = "dev"

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not load config: %v\n", err)
		cfg = config.Default()
	}

	app := &cli.App{
		Name:                 "gc-cli",
		Version:              Version,
		Usage:                "Google Classroom CLI for students",
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "enable verbose output",
			},
			&cli.StringFlag{
				Name:        "config",
				Usage:       "path to config file",
				DefaultText: cfg.ConfigPath,
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "auth",
				Usage: "manage authentication",
				Subcommands: []*cli.Command{
					{
						Name:  "login",
						Usage: "authenticate with Google",
						Action: func(c *cli.Context) error {
							return handleLogin(ctx, cfg)
						},
					},
					{
						Name:  "status",
						Usage: "check authentication status",
						Action: func(c *cli.Context) error {
							return handleAuthStatus(ctx, cfg)
						},
					},
				},
			},
			{
				Name:  "login",
				Usage: "authenticate with Google (alias for auth login)",
				Action: func(c *cli.Context) error {
					return handleLogin(ctx, cfg)
				},
			},
			CoursesCmd(cfg),
			{
				Name:  "course",
				Usage: "view course details",
				Subcommands: []*cli.Command{
					{
						Name:      "view",
						Usage:     "view course details",
						ArgsUsage: "<course-id>",
						Action: func(c *cli.Context) error {
							if c.Args().Len() < 1 {
								return fmt.Errorf("course ID required")
							}
							fmt.Printf("Viewing course: %s\n", c.Args().First())
							return nil
						},
					},
				},
			},
			{
				Name:  "assignments",
				Usage: "list assignments for a course",
				Action: func(c *cli.Context) error {
					fmt.Println("Assignments list - to be implemented")
					return nil
				},
			},
			CourseworkCmd(cfg),
			SubmitCmd(cfg),
			GradesCmd(cfg),
			AnnouncementsCmd(cfg),
			{
				Name:  "tui",
				Usage: "launch interactive TUI mode",
				Action: func(c *cli.Context) error {
					return tui.Run(cfg)
				},
			},
		},
		Before: func(c *cli.Context) error {
			if c.String("config") != "" {
				cfg.ConfigPath = c.String("config")
			}
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleLogin(ctx context.Context, cfg *config.Config) error {
	if cfg.Auth.ClientID == "" || cfg.Auth.ClientSecret == "" {
		fmt.Println("OAuth credentials not configured.")
		fmt.Printf("Please configure your OAuth credentials in %s\n", cfg.ConfigPath)
		fmt.Printf("Or visit: %s\n", auth.GetConfigURL())
		fmt.Println("\nRequired configuration:")
		fmt.Println("  auth:")
		fmt.Println("    client_id: YOUR_CLIENT_ID")
		fmt.Println("    client_secret: YOUR_CLIENT_SECRET")
		return nil
	}

	authCfg := auth.NewConfig(cfg.Auth.ClientID, cfg.Auth.ClientSecret, cfg.Auth.TokenFile)

	fmt.Println("Starting OAuth authentication flow...")
	fmt.Println("A browser window will open for you to sign in with your Google account.")

	token, err := auth.BrowserFlow(ctx, authCfg)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	if err := auth.TokenToFile(cfg.Auth.TokenFile, token); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	fmt.Println("\n✓ Authentication successful!")
	fmt.Printf("Token saved to: %s\n", cfg.Auth.TokenFile)

	return nil
}

func handleAuthStatus(ctx context.Context, cfg *config.Config) error {
	if !auth.TokenExists(cfg.Auth.TokenFile) {
		fmt.Println("Status: Not logged in")
		fmt.Println("Run 'gc-cli auth login' to authenticate")
		return nil
	}

	token, err := auth.TokenFromFile(cfg.Auth.TokenFile)
	if err != nil {
		fmt.Println("Status: Not logged in (invalid token file)")
		fmt.Println("Run 'gc-cli auth login' to authenticate")
		return nil
	}

	if token.Expiry.After(time.Now()) {
		fmt.Println("Status: Logged in")
		fmt.Printf("Token expires: %s\n", token.Expiry.Format("2006-01-02 15:04:05"))
	} else if token.RefreshToken != "" {
		fmt.Println("Status: Logged in (token expired, refresh available)")
		fmt.Printf("Token expired: %s\n", token.Expiry.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Println("Status: Not logged in (token expired)")
		fmt.Println("Run 'gc-cli auth login' to authenticate")
	}

	return nil
}
