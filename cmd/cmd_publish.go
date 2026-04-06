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
	RunE: runPublish,
}

// runPublish uploads a profile to GitHub Gist.
func runPublish(_ *cobra.Command, args []string) error {
	name := selectProfileName(args)

	// Check authentication
	token, err := github.ResolveToken()
	if err != nil {
		return enhanceError(err)
	}

	// Load profile
	p, err := loadAnyProfile(name)
	if err != nil {
		return err
	}

	printCommandHeader("publish", fmt.Sprintf("Publishing profile %s", bold("'"+p.Name+"'")))

	// Prepare profile for publishing (with optional redaction)
	pub := prepareForPublishing(p, publishNoRedact)

	// Encode to JSON
	data, err := json.MarshalIndent(pub, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding profile: %w", err)
	}

	// Upload to GitHub
	gist, err := uploadToGitHub(token, p.Name, string(data))
	if err != nil {
		return enhanceError(err)
	}

	printPublishSuccess(gist)
	printNextSteps(
		nextStep("skel clone "+gist.HTMLURL, "on another Mac"),
	)

	return nil
}

// prepareForPublishing optionally redacts sensitive data.
func prepareForPublishing(p *profile.Profile, noRedact bool) *profile.Profile {
	if noRedact {
		fmt.Printf("  %s Publishing without redaction - git identity and hostname will be visible.\n\n", iconWarn())
		tmp := *p
		return &tmp
	}

	pub := p.Redact()
	fmt.Printf("  %s Redacted before publishing: %s\n\n",
		iconDot(),
		dim("git identity · hostname · SSH key comments"),
	)
	return pub
}

// uploadToGitHub creates a gist with the profile.
func uploadToGitHub(token, profileName, data string) (*github.Gist, error) {
	spin := newSpinner("Publishing to GitHub Gist...")
	spin.Start()

	filename := profileName + "-skel.json"
	gist, err := github.CreateGist(token, &github.CreateGistRequest{
		Description: fmt.Sprintf("skel profile: %s", profileName),
		Public:      true,
		Files: map[string]github.CreateGistFile{
			filename: {Content: data},
		},
	})

	spin.Stop()

	if err != nil {
		return nil, fmt.Errorf("uploading to GitHub: %w", err)
	}

	return gist, nil
}

// printPublishSuccess displays the published gist URL.
func printPublishSuccess(gist *github.Gist) {
	fmt.Printf("\n  %s %s\n", iconCheck(), randomMessage(publishCompleteMsgs))
	fmt.Printf("  %s\n", dim(gist.HTMLURL))
	fmt.Printf("  %s\n\n", dim("Others can clone it with: skel clone "+gist.HTMLURL))
}

func init() {
	publishCmd.Flags().BoolVar(&publishNoRedact, "no-redact", false, "Publish without redacting PII (not recommended)")
}
