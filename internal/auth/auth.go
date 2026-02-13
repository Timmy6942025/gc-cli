package auth

import (
	"context"
	"fmt"
	"net"
	"net/http"
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
		RedirectURL:  "http://localhost:8080/callback",
		TokenFile:    tokenFile,
	}
}

func BrowserFlow(ctx context.Context, cfg *Config) (*oauth2.Token, error) {
	oauthCfg := cfg.OAuth2Config()

	state := fmt.Sprintf("gc-cli-%d", time.Now().UnixNano())
	authURL := oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOffline)

	fmt.Println("Opening browser for authentication...")
	if err := openBrowser(authURL); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open browser: %v\n", err)
		fmt.Printf("Please open the following URL in your browser:\n%s\n", authURL)
	}

	code, err := startCallbackServer(ctx, state)
	if err != nil {
		return nil, fmt.Errorf("callback failed: %w", err)
	}

	token, err := oauthCfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	return token, nil
}

func startCallbackServer(ctx context.Context, expectedState string) (string, error) {
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", fmt.Errorf("failed to find available port: %w", err)
	}
	defer listener.Close()

	addr := listener.Addr().String()
	port := strings.Split(addr, ":")[1]

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		if errMsg := query.Get("error"); errMsg != "" {
			errChan <- fmt.Errorf("oauth error: %s - %s", errMsg, query.Get("error_description"))
			return
		}

		state := query.Get("state")
		if state != expectedState {
			errChan <- fmt.Errorf("state mismatch: expected %s, got %s", expectedState, state)
			return
		}

		code := query.Get("code")
		if code == "" {
			errChan <- fmt.Errorf("no authorization code received")
			return
		}

		codeChan <- code
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html><html><head><title>gc-cli - Authentication Successful</title></head><body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; background: #f5f5f5;"><div style="text-align: center; padding: 40px; background: white; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1);"><h1 style="color: #4CAF50;">✓ Authentication Successful</h1><p>You can close this window and return to the terminal.</p></div></body></html>`))
	})

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case code := <-codeChan:
		server.Shutdown(ctx)
		return code, nil
	case err := <-errChan:
		server.Shutdown(ctx)
		return "", err
	case <-ctx.Done():
		server.Shutdown(ctx)
		return "", ctx.Err()
	}
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
