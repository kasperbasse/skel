package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kasperbasse/skel/internal/github"
	"github.com/kasperbasse/skel/internal/profile"
)

var publishCmd = &cobra.Command{
	Use:   "publish <profile-name>",
	Short: "Publish a profile as a GitHub Gist",
	Long: `Publish a profile as a public GitHub Gist.

Requires authentication via GITHUB_TOKEN env var or the gh CLI (gh auth login).
Your hostname is automatically redacted before publishing.

Examples:
  skel publish my-setup
  GITHUB_TOKEN=ghp_xxx skel publish my-setup`,
	Args: requireArgs("publish <profile-name>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := github.ResolveToken()
		if err != nil {
			return err
		}

		p, err := profile.Load(args[0])
		if err != nil {
			return err
		}

		// Redact machine-specific info before publishing.
		pub := *p
		pub.Machine = "shared"

		data, err := json.MarshalIndent(&pub, "", "  ")
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
		fmt.Printf("  %s\n\n", dim(gist.HTMLURL))
		fmt.Printf("  %s\n", dim("Others can clone it with: skel clone "+gist.HTMLURL))

		return nil
	},
}
