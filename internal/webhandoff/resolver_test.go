package webhandoff

import "testing"

func TestResolveCourseSettings(t *testing.T) {
	r := Resolver{}
	h := r.Resolve("invite-code", "123456", nil)
	if h.Blocked {
		t.Fatalf("expected non-blocked handoff, got blocked: %s", h.Reason)
	}
	if h.URL == "" {
		t.Fatalf("expected handoff URL")
	}
}

func TestResolveUnknownFeatureBlocked(t *testing.T) {
	r := Resolver{}
	h := r.Resolve("unknown-feature", "", nil)
	if !h.Blocked {
		t.Fatalf("expected blocked unknown handoff")
	}
}
