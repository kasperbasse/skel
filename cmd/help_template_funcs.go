package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	cobra.AddTemplateFunc("flagUsagesSpaced", flagUsagesSpaced)
}

func flagUsagesSpaced(usages string) string {
	usages = strings.TrimRight(usages, "\n")
	if usages == "" {
		return usages
	}

	lines := strings.Split(usages, "\n")
	var b strings.Builder
	for i, line := range lines {
		if i > 0 {
			if isFlagLine(lines[i-1]) && isFlagLine(line) {
				b.WriteString("\n")
			}
			b.WriteString("\n")
		}
		b.WriteString(line)
	}
	return b.String()
}

func isFlagLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "-")
}
