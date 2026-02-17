package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/timothy/gc-cli/internal/config"
	"github.com/timothy/gc-cli/internal/platform"
)

type LoginRequest struct {
	Profile     string
	Scopes      []string
	OpenBrowser bool
	Timeout     time.Duration
	OnAuthURL   func(url string, browserOpened bool)
}

type LoginResult struct {
	Profile       string    `json:"profile"`
	Email         string    `json:"email,omitempty"`
	ScopesGranted []string  `json:"scopes_granted"`
	AuthURL       string    `json:"auth_url"`
	BrowserOpened bool      `json:"browser_opened"`
	ExpiresAt     time.Time `json:"expires_at"`
}

type Status struct {
	Profile        string    `json:"profile"`
	Email          string    `json:"email,omitempty"`
	Authenticated  bool      `json:"authenticated"`
	ScopesGranted  []string  `json:"scopes_granted"`
	TokenExpiresAt time.Time `json:"token_expires_at,omitempty"`
}

type Manager struct {
	ConfigStore *config.Store
	TokenStore  TokenStore
	OpenURL     func(string) error
}

func NewManager(configStore *config.Store, tokenStore TokenStore) *Manager {
	return &Manager{
		ConfigStore: configStore,
		TokenStore:  tokenStore,
		OpenURL:     platform.OpenURL,
	}
}

func (m *Manager) Login(ctx context.Context, req LoginRequest) (*LoginResult, error) {
	cfg, err := m.ConfigStore.Load()
	if err != nil {
		return nil, err
	}
	profileName := req.Profile
	if profileName == "" {
		profileName = cfg.ActiveProfile
		if profileName == "" {
			profileName = "default"
		}
	}
	scopes := NormalizeScopes(req.Scopes)
	if len(scopes) == 0 {
		scopes = DefaultReadScopes
	}
	clientID, clientSecret := m.clientCredentials(cfg)
	if clientID == "" {
		return nil, ErrMissingClientID
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("create callback listener: %w", err)
	}
	defer listener.Close()

	redirectURL := fmt.Sprintf("http://%s/oauth2/callback", listener.Addr().String())
	oauthCfg := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
	}
	state, err := randomURLSafe(32)
	if err != nil {
		return nil, err
	}
	verifier, err := randomURLSafe(64)
	if err != nil {
		return nil, err
	}
	challenge := codeChallenge(verifier)
	authURL := oauthCfg.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("prompt", "consent"),
	)

	type callbackPayload struct {
		code string
		err  error
	}
	cbCh := make(chan callbackPayload, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2/callback", func(w http.ResponseWriter, r *http.Request) {
		if queryErr := r.URL.Query().Get("error"); queryErr != "" {
			select {
			case cbCh <- callbackPayload{err: fmt.Errorf("oauth callback error: %s", queryErr)}:
			default:
			}
			http.Error(w, "Authentication failed. Return to terminal.", http.StatusBadRequest)
			return
		}
		if got := r.URL.Query().Get("state"); got != state {
			select {
			case cbCh <- callbackPayload{err: fmt.Errorf("oauth state mismatch")}:
			default:
			}
			http.Error(w, "Invalid state. Return to terminal.", http.StatusBadRequest)
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			select {
			case cbCh <- callbackPayload{err: fmt.Errorf("missing authorization code")}:
			default:
			}
			http.Error(w, "Missing code. Return to terminal.", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte("Authentication complete. You can close this tab and return to gc."))
		select {
		case cbCh <- callbackPayload{code: code}:
		default:
		}
	})
	srv := &http.Server{Handler: mux}
	serveErrCh := make(chan error, 1)
	go func() {
		err := srv.Serve(listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErrCh <- err
		}
	}()

	browserOpened := false
	if req.OpenBrowser {
		if err := m.OpenURL(authURL); err == nil {
			browserOpened = true
		}
	}
	if req.OnAuthURL != nil {
		req.OnAuthURL(authURL, browserOpened)
	}

	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}

	var payload callbackPayload
	select {
	case payload = <-cbCh:
	case err := <-serveErrCh:
		_ = srv.Shutdown(context.Background())
		return nil, fmt.Errorf("oauth callback server failed: %w", err)
	case <-time.After(timeout):
		_ = srv.Shutdown(context.Background())
		return nil, fmt.Errorf("timed out waiting for OAuth callback")
	case <-ctx.Done():
		_ = srv.Shutdown(context.Background())
		return nil, ctx.Err()
	}
	_ = srv.Shutdown(context.Background())
	if payload.err != nil {
		return nil, payload.err
	}

	token, err := oauthCfg.Exchange(ctx, payload.code, oauth2.SetAuthURLParam("code_verifier", verifier))
	if err != nil {
		return nil, fmt.Errorf("exchange authorization code: %w", err)
	}
	if err := m.TokenStore.SaveToken(profileName, token); err != nil {
		return nil, err
	}

	email := emailFromToken(token)
	if _, err := m.ConfigStore.Update(func(cfg *config.Config) error {
		if cfg.OAuthClientID == "" {
			cfg.OAuthClientID = clientID
		}
		if cfg.OAuthClientSecret == "" && clientSecret != "" {
			cfg.OAuthClientSecret = clientSecret
		}
		cfg.ActiveProfile = profileName
		profile := cfg.EnsureProfile(profileName)
		if email != "" {
			profile.Email = email
		}
		cfg.AddScopes(profileName, scopes)
		return nil
	}); err != nil {
		return nil, err
	}

	return &LoginResult{
		Profile:       profileName,
		Email:         email,
		ScopesGranted: scopes,
		AuthURL:       authURL,
		BrowserOpened: browserOpened,
		ExpiresAt:     token.Expiry,
	}, nil
}

