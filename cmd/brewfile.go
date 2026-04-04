package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kasperbasse/skel/internal/brewfile"
	"github.com/kasperbasse/skel/internal/profile"
)

var brewfileOutputFile string

var brewfileCmd = &cobra.Command{
	Use:   "brewfile",
	Short: "Import and export Brewfiles",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Import and export Brewfiles")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  skel brewfile <command>")
		fmt.Println()
		fmt.Println("Available Commands:")
		fmt.Println("  export      Export a profile's Homebrew packages as a Brewfile")
		fmt.Println("  import      Import a Brewfile into a profile")
		fmt.Println()
		fmt.Println("Use \"skel brewfile <command> --help\" for more info.")
	},
}

var brewfileExportCmd = &cobra.Command{
	Use:   "export [profile-name]",
	Short: "Export a profile's Homebrew packages as a Brewfile",
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

		content := brewfile.Generate(p.Homebrew)
		if content == "" {
			fmt.Printf("\n  %s Profile %s has no Homebrew packages to export.\n\n",
				yellow("!"), bold("'"+p.Name+"'"))
			return nil
		}

		output := brewfileOutputFile
		if output == "" {
			output = "Brewfile"
		}

		if err := os.WriteFile(output, []byte(content), 0600); err != nil {
			return fmt.Errorf("writing %s: %w", output, err)
		}

		total := len(p.Homebrew.Taps) + len(p.Homebrew.Formulas) + len(p.Homebrew.Casks) + len(p.Homebrew.MasApps)
		fmt.Printf("\n  %s Exported to %s %s\n",
			green("✓"), bold(output), dim(fmt.Sprintf("(%d entries)", total)))
		fmt.Printf("  %s\n\n", dim("Compatible with 'brew bundle install'"))
		return nil
	},
}

var brewfileImportCmd = &cobra.Command{
	Use:   "import [file]",
	Short: "Import a Brewfile into a profile",
	Args:  requireArgs("brewfile import <file>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		fi, err := os.Stat(args[0])
		if err != nil {
			return fmt.Errorf("could not read file: %w", err)
		}
		if fi.Size() > int64(brewfile.MaxBrewfileSize) {
			return fmt.Errorf("Brewfile too large (%d bytes, max %d)", fi.Size(), brewfile.MaxBrewfileSize)
		}

		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("could not read file: %w", err)
		}

		h, err := brewfile.Parse(string(data))
		if err != nil {
			return fmt.Errorf("invalid Brewfile: %w", err)
		}

		profileName, _ := cmd.Flags().GetString("profile")
		if profileName == "" {
			profileName = "brewfile"
		}

		var p *profile.Profile
		if profile.Exists(profileName) {
			p, err = profile.Load(profileName)
			if err != nil {
				return err
			}
			// Merge: add new items, deduplicate
			p.Homebrew = mergeHomebrew(p.Homebrew, h)
		} else {
			p = &profile.Profile{
				Name:     profileName,
				Machine:  "brewfile-import",
				Homebrew: h,
			}
		}

		if _, err := profile.Save(p); err != nil {
			return err
		}

		fmt.Printf("\n  %s Imported Brewfile into profile %s\n", green("✓"), bold("'"+profileName+"'"))

		var parts []string
		if len(h.Taps) > 0 {
			parts = append(parts, fmt.Sprintf("%d taps", len(h.Taps)))
		}
		if len(h.Formulas) > 0 {
			parts = append(parts, fmt.Sprintf("%d formulas", len(h.Formulas)))
		}
		if len(h.Casks) > 0 {
			parts = append(parts, fmt.Sprintf("%d casks", len(h.Casks)))
		}
		if len(h.MasApps) > 0 {
			parts = append(parts, fmt.Sprintf("%d App Store apps", len(h.MasApps)))
		}
		if len(parts) > 0 {
			fmt.Printf("  %s\n", dim(strings.Join(parts, " · ")))
		}
		fmt.Println()
		return nil
	},
}

func init() {
	brewfileExportCmd.Flags().StringVarP(&brewfileOutputFile, "output", "o", "", "Output file (default: Brewfile)")
	brewfileImportCmd.Flags().StringP("profile", "p", "", "Profile name to import into (default: brewfile)")

	brewfileCmd.AddCommand(brewfileExportCmd)
	brewfileCmd.AddCommand(brewfileImportCmd)
}

// mergeHomebrew combines two HomebrewProfiles, deduplicating entries.
func mergeHomebrew(existing, incoming profile.HomebrewProfile) profile.HomebrewProfile {
	existing.Taps = mergeStrings(existing.Taps, incoming.Taps)
	existing.Formulas = mergeStrings(existing.Formulas, incoming.Formulas)
	existing.Casks = mergeStrings(existing.Casks, incoming.Casks)

	// Deduplicate MAS apps by ID
	masSet := make(map[string]bool)
	for _, app := range existing.MasApps {
		masSet[app.ID] = true
	}
	for _, app := range incoming.MasApps {
		if !masSet[app.ID] {
			existing.MasApps = append(existing.MasApps, app)
		}
	}

	return existing
}

func mergeStrings(a, b []string) []string {
	set := make(map[string]bool, len(a))
	for _, s := range a {
		set[s] = true
	}
	for _, s := range b {
		if !set[s] {
			a = append(a, s)
			set[s] = true
		}
	}
	return a
}
