package cmd

import "github.com/kasperbasse/skel/internal/profile"

// collectImportWarnings gathers warnings for profile fields that execute user code.
func collectImportWarnings(p *profile.Profile) []string {
	warnings := make([]string, 0)
	for _, g := range scanGroups {
		if g.ImportWarnings == nil {
			continue
		}
		warnings = append(warnings, g.ImportWarnings(p)...)
	}
	return warnings
}
