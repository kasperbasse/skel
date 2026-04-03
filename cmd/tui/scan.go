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

type scanDoneMsg struct {
	profile  *profile.Profile
	warnings []string
	err      error
}

type scanProgressMsg struct {
	section string
}

// ScanModel is the Bubble Tea model for the scan command.
type ScanModel struct {
	name       string
	spinner    spinner.Model
	done       bool
	result     *ScanResult
	startMsg   string
	section    string
	progressCh chan string
	doneCh     chan scanDoneMsg
}

func NewScanModel(name, startMsg string) ScanModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))

	return ScanModel{
		name:       name,
		spinner:    s,
		startMsg:   startMsg,
		progressCh: make(chan string, 12),
		doneCh:     make(chan scanDoneMsg, 1),
	}
}

func (m ScanModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.startScan(), m.waitForProgress())
}

func (m ScanModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case scanProgressMsg:
		m.section = msg.section
		return m, tea.Batch(m.spinner.Tick, m.waitForProgress())
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
	if m.done {
		return ""
	}
	label := "scanning..."
	if m.section != "" {
		label = "scanning " + strings.ToLower(m.section) + "..."
	}
	return fmt.Sprintf("  %s  %s\n", m.spinner.View(), Dim.Render(label))
}

func (m ScanModel) Result() *ScanResult {
	return m.result
}

func (m ScanModel) startScan() tea.Cmd {
	return func() tea.Msg {
		go func() {
			p, warnings, err := scanner.RunWithProgress(m.name, func(section string) {
				m.progressCh <- section
			})
			close(m.progressCh)
			m.doneCh <- scanDoneMsg{profile: p, warnings: warnings, err: err}
		}()
		return nil
	}
}

func (m ScanModel) waitForProgress() tea.Cmd {
	return func() tea.Msg {
		section, ok := <-m.progressCh
		if !ok {
			// Channel closed - scanner finished, collect the result.
			return <-m.doneCh
		}
		return scanProgressMsg{section: section}
	}
}
