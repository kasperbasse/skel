package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kasperbasse/skel/internal/profile"
)

// CommandUI provides reusable UI patterns for all commands.

// PrintCommandHeader prints a standard command header with optional subtitle text.
func PrintCommandHeader(commandName, subject string, subtitle ...string) {
	icon := headlineIcon(commandName)
	fmt.Printf("\n  %s %s\n", cyan(icon), subject)
	fmt.Printf("  %s\n", dividerStyle.Render(dividerLine))
	if len(subtitle) > 0 && subtitle[0] != "" {
		fmt.Printf("  %s\n\n", dim(subtitle[0]))
	}
}

// ConfirmOverwrite prompts user to confirm overwriting an existing profile.
// Returns true if user wants to proceed, false to cancel.
func ConfirmOverwrite(name string) (bool, error) {
	// Check if profile already exists
	if !profile.Exists(name) {
		// Profile doesn't exist, no confirmation needed
		return true, nil
	}

	existing, err := profile.Load(name)
	if err != nil {
		return false, err
	}

	fmt.Printf("\n  %s Profile %s already exists (saved %s)\n\n",
		iconWarn(), bold("'"+name+"'"), existing.CreatedAt.Format(dateTimeFormat))
	fmt.Printf("  Overwrite? [y/N] ")

	answer, readErr := readLine()
	if readErr != nil {
		if errors.Is(readErr, io.EOF) {
			// Non-interactive / stdin closed — treat as default No.
			return false, nil
		}
		return false, readErr
	}

	return isAffirmative(answer), nil
}

func isAffirmative(answer string) bool {
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes"
}

// readLine reads a single line from stdin (helper for testing).
// Returns ("", nil) for a blank line (user just pressed Enter),
// and ("", io.EOF) when stdin is closed / non-interactive.
var readLine = func() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) {
			return "", io.EOF
		}
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

// PrintWarnings prints a list of warnings.
func PrintWarnings(warnings []string) {
	if len(warnings) == 0 {
		return
	}
	fmt.Println()
	for _, w := range warnings {
		fmt.Printf("  %s %s\n", iconWarn(), dim(w))
	}
}

// PrintError prints a formatted error message.
func PrintError(err error) {
	fmt.Printf("\n  %s %s\n\n", iconCross(), red(err.Error()))
}
