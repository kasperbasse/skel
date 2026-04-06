package cmd

import (
	"reflect"
	"testing"
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
