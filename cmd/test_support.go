package cmd

import (
	"github.com/kasperbasse/skel/internal/profile"
)

// TestProfile creates a minimal Profile for testing.
// Use this in tests instead of manually building profiles.
// Example: p := NewTestProfile().WithHomebrew([]string{"git", "vim"}).Build()
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
