package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/kasperbasse/skel/internal/profile"
)

func TestRunDeleteEnhancesMissingProfileWithSuggestion(t *testing.T) {
	profile.SetProfileDirOverride(t.TempDir())
	t.Cleanup(func() { profile.SetProfileDirOverride("") })

	if _, err := profile.Save(&profile.Profile{Name: "default", CreatedAt: time.Now()}); err != nil {
		t.Fatalf("Save(default): %v", err)
	}

	err := runDelete(nil, []string{"defalt"})
	if err == nil {
		t.Fatal("runDelete() error = nil, want enhanced not found error")
	}

	msg := err.Error()
	if !strings.Contains(msg, "Did you mean 'default'?") {
		t.Fatalf("runDelete() error = %q, want fuzzy suggestion", msg)
	}
	if !strings.Contains(msg, "Use 'skel list' to see all profiles") {
		t.Fatalf("runDelete() error = %q, want profiles guidance", msg)
	}
}

func TestRunDeleteEnhancesMissingProfileWithoutSuggestion(t *testing.T) {
	profile.SetProfileDirOverride(t.TempDir())
	t.Cleanup(func() { profile.SetProfileDirOverride("") })

	err := runDelete(nil, []string{"missing"})
	if err == nil {
		t.Fatal("runDelete() error = nil, want enhanced not found error")
	}

	msg := err.Error()
	if !strings.Contains(msg, "profile 'missing' not found") {
		t.Fatalf("runDelete() error = %q, want missing profile message", msg)
	}
	if !strings.Contains(msg, "Use 'skel list' to see available profiles") {
		t.Fatalf("runDelete() error = %q, want fallback guidance", msg)
	}
}

