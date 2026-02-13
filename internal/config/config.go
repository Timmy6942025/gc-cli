package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/timboy697/gc-cli/internal/auth"
)

type Config struct {
	ConfigPath      string          `mapstructure:"-"`
	Auth            AuthConfig      `mapstructure:"auth"`
	GoogleClassroom ClassroomConfig `mapstructure:"google_classroom"`
}

type AuthConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	TokenFile    string `mapstructure:"token_file"`
}

type ClassroomConfig struct {
	CourseID string `mapstructure:"course_id"`
}

func Default() *Config {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".config", "gc-cli")
	defaultAuth := auth.DefaultAuthConfig()

	return &Config{
		ConfigPath: filepath.Join(configDir, "config.yaml"),
		Auth: AuthConfig{
			ClientID:     defaultAuth.ClientID,
			ClientSecret: defaultAuth.ClientSecret,
			TokenFile:    filepath.Join(configDir, "token.json"),
		},
		GoogleClassroom: ClassroomConfig{},
	}
}

func Load() (*Config, error) {
	cfg := Default()

	viper.SetConfigType("yaml")
	viper.SetConfigFile(cfg.ConfigPath)

	viper.SetDefault("auth.client_id", cfg.Auth.ClientID)
	viper.SetDefault("auth.client_secret", cfg.Auth.ClientSecret)
	viper.SetDefault("auth.token_file", cfg.Auth.TokenFile)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	cfg.ConfigPath = viper.ConfigFileUsed()

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}

func EnsureConfigDir(cfg *Config) error {
	configDir := filepath.Dir(cfg.ConfigPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	return nil
}

func Save(cfg *Config) error {
	if err := EnsureConfigDir(cfg); err != nil {
		return err
	}

	viper.SetConfigFile(cfg.ConfigPath)
	viper.Set("auth", cfg.Auth)
	viper.Set("google_classroom", cfg.GoogleClassroom)

	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	return nil
}
