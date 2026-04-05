package cmd

import (
	"testing"

	"github.com/kasperbasse/skel/internal/profile"
	"github.com/kasperbasse/skel/internal/restore"
)

// TestProfile creates a minimal Profile for testing.
// Use this in tests instead of manually building profiles.
// Example: p := TestProfile().WithHomebrew([]string{"git", "vim"}).Build()
type TestProfile struct {
	p *profile.Profile
}

// NewTestProfile returns a minimal, safe Profile for testing.
func NewTestProfile() *TestProfile {
	return &TestProfile{
		p: &profile.Profile{
			Name:    "test-profile",
			Machine: "test-machine",
		},
	}
}

// WithHomebrew adds formulas to the test profile.
func (tp *TestProfile) WithHomebrew(formulas []string) *TestProfile {
	tp.p.Homebrew.Formulas = formulas
	return tp
}

// WithGit configures git data.
func (tp *TestProfile) WithGit(username, email string) *TestProfile {
	tp.p.Git.UserName = username
	tp.p.Git.UserEmail = email
	return tp
}

// Build returns the constructed Profile.
func (tp *TestProfile) Build() *profile.Profile {
	return tp.p
}

// TestOptions creates minimal RestoreOptions for testing.
func TestOptions(sections map[string]bool) *restore.Options {
	return &restore.Options{
		Sections: sections,
	}
}

// AssertError checks if an error occurred when it should have.
// Use in tests for consistent error checking.
func AssertError(t *testing.T, err error, shouldError bool, msg string) {
	t.Helper()
	if shouldError && err == nil {
		t.Errorf("%s: expected error, got nil", msg)
	}
	if !shouldError && err != nil {
		t.Errorf("%s: unexpected error: %v", msg, err)
	}
}

// AssertEqual checks if two values are equal.
// Generic helper to reduce boilerplate in tests.
func AssertEqual(t *testing.T, got, want interface{}, msg string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %v, want %v", msg, got, want)
	}
}
