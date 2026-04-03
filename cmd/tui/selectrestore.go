package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// SelectItem represents a restorable section in the checklist.
type SelectItem struct {
	Icon     string
	Label    string
	Keys     []string // RestoreKeys from the ScanGroup
	Summary  string   // short description of what's in this section
	Selected bool
}

// SelectRestoreModel is a Bubble Tea model for choosing which sections to restore.
type SelectRestoreModel struct {
	items     []SelectItem
	cursor    int
	confirmed bool
	canceled  bool
}

// NewSelectRestoreModel creates a new checklist with all items selected by default.
func NewSelectRestoreModel(items []SelectItem) SelectRestoreModel {
	return SelectRestoreModel{items: items}
}

func (m SelectRestoreModel) Init() tea.Cmd { return nil }

func (m SelectRestoreModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "escape":
			m.canceled = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "x", " ":
			m.items[m.cursor].Selected = !m.items[m.cursor].Selected
		case "a":
			allOn := true
			for _, it := range m.items {
				if !it.Selected {
					allOn = false
					break
				}
			}
			for i := range m.items {
				m.items[i].Selected = !allOn
			}
		case "enter":
			m.confirmed = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m SelectRestoreModel) View() string {
	var b strings.Builder

	b.WriteString("\n  Select sections to restore\n\n")

	for i, it := range m.items {
		cursor := "   "
		if i == m.cursor {
			cursor = Green.Render("▸") + " "
		}

		check := Dim.Render("○")
		if it.Selected {
			check = Green.Render("◉")
		}

		label := it.Icon + " " + it.Label
		if i == m.cursor {
			label = it.Icon + " " + selectedStyle.Render(it.Label)
		}

		summary := ""
		if it.Summary != "" {
			summary = "  " + Dim.Render(it.Summary)
		}

		b.WriteString(fmt.Sprintf("  %s%s  %s%s\n", cursor, check, label, summary))
	}

	b.WriteString("\n")
	hints := []string{
		hintStyle.Render("enter") + " restore",
		hintStyle.Render("x") + " toggle",
		hintStyle.Render("a") + " all",
		hintStyle.Render("q") + " cancel",
	}
	b.WriteString("  " + strings.Join(hints, Dim.Render("  ·  ")) + "\n\n")

	return b.String()
}

// Confirmed returns true if the user pressed enter to proceed.
func (m SelectRestoreModel) Confirmed() bool { return m.confirmed }

// SelectedKeys returns a map of restore keys from all selected items.
func (m SelectRestoreModel) SelectedKeys() map[string]bool {
	keys := make(map[string]bool)
	for _, it := range m.items {
		if it.Selected {
			for _, k := range it.Keys {
				keys[k] = true
			}
		}
	}
	return keys
}
