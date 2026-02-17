package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/timothy/gc-cli/internal/platform"
)

type Store struct {
	mu   sync.Mutex
	path string
}

func NewStore() (*Store, error) {
	dir, err := platform.ConfigDir()
	if err != nil {
		return nil, err
	}
	if err := platform.EnsureDir(dir); err != nil {
		return nil, err
	}
	return &Store{path: filepath.Join(dir, "config.json")}, nil
}

func (s *Store) Path() string {
	return s.path
}

func (s *Store) Load() (*Config, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		cfg := Default()
		if err := s.writeLocked(cfg); err != nil {
			return nil, err
		}
		return cfg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if cfg.Version == "" {
		cfg.Version = Version
	}
	if len(cfg.Profiles) == 0 {
		cfg.Profiles = []Profile{{Name: "default", ScopesGranted: []string{}}}
	}
	if cfg.ActiveProfile == "" {
		cfg.ActiveProfile = cfg.Profiles[0].Name
	}
	return &cfg, nil
}

func (s *Store) Save(cfg *Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.writeLocked(cfg)
}

func (s *Store) Update(fn func(cfg *Config) error) (*Config, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.loadLocked()
	if err != nil {
		return nil, err
	}
	if err := fn(cfg); err != nil {
		return nil, err
	}
	if err := s.writeLocked(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (s *Store) loadLocked() (*Config, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		cfg := Default()
		return cfg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if cfg.Version == "" {
		cfg.Version = Version
	}
	if len(cfg.Profiles) == 0 {
		cfg.Profiles = []Profile{{Name: "default", ScopesGranted: []string{}}}
	}
	if cfg.ActiveProfile == "" {
		cfg.ActiveProfile = cfg.Profiles[0].Name
	}
	return &cfg, nil
}

func (s *Store) writeLocked(cfg *Config) error {
	if cfg.Version == "" {
		cfg.Version = Version
	}
	if len(cfg.Profiles) == 0 {
		cfg.Profiles = []Profile{{Name: "default", ScopesGranted: []string{}}}
		cfg.ActiveProfile = "default"
	}
	if cfg.ActiveProfile == "" {
		cfg.ActiveProfile = cfg.Profiles[0].Name
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(s.path, data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}
