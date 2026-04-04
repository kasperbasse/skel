package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/kasperbasse/skel/internal/profile"
)

type check struct {
	label string
	ok    bool
	fix   string // install hint shown when not ok
}

var doctorCmd = &cobra.Command{
	Use:   "doctor [profile-name]",
	Short: "Check that a profile can be restored on this machine",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := "default"
		if len(args) > 0 {
			name = args[0]
		}

		p, err := profile.Load(name)
		if err != nil {
			return err
		}

		fmt.Printf("\n  %s Checking %s\n", cyan("🩺"), bold("'"+p.Name+"'"))
		fmt.Printf("  %s\n\n", dividerStyle.Render("────────────────────────────────────────────"))

		checks := buildChecks(p)
		if len(checks) == 0 {
			fmt.Printf("  %s Profile has no restorable sections.\n\n", dim("-"))
			return nil
		}

		issues := 0
		for _, c := range checks {
			printCheck(c)
			if !c.ok {
				issues++
			}
		}

		fmt.Println()
		if issues == 0 {
			fmt.Printf("  %s All tools present. Ready to restore.\n\n", green("✓"))
		} else {
			fmt.Printf("  %s %s - install missing tools then run %s\n\n",
				yellow("⚠"),
				bold(fmt.Sprintf("%d issue%s found", issues, pluralS(issues))),
				cyan("skel restore "+name),
			)
		}
		return nil
	},
}

func buildChecks(p *profile.Profile) []check {
	var checks []check

	need := func(label, cmd, fix string) {
		checks = append(checks, check{label: label, ok: toolExists(cmd), fix: fix})
	}

	if len(p.Homebrew.Formulas) > 0 || len(p.Homebrew.Casks) > 0 || len(p.Homebrew.Taps) > 0 {
		need("Homebrew", "brew", "https://brew.sh")
	}
	if len(p.Homebrew.MasApps) > 0 {
		need("mas (App Store)", "mas", "brew install mas")
	}
	if p.Editor.VSCode {
		need("VS Code", "code", "brew install --cask visual-studio-code")
	}
	if p.Editor.Cursor {
		need("Cursor", "cursor", "brew install --cask cursor")
	}
	if p.Editor.Neovim {
		need("Neovim", "nvim", "brew install neovim")
	}
	if p.Git.UserName != "" || p.Git.GitConfigContent != "" {
		need("Git", "git", "brew install git")
	}
	if p.Languages.NodeVersion != "" || len(p.Languages.NpmGlobals) > 0 {
		need("Node.js", "node", "https://nodejs.org  or  brew install node")
	}
	if len(p.Languages.NpmGlobals) > 0 {
		need("npm", "npm", "included with Node.js")
	}
	if len(p.Languages.YarnGlobals) > 0 {
		need("Yarn", "yarn", "npm install -g yarn")
	}
	if len(p.Languages.PnpmGlobals) > 0 {
		need("pnpm", "pnpm", "npm install -g pnpm")
	}
	if len(p.Languages.PipGlobals) > 0 {
		need("pip3", "pip3", "brew install python3")
	}
	if len(p.Languages.GemGlobals) > 0 {
		need("gem (Ruby)", "gem", "brew install ruby")
	}
	if len(p.Languages.CargoPackages) > 0 {
		need("cargo (Rust)", "cargo", "https://rustup.rs")
	}
	if len(p.Languages.ComposerGlobals) > 0 {
		need("Composer", "composer", "brew install composer")
	}

	return checks
}

func printCheck(c check) {
	if c.ok {
		fmt.Printf("  %s  %s\n", green("✓"), c.label)
	} else {
		fmt.Printf("  %s  %s\n", red("✗"), bold(c.label))
		fmt.Printf("       %s  %s\n", dim("→"), dim(c.fix))
	}
}

func toolExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func pluralS(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func init() {
	doctorCmd.ValidArgsFunction = singleProfileCompletion
}
