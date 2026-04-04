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
	RunE: func(cmd *cobra.Command, args []string) error {
		name := "default"
		if len(args) > 0 {
			name = args[0]
		}

		p, err := profile.Load(name)
		if err != nil {
			return err
		}

		filename := name + "-skel.json"
		data, err := json.MarshalIndent(p, "", "  ")
		if err != nil {
			return fmt.Errorf("encoding profile: %w", err)
		}

		if err := os.WriteFile(filename, data, 0600); err != nil {
			return fmt.Errorf("writing %s: %w", filename, err)
		}

		fmt.Printf("\n  %s Exported profile %s to %s\n", green("✓"), bold("'"+p.Name+"'"), bold(filename))
		fmt.Printf("  %s\n", dividerStyle.Render("────────────────────────────────────────────"))
		fmt.Printf("  %s\n\n", dim("Share this file and others can run: skel import "+filename))
		return nil
	},
}
