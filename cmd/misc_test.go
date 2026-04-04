package cmd

import (
	"testing"
)

// ---------------------------------------------------------------------------
// headlineIcon
// ---------------------------------------------------------------------------

func TestHeadlineIconKnown(t *testing.T) {
	known := map[string]string{
		"scan":    "🔍",
		"restore": "🚀",
		"doctor":  "🩺",
		"import":  "📥",
		"delete":  "🗑",
		"update":  "🔄",
		"clone":   "🧬",
	}
	for key, want := range known {
		if got := headlineIcon(key); got != want {
			t.Errorf("headlineIcon(%q) = %q, want %q", key, got, want)
		}
	}
}

func TestHeadlineIconUnknown(t *testing.T) {
	// Unknown keys fall back to the default package icon.
	if got := headlineIcon("nonexistent"); got != "📦" {
		t.Errorf("headlineIcon(unknown) = %q, want '📦'", got)
	}
}

// ---------------------------------------------------------------------------
// randomMessage
// ---------------------------------------------------------------------------

func TestRandomMessage(t *testing.T) {
	msgs := []string{"alpha", "bravo", "charlie"}
	seen := make(map[string]bool)
	// Run enough times to expect every message to appear.
	for i := 0; i < 200; i++ {
		m := randomMessage(msgs)
		if !seen[m] {
			seen[m] = true
		}
		found := false
		for _, v := range msgs {
			if v == m {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("randomMessage returned unexpected value %q", m)
		}
	}
}

func TestRandomMessageSingle(t *testing.T) {
	msgs := []string{"only one"}
	for i := 0; i < 10; i++ {
		if got := randomMessage(msgs); got != "only one" {
			t.Errorf("randomMessage = %q, want %q", got, "only one")
		}
	}
}
