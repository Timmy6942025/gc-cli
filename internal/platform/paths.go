package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const appDirName = "gc-cli"

func ConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve config dir: %w", err)
	}
	return filepath.Join(base, appDirName), nil
}

func StateDir() (string, error) {
	base := os.Getenv("XDG_STATE_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home dir for state path: %w", err)
		}
		switch runtime.GOOS {
		case "darwin":
			base = filepath.Join(home, "Library", "Application Support")
		case "windows":
			appData := os.Getenv("AppData")
			if appData != "" {
				base = appData
			} else {
				base = filepath.Join(home, "AppData", "Roaming")
			}
		default:
			base = filepath.Join(home, ".local", "state")
		}
	}
	return filepath.Join(base, appDirName), nil
}

func EnsureDir(path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("create dir %s: %w", path, err)
	}
	return nil
}
