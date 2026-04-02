package scanner

import (
	"os/exec"
	"strings"

	"github.com/kasperbasse/skel/internal/profile"
)

// curatedDefault defines a macOS preference we capture.
type curatedDefault struct {
	Domain string
	Key    string
	Type   string // "string", "int", "float", "bool"
}

// curatedDefaults is the curated list of safe, developer-relevant macOS settings.
var curatedDefaults = []curatedDefault{
	// Dock
	{"com.apple.dock", "tilesize", "int"},
	{"com.apple.dock", "autohide", "bool"},
	{"com.apple.dock", "autohide-delay", "float"},
	{"com.apple.dock", "orientation", "string"},
	{"com.apple.dock", "show-recents", "bool"},
	{"com.apple.dock", "mineffect", "string"},
	{"com.apple.dock", "magnification", "bool"},

	// Keyboard
	{"NSGlobalDomain", "KeyRepeat", "int"},
	{"NSGlobalDomain", "InitialKeyRepeat", "int"},
	{"NSGlobalDomain", "ApplePressAndHoldEnabled", "bool"},

	// Trackpad
	{"com.apple.AppleMultitouchTrackpad", "Clicking", "bool"},
	{"NSGlobalDomain", "com.apple.trackpad.scaling", "float"},

	// Finder
	{"com.apple.finder", "AppleShowAllExtensions", "bool"},
	{"com.apple.finder", "ShowPathbar", "bool"},
	{"com.apple.finder", "ShowStatusBar", "bool"},
	{"com.apple.finder", "_FXSortFoldersFirst", "bool"},
	{"com.apple.finder", "FXDefaultSearchScope", "string"},
	{"com.apple.finder", "FXPreferredViewStyle", "string"},

	// Screenshots
	{"com.apple.screencapture", "location", "string"},
	{"com.apple.screencapture", "type", "string"},
	{"com.apple.screencapture", "disable-shadow", "bool"},

	// Misc developer-friendly
	{"NSGlobalDomain", "NSAutomaticSpellingCorrectionEnabled", "bool"},
	{"NSGlobalDomain", "NSAutomaticCapitalizationEnabled", "bool"},
	{"NSGlobalDomain", "NSAutomaticPeriodSubstitutionEnabled", "bool"},
	{"NSGlobalDomain", "NSAutomaticDashSubstitutionEnabled", "bool"},
	{"NSGlobalDomain", "NSAutomaticQuoteSubstitutionEnabled", "bool"},
}

func scanDefaults(warn func(string)) profile.DefaultsProfile {
	if !which("defaults") {
		return profile.DefaultsProfile{}
	}

	var settings []profile.DefaultsSetting
	for _, d := range curatedDefaults {
		val, ok := readDefault(d.Domain, d.Key)
		if !ok {
			continue // key not set - user uses system default
		}
		settings = append(settings, profile.DefaultsSetting{
			Domain: d.Domain,
			Key:    d.Key,
			Type:   d.Type,
			Value:  val,
		})
	}
	return profile.DefaultsProfile{Settings: settings}
}

// readDefault reads a single macOS preference. Returns the value and true if
// the key exists, or ("", false) if the key is unset or the read fails.
func readDefault(domain, key string) (string, bool) {
	out, err := exec.Command("defaults", "read", domain, key).Output()
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(string(out)), true
}
