package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kasperbasse/skel/internal/github"
	"github.com/kasperbasse/skel/internal/profile"
)

var cloneForce bool

var cloneCmd = &cobra.Command{
	Use:   "clone <source>",
	Short: "Clone a profile from a GitHub Gist",
	Long: `Clone a profile from a GitHub Gist URL or shorthand.

Examples:
  skel clone https://gist.github.com/user/abc123
  skel clone github:user/abc123
  skel clone github:user/abc123 --force`,
	Args: requireArgs("clone <source>  (URL or github:user/gist-id)"),
	RunE: func(cmd *cobra.Command, args []string) error {
		gistID, err := github.ParseSource(args[0])
		if err != nil {
			return err
		}

		spin := NewSpinner("Fetching gist...")
		spin.Start()

		gist, err := github.FetchGist(gistID)
		spin.Stop()
		if err != nil {
			return err
		}

		content, err := github.FindProfileJSON(gist, profile.MaxImportSize)
		if err != nil {
			return err
		}

		var p profile.Profile
		if err := json.Unmarshal([]byte(content), &p); err != nil {
			return fmt.Errorf("that doesn't look like an skel profile: %w", err)
		}

		if p.Name == "" {
			return fmt.Errorf("profile is missing a name - this might not be an skel gist")
		}

		if err := p.Validate(); err != nil {
			return fmt.Errorf("this profile failed safety checks: %w", err)
		}

		// Check for shell/git configs that execute as the user.
		var warnings []string
		for _, g := range scanGroups {
			if g.ImportWarnings == nil {
				continue
			}
			warnings = append(warnings, g.ImportWarnings(&p)...)
		}

		if len(warnings) > 0 && !cloneForce {
			fmt.Printf("\n  %s This profile contains configs that run as your user:\n", yellow("⚠"))
			for _, w := range warnings {
				fmt.Printf("     %s %s\n", yellow("·"), w)
			}
			fmt.Printf("\n  %s\n", dim("These files execute code when your shell starts or git runs."))
			fmt.Printf("  %s\n\n", dim("Use --force to skip this check, or review after cloning with 'skel show'."))

			if IsInteractive() {
				fmt.Printf("  Continue? [y/N] ")
				reader := bufio.NewReader(os.Stdin)
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(strings.ToLower(answer))
				if answer != "y" && answer != "yes" {
					fmt.Printf("\n  %s Clone canceled. Better safe than sorry!\n\n", dim("-"))
					return nil
				}
			} else {
				return fmt.Errorf("profile contains shell/git configs (%s) - use --force to accept", strings.Join(warnings, ", "))
			}
		}

		if _, err := profile.Save(&p); err != nil {
			return err
		}

		fmt.Printf("\n  %s %s\n\n", green("✓"), randomMessage(cloneCompleteMsgs))
		fmt.Printf("  %s\n", dim(fmt.Sprintf(
			"Saved as '%s' (%d formulas, %d casks)",
			p.Name, len(p.Homebrew.Formulas), len(p.Homebrew.Casks),
		)))
		fmt.Printf("  %s\n\n", dim("Run 'skel show "+p.Name+"' to review, then 'skel restore "+p.Name+"' to apply."))

		return nil
	},
}

func init() {
	cloneCmd.Flags().BoolVar(&cloneForce, "force", false, "Skip confirmation for profiles with shell/git configs")
}
