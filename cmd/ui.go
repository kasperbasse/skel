package cmd

import (
	"fmt"
	"os"

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
