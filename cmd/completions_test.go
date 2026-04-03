package cmd

import (
	"testing"
	"time"

	"github.com/kasperbasse/skel/internal/profile"
)

func TestProfileNameCompletion(t *testing.T) {
	profile.SetProfileDirOverride(t.TempDir())
	t.Cleanup(func() { profile.SetProfileDirOverride("") })

	names := []string{"alpha", "beta", "gamma"}
	for _, n := range names {
		if _, err := profile.Save(&profile.Profile{Name: n, CreatedAt: time.Now()}); err != nil {
			t.Fatalf("Save(%s): %v", n, err)
		}
	}

	completions, directive := profileNameCompletion(nil, nil, "")
	if directive != 4 { // cobra.ShellCompDirectiveNoFileComp = 4
		t.Errorf("directive = %d, want 4 (NoFileComp)", directive)
	}
	if len(completions) != 3 {
		t.Fatalf("expected 3 completions, got %d: %v", len(completions), completions)
	}
	got := make(map[string]bool)
	for _, c := range completions {
		got[c] = true
	}
	for _, want := range names {
		if !got[want] {
			t.Errorf("missing completion %q", want)
		}
	}
}

func TestProfileNameCompletionEmpty(t *testing.T) {
	profile.SetProfileDirOverride(t.TempDir())
	t.Cleanup(func() { profile.SetProfileDirOverride("") })

	completions, _ := profileNameCompletion(nil, nil, "")
	if len(completions) != 0 {
		t.Errorf("expected 0 completions for empty dir, got %d", len(completions))
	}
}

func TestSingleProfileCompletionSkipsAfterFirst(t *testing.T) {
	profile.SetProfileDirOverride(t.TempDir())
	t.Cleanup(func() { profile.SetProfileDirOverride("") })

	if _, err := profile.Save(&profile.Profile{Name: "myprofile", CreatedAt: time.Now()}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// With args already provided, should return nothing.
	completions, _ := singleProfileCompletion(nil, []string{"already"}, "")
	if len(completions) != 0 {
		t.Errorf("expected 0 completions when arg already given, got %d", len(completions))
	}

	// Without args, should return profiles.
	completions, _ = singleProfileCompletion(nil, []string{}, "")
	if len(completions) != 1 {
		t.Errorf("expected 1 completion, got %d", len(completions))
	}
}

func TestTwoProfileCompletionLimit(t *testing.T) {
	profile.SetProfileDirOverride(t.TempDir())
	t.Cleanup(func() { profile.SetProfileDirOverride("") })

	for _, n := range []string{"a", "b", "c"} {
		if _, err := profile.Save(&profile.Profile{Name: n, CreatedAt: time.Now()}); err != nil {
			t.Fatalf("Save(%s): %v", n, err)
		}
	}

	// First arg: all profiles offered.
	completions, _ := twoProfileCompletion(nil, []string{}, "")
	if len(completions) != 3 {
		t.Errorf("first arg: expected 3, got %d", len(completions))
	}

	// Second arg: all profiles offered.
	completions, _ = twoProfileCompletion(nil, []string{"a"}, "")
	if len(completions) != 3 {
		t.Errorf("second arg: expected 3, got %d", len(completions))
	}

	// Third arg: nothing offered.
	completions, _ = twoProfileCompletion(nil, []string{"a", "b"}, "")
	if len(completions) != 0 {
		t.Errorf("third arg: expected 0, got %d", len(completions))
	}
}
