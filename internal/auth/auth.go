package auth

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
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

const (
	envClientID     = "GC_CLI_CLIENT_ID"
	envClientSecret = "GC_CLI_CLIENT_SECRET"

	DefaultClientID     = "597878429548-jc1e74sbsk9fhqg6fbivmclrnq70hs3t.apps.googleusercontent.com"
	DefaultClientSecret = "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"
)

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
		TokenFile:    tokenFile,
	}
}

func BrowserFlow(ctx context.Context, cfg *Config) (*oauth2.Token, error) {
	token, err := tryAutoCallback(ctx, cfg)
	if err == nil {
		return token, nil
	}
	fmt.Println("Using fallback method...")
	return manualFlow(ctx, cfg)
}

func tryAutoCallback(ctx context.Context, cfg *Config) (*oauth2.Token, error) {
	oauthCfg := cfg.OAuth2Config()

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, fmt.Errorf("listen: %w", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	redirectURL := fmt.Sprintf("http://localhost:%d", port)

	oauthCfg.RedirectURL = redirectURL

	state := fmt.Sprintf("gc-cli-%d", time.Now().UnixNano())
	authURL := oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOffline)

	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("error") != "" {
			errChan <- fmt.Errorf("oauth error: %s", q.Get("error_description"))
			return
		}
		if q.Get("state") != state {
			errChan <- fmt.Errorf("state mismatch")
			return
		}
		code := q.Get("code")
		if code == "" {
			errChan <- fmt.Errorf("no code")
			return
		}
		codeChan <- code
		io.WriteString(w, "<html><body><h1>✓ Success! You can close this window.</h1></body></html>")
	})

	server := &http.Server{Addr: redirectURL, Handler: mux}
	go server.Serve(listener)

	fmt.Println("🌐 Opening browser...")
	_ = openBrowser(authURL)
	fmt.Printf("📋 Or visit: %s\n", authURL)
	fmt.Println("⏳ Waiting...")

	select {
	case code := <-codeChan:
		server.Close()
		token, err := oauthCfg.Exchange(ctx, code)
		if err != nil {
			return nil, fmt.Errorf("exchange: %w", err)
		}
		fmt.Println("✓ Logged in!")
		return token, nil
	case err := <-errChan:
		server.Close()
		return nil, err
	case <-time.After(60 * time.Second):
		server.Close()
		return nil, fmt.Errorf("timeout")
	}
}

func manualFlow(ctx context.Context, cfg *Config) (*oauth2.Token, error) {
	oauthCfg := cfg.OAuth2Config()

	state := fmt.Sprintf("gc-cli-%d", time.Now().UnixNano())
	authURL := oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOffline)

	fmt.Println("╔═══════════════════════════════════════════╗")
	fmt.Println("║     GOOGLE CLASSROOM AUTHENTICATION       ║")
	fmt.Println("╚═══════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("1. Open this URL:")
	fmt.Printf("   %s\n", authURL)
	fmt.Println()
	fmt.Println("2. Sign in and authorize")
	fmt.Println()
	fmt.Print("3. Paste the URL you're redirected to: ")

	var redirectURL string
	fmt.Scanln(&redirectURL)
	if redirectURL == "" {
		return nil, fmt.Errorf("no URL")
	}

	parsed, err := url.Parse(redirectURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	code := parsed.Query().Get("code")
	if code == "" {
		return nil, fmt.Errorf("no code in URL")
	}

	token, err := oauthCfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchange failed: %w", err)
	}

	fmt.Println("✓ Logged in!")
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
		return fmt.Errorf("no browser")
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

func DefaultAuthConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	tokenFile := filepath.Join(homeDir, ".config", "gc-cli", "token.json")
	clientID := os.Getenv(envClientID)
	clientSecret := os.Getenv(envClientSecret)
	if clientID == "" {
		clientID = DefaultClientID
	}
	if clientSecret == "" {
		clientSecret = DefaultClientSecret
	}
	return &Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost",
		TokenFile:    tokenFile,
	}
}
