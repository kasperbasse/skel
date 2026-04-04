package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestRequireArgs(t *testing.T) {
	validator := requireArgs("show <profile-name>")
	if err := validator(&cobra.Command{}, []string{}); err == nil {
		t.Fatal("expected error for missing args")
	}

	err := validator(&cobra.Command{}, []string{})
	want := "missing required argument\n\nUsage: skel show <profile-name>"
	if err.Error() != want {
		t.Fatalf("unexpected error: %q", err.Error())
	}

	if err := validator(&cobra.Command{}, []string{"ok"}); err != nil {
		t.Fatalf("expected nil error for provided arg, got: %v", err)
	}
}

func TestRequireExactArgs(t *testing.T) {
	validator := requireExactArgs(2, "diff <profile-a> <profile-b>")

	if err := validator(&cobra.Command{}, []string{"only-one"}); err == nil {
		t.Fatal("expected missing args error")
	} else if err.Error() != "missing required argument(s)\n\nUsage: skel diff <profile-a> <profile-b>" {
		t.Fatalf("unexpected error: %q", err.Error())
	}

	if err := validator(&cobra.Command{}, []string{"a", "b", "c"}); err == nil {
		t.Fatal("expected too many args error")
	} else if err.Error() != "too many arguments\n\nUsage: skel diff <profile-a> <profile-b>" {
		t.Fatalf("unexpected error: %q", err.Error())
	}

	if err := validator(&cobra.Command{}, []string{"a", "b"}); err != nil {
		t.Fatalf("expected nil error for exact arg count, got: %v", err)
	}
}
