package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/kasperbasse/skel/internal/profile"
)

var exportCmd = &cobra.Command{
	Use:   "export [profile-name]",
	Short: "Export a profile to a shareable JSON file",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runExport,
}

// runExport exports a profile to a JSON file.
func runExport(_ *cobra.Command, args []string) error {
	name := SelectProfileName(args)

	p, err := LoadAnyProfile(name)
	if err != nil {
		return err
	}

	return exportProfileToFile(p)
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
	fmt.Printf("  %s\n", dividerStyle.Render("────────────────────────────────────────────"))
	fmt.Printf("  %s\n\n", dim("Share this file and others can run: skel import "+filename))
	printNextSteps(
		nextStep("skel list", "to see all profiles"),
	)
}
