package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kasperbasse/skel/internal/github"
	"github.com/kasperbasse/skel/internal/profile"
)

var publishNoRedact bool

var publishCmd = &cobra.Command{
	Use:   "publish <profile-name>",
	Short: "Publish a profile as a GitHub Gist",
	Long: `Publish a profile as a public GitHub Gist.

Requires authentication via GITHUB_TOKEN env var or the gh CLI (gh auth login).
PII (git identity, hostname, SSH key comments) is redacted before publishing.
Use --no-redact to skip redaction (not recommended for public gists).

Examples:
  skel publish my-setup
  GITHUB_TOKEN=ghp_xxx skel publish my-setup
  skel publish my-setup --no-redact`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := "default"
		if len(args) > 0 {
			name = args[0]
		}

		token, err := github.ResolveToken()
		if err != nil {
			return err
		}

		p, err := profile.Load(name)
		if err != nil {
			return enhanceError(err)
		}

		var pub *profile.Profile

		fmt.Printf("\n  %s Publishing profile %s\n", cyan("🚀"), bold("'"+p.Name+"'"))
		fmt.Printf("  %s\n", dividerStyle.Render("────────────────────────────────────────────"))
		fmt.Printf("  %s\n\n", dim(fmt.Sprintf("Saved %s from %s", p.CreatedAt.Format("Jan 02 2006"), p.Machine)))

		if publishNoRedact {
			fmt.Printf("  %s Publishing without redaction — git identity and hostname will be visible.\n\n", yellow("⚠"))
			tmp := *p
			pub = &tmp
		} else {
			pub = p.Redact()
			fmt.Printf("  %s Redacted before publishing: %s\n\n",
				dim("·"),
				dim("git identity · hostname · SSH key comments"),
			)
		}

		data, err := json.MarshalIndent(pub, "", "  ")
		if err != nil {
			return fmt.Errorf("encoding profile: %w", err)
		}

		filename := p.Name + "-skel.json"

		spin := NewSpinner("Publishing to GitHub Gist...")
		spin.Start()

		gist, err := github.CreateGist(token, &github.CreateGistRequest{
			Description: fmt.Sprintf("skel profile: %s", p.Name),
			Public:      true,
			Files: map[string]github.CreateGistFile{
				filename: {Content: string(data)},
			},
		})
		spin.Stop()

		if err != nil {
			return err
		}

		fmt.Printf("\n  %s %s\n", green("✓"), randomMessage(publishCompleteMsgs))
		fmt.Printf("  %s\n", dim(gist.HTMLURL))
		fmt.Printf("  %s\n\n", dim("Others can clone it with: skel clone "+gist.HTMLURL))
		printNextSteps(
			nextStep("skel clone "+gist.HTMLURL, "on another Mac"),
		)

		return nil
	},
}

func init() {
	publishCmd.Flags().BoolVar(&publishNoRedact, "no-redact", false, "Publish without redacting PII (not recommended)")
}
