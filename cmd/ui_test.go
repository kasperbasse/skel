package cmd

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestEnhanceError(t *testing.T) {
	tests := []struct {
		name     string
		input    error
		expected string
	}{
		{
			name:     "profile not found",
			input:    errors.New("profile 'work' not found"),
			expected: "profile 'work' not found\n\nUse 'skel list' to see available profiles",
		},
		{
			name:     "brew not found",
			input:    errors.New("brew not found"),
			expected: "brew not found\n\nInstall Homebrew first: https://brew.sh",
		},
		{
			name:     "gh not found",
			input:    errors.New("gh: command not found"),
			expected: "gh: command not found\n\nInstall GitHub CLI: https://cli.github.com",
		},
		{
			name:     "mas not found",
			input:    errors.New("mas not found"),
			expected: "mas not found\n\nInstall mas for Mac App Store: brew install mas",
		},
		{
			name:     "code not found",
			input:    errors.New("code not found"),
			expected: "code not found\n\nInstall VS Code or ensure it's in your PATH",
		},
		{
			name:     "cursor not found",
			input:    errors.New("cursor not found"),
			expected: "cursor not found\n\nInstall Cursor or ensure it's in your PATH",
		},
		{
			name:  "invalid section",
			input: errors.New("unknown section \"bad\""),
			expected: fmt.Sprintf(
				"unknown section \"bad\"\n\nValid sections: %s",
				strings.Join(allRestoreKeys(), ", "),
			),
		},
		{
			name:     "json parsing error",
			input:    errors.New("invalid json: unexpected token"),
			expected: "invalid json: unexpected token\n\nCheck that the file contains valid JSON. For profiles, use 'skel export' to create valid files",
		},
		{
			name:     "rate limit error",
			input:    errors.New("GitHub API rate limit exceeded"),
			expected: "GitHub API rate limit exceeded\n\nSet GITHUB_TOKEN environment variable or run 'gh auth login' to increase limits",
		},
		{
			name:     "authentication error",
			input:    errors.New("authentication failed - check your GITHUB_TOKEN"),
			expected: "authentication failed - check your GITHUB_TOKEN\n\nSet GITHUB_TOKEN or run 'gh auth login' to authenticate",
		},
		{
			name:     "file too large",
			input:    errors.New("profile file too large (5MB, max 1MB)"),
			expected: "profile file too large (5MB, max 1MB)\n\nTry reducing config file sizes or split into multiple profiles",
		},
		{
			name:     "permission denied",
			input:    errors.New("permission denied"),
			expected: "permission denied\n\nTry running with sudo if you have admin privileges",
		},
		{
			name:     "network error",
			input:    errors.New("connection refused"),
			expected: "connection refused\n\nCheck your internet connection and try again",
		},
		{
			name:     "github gist error",
			input:    errors.New("gist not found (404)"),
			expected: "gist not found (404)\n\nCheck the gist URL and your GitHub authentication",
		},
		{
			name:     "unknown error",
			input:    errors.New("some unknown error"),
			expected: "some unknown error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := enhanceError(tt.input)
			if result.Error() != tt.expected {
				t.Errorf("enhanceError() = %q, want %q", result.Error(), tt.expected)
			}
		})
	}
}
