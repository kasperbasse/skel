package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	appmeta "github.com/kasperbasse/skel/internal/app/profilemeta"
	"github.com/kasperbasse/skel/internal/profile"
	internalui "github.com/kasperbasse/skel/internal/ui"
)

var (
	selectedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	unselectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	markedIconStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	markedNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
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
		case "q", "esc", "ctrl+c":
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
	_, _ = fmt.Fprintf(&b, "\n  %s\n", fmt.Sprintf("Profiles (%s)", styleCyan.Render(fmt.Sprintf("%d", n))))
	_, _ = fmt.Fprintf(&b, "  %s\n", dividerStyle.Render(dividerLine))
	b.WriteString("  " + Dim.Render("Overview: profile · status · modified · machine") + "\n\n")
	b.WriteString(m.renderHeader())

	for i, p := range m.profiles {
		cursor := "  "
		if i == m.cursor {
			cursor = Green.Render("▸ ")
		}

		mark := Dim.Render("·")
		if m.marked[i] {
			mark = markedIconStyle.Render("✓")
		}

		b.WriteString(m.renderRow(i, cursor, mark, p))
	}

	b.WriteString("\n")

	if m.deleting {
		names := make([]string, 0, len(m.marked))
		for idx := range m.marked {
			names = append(names, m.profiles[idx].Name)
		}
		_, _ = fmt.Fprintf(&b, "  %s Delete %s? %s\n\n",
			Yellow.Render("⚠"),
			markedNameStyle.Render(strings.Join(names, ", ")),
			Dim.Render("[y/n]"))
	} else {
		hints := []string{
			hintStyle.Render("enter") + " show",
			hintStyle.Render("x") + " mark for delete",
			hintStyle.Render("d") + " delete marked",
			hintStyle.Render("q") + " quit",
		}
		b.WriteString("  " + strings.Join(hints, Dim.Render("  ·  ")) + "\n\n")
	}

	return b.String()
}

func (m ListModel) renderHeader() string {
	headers := []string{"", "", "PROFILE", "STATUS", "MODIFIED", "MACHINE"}
	widths := []int{3, 4, 18, 14, 14, 18}

	var parts []string
	for i, h := range headers {
		parts = append(parts, hintStyle.Bold(true).Render(padCell(h, widths[i], false)))
	}

	var dividers []string
	for _, w := range widths {
		dividers = append(dividers, dividerStyle.Render(strings.Repeat("─", w)))
	}

	return "  " + strings.Join(parts, "  ") + "\n  " + strings.Join(dividers, "  ") + "\n"
}

// renderRow prints one aligned list row using fixed-width plain text fields,
// then applies styles to avoid ANSI width drift.
func (m ListModel) renderRow(i int, cursor, mark string, p *profile.Profile) string {
	status := readinessBadge(p)
	displayName := truncateText(p.Name, 18)
	paddedName := padCell(displayName, 18, false)
	name := unselectedStyle.Render(paddedName)
	if i == m.cursor {
		name = selectedStyle.Render(paddedName)
	}
	if m.marked[i] {
		name = markedNameStyle.Render(paddedName)
	}

	modified := Dim.Render(padCell(relativeTime(p.CreatedAt), 14, false))
	machine := Dim.Render(padCell(truncateText(p.Machine, 18), 18, false))

	cols := []string{
		padCell(cursor, 3, false),
		padCell(mark, 4, false),
		name,
		padCell(status, 14, false),
		modified,
		machine,
	}

	return "  " + strings.Join(cols, "  ") + "\n"
}

func readinessBadge(p *profile.Profile) string {
	return internalui.ReadinessBadge(string(appmeta.ReadinessForProfile(p)))
}

func padCell(value string, width int, rightAlign bool) string {
	padding := width - lipgloss.Width(value)
	if padding <= 0 {
		return value
	}
	spaces := strings.Repeat(" ", padding)
	if rightAlign {
		return spaces + value
	}
	return value + spaces
}

func truncateText(value string, width int) string {
	if lipgloss.Width(value) <= width {
		return value
	}
	runes := []rune(value)
	for len(runes) > 0 && lipgloss.Width(string(runes)) > width-1 {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "…"
}

func (m ListModel) Action() ListAction { return m.action }
func (m ListModel) Chosen() string     { return m.chosen }
func (m ListModel) Deleted() []string  { return m.deleted }
