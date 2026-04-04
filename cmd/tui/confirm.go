package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type confirmModel struct {
	message   string
	confirmed bool
	done      bool
}

func (m confirmModel) Init() tea.Cmd {
	return nil
}

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			m.confirmed = true
			m.done = true
			return m, tea.Quit
		case "n", "N", "q", "ctrl+c", "esc":
			m.confirmed = false
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m confirmModel) View() string {
	if m.done {
		return ""
	}
	return fmt.Sprintf("  %s [y/N]: \n\n", m.message)
}

func Confirm(message string) (bool, error) {
	m := confirmModel{message: message}
	p := tea.NewProgram(m)

	result, err := p.Run()
	if err != nil {
		return false, err
	}

	finalModel := result.(confirmModel)
	return finalModel.confirmed, nil
}
