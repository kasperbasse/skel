package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
)

// IsInteractive returns true when stdout is a terminal (not piped).
func IsInteractive() bool {
	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

// Lipgloss-based color helpers that work across terminals.
var (
	styleBold   = lipgloss.NewStyle().Bold(true)
	styleDim    = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	styleGreen  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	styleRed    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	styleYellow = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	styleCyan   = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
)

func bold(s string) string   { return styleBold.Render(s) }
func dim(s string) string    { return styleDim.Render(s) }
func green(s string) string  { return styleGreen.Render(s) }
func red(s string) string    { return styleRed.Render(s) }
func yellow(s string) string { return styleYellow.Render(s) }
func cyan(s string) string   { return styleCyan.Render(s) }

// addColor adds optional color if the value is non-empty
func colorize(value, noValue string, style lipgloss.Style) string {
	if value != "" {
		return style.Render(value)
	}
	return noValue
}

// printBox prints a styled info box with border, useful for warnings/tips
func printBox(title, content string) {
	fmt.Printf("  ┌─ %s\n", bold(title))
	for _, line := range strings.Split(content, "\n") {
		if line != "" {
			fmt.Printf("  │  %s\n", line)
		}
	}
	fmt.Printf("  └─────────────────────────────────────────────────────\n\n")
}

// printWarningBox prints a warning box with ⚠ icon
func printWarningBox(title, content string) {
	fmt.Printf("  %s %s\n", yellow("┌─"), bold(title))
	for _, line := range strings.Split(content, "\n") {
		if line != "" {
			fmt.Printf("  %s  %s\n", yellow("│"), line)
		}
	}
	fmt.Printf("  %s\n\n", yellow("└─────────────────────────────────────────────────────"))
}

// printTip prints a subtle tip/hint line
func printTip(tip string) {
	fmt.Printf("  %s %s\n", dim("💡"), dim(tip))
}

// printNextSteps prints suggested next commands
func printNextSteps(steps ...string) {
	if len(steps) == 0 {
		return
	}
	fmt.Printf("  %s Next steps:\n", dim(cyan("→")))
	for _, step := range steps {
		fmt.Printf("    %s\n", dim(step))
	}
	fmt.Println()
}

// nextStep formats a next step with a styled command name
func nextStep(command string, description string) string {
	return fmt.Sprintf("%s %s", cyan(command), dim(description))
}

// printFirstRun prints an onboarding prompt when no profiles exist yet.
func printFirstRun() {
	fmt.Printf("\n  %s\n", bold("💀 Welcome to skel!"))
	fmt.Printf("  %s\n\n", dividerStyle.Render("────────────────────────────────────────────"))
	fmt.Printf("  Capture your current Mac setup to get started:\n\n")
	fmt.Printf("    %s\n\n", cyan("skel scan"))
	fmt.Printf("  %s\n", dim("This saves your Homebrew packages, shell config, editors,"))
	fmt.Printf("  %s\n\n", dim("git settings, and more into a portable profile."))
}

// printErr writes a formatted message to stderr.
func printErr(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
}

// Spinner shows an animated progress indicator (non-TUI fallback).
type Spinner struct {
	msg    string
	frames []string
	done   chan struct{}
}

func NewSpinner(msg string) *Spinner {
	return &Spinner{
		msg:    msg,
		frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		done:   make(chan struct{}),
	}
}

func (s *Spinner) Start() {
	_, _ = fmt.Fprint(os.Stdout, "\033[?25l") // hide cursor
	go func() {
		i := 0
		for {
			select {
			case <-s.done:
				_, _ = fmt.Fprintf(os.Stdout, "\r%-60s\r", "") // clear line
				_, _ = fmt.Fprint(os.Stdout, "\033[?25h")      // restore cursor
				return
			default:
				_, _ = fmt.Fprintf(os.Stdout, "\r  %s %s", cyan(s.frames[i%len(s.frames)]), s.msg)
				i++
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
}

func (s *Spinner) Stop() {
	s.done <- struct{}{}
	fmt.Println() // add space after spinner
}

// requireArgs returns a Cobra args validator that shows a friendly error with usage hint.
func requireArgs(usage string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("missing required argument\n\nUsage: skel %s", usage)
		}
		return nil
	}
}

// requireExactArgs returns a Cobra args validator for exactly n args with a friendly error.
func requireExactArgs(n int, usage string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < n {
			return fmt.Errorf("missing required argument(s)\n\nUsage: skel %s", usage)
		}
		if len(args) > n {
			return fmt.Errorf("too many arguments\n\nUsage: skel %s", usage)
		}
		return nil
	}
}
