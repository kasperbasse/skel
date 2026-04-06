package cmd

import (
	"bytes"
	"os"

	"github.com/kasperbasse/skel/internal/profile"
)

// ---------------------------------------------------------------------------
// profile
// ---------------------------------------------------------------------------

type testProfile struct {
	p *profile.Profile
}

func newTestProfile() *testProfile {
	return &testProfile{
		p: &profile.Profile{
			Name:    "test-profile",
			Machine: "test-machine",
		},
	}
}

func (tp *testProfile) withHomebrew(formulas []string) *testProfile {
	tp.p.Homebrew.Formulas = formulas
	return tp
}

func (tp *testProfile) withGit(username, email string) *testProfile {
	tp.p.Git.UserName = username
	tp.p.Git.UserEmail = email
	return tp
}

func (tp *testProfile) build() *profile.Profile {
	return tp.p
}

// ---------------------------------------------------------------------------
// output
// ---------------------------------------------------------------------------

func captureOutput(f func()) string {
	var buf bytes.Buffer
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	done := make(chan struct{})
	go func() {
		_, _ = buf.ReadFrom(r)
		done <- struct{}{}
	}()
	f()
	_ = w.Close()
	os.Stdout = stdout
	<-done
	_ = r.Close()
	return buf.String()
}
