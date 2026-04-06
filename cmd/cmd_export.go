package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/kasperbasse/skel/internal/profile"
)

var exportNoRedact bool

var exportCmd = &cobra.Command{
	Use:   "export [profile-name]",
	Short: "Export a profile to a shareable JSON file",
	Long: `Export a profile to a shareable JSON file.

PII (git identity, hostname, SSH key comments) is redacted before exporting.
Use --no-redact to export the full profile including personal data.

Examples:
  skel export my-setup
  skel export my-setup --no-redact`,
	Args: cobra.MaximumNArgs(1),
	RunE: runExport,
}

// runExport exports a profile to a JSON file.
func runExport(_ *cobra.Command, args []string) error {
	name := selectProfileName(args)
	printCommandHeader("export", fmt.Sprintf("Exporting profile %s", bold("'"+name+"'")), randomMessage(exportStartMsgs))

	p, err := loadAnyProfile(name)
	if err != nil {
		return err
	}

	out := prepareForExport(p, exportNoRedact)
	return enhanceError(exportProfileToFile(out))
}

// prepareForExport optionally redacts sensitive data before export.
func prepareForExport(p *profile.Profile, noRedact bool) *profile.Profile {
	if noRedact {
		fmt.Printf("  %s Exporting without redaction - git identity and hostname will be visible.\n\n", iconWarn())
		// Shallow copy is sufficient here: the returned profile is immediately
		// marshaled to JSON and never mutated further.
		tmp := *p
		return &tmp
	}

	out := p.Redact()
	fmt.Printf("  %s Redacted before exporting: %s\n\n",
		iconDot(),
		dim("git identity · hostname · SSH key comments"),
	)
	return out
}

// exportProfileToFile saves a profile to a JSON file.
func exportProfileToFile(p *profile.Profile) error {
	filename := p.Name + "-skel.json"

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding profile: %w", err)
	}

	if err := os.WriteFile(filename, data, 0600); err != nil {
		return fmt.Errorf("writing %s: %w", filename, err)
	}

	printExportSuccess(p.Name, filename)
	return nil
}

// printExportSuccess displays export completion message.
func printExportSuccess(profileName, filename string) {
	fmt.Printf("\n  %s Exported profile %s to %s\n", iconCheck(), bold("'"+profileName+"'"), bold(filename))
	fmt.Printf("  %s\n", dividerStyle.Render(dividerLine))
	fmt.Printf("  %s\n\n", dim("Share this file and others can run: skel import "+filename))
	printNextSteps(
		nextStep("skel list", "to see all profiles"),
	)
}

func init() {
	exportCmd.Flags().BoolVar(&exportNoRedact, "no-redact", false, "Export without redacting PII (not recommended for sharing)")
}
