package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
)

func EnsureTokenDir(tokenFile string) error {
	dir := filepath.Dir(tokenFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create token directory: %w", err)
	}
	return nil
}

func TokenFromFile(tokenFile string) (*oauth2.Token, error) {
	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	return &token, nil
}

func TokenToFile(tokenFile string, token *oauth2.Token) error {
	if err := EnsureTokenDir(tokenFile); err != nil {
		return err
	}

	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	if err := os.WriteFile(tokenFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

func TokenExists(tokenFile string) bool {
	_, err := os.Stat(tokenFile)
	return err == nil
}

func ValidateToken(ctx context.Context, cfg *Config, token *oauth2.Token) bool {
	if token == nil {
		return false
	}

	if token.Expiry.After(time.Now()) {
		return true
	}

	if token.RefreshToken != "" {
		return true
	}

	return false
}

func GetTokenSource(ctx context.Context, cfg *Config, token *oauth2.Token) oauth2.TokenSource {
	oauthCfg := cfg.OAuth2Config()
	return oauthCfg.TokenSource(ctx, token)
}

func RefreshToken(ctx context.Context, cfg *Config, token *oauth2.Token) (*oauth2.Token, error) {
	if token.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	oauthCfg := cfg.OAuth2Config()
	ts := oauthCfg.TokenSource(ctx, token)

	newToken, err := ts.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	return newToken, nil
}

func GetValidToken(ctx context.Context, cfg *Config) (*oauth2.Token, error) {
	token, err := TokenFromFile(cfg.TokenFile)
	if err != nil {
		return nil, fmt.Errorf("no valid token found, please run 'gc-cli auth login': %w", err)
	}

	if token.Expiry.After(time.Now()) {
		return token, nil
	}

	if token.RefreshToken != "" {
		fmt.Println("Token expired, refreshing...")
		newToken, err := RefreshToken(ctx, cfg, token)
		if err == nil {
			if err := TokenToFile(cfg.TokenFile, newToken); err != nil {
				return nil, fmt.Errorf("failed to save refreshed token: %w", err)
			}
			return newToken, nil
		}
		fmt.Printf("Token refresh failed: %v\n", err)
	}

	return nil, fmt.Errorf("token expired, please run 'gc-cli auth login'")
}
