package doctor

import "github.com/kasperbasse/skel/internal/profile"

type sectionToolRule struct {
	Section string
	Tool    string
	Needed  func(p *profile.Profile) bool
}

type SectionFilter func(section string) bool

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
	{
		Section: "editors",
		Tool:    "nvim",
		Needed: func(p *profile.Profile) bool {
			return p.Editor.Neovim
		},
	},
	{
		Section: "git",
		Tool:    "git",
		Needed: func(p *profile.Profile) bool {
			return p.Git.UserName != "" || p.Git.GitConfigContent != ""
		},
	},
	{
		Section: "languages",
		Tool:    "node",
		Needed: func(p *profile.Profile) bool {
			return p.Languages.NodeVersion != "" || len(p.Languages.NpmGlobals) > 0
		},
	},
	{
		Section: "languages",
		Tool:    "npm",
		Needed: func(p *profile.Profile) bool {
			return len(p.Languages.NpmGlobals) > 0
		},
	},
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
	{
		Section: "languages",
		Tool:    "pip3",
		Needed: func(p *profile.Profile) bool {
			return len(p.Languages.PipGlobals) > 0
		},
	},
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

// requiredToolsFor is a validator which will be using rules to determine which tools are required for a given profile, based on the sections that should be included.
func requiredToolsFor(p *profile.Profile, shouldInclude SectionFilter) []string {
	if p == nil {
		return nil
	}
	seen := map[string]struct{}{}
	var tools []string
	add := func(t string) {
		if _, ok := seen[t]; ok {
			return
		}
		seen[t] = struct{}{}
		tools = append(tools, t)
	}
	for _, r := range rules {
		if !shouldInclude(r.Section) {
			continue
		}
		if r.Needed != nil && r.Needed(p) {
			add(r.Tool)
		}
	}
	return tools
}

// RequiredTools returns a list of tools that are required for the given profile, based on all sections.
func RequiredTools(p *profile.Profile) []string {
	return requiredToolsFor(p, includeAllSections)
}

// RequiredToolsForSections returns a list of tools that are required for the given profile,
// based on the sections that should be included as determined by the provided SectionFilter.
// If shouldInclude is nil, all sections will be included.
func RequiredToolsForSections(p *profile.Profile, shouldInclude SectionFilter) []string {
	if shouldInclude == nil {
		shouldInclude = includeAllSections
	}
	return requiredToolsFor(p, shouldInclude)
}

// BlockedSectionTools returns missing tool commands grouped by restore section key.
func BlockedSectionTools(p *profile.Profile) map[string][]string {
	if p == nil {
		return map[string][]string{}
	}

	missingBySection := make(map[string][]string)
	seen := make(map[string]map[string]struct{})

	for _, r := range rules {
		if r.Needed == nil || !r.Needed(p) || CommandExists(r.Tool) {
			continue
		}
		if _, ok := seen[r.Section]; !ok {
			seen[r.Section] = make(map[string]struct{})
		}
		if _, ok := seen[r.Section][r.Tool]; ok {
			continue
		}
		seen[r.Section][r.Tool] = struct{}{}
		missingBySection[r.Section] = append(missingBySection[r.Section], r.Tool)
	}

	return missingBySection
}
