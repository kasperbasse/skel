package cmd

import (
	internalui "github.com/kasperbasse/skel/internal/ui"
)

// Date format constants used when displaying profile timestamps throughout the UI.
const (
	dateFormat     = "Jan 02 2006"
	dateTimeFormat = "Jan 02 2006 15:04"
)

// IsInteractive returns true when stdout is a terminal (not piped).
func IsInteractive() bool {
	return internalui.IsInteractive()
}

func bold(s string) string   { return internalui.Bold(s) }
func dim(s string) string    { return internalui.Dim(s) }
func green(s string) string  { return internalui.Green(s) }
func red(s string) string    { return internalui.Red(s) }
func yellow(s string) string { return internalui.Yellow(s) }
func cyan(s string) string   { return internalui.Cyan(s) }

func iconCheck() string { return internalui.IconCheck() }
func iconWarn() string  { return internalui.IconWarn() }
func iconCross() string { return internalui.IconCross() }
func iconDash() string  { return internalui.IconDash() }
func iconDot() string   { return internalui.IconDot() }

// printWarningBox prints a warning box with ⚠ icon
func printWarningBox(title, content string) {
	internalui.PrintWarningBox(title, content)
}

// printNextSteps prints suggested next commands
func printNextSteps(steps ...string) {
	internalui.PrintNextSteps(steps...)
}

// nextStep formats a next step with a styled command name
func nextStep(command string, description string) string {
	return internalui.NextStep(command, description)
}

// printFirstRun prints an onboarding prompt when no profiles exist yet.
func printFirstRun() {
	internalui.PrintFirstRun(dividerStyle.Render(dividerLine), "skel scan")
}

// Spinner shows an animated progress indicator (non-TUI fallback).
type Spinner = internalui.Spinner

func newSpinner(msg string) *Spinner { return internalui.NewSpinner(msg) }

// enhanceError wraps errors with helpful context and suggestions
func enhanceError(err error) error {
	if err == nil {
		return nil
	}

	if enhanced := applyErrorEnhancementRules(err.Error()); enhanced != nil {
		return enhanced
	}

	return err
}
