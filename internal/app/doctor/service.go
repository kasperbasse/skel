package doctor

import "github.com/kasperbasse/skel/internal/profile"

// Package doctor validates that required tools are installed to restore a profile.
// It uses a rules-based system: data-driven definitions of which tools each section needs.
// Adding a new tool requires only a new rule; no code changes to restore logic.

// sectionToolRule maps a restore section to the tools it needs.
type sectionToolRule struct {
	Section string
	Tool    string
	Needed  func(p *profile.Profile) bool
}

type SectionFilter func(section string) bool

// rules defines which tools are required for each section.
// Each rule says: if a profile has data in this section AND this condition passes,
// then this tool is required. Sections may appear in multiple rules (e.g., homebrew needs brew + mas).
var rules = []sectionToolRule{
	{
		Section: "homebrew",
		Tool:    "brew",
		Needed: func(p *profile.Profile) bool {
			return len(p.Homebrew.Formulas) > 0 || len(p.Homebrew.Casks) > 0 || len(p.Homebrew.Taps) > 0
		},
	},
	{
		Section: "mas",
		Tool:    "mas",
		Needed: func(p *profile.Profile) bool {
			return len(p.Homebrew.MasApps) > 0
		},
	},
	{
		Section: "editors",
		Tool:    "code",
		Needed: func(p *profile.Profile) bool {
			return p.Editor.VSCode
		},
	},
	{
		Section: "editors",
		Tool:    "cursor",
		Needed: func(p *profile.Profile) bool {
			return p.Editor.Cursor
		},
	},
	// nvim is intentionally omitted: the restore implementation does not invoke nvim
	// and only writes config files, so its absence should not block the editors section.
	// git is intentionally omitted: restoreGitConfigs only writes files and never
	// invokes git, so its absence should not block restoring Git config.
	{
		Section: "languages",
		Tool:    "npm",
		Needed: func(p *profile.Profile) bool {
			return len(p.Languages.NpmGlobals) > 0
		},
	},
	// node is intentionally omitted: restore does not manage node versions; npm
	// existence is already checked by the rule above and at runtime in restoreLanguageTools.
	{
		Section: "languages",
		Tool:    "yarn",
		Needed: func(p *profile.Profile) bool {
			return len(p.Languages.YarnGlobals) > 0
		},
	},
	{
		Section: "languages",
		Tool:    "pnpm",
		Needed: func(p *profile.Profile) bool {
			return len(p.Languages.PnpmGlobals) > 0
		},
	},
	// pip3 is intentionally omitted: restore does not install pip packages,
	// so its absence should not block the languages section.
	{
		Section: "languages",
		Tool:    "gem",
		Needed: func(p *profile.Profile) bool {
			return len(p.Languages.GemGlobals) > 0
		},
	},
	{
		Section: "languages",
		Tool:    "cargo",
		Needed: func(p *profile.Profile) bool {
			return len(p.Languages.CargoPackages) > 0
		},
	},
	{
		Section: "languages",
		Tool:    "composer",
		Needed: func(p *profile.Profile) bool {
			return len(p.Languages.ComposerGlobals) > 0
		},
	},
}

// includeAllSections is a SectionFilter that includes all sections.
func includeAllSections(string) bool { return true }

// requiredToolsFor collects all tools required for a profile based on which sections to include.
func requiredToolsFor(p *profile.Profile, shouldInclude SectionFilter) []string {
	if p == nil {
		return nil
	}

	seen := make(map[string]struct{})
	var tools []string

	for _, r := range rules {
		if !shouldInclude(r.Section) {
			continue
		}
		if r.Needed != nil && r.Needed(p) {
			if _, alreadyAdded := seen[r.Tool]; !alreadyAdded {
				seen[r.Tool] = struct{}{}
				tools = append(tools, r.Tool)
			}
		}
	}

	return tools
}

// RequiredTools returns tools required to restore a profile (all sections).
func RequiredTools(p *profile.Profile) []string {
	return requiredToolsFor(p, includeAllSections)
}

// RequiredToolsForSections returns tools required based on a section filter.
// The callback determines which sections' tool requirements to include.
func RequiredToolsForSections(p *profile.Profile, shouldInclude SectionFilter) []string {
	if shouldInclude == nil {
		shouldInclude = includeAllSections
	}
	return requiredToolsFor(p, shouldInclude)
}

// BlockedSectionTools returns tools that are missing, grouped by section.
// Used by the UI layer to show which sections can't be restored due to missing tools.
func BlockedSectionTools(p *profile.Profile) map[string][]string {
	if p == nil {
		return map[string][]string{}
	}

	missingBySection := make(map[string][]string)
	seen := make(map[string]map[string]struct{})

	for _, r := range rules {
		if r.Needed == nil || !r.Needed(p) || CommandExists(r.Tool) {
			continue // tool exists or isn't needed
		}

		// Initialize deduplication set for this section
		if _, ok := seen[r.Section]; !ok {
			seen[r.Section] = make(map[string]struct{})
		}

		// Skip if we've already added this tool for this section
		if _, ok := seen[r.Section][r.Tool]; ok {
			continue
		}

		// Record this tool as missing for this section
		seen[r.Section][r.Tool] = struct{}{}
		missingBySection[r.Section] = append(missingBySection[r.Section], r.Tool)
	}

	return missingBySection
}
