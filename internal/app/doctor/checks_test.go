package doctor

import (
	"testing"

	"github.com/kasperbasse/skel/internal/profile"
)

func TestBuildChecksEmptyProfile(t *testing.T) {
	p := &profile.Profile{Name: "empty"}
	checks := buildChecksWith(RequiredTools(p),
		func(command string) (string, string, string, bool) { return command, command, "fix", true },
		func(command string) bool { return true },
	)
	if len(checks) != 0 {
		t.Fatalf("expected 0 checks, got %d", len(checks))
	}
}

func TestBuildChecksUsesResolverAndExists(t *testing.T) {
	p := &profile.Profile{Homebrew: profile.HomebrewProfile{Formulas: []string{"git"}}}
	checks := buildChecksWith(RequiredTools(p),
		func(command string) (string, string, string, bool) {
			if command != "brew" {
				t.Fatalf("unexpected command passed to resolver: %s", command)
			}
			return "Homebrew", "brew", "https://brew.sh", true
		},
		func(command string) bool {
			if command != "brew" {
				t.Fatalf("unexpected command passed to exists: %s", command)
			}
			return false
		},
	)

	if len(checks) != 1 {
		t.Fatalf("expected 1 check, got %d", len(checks))
	}
	if checks[0].Label != "Homebrew" || checks[0].OK || checks[0].Fix != "https://brew.sh" {
		t.Fatalf("unexpected check: %+v", checks[0])
	}
}

func TestBuildChecksFallbackWhenResolverMissing(t *testing.T) {
	p := &profile.Profile{Homebrew: profile.HomebrewProfile{Formulas: []string{"git"}}}
	checks := buildChecksWith(RequiredTools(p),
		func(command string) (string, string, string, bool) { return "", "", "", false },
		func(command string) bool { return true },
	)

	if len(checks) != 1 {
		t.Fatalf("expected 1 check, got %d", len(checks))
	}
	if checks[0].Label != "brew" {
		t.Fatalf("expected fallback label brew, got %q", checks[0].Label)
	}
	if checks[0].Fix == "" {
		t.Fatal("expected fallback fix hint")
	}
}

func TestBuildChecksNilCallbacks(t *testing.T) {
	p := &profile.Profile{Homebrew: profile.HomebrewProfile{Formulas: []string{"git"}}}

	if checks := buildChecksWith(RequiredTools(p), nil, func(command string) bool { return true }); checks != nil {
		t.Fatalf("expected nil checks when resolver is nil, got %v", checks)
	}

	if checks := buildChecksWith(RequiredTools(p), func(command string) (string, string, string, bool) { return command, command, "fix", true }, nil); checks != nil {
		t.Fatalf("expected nil checks when exists callback is nil, got %v", checks)
	}
}
