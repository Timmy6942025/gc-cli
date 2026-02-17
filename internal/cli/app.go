package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/timothy/gc-cli/internal/auth"
	"github.com/timothy/gc-cli/internal/classroom"
	"github.com/timothy/gc-cli/internal/config"
	"github.com/timothy/gc-cli/internal/drive"
	"github.com/timothy/gc-cli/internal/output"
	"github.com/timothy/gc-cli/internal/preview"
	"github.com/timothy/gc-cli/internal/store"
	"github.com/timothy/gc-cli/internal/sync"
	"github.com/timothy/gc-cli/internal/webhandoff"
)

type App struct {
	ConfigStore *config.Store
	Auth        *auth.Manager
	Resolver    webhandoff.HandoffResolver
	Detector    preview.CapabilityDetector
	Store       *store.SQLiteStore
	SyncStore   *sync.Engine
	Printer     output.Printer
	Out         io.Writer
	Err         io.Writer
}

func NewApp(out io.Writer, errOut io.Writer) (*App, error) {
	cfgStore, err := config.NewStore()
	if err != nil {
		return nil, err
	}
	cfg, err := cfgStore.Load()
	if err != nil {
		return nil, err
	}
	tokenStore := auth.NewKeyringTokenStore()
	authManager := auth.NewManager(cfgStore, tokenStore)
	cacheStore, err := store.OpenSQLite()
	if err != nil {
		return nil, err
	}
	app := &App{
		ConfigStore: cfgStore,
		Auth:        authManager,
		Resolver:    webhandoff.Resolver{},
		Detector:    preview.Detector{PreviewEnabled: cfg.PreviewEnabled},
		Store:       cacheStore,
		Printer: output.Printer{
			JSON: false,
			Out:  out,
		},
		Out: out,
		Err: errOut,
	}
	return app, nil
}

func (a *App) Close() error {
	if a.Store != nil {
		return a.Store.Close()
	}
	return nil
}

func (a *App) ClassroomClient(ctx context.Context, requiredScopes []string) (classroom.ClassroomClient, error) {
	cfg, err := a.ConfigStore.Load()
	if err != nil {
		return nil, err
	}
	profile := cfg.ActiveProfile
	if profile == "" {
		profile = "default"
	}
	requiredScopes = auth.NormalizeScopes(requiredScopes)
	if len(requiredScopes) == 0 {
		requiredScopes = auth.DefaultReadScopes
	}

	if err := a.Auth.EnsureScopes(ctx, profile, requiredScopes, true, a.onAuthURL); err != nil {
		if errors.Is(err, auth.ErrMissingClientID) {
			return nil, output.CLIError{Code: "auth_config_missing", Reason: err.Error(), ActionableHint: "Set GC_OAUTH_CLIENT_ID and run gc auth login"}
		}
		return nil, err
	}

	ts, _, err := a.Auth.TokenSource(ctx, profile, requiredScopes)
	if err != nil {
		if errors.Is(err, auth.ErrNoToken) {
			if _, loginErr := a.Auth.Login(ctx, auth.LoginRequest{
				Profile:     profile,
				Scopes:      requiredScopes,
				OpenBrowser: true,
				OnAuthURL:   a.onAuthURL,
			}); loginErr != nil {
				return nil, loginErr
			}
			ts, _, err = a.Auth.TokenSource(ctx, profile, requiredScopes)
		}
		if err != nil {
			if errors.Is(err, auth.ErrMissingClientID) {
				return nil, output.CLIError{Code: "auth_config_missing", Reason: err.Error(), ActionableHint: "Set GC_OAUTH_CLIENT_ID and run gc auth login"}
			}
			return nil, err
		}
	}
	return classroom.NewClient(ctx, ts)
}

func (a *App) DriveClient(ctx context.Context, requiredScopes []string) (drive.DriveClient, error) {
	cfg, err := a.ConfigStore.Load()
	if err != nil {
		return nil, err
	}
	profile := cfg.ActiveProfile
	if profile == "" {
		profile = "default"
	}
	requiredScopes = auth.NormalizeScopes(requiredScopes)
	if err := a.Auth.EnsureScopes(ctx, profile, requiredScopes, true, a.onAuthURL); err != nil {
		return nil, err
	}
	ts, _, err := a.Auth.TokenSource(ctx, profile, requiredScopes)
	if err != nil {
		return nil, err
	}
	return drive.NewClient(ctx, ts)
}

func (a *App) SetJSONOutput(enabled bool) {
	a.Printer.JSON = enabled
}

func (a *App) onAuthURL(url string, browserOpened bool) {
	if browserOpened {
		_, _ = fmt.Fprintf(a.Out, "Authorizing in browser: %s\n", redactedURL(url))
		return
	}
	_, _ = fmt.Fprintf(a.Out, "Open this URL to authorize: %s\n", url)
}

func redactedURL(raw string) string {
	if idx := strings.Index(raw, "&code_challenge="); idx > 0 {
		return raw[:idx] + "&code_challenge=..."
	}
	return raw
}
