package cmd

import (
	"fmt"
	"strings"

	"github.com/kasperbasse/skel/internal/profile"
)

// CommandUI provides reusable UI patterns for all commands.

// PrintCommandHeader prints a standard command header.
func PrintCommandHeader(commandName, subject string) {
	icon := headlineIcon(commandName)
	fmt.Printf("\n  %s %s\n", cyan(icon), subject)
	fmt.Printf("  %s\n", dividerStyle.Render(dividerLine))
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
		return false, readErr
	}

	return isAffirmative(answer), nil
}

func isAffirmative(answer string) bool {
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes"
}

// readLine reads a single line from stdin (helper for testing).
var readLine = func() (string, error) {
	var input string
	_, err := fmt.Scanln(&input)
	if err != nil {
		return "", err
	}
	return input, nil
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
