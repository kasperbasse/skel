package doctor

import "github.com/kasperbasse/skel/internal/profile"

// RequiredTools returns the ordered list of tool commands needed to restore a profile.
func RequiredTools(p *profile.Profile) []string {
	if p == nil {
		return nil
	}

	tools := make([]string, 0, 14)
	seen := make(map[string]struct{}, 14)
	add := func(cmd string) {
		if _, ok := seen[cmd]; ok {
			return
		}
		seen[cmd] = struct{}{}
		tools = append(tools, cmd)
	}

	if len(p.Homebrew.Formulas) > 0 || len(p.Homebrew.Casks) > 0 || len(p.Homebrew.Taps) > 0 {
		add("brew")
	}
	if len(p.Homebrew.MasApps) > 0 {
		add("mas")
	}
	if p.Editor.VSCode {
		add("code")
	}
	if p.Editor.Cursor {
		add("cursor")
	}
	if p.Editor.Neovim {
		add("nvim")
	}
	if p.Git.UserName != "" || p.Git.GitConfigContent != "" {
		add("git")
	}
	if p.Languages.NodeVersion != "" || len(p.Languages.NpmGlobals) > 0 {
		add("node")
	}
	if len(p.Languages.NpmGlobals) > 0 {
		add("npm")
	}
	if len(p.Languages.YarnGlobals) > 0 {
		add("yarn")
	}
	if len(p.Languages.PnpmGlobals) > 0 {
		add("pnpm")
	}
	if len(p.Languages.PipGlobals) > 0 {
		add("pip3")
	}
	if len(p.Languages.GemGlobals) > 0 {
		add("gem")
	}
	if len(p.Languages.CargoPackages) > 0 {
		add("cargo")
	}
	if len(p.Languages.ComposerGlobals) > 0 {
		add("composer")
	}

	return tools
}
