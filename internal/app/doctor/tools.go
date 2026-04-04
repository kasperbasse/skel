package doctor

// toolMetadata centralizes command metadata used by doctor checks and error hints.
// command key is the canonical tool command (for example "brew").
type toolMetadata struct {
	DisplayName  string
	ValidatorCmd string // optional: defaults to the map key when empty
	DoctorFix    string // optional: overrides fallback for doctor output
	InstallHint  string
	DocsURL      string
}

var toolCatalog = map[string]toolMetadata{
	"brew": {
		DisplayName: "Homebrew",
		DoctorFix:   "https://brew.sh",
		InstallHint: "Install Homebrew first: https://brew.sh",
		DocsURL:     "https://brew.sh",
	},
	"mas": {
		DisplayName: "mas (App Store)",
		DoctorFix:   "brew install mas",
		InstallHint: "Install mas for Mac App Store: brew install mas",
	},
	"code": {
		DisplayName: "VS Code",
		DoctorFix:   "brew install --cask visual-studio-code",
		InstallHint: "Install VS Code or ensure it's in your PATH",
	},
	"cursor": {
		DisplayName: "Cursor",
		DoctorFix:   "brew install --cask cursor",
		InstallHint: "Install Cursor or ensure it's in your PATH",
	},
	"nvim": {
		DisplayName: "Neovim",
		DoctorFix:   "brew install neovim",
		InstallHint: "brew install neovim",
	},
	"git": {
		DisplayName: "Git",
		DoctorFix:   "brew install git",
		InstallHint: "brew install git",
	},
	"node": {
		DisplayName: "Node.js",
		DoctorFix:   "https://nodejs.org  or  brew install node",
		InstallHint: "https://nodejs.org  or  brew install node",
		DocsURL:     "https://nodejs.org",
	},
	"npm": {
		DisplayName: "npm",
		DoctorFix:   "included with Node.js",
		InstallHint: "included with Node.js",
	},
	"yarn": {
		DisplayName: "Yarn",
		DoctorFix:   "npm install -g yarn",
		InstallHint: "npm install -g yarn",
	},
	"pnpm": {
		DisplayName: "pnpm",
		DoctorFix:   "npm install -g pnpm",
		InstallHint: "npm install -g pnpm",
	},
	"pip3": {
		DisplayName: "pip3",
		DoctorFix:   "brew install python3",
		InstallHint: "brew install python3",
	},
	"gem": {
		DisplayName: "gem (Ruby)",
		DoctorFix:   "brew install ruby",
		InstallHint: "brew install ruby",
	},
	"cargo": {
		DisplayName: "cargo (Rust)",
		DoctorFix:   "https://rustup.rs",
		InstallHint: "https://rustup.rs",
		DocsURL:     "https://rustup.rs",
	},
	"composer": {
		DisplayName: "Composer",
		DoctorFix:   "brew install composer",
		InstallHint: "brew install composer",
	},
	"gh": {
		DisplayName: "GitHub CLI",
		InstallHint: "Install GitHub CLI: https://cli.github.com",
		DocsURL:     "https://cli.github.com",
	},
}

// ToolDoctorInfo returns doctor-facing metadata for a tool command.
func ToolDoctorInfo(command string) (label string, validatorCmd string, fix string, ok bool) {
	meta, ok := toolCatalog[command]
	if !ok {
		return "", "", "", false
	}
	if meta.DisplayName == "" {
		return "", "", "", false
	}

	validatorCmd = meta.ValidatorCmd
	if validatorCmd == "" {
		validatorCmd = command
	}

	fix = meta.DoctorFix
	if fix == "" {
		fix = meta.InstallHint
	}
	if fix == "" {
		fix = meta.DocsURL
	}

	if fix == "" {
		return "", "", "", false
	}

	return meta.DisplayName, validatorCmd, fix, true
}

// ToolNotFoundHint returns the user-facing install hint for a missing tool command.
func ToolNotFoundHint(command string) (hint string, ok bool) {
	meta, ok := toolCatalog[command]
	if !ok {
		return "", false
	}
	hint = meta.InstallHint
	if hint == "" {
		hint = meta.DocsURL
	}
	if hint == "" {
		return "", false
	}
	return hint, true
}
