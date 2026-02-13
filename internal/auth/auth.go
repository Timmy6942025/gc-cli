package auth

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

var Scopes = []string{
	"https://www.googleapis.com/auth/classroom.courses.readonly",
	"https://www.googleapis.com/auth/classroom.coursework.me",
	"https://www.googleapis.com/auth/classroom.coursework.students",
	"https://www.googleapis.com/auth/classroom.announcements.readonly",
}

type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	TokenFile    string
}

func (c *Config) OAuth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		Scopes:       Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
		RedirectURL: c.RedirectURL,
	}
}

func NewConfig(clientID, clientSecret, tokenFile string) *Config {
	return &Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost",
		TokenFile:   tokenFile,
	}
}

func BrowserFlow(ctx context.Context, cfg *Config) (*oauth2.Token, error) {
	oauthCfg := cfg.OAuth2Config()

	state := fmt.Sprintf("gc-cli-%d", time.Now().UnixNano())
	authURL := oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOffline)

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║           GOOGLE CLASSROOM AUTHENTICATION              ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("1. Open this URL in your browser:")
	fmt.Printf("\n  %s\n\n", authURL)
	fmt.Println("2. After signing in, you'll be redirected to a page that says")
	fmt.Println("   'This site can't be reached' - that's OK!")
	fmt.Println()
	fmt.Println("3. Copy the ENTIRE URL from your browser address bar")
	fmt.Println("   (it will look like: http://localhost/?code=...&state=...)")
	fmt.Println()
	fmt.Print("4. Paste the URL here: ")

	var redirectURL string
	fmt.Scanln(&redirectURL)
	if redirectURL == "" {
		return nil, fmt.Errorf("no URL provided")
	}

	parsed, err := url.Parse(redirectURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	code := parsed.Query().Get("code")
	if code == "" {
		return nil, fmt.Errorf("no authorization code found in URL")
	}

	token, err := oauthCfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	fmt.Println("✓ Authentication successful!")
	return token, nil
}

func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch {
	case isWsl():
		cmd = exec.Command("cmd.exe", "/c", "start", "", url)
	case isWindows():
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case isMac():
		cmd = exec.Command("open", url)
	default:
		browsers := []string{"xdg-open", "gnome-open", "firefox", "google-chrome", "chromium-browser"}
		for _, b := range browsers {
			if _, err := exec.LookPath(b); err == nil {
				cmd = exec.Command(b, url)
				break
			}
		}
	}

	if cmd == nil {
		return fmt.Errorf("no suitable browser found")
	}

	_ = cmd.Start()
	return nil
}

func isWindows() bool {
	return os.PathSeparator == '\\'
}

func isMac() bool {
	return strings.Contains(strings.ToLower(os.Getenv("GOOS")), "darwin")
}

func isWsl() bool {
	if wsl, ok := os.LookupEnv("WSL_DISTRO_NAME"); ok && wsl != "" {
		return true
	}
	return false
}

func GetConfigURL() string {
	return "https://console.cloud.google.com/apis/credentials"
}

func Configured(cfg *Config) bool {
	return cfg.ClientID != "" && cfg.ClientSecret != ""
}
