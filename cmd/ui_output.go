package cmd

import (
	"fmt"

	"github.com/kasperbasse/skel/internal/profile"
)

// CommandUI provides reusable UI patterns for all commands.

// PrintCommandHeader prints a standard command header.
func PrintCommandHeader(commandName, subject string) {
	icon := headlineIcon(commandName)
	fmt.Printf("\n  %s %s\n", cyan(icon), subject)
	fmt.Printf("  %s\n", dividerStyle.Render(dividerLine))
}

// PrintCommandComplete prints a success message with next steps.
func PrintCommandComplete(commandName, message string, steps ...string) {
	fmt.Printf("\n  %s %s\n", cyan(headlineIcon(commandName)), message)
	if len(steps) > 0 {
		fmt.Println()
		for _, step := range steps {
			fmt.Printf("  %s %s\n", dim("→"), step)
		}
	}
	fmt.Println()
}

// PrintProfileInfo prints profile metadata.
func PrintProfileInfo(p *profile.Profile) {
	fmt.Printf("  %s Profile: %s\n", cyan("→"), bold(p.Name))
	fmt.Printf("  %s Saved: %s from %s\n", cyan("→"), p.CreatedAt.Format(dateTimeFormat), p.Machine)
	fmt.Printf("  %s macOS: %s\n", cyan("→"), p.System.MacOSVersion)
}

// PrintItemCounts prints item counts for a profile.
func PrintItemCounts(label string, count int) {
	fmt.Printf("  %s %s: %s\n", cyan("→"), label, num(count))
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

	fmt.Printf("\n  %s Profile %s already exists (saved %s)\n",
		iconWarn(), bold("'"+name+"'"), existing.CreatedAt.Format(dateTimeFormat))
	fmt.Printf("  Overwrite? [y/N] ")

	answer, readErr := readUserConfirmation()
	if readErr != nil {
		return false, readErr
	}

	return answer == "yes", nil
}

// readUserConfirmation reads yes/no from user.
func readUserConfirmation() (string, error) {
	return readLine()
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

// PrintSection prints a section header and content.
func PrintSection(icon, title string) {
	fmt.Printf("\n  %s %s\n", icon, bold(title))
	fmt.Printf("  %s\n", dim("─────────────────────────────────────────"))
}

// PrintComparisonResult prints added/removed items in a comparison.
func PrintComparisonResult(icon, section string, added, removed []string) {
	fmt.Println()
	fmt.Printf("  %s %s\n", icon, section)

	if len(added) > 0 {
		fmt.Printf("    %s Added:\n", green("+"))
		for _, item := range added {
			fmt.Printf("      %s\n", green(item))
		}
	}

	if len(removed) > 0 {
		fmt.Printf("    %s Removed:\n", red("−"))
		for _, item := range removed {
			fmt.Printf("      %s\n", red(item))
		}
	}
}

// PrintError prints a formatted error message.
func PrintError(err error) {
	fmt.Printf("\n  %s %s\n\n", iconCross(), red(err.Error()))
}

// PrintSuccess prints a success message.
func PrintSuccess(message string) {
	fmt.Printf("\n  %s %s\n\n", iconCheck(), message)
}

// PrintInfo prints an informational message.
func PrintInfo(message string) {
	fmt.Printf("  %s %s\n", iconDot(), message)
}
