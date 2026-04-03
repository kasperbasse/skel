package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/kasperbasse/skel/internal/profile"
	"github.com/kasperbasse/skel/internal/restore"
)

type restoreStepMsg struct {
	result restore.Result
}

type restoreDoneMsg struct{}

// RestoreModel is the Bubble Tea model for the restore command.
type RestoreModel struct {
	profile    *profile.Profile
	opts       *restore.Options
	spinner    spinner.Model
	steps      []restore.Result
	done       bool
	failed     int
	startMsg   string
	maxVisible int
	stepCh     chan restore.Result
}

func NewRestoreModel(p *profile.Profile, opts *restore.Options, startMsg string) RestoreModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))

	return RestoreModel{
		profile:    p,
		opts:       opts,
		spinner:    s,
		startMsg:   startMsg,
		maxVisible: 20,
		stepCh:     make(chan restore.Result),
	}
}

func (m RestoreModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.startRestore(), m.waitForStep())
}

func (m RestoreModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case restoreStepMsg:
		m.steps = append(m.steps, msg.result)
		if !msg.result.Success {
			m.failed++
		}
		// Keep only the last maxVisible steps in memory.
		if len(m.steps) > m.maxVisible*2 {
			m.steps = m.steps[len(m.steps)-m.maxVisible:]
		}
		return m, m.waitForStep()
	case restoreDoneMsg:
		m.done = true
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

var barFilled = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))

func progressBar(current, total int) string {
	const barWidth = 22
	if total <= 0 {
		return ""
	}
	pct := float64(current) / float64(total)
	filled := int(pct * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	bar := barFilled.Render(strings.Repeat("█", filled)) + Dim.Render(strings.Repeat("░", barWidth-filled))
	counter := Dim.Render(fmt.Sprintf("%d/%d", current, total))
	return fmt.Sprintf("  %s  %s", bar, counter)
}

func (m RestoreModel) View() string {
	var b strings.Builder

	b.WriteString("\n")

	start := 0
	if len(m.steps) > m.maxVisible {
		start = len(m.steps) - m.maxVisible
	}
	for _, step := range m.steps[start:] {
		progress := Dim.Render(fmt.Sprintf("[%d/%d]", step.Index, step.Total))
		if step.Success {
			if step.Message == "already installed" {
				b.WriteString(fmt.Sprintf("  %s %s %s  %s\n", progress, Checkmark, step.Step, StatusSkipped))
			} else {
				b.WriteString(fmt.Sprintf("  %s %s %s\n", progress, Checkmark, step.Step))
			}
		} else {
			b.WriteString(fmt.Sprintf("  %s %s %s  %s\n", progress, Cross, step.Step, Dim.Render(step.Message)))
		}
	}

	if !m.done {
		if len(m.steps) > 0 {
			last := m.steps[len(m.steps)-1]
			b.WriteString("\n" + progressBar(last.Index, last.Total) + "\n")
		}
		b.WriteString(fmt.Sprintf("\n  %s Working...\n", m.spinner.View()))
	} else {
		b.WriteString("\n")
		if m.failed == 0 {
			b.WriteString(fmt.Sprintf("  %s All done! Your Mac is feeling like home again.\n", Green.Render("🎉")))
		} else {
			b.WriteString(fmt.Sprintf("  %s Done with %s. Check the output above.\n",
				Warning, Red.Render(fmt.Sprintf("%d errors", m.failed))))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (m RestoreModel) startRestore() tea.Cmd {
	return func() tea.Msg {
		go func() {
			defer close(m.stepCh)
			restore.Run(m.profile, m.opts, func(r restore.Result) {
				m.stepCh <- r
			})
		}()
		return nil
	}
}

func (m RestoreModel) waitForStep() tea.Cmd {
	return func() tea.Msg {
		r, ok := <-m.stepCh
		if !ok {
			return restoreDoneMsg{}
		}
		return restoreStepMsg{result: r}
	}
}
