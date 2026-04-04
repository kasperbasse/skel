package profilemeta

import (
	appdoctor "github.com/kasperbasse/skel/internal/app/doctor"
	"github.com/kasperbasse/skel/internal/profile"
)

// Counts summarizes item totals for major profile categories.
type Counts struct {
	Brew    int
	Editors int
	Langs   int
	Configs int
	Total   int
}

// Readiness represents whether required tools are available for a profile.
type Readiness string

const (
	ReadinessReady        Readiness = "ready"
	ReadinessNeedsInstall Readiness = "needs-install"
	ReadinessMissing      Readiness = "missing"
)

// CountsForProfile returns categorized item counts for a profile.
func CountsForProfile(p *profile.Profile) Counts {
	brew := len(p.Homebrew.Formulas) + len(p.Homebrew.Casks) + len(p.Homebrew.Taps) + len(p.Homebrew.MasApps)

	editors := len(p.Editor.VSCodeExts) + len(p.Editor.CursorExts) + len(p.Editor.NeovimPlugins)
	for _, jb := range p.Editor.JetBrains {
		editors += len(jb.Plugins)
	}

	langs := len(p.Languages.NpmGlobals) + len(p.Languages.YarnGlobals) + len(p.Languages.PnpmGlobals) +
		len(p.Languages.PipGlobals) + len(p.Languages.ComposerGlobals) + len(p.Languages.GemGlobals) + len(p.Languages.CargoPackages)
	if p.Languages.NodeVersion != "" {
		langs++
	}
	if p.Languages.GoVersion != "" {
		langs++
	}
	if p.Languages.PythonVersion != "" {
		langs++
	}
	if p.Languages.RubyVersion != "" {
		langs++
	}
	if p.Languages.PHPVersion != "" {
		langs++
	}
	if p.Languages.RustVersion != "" {
		langs++
	}
	if p.Languages.JavaVersion != "" {
		langs++
	}

	configs := len(p.ConfigFiles)

	return Counts{
		Brew:    brew,
		Editors: editors,
		Langs:   langs,
		Configs: configs,
		Total:   brew + editors + langs + configs,
	}
}

// ReadinessForProfile reports whether all required tools are installed.
func ReadinessForProfile(p *profile.Profile) Readiness {
	required := appdoctor.RequiredTools(p)
	if len(required) == 0 {
		return ReadinessReady
	}

	missing := 0
	for _, cmd := range required {
		_, validator, _, ok := appdoctor.ToolDoctorInfo(cmd)
		if !ok {
			validator = cmd
		}
		if !appdoctor.CommandExists(validator) {
			missing++
		}
	}

	switch {
	case missing == 0:
		return ReadinessReady
	case missing == len(required):
		return ReadinessMissing
	default:
		return ReadinessNeedsInstall
	}
}
