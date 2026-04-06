package cmd

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kasperbasse/skel/internal/profile"
	"github.com/kasperbasse/skel/internal/restore"
)

func TestGatherMissingToolsForSection(t *testing.T) {
	group := scanGroup{RestoreKeys: []string{"homebrew", "mas"}}
	missing := map[string][]string{
		"homebrew": {"brew", "mas"},
		"mas":      {"mas", "brew"},
	}

	blocked, tools := gatherMissingToolsForSection(group, missing)
	if !blocked {
		t.Fatal("expected section to be blocked when tools are missing")
	}

	want := []string{"brew", "mas"}
	if !reflect.DeepEqual(tools, want) {
		t.Fatalf("gatherMissingToolsForSection() tools = %v, want %v", tools, want)
	}
}

func TestGatherMissingToolsForSectionUnblocked(t *testing.T) {
	group := scanGroup{RestoreKeys: []string{"shell"}}

	blocked, tools := gatherMissingToolsForSection(group, map[string][]string{})
	if blocked {
		t.Fatal("expected section to remain unblocked when no tools are missing")
	}
	if len(tools) != 0 {
		t.Fatalf("expected no missing tools, got %v", tools)
	}
}

func TestHasRestorableDataSkipsNoRestoreKeys(t *testing.T) {
	// A profile with only SSH keys and system data — these groups have no RestoreKeys.
	p := &profile.Profile{
		SSH: profile.SSHProfile{
			Keys: []profile.SSHKey{{Filename: "id_ed25519", Fingerprint: "SHA256:abc"}},
		},
		System: profile.SystemProfile{
			Hostname:     "mymac",
			MacOSVersion: "14.0",
			ChipArch:     "arm64",
		},
	}
	opts := &restore.Options{}
	if hasRestorableData(p, opts) {
		t.Error("hasRestorableData should return false when only groups without RestoreKeys have data")
	}
}

func TestHasRestorableDataWithRestorableSection(t *testing.T) {
	p := &profile.Profile{
		Homebrew: profile.HomebrewProfile{
			Formulas: []string{"git"},
		},
	}
	opts := &restore.Options{}
	if !hasRestorableData(p, opts) {
		t.Error("hasRestorableData should return true when a restorable section has data")
	}
}

func TestHasRestorableDataOnlyFlagRespected(t *testing.T) {
	p := &profile.Profile{
		Homebrew: profile.HomebrewProfile{Formulas: []string{"git"}},
		Git:      profile.GitProfile{UserName: "Test"},
	}

	// --only shell: no shell data → should be false
	opts := &restore.Options{Sections: map[string]bool{"shell": true}}
	if hasRestorableData(p, opts) {
		t.Error("hasRestorableData should return false when --only matches no sections with data")
	}

	// --only homebrew: homebrew has data → should be true
	opts2 := &restore.Options{Sections: map[string]bool{"homebrew": true}}
	if !hasRestorableData(p, opts2) {
		t.Error("hasRestorableData should return true when --only matches a section with data")
	}
}

func TestPrintMissingToolsWarningAllSections(t *testing.T) {
	out := captureStdout(func() { printMissingToolsWarning(1, false) })
	if !strings.Contains(out, "all sections") {
		t.Errorf("expected 'all sections' when scoped=false, got: %q", out)
	}
	if strings.Contains(out, "selected sections") {
		t.Errorf("unexpected 'selected sections' when scoped=false, got: %q", out)
	}
}

func TestPrintMissingToolsWarningSelectedSections(t *testing.T) {
	out := captureStdout(func() { printMissingToolsWarning(2, true) })
	if !strings.Contains(out, "selected sections") {
		t.Errorf("expected 'selected sections' when scoped=true, got: %q", out)
	}
	if strings.Contains(out, "all sections") {
		t.Errorf("unexpected 'all sections' when scoped=true, got: %q", out)
	}
}

func TestOnlyFlagHelpIncludesDefaults(t *testing.T) {
	flag := restoreCmd.Flags().Lookup("only")
	if flag == nil {
		t.Fatal("--only flag not found")
	}
	if !strings.Contains(flag.Usage, "defaults") {
		t.Errorf("--only flag usage should mention 'defaults', got: %q", flag.Usage)
	}
}
