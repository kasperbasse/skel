package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/kasperbasse/skel/internal/profile"
	"github.com/kasperbasse/skel/internal/scanner"
)

// ScanResult is returned when the scan completes.
type ScanResult struct {
	Profile  *profile.Profile
	Warnings []string
	Err      error
}

// scanDoneMsg is sent when the background scan completes.
type scanDoneMsg struct {
	profile  *profile.Profile
	warnings []string
	err      error
}

// ScanModel is the Bubble Tea model for the scan command.
type ScanModel struct {
	name     string
	spinner  spinner.Model
	done     bool
	result   *ScanResult
	startMsg string
}

func NewScanModel(name, startMsg string) ScanModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))

	return ScanModel{
		name:     name,
		spinner:  s,
		startMsg: startMsg,
	}
}

func (m ScanModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.runScan())
}

func (m ScanModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case scanDoneMsg:
		m.done = true
		m.result = &ScanResult{
			Profile:  msg.profile,
			Warnings: msg.warnings,
			Err:      msg.err,
		}
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m ScanModel) View() string {
	var b strings.Builder

	b.WriteString("\n")

	if !m.done {
		b.WriteString(fmt.Sprintf("  %s %s\n", m.spinner.View(), m.startMsg))
		return b.String()
	}

	return b.String()
}

func (m ScanModel) Result() *ScanResult {
	return m.result
}

func (m ScanModel) runScan() tea.Cmd {
	return func() tea.Msg {
		p, warnings, err := scanner.Run(m.name)
		return scanDoneMsg{profile: p, warnings: warnings, err: err}
	}
}
