package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	appmeta "github.com/kasperbasse/skel/internal/app/profilemeta"
	"github.com/kasperbasse/skel/internal/profile"
	internalui "github.com/kasperbasse/skel/internal/ui"
)

func readinessBadge(p *profile.Profile) string {
	return internalui.ReadinessBadge(string(appmeta.ReadinessForProfile(p)))
}

func renderProfilesTable(profiles []*profile.Profile) string {
	rows := make([][]string, 0, len(profiles))
	for _, p := range profiles {
		rows = append(rows, []string{
			p.Name,
			readinessBadge(p),
			timeAgo(p.CreatedAt),
			p.Machine,
		})
	}

	return renderAlignedTable(
		[]string{"PROFILE", "STATUS", "MODIFIED", "MACHINE"},
		rows,
		map[int]int{0: 18, 3: 18},
		map[int]bool{2: true, 3: true},
	)
}

func printProfilesTable(profiles []*profile.Profile) {
	fmt.Printf("%s\n", renderProfilesTable(profiles))
}

func renderAlignedTable(headers []string, rows [][]string, maxWidths map[int]int, metaCols map[int]bool) string {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}

	for _, row := range rows {
		for i, cell := range row {
			if maxWidth, ok := maxWidths[i]; ok && lipgloss.Width(cell) > maxWidth {
				cell = truncateCell(cell, maxWidth)
				row[i] = cell
			}
			if w := lipgloss.Width(cell); w > widths[i] {
				widths[i] = w
			}
		}
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("250"))
	separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))

	var b strings.Builder
	b.WriteString("  ")
	for i, h := range headers {
		if i > 0 {
			b.WriteString("  ")
		}
		b.WriteString(headerStyle.Render(padCell(h, widths[i])))
	}
	b.WriteString("\n  ")
	for i, w := range widths {
		if i > 0 {
			b.WriteString("  ")
		}
		b.WriteString(separatorStyle.Render(strings.Repeat("─", w)))
	}

	for _, row := range rows {
		b.WriteString("\n  ")
		for i, cell := range row {
			if i > 0 {
				b.WriteString("  ")
			}
			styled := cell
			if metaCols[i] {
				styled = metaStyle.Render(cell)
			}
			b.WriteString(padStyledCell(styled, widths[i]))
		}
	}

	return b.String()
}

func padCell(value string, width int) string {
	padding := width - lipgloss.Width(value)
	if padding <= 0 {
		return value
	}
	return value + strings.Repeat(" ", padding)
}

func padStyledCell(value string, width int) string {
	padding := width - lipgloss.Width(value)
	if padding <= 0 {
		return value
	}
	return value + strings.Repeat(" ", padding)
}

func truncateCell(value string, maxWidth int) string {
	if lipgloss.Width(value) <= maxWidth {
		return value
	}
	if maxWidth <= 1 {
		return "…"
	}
	runes := []rune(value)
	trimmed := string(runes)
	for lipgloss.Width(trimmed) > maxWidth-1 && len(runes) > 0 {
		runes = runes[:len(runes)-1]
		trimmed = string(runes)
	}
	return trimmed + "…"
}
