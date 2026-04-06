package cmd

import (
	"io"
	"strings"
	"testing"

	"github.com/kasperbasse/skel/internal/profile"
)

func TestIsAffirmative(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{name: "y", in: "y", want: true},
		{name: "yes", in: "yes", want: true},
		{name: "uppercase", in: "Y", want: true},
		{name: "spaced", in: "  yes  ", want: true},
		{name: "no", in: "n", want: false},
		{name: "empty", in: "", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isAffirmative(tc.in); got != tc.want {
				t.Fatalf("isAffirmative(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestPrintCommandHeaderWithoutSubtitle(t *testing.T) {
	out := captureStdout(func() {
		PrintCommandHeader("scan", "Scanning your Mac setup...")
	})
	if !strings.Contains(out, "Scanning your Mac setup...") {
		t.Fatalf("PrintCommandHeader() output missing subject: %q", out)
	}
}

func TestPrintCommandHeaderWithSubtitle(t *testing.T) {
	out := captureStdout(func() {
		PrintCommandHeader("scan", "Scanning your Mac setup...", "Fun subtitle")
	})
	if !strings.Contains(out, "Fun subtitle") {
		t.Fatalf("PrintCommandHeader() output missing subtitle: %q", out)
	}
}

func TestConfirmOverwriteEOFReturnsQuietCancel(t *testing.T) {
	profile.SetProfileDirOverride(t.TempDir())
	t.Cleanup(func() { profile.SetProfileDirOverride("") })
	if _, err := profile.Save(&profile.Profile{Name: "eoftest"}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	orig := readLine
	t.Cleanup(func() { readLine = orig })
	readLine = func() (string, error) { return "", io.EOF }

	ok, err := ConfirmOverwrite("eoftest")
	if err != nil {
		t.Fatalf("ConfirmOverwrite() returned error on EOF: %v", err)
	}
	if ok {
		t.Fatal("ConfirmOverwrite() returned true on EOF, want false (default No)")
	}
}