func (m *Manager) EnsureScopes(ctx context.Context, profile string, required []string, openBrowser bool, onURL func(string, bool)) error {
	required = NormalizeScopes(required)
	if len(required) == 0 {
		return nil
	}
	cfg, err := m.ConfigStore.Load()
	if err != nil {
		return err
	}
	if profile == "" {
		profile = cfg.ActiveProfile
	}
	if profile == "" {
		profile = "default"
	}
	if cfg.HasScopes(profile, required) {
		return nil
	}
	activeProfile := cfg.EnsureProfile(profile)
	union := make([]string, 0, len(activeProfile.ScopesGranted)+len(required))
	union = append(union, activeProfile.ScopesGranted...)
	union = append(union, required...)
	union = NormalizeScopes(union)
	_, err = m.Login(ctx, LoginRequest{
		Profile:     profile,
		Scopes:      union,
		OpenBrowser: openBrowser,
		OnAuthURL:   onURL,
	})
	return err
}

func (m *Manager) TokenSource(ctx context.Context, profile string, required []string) (oauth2.TokenSource, *config.Profile, error) {
	cfg, err := m.ConfigStore.Load()
	if err != nil {
		return nil, nil, err
	}
	if profile == "" {
		profile = cfg.ActiveProfile
	}
	if profile == "" {
		profile = "default"
	}
	p := cfg.EnsureProfile(profile)

	missing := missingScopes(p.ScopesGranted, required)
	if len(missing) > 0 {
		return nil, nil, ScopesRequiredError{Missing: missing}
	}

	clientID, clientSecret := m.clientCredentials(cfg)
	if clientID == "" {
		return nil, nil, ErrMissingClientID
	}
	tok, err := m.TokenStore.LoadToken(profile)
	if err != nil {
		return nil, nil, err
	}
	oauthCfg := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       p.ScopesGranted,
	}
	source := oauthCfg.TokenSource(ctx, tok)
	refreshed, err := source.Token()
	if err != nil {
		return nil, nil, fmt.Errorf("refresh token: %w", err)
	}
	if err := m.TokenStore.SaveToken(profile, refreshed); err != nil {
		return nil, nil, err
	}
	return oauth2.ReuseTokenSource(refreshed, source), p, nil
}

func (m *Manager) Status(profile string) (*Status, error) {
	cfg, err := m.ConfigStore.Load()
	if err != nil {
		return nil, err
	}
	if profile == "" {
		profile = cfg.ActiveProfile
	}
	if profile == "" {
		profile = "default"
	}
	p := cfg.EnsureProfile(profile)
	status := &Status{
		Profile:       profile,
		Email:         p.Email,
		Authenticated: false,
		ScopesGranted: append([]string(nil), p.ScopesGranted...),
	}
	tok, err := m.TokenStore.LoadToken(profile)
	if err != nil {
		if errors.Is(err, ErrNoToken) {
			return status, nil
		}
		return nil, err
	}
	status.Authenticated = true
	status.TokenExpiresAt = tok.Expiry
	return status, nil
}

func (m *Manager) Logout(profile string) error {
	cfg, err := m.ConfigStore.Load()
	if err != nil {
		return err
	}
	if profile == "" {
		profile = cfg.ActiveProfile
	}
	if profile == "" {
		profile = "default"
	}
	if err := m.TokenStore.DeleteToken(profile); err != nil {
		return err
	}
	_, err = m.ConfigStore.Update(func(cfg *config.Config) error {
		p := cfg.EnsureProfile(profile)
		p.ScopesGranted = nil
		return nil
	})
	return err
}

func (m *Manager) clientCredentials(cfg *config.Config) (clientID, clientSecret string) {
	clientID = strings.TrimSpace(os.Getenv("GC_OAUTH_CLIENT_ID"))
	clientSecret = strings.TrimSpace(os.Getenv("GC_OAUTH_CLIENT_SECRET"))
	if clientID == "" {
		clientID = strings.TrimSpace(cfg.OAuthClientID)
	}
	if clientSecret == "" {
		clientSecret = strings.TrimSpace(cfg.OAuthClientSecret)
	}
	if clientID == "" {
		clientID = DefaultOAuthClientID
	}
	if clientSecret == "" {
		clientSecret = DefaultOAuthClientSecret
	}
	return clientID, clientSecret
}

func randomURLSafe(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func codeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func emailFromToken(tok *oauth2.Token) string {
	raw, _ := tok.Extra("id_token").(string)
	if raw == "" {
		return ""
	}
	parts := strings.Split(raw, ".")
	if len(parts) < 2 {
		return ""
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}
	var claims struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return ""
	}
	return claims.Email
}

func missingScopes(granted, required []string) []string {
	if len(required) == 0 {
		return nil
	}
	set := map[string]struct{}{}
	for _, scope := range granted {
		set[scope] = struct{}{}
	}
	missing := make([]string, 0)
	for _, scope := range required {
		if _, ok := set[scope]; !ok {
			missing = append(missing, scope)
		}
	}
	sort.Strings(missing)
	return missing
}
