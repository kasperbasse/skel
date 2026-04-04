package cmd

import (
	"bytes"
	"errors"
	"testing"
)

func TestPrintCLIErrorAddsBlankLineAbove(t *testing.T) {
	var buf bytes.Buffer

	printCLIError(&buf, errors.New("profile 'default' not found\n\nUse 'skel list' to see available profiles"))

	want := "\nError: profile 'default' not found\n\nUse 'skel list' to see available profiles\n"
	if got := buf.String(); got != want {
		t.Fatalf("printCLIError() = %q, want %q", got, want)
	}
}

func TestPrintCLIErrorNil(t *testing.T) {
	var buf bytes.Buffer

	printCLIError(&buf, nil)

	if got := buf.String(); got != "" {
		t.Fatalf("printCLIError(nil) wrote %q, want empty string", got)
	}
}
