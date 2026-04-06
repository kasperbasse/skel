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
	RunE: runClone,
}

func runClone(_ *cobra.Command, args []string) error {
	printCommandHeader("clone", "Cloning profile...", randomMessage(cloneStartMsgs))

	p, err := loadProfileFromSource(args[0])
	if err != nil {
		return enhanceError(err)
	}

	proceed, err := confirmCloneWarnings(p, cloneForce)
	if err != nil {
		return enhanceError(err)
	}
	if !proceed {
		fmt.Printf("  %s Canceled.\n\n", iconDash())
		return nil
	}

	if _, err := profile.Save(p); err != nil {
		return enhanceError(fmt.Errorf("saving profile: %w", err))
	}

	printCloneSummary(p)
	printNextSteps(
		nextStep("skel show "+p.Name, "to review before restoring"),
		nextStep("skel restore "+p.Name, "to apply this setup"),
	)

	return nil
}

func loadProfileFromSource(source string) (*profile.Profile, error) {
	gistID, err := github.ParseSource(source)
	if err != nil {
		return nil, err
	}

	spin := newSpinner("Fetching gist...")
	spin.Start()

	gist, err := github.FetchGist(gistID)
	spin.Stop()
	if err != nil {
		return nil, fmt.Errorf("fetching gist: %w", err)
	}

	content, err := github.FindProfileJSON(gist, profile.MaxImportSize)
	if err != nil {
		return nil, fmt.Errorf("finding profile: %w", err)
	}

	var p profile.Profile
	if err := json.Unmarshal([]byte(content), &p); err != nil {
		return nil, fmt.Errorf("parsing profile: %w", err)
	}

	if p.Name == "" {
		return nil, fmt.Errorf("profile is missing a name - this might not be a skel gist")
	}

	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("this profile failed safety checks: %w", err)
	}

	return &p, nil
}

func confirmCloneWarnings(p *profile.Profile, force bool) (bool, error) {
	warnings := collectImportWarnings(p)
	if len(warnings) == 0 || force {
		return true, nil
	}

	warningText := "This profile contains shell/git configs that execute code when your shell starts or git runs.\n" +
		"These files will run as your user. Review them before proceeding.\n\n" +
		"Use " + cyan("--force") + " to skip this check."
	printWarningBox("Security Check Required", warningText)

	if !IsInteractive() {
		return false, fmt.Errorf("profile contains shell/git configs (%s) - use --force to accept", strings.Join(warnings, ", "))
	}

	fmt.Printf("  Continue? [y/N] ")
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer != "y" && answer != "yes" {
		fmt.Printf("  %s Clone canceled. Better safe than sorry!\n\n", iconDash())
		return false, nil
	}
	return true, nil
}

func printCloneSummary(p *profile.Profile) {
	fmt.Printf("\n  %s %s\n", iconCheck(), fmt.Sprintf(
		"Saved as '%s' (%s, %s)",
		bold(p.Name),
		countLabel(len(p.Homebrew.Formulas), "formula", "formulas"),
		countLabel(len(p.Homebrew.Casks), "cask", "casks"),
	))
	fmt.Printf("  %s\n", dividerStyle.Render(dividerLine))
	fmt.Printf("  %s\n\n", randomMessage(cloneCompleteMsgs))
}

func init() {
	cloneCmd.Flags().BoolVar(&cloneForce, "force", false, "Skip confirmation for profiles with shell/git configs")
}
