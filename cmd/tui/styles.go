package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/kasperbasse/skel/internal/restore"
)

// relativeTime returns a human-readable relative duration for a past time.
func relativeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	case d < 2*24*time.Hour:
		return "yesterday"
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%d days ago", int(d.Hours()/24))
	case d < 14*24*time.Hour:
		return "1 week ago"
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%d weeks ago", int(d.Hours()/(24*7)))
	case d < 60*24*time.Hour:
		return "1 month ago"
	case d < 365*24*time.Hour:
		return fmt.Sprintf("%d months ago", int(d.Hours()/(24*30)))
	case d < 2*365*24*time.Hour:
		return "1 year ago"
	default:
		return fmt.Sprintf("%d years ago", int(d.Hours()/(24*365)))
	}
}

// dividerLine is the standard horizontal rule rendered throughout the TUI.
const dividerLine = "────────────────────────────────────────────"

var (
	Green  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	Red    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	Yellow = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	Dim    = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))

	Checkmark = Green.Render("✓")
	Cross     = Red.Render("✗")
	Warning   = Yellow.Render("⚠")

	StatusSkipped = Dim.Render(restore.MsgAlreadyInstalled)
)
