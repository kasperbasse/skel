package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
)

var (
	styleBold   = lipgloss.NewStyle().Bold(true)
	styleDim    = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	styleGreen  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	styleRed    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	styleYellow = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	styleCyan   = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
)

func IsInteractive() bool {
	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

func Bold(s string) string   { return styleBold.Render(s) }
func Dim(s string) string    { return styleDim.Render(s) }
func Green(s string) string  { return styleGreen.Render(s) }
func Red(s string) string    { return styleRed.Render(s) }
func Yellow(s string) string { return styleYellow.Render(s) }
func Cyan(s string) string   { return styleCyan.Render(s) }

func IconCheck() string { return Green("✓") }
func IconWarn() string  { return Yellow("⚠") }
func IconCross() string { return Red("✗") }
func IconDash() string  { return Dim("-") }
func IconDot() string   { return Dim("·") }

func ReadinessBadge(status string) string {
	switch status {
	case "ready":
		return Green("✓ READY")
	case "missing":
		return Red("✗ MISSING")
	case "needs-install":
		return Yellow("⚠ NEEDS INSTALL")
	default:
		return status
	}
}

func PrintWarningBox(title, content string) {
	fmt.Printf("  %s %s\n", Yellow("┌─"), Bold(title))
	for _, line := range strings.Split(content, "\n") {
		if line != "" {
			fmt.Printf("  %s  %s\n", Yellow("│"), line)
		}
	}
	fmt.Printf("  %s\n\n", Yellow("└─────────────────────────────────────────────────────"))
}

func PrintNextSteps(steps ...string) {
	if len(steps) == 0 {
		return
	}
	fmt.Printf("  %s Next steps:\n", Dim(Cyan("→")))
	for _, step := range steps {
		fmt.Printf("    %s\n", Dim(step))
	}
	fmt.Println()
}

func NextStep(command, description string) string {
	return fmt.Sprintf("%s %s", Cyan(command), Dim(description))
}

func PrintFirstRun(divider, scanCommand string) {
	fmt.Printf("\n  %s\n", Bold("💀 Welcome to skel!"))
	fmt.Printf("  %s\n\n", divider)
	fmt.Printf("  Capture your current Mac setup to get started:\n\n")
	fmt.Printf("    %s\n\n", Cyan(scanCommand))
	fmt.Printf("  %s\n", Dim("This saves your Homebrew packages, shell config, editors,"))
	fmt.Printf("  %s\n\n", Dim("git settings, and more into a portable profile."))
}

func PrintErr(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
}

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
				_, _ = fmt.Fprint(os.Stdout, "\r\033[2K") // clear entire line (no trailing spaces)
				_, _ = fmt.Fprint(os.Stdout, "\033[?25h") // restore cursor
				return
			default:
				_, _ = fmt.Fprintf(os.Stdout, "\r  %s %s", Cyan(s.frames[i%len(s.frames)]), s.msg)
				i++
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
}

func (s *Spinner) Stop() {
	s.done <- struct{}{}
	// The goroutine clears its line with ANSI; cursor stays at start of that line.
	// Callers print their next output there directly.
}
