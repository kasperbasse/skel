package doctor

import "testing"

func TestCommandExistsUnknown(t *testing.T) {
	if CommandExists("skel-this-command-should-not-exist-xyz") {
		t.Fatal("expected unknown command to be missing")
	}
}
