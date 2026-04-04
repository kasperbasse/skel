package doctor

import "testing"

func TestToolNotFoundHint(t *testing.T) {
	hint, ok := ToolNotFoundHint("gh")
	if !ok {
		t.Fatal("expected gh to have a not-found hint")
	}
	if hint != "Install GitHub CLI: https://cli.github.com" {
		t.Fatalf("unexpected hint: %q", hint)
	}
}

func TestToolDoctorInfo(t *testing.T) {
	label, validatorCmd, fix, ok := ToolDoctorInfo("brew")
	if !ok {
		t.Fatal("expected brew to have doctor info")
	}
	if label != "Homebrew" {
		t.Fatalf("unexpected label: %q", label)
	}
	if validatorCmd != "brew" {
		t.Fatalf("unexpected validator command: %q", validatorCmd)
	}
	if fix != "https://brew.sh" {
		t.Fatalf("unexpected fix: %q", fix)
	}
}
