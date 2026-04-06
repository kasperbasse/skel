package cmd

import (
	"strings"
	"testing"

	"github.com/kasperbasse/skel/internal/profile"
)

// ---------------------------------------------------------------------------
// prepareForExport
// ---------------------------------------------------------------------------

func TestPrepareForExportRedactsByDefault(t *testing.T) {
	p := NewTestProfile().
		WithGit("Jane Doe", "jane@example.com").
		Build()
	p.System.Hostname = "my-macbook"
	p.SSH.Keys = []profile.SSHKey{
		{Filename: "id_ed25519", Comment: "jane@example.com"},
	}

	out := captureStdout(func() {
		result := prepareForExport(p, false)
		if result.Git.UserName != "" {
			t.Errorf("expected git user name to be redacted, got %q", result.Git.UserName)
		}
		if result.Git.UserEmail != "" {
			t.Errorf("expected git user email to be redacted, got %q", result.Git.UserEmail)
		}
		if result.System.Hostname != "" {
			t.Errorf("expected hostname to be redacted, got %q", result.System.Hostname)
		}
		if len(result.SSH.Keys) > 0 && result.SSH.Keys[0].Comment != "" {
			t.Errorf("expected SSH key comment to be redacted, got %q", result.SSH.Keys[0].Comment)
		}
		if result.Machine != "shared" {
			// Redact() sets Machine to "shared" to strip the original hostname identifier.
			t.Errorf("expected machine to be %q, got %q", "shared", result.Machine)
		}
	})

	if !strings.Contains(out, "Redacted") {
		t.Errorf("expected redaction message in output, got: %q", out)
	}
}

func TestPrepareForExportNoRedact(t *testing.T) {
	p := NewTestProfile().
		WithGit("Jane Doe", "jane@example.com").
		Build()
	p.System.Hostname = "my-macbook"

	out := captureStdout(func() {
		result := prepareForExport(p, true)
		if result.Git.UserName != "Jane Doe" {
			t.Errorf("expected git user name to be preserved, got %q", result.Git.UserName)
		}
		if result.Git.UserEmail != "jane@example.com" {
			t.Errorf("expected git user email to be preserved, got %q", result.Git.UserEmail)
		}
		if result.System.Hostname != "my-macbook" {
			t.Errorf("expected hostname to be preserved, got %q", result.System.Hostname)
		}
	})

	if !strings.Contains(out, "without redaction") {
		t.Errorf("expected no-redact warning in output, got: %q", out)
	}
}

func TestPrepareForExportDoesNotMutateOriginal(t *testing.T) {
	p := NewTestProfile().
		WithGit("Jane Doe", "jane@example.com").
		Build()
	p.System.Hostname = "my-macbook"

	captureStdout(func() {
		_ = prepareForExport(p, false)
	})

	// Original profile should be unchanged.
	if p.Git.UserName != "Jane Doe" {
		t.Errorf("prepareForExport mutated original: git user name = %q, want %q", p.Git.UserName, "Jane Doe")
	}
	if p.System.Hostname != "my-macbook" {
		t.Errorf("prepareForExport mutated original: hostname = %q, want %q", p.System.Hostname, "my-macbook")
	}
}
