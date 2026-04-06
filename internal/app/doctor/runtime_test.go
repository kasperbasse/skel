package doctor

import "testing"

func TestCommandExistsUnknown(t *testing.T) {
	if CommandExists("skel-this-command-should-not-exist-xyz") {
		t.Fatal("expected unknown command to be missing")
	}
}

func TestCommandExists_EnvironmentInvariant(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		// Commands that should NOT exist (invariant across environments)
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
		// Edge cases (invariant)
		{
			name:     "command with spaces",
			command:  "this_command_does_not_exist 12345",
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
