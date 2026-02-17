package config

import (
	"errors"
	"sort"
)

const Version = "v1"

var ErrNoActiveProfile = errors.New("no active profile configured")

// Config is stored as JSON and controls auth/profile/runtime defaults.
type Config struct {
	Version           string    `json:"version"`
	Profiles          []Profile `json:"profiles"`
	ActiveProfile     string    `json:"active_profile"`
	OAuthClientID     string    `json:"oauth_client_id"`
	OAuthClientSecret string    `json:"oauth_client_secret,omitempty"`
	DefaultOrgDomain  string    `json:"default_org_domain,omitempty"`
	PreviewEnabled    bool      `json:"preview_enabled"`
	TelemetryOptIn    bool      `json:"telemetry_opt_in"`
}

type Profile struct {
	Name          string   `json:"name"`
	Email         string   `json:"email,omitempty"`
	ScopesGranted []string `json:"scopes_granted"`
}

func Default() *Config {
	return &Config{
		Version:        Version,
		Profiles:       []Profile{{Name: "default", ScopesGranted: []string{}}},
		ActiveProfile:  "default",
		PreviewEnabled: false,
		TelemetryOptIn: false,
	}
}

func (c *Config) Active() (*Profile, error) {
	for i := range c.Profiles {
		if c.Profiles[i].Name == c.ActiveProfile {
			return &c.Profiles[i], nil
		}
	}
	return nil, ErrNoActiveProfile
}

func (c *Config) EnsureProfile(name string) *Profile {
	for i := range c.Profiles {
		if c.Profiles[i].Name == name {
			return &c.Profiles[i]
		}
	}
	c.Profiles = append(c.Profiles, Profile{Name: name, ScopesGranted: []string{}})
	return &c.Profiles[len(c.Profiles)-1]
}

func (c *Config) AddScopes(profileName string, scopes []string) {
	p := c.EnsureProfile(profileName)
	existing := make(map[string]struct{}, len(p.ScopesGranted))
	for _, s := range p.ScopesGranted {
		existing[s] = struct{}{}
	}
	for _, s := range scopes {
		if s == "" {
			continue
		}
		existing[s] = struct{}{}
	}
	p.ScopesGranted = p.ScopesGranted[:0]
	for s := range existing {
		p.ScopesGranted = append(p.ScopesGranted, s)
	}
	sort.Strings(p.ScopesGranted)
}

func (c *Config) HasScopes(profileName string, scopes []string) bool {
	for i := range c.Profiles {
		if c.Profiles[i].Name != profileName {
			continue
		}
		set := map[string]struct{}{}
		for _, granted := range c.Profiles[i].ScopesGranted {
			set[granted] = struct{}{}
		}
		for _, needed := range scopes {
			if _, ok := set[needed]; !ok {
				return false
			}
		}
		return true
	}
	return false
}
