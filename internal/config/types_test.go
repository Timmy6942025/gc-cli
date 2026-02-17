package config

import "testing"

func TestAddScopesAndHasScopes(t *testing.T) {
	cfg := Default()
	cfg.AddScopes("default", []string{"b", "a", "a"})
	if !cfg.HasScopes("default", []string{"a", "b"}) {
		t.Fatalf("expected scopes to be present")
	}
	if cfg.HasScopes("default", []string{"c"}) {
		t.Fatalf("expected missing scope to be absent")
	}
}

func TestEnsureProfile(t *testing.T) {
	cfg := Default()
	p := cfg.EnsureProfile("new-profile")
	if p.Name != "new-profile" {
		t.Fatalf("expected profile name to be new-profile, got %q", p.Name)
	}
	if len(cfg.Profiles) < 2 {
		t.Fatalf("expected profile to be appended")
	}
}
