package tui

import "github.com/charmbracelet/lipgloss"

var (
	Green  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	Red    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	Yellow = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	Dim    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	Checkmark = Green.Render("✓")
	Cross     = Red.Render("✗")
	Warning   = Yellow.Render("⚠")

	StatusSkipped = Dim.Render("already installed")
)
