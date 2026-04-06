package errorx

import "testing"

func TestEnhanceMessageCommandNotFound(t *testing.T) {
	msg, ok := EnhanceMessage("gh: command not found", EnhanceOptions{
		ToolHint: func(command string) (string, bool) {
			if command == "gh" {
				return "Install GitHub CLI: https://cli.github.com", true
			}
			return "", false
		},
	})
	if !ok {
		t.Fatal("expected message to be enhanced")
	}
	want := "gh: command not found\n\n  Install GitHub CLI: https://cli.github.com"
	if msg != want {
		t.Fatalf("unexpected message: %q", msg)
	}
}

func TestEnhanceMessageUnknownError(t *testing.T) {
	if msg, ok := EnhanceMessage("something else", EnhanceOptions{}); ok || msg != "" {
		t.Fatalf("expected no enhancement, got ok=%v msg=%q", ok, msg)
	}
}
