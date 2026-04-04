package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/kasperbasse/skel/internal/profile"
)

var (
	selectedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	unselectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	markedStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
	hintStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	dividerStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("237"))
	styleCyan       = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
)

// ListAction indicates what the user chose to do.
type ListAction int

const (
	ListActionNone   ListAction = iota
	ListActionShow              // Enter on a profile
	ListActionDelete            // confirmed deletion
)

// ListModel is the Bubble Tea model for interactive profile listing.
type ListModel struct {
	profiles []*profile.Profile
	cursor   int
	marked   map[int]bool // marked for deletion
	deleting bool         // in delete-confirm mode
	action   ListAction
	chosen   string // profile name for Show
	deleted  []string
}

func NewListModel(profiles []*profile.Profile) ListModel {
	return ListModel{
		profiles: profiles,
		marked:   make(map[int]bool),
	}
}

func (m ListModel) Init() tea.Cmd { return nil }

func (m ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.deleting {
			switch msg.String() {
			case "y", "Y":
				for idx := range m.marked {
					name := m.profiles[idx].Name
					if err := profile.Delete(name); err == nil {
						m.deleted = append(m.deleted, name)
					}
				}
				m.action = ListActionDelete
				return m, tea.Quit
			case "n", "N", "escape":
				m.deleting = false
				return m, nil
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "escape":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.profiles)-1 {
				m.cursor++
			}
		case "enter":
			m.action = ListActionShow
			m.chosen = m.profiles[m.cursor].Name
			return m, tea.Quit
		case "x", " ":
			if m.marked[m.cursor] {
				delete(m.marked, m.cursor)
			} else {
				m.marked[m.cursor] = true
			}
		case "d", "D":
			if len(m.marked) > 0 {
				m.deleting = true
			}
		}
	}
	return m, nil
}

func (m ListModel) View() string {
	var b strings.Builder

	n := len(m.profiles)
	label := "profiles"
	if n == 1 {
		label = "profile"
	}

	b.WriteString(fmt.Sprintf("\n  %s\n", fmt.Sprintf("%s %s", styleCyan.Render(fmt.Sprintf("%d", n)), label)))
	b.WriteString(fmt.Sprintf("  %s\n", dividerStyle.Render("────────────────────────────────────────────")))

	for i, p := range m.profiles {
		cursor := "   "
		if i == m.cursor {
			cursor = Green.Render("▸") + " "
		}

		mark := Dim.Render("·")
		if m.marked[i] {
			mark = markedStyle.Render("✕")
		}

		name := unselectedStyle.Render(p.Name)
		if i == m.cursor {
			name = selectedStyle.Render(p.Name)
		}
		if m.marked[i] {
			name = markedStyle.Render(p.Name)
		}

		date := Dim.Render(relativeTime(p.CreatedAt))
		summary := fmt.Sprintf("%d formulas · %d casks", len(p.Homebrew.Formulas), len(p.Homebrew.Casks))

		b.WriteString(fmt.Sprintf("  %s%s %s  %s  %s\n", cursor, mark, name, date, Dim.Render(summary)))
	}

	b.WriteString("\n")

	if m.deleting {
		names := make([]string, 0, len(m.marked))
		for idx := range m.marked {
			names = append(names, m.profiles[idx].Name)
		}
		b.WriteString(fmt.Sprintf("  %s Delete %s? %s\n\n",
			Yellow.Render("⚠"),
			markedStyle.Render(strings.Join(names, ", ")),
			Dim.Render("[y/n]")))
	} else {
		hints := []string{
			hintStyle.Render("enter") + " show",
			hintStyle.Render("x") + " mark",
			hintStyle.Render("d") + " delete marked",
			hintStyle.Render("q") + " quit",
		}
		b.WriteString("  " + strings.Join(hints, Dim.Render("  ·  ")) + "\n\n")
	}

	return b.String()
}

func (m ListModel) Action() ListAction { return m.action }
func (m ListModel) Chosen() string     { return m.chosen }
func (m ListModel) Deleted() []string  { return m.deleted }
