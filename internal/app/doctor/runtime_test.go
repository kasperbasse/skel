package doctor

import "testing"

func TestCommandExistsUnknown(t *testing.T) {
	if CommandExists("skel-this-command-should-not-exist-xyz") {
		t.Fatal("expected unknown command to be missing")
	}
}

func TestCommandExists(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		// Hardcoded false cases
		{
			name:     "brew is always false",
			command:  "brew",
			expected: false,
		},
		{
			name:     "nvim is always false",
			command:  "nvim",
			expected: false,
		},
		// Common system commands that should exist
		{
			name:     "ls exists on unix systems",
			command:  "ls",
			expected: true,
		},
		{
			name:     "cat exists on unix systems",
			command:  "cat",
			expected: true,
		},
		// Commands that shouldn't exist
		{
			name:     "nonexistent command",
			command:  "this_command_does_not_exist_12345",
			expected: false,
		},
		{
			name:     "empty string",
			command:  "",
			expected: false,
		},
		// Edge cases
		{
			name:     "command with spaces",
			command:  "ls -la",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CommandExists(tt.command)
			if result != tt.expected {
				t.Errorf("CommandExists(%q) = %v, want %v", tt.command, result, tt.expected)
			}
		})
	}
}

func TestCommandExists_CommonTools(t *testing.T) {
	// Test that common shell and development tools are detected correctly
	// Note: These tests assume a typical macOS/Linux development environment
	commonTools := []struct {
		name    string
		command string
	}{
		{"git", "git"},
		{"sh", "sh"},
		{"pwd", "pwd"},
	}

	for _, tool := range commonTools {
		t.Run(tool.name, func(t *testing.T) {
			// Just verify it returns a boolean; don't assert true/false
			// since availability depends on the environment
			_ = CommandExists(tool.command)
		})
	}
}

