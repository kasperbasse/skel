package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kasperbasse/skel/internal/profile"
	"github.com/kasperbasse/skel/internal/scanner"
)

var importCmd = &cobra.Command{
	Use:   "import [file]",
	Short: "Import a profile from an exported JSON file",
	Args:  requireArgs("import <file>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		fi, err := os.Stat(args[0])
		if err != nil {
			return fmt.Errorf("could not read file: %w", err)
		}
		if fi.Size() > profile.MaxImportSize {
			return fmt.Errorf("profile file too large (%d bytes, max %d)", fi.Size(), profile.MaxImportSize)
		}

		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("could not read file: %w", err)
		}

		var p profile.Profile
		if err := json.Unmarshal(data, &p); err != nil {
			return fmt.Errorf("invalid profile file: %w", err)
		}

		if err := p.Validate(); err != nil {
			return fmt.Errorf("unsafe profile: %w", err)
		}

		if _, err := profile.Save(&p); err != nil {
			return err
		}

		fmt.Printf("\n  %s Imported profile %s\n", green("✓"), bold("'"+p.Name+"'"))
		fmt.Printf("  %s\n\n", dim(fmt.Sprintf(
			"%d formulas · %d casks · originally saved from %s",
			len(p.Homebrew.Formulas), len(p.Homebrew.Casks), p.Machine,
		)))

		// Warn about shell/git configs that will execute as the user.
		var warnings []string
		for _, g := range scanGroups {
			if g.ImportWarnings == nil {
				continue
			}
			warnings = append(warnings, g.ImportWarnings(&p)...)
		}
		if len(warnings) > 0 {
			fmt.Printf("  %s This profile contains shell/git configs (%s)\n", yellow("⚠"), strings.Join(warnings, ", "))
			fmt.Printf("  %s\n\n", dim("Review with 'skel show "+p.Name+"' before restoring - these files run as your user."))
		}

		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete [profile-name]",
	Short: "Delete a saved profile",
	Args:  requireArgs("delete <profile-name>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		p, err := profile.Load(name)
		if err != nil {
			return err
		}

		if err := profile.Delete(name); err != nil {
			return err
		}

		fmt.Printf("\n  %s Deleted profile %s\n", green("✓"), bold("'"+p.Name+"'"))
		fmt.Printf("  %s\n\n", dim(fmt.Sprintf(
			"Saved %s from %s", p.CreatedAt.Format("Jan 02 2006"), p.Machine,
		)))
		return nil
	},
}

var updateCmd = &cobra.Command{
	Use:   "update [profile-name]",
	Short: "Re-scan your Mac and update an existing profile",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := "default"
		if len(args) > 0 {
			name = args[0]
		}

		old, _ := profile.Load(name) // best-effort, nil if it doesn't exist yet

		fmt.Printf("\n  %s Updating profile %s...\n", cyan("🔄"), bold("'"+name+"'"))

		spin := NewSpinner("Re-scanning your environment...")
		spin.Start()

		p, warnings, err := scanner.Run(name)
		spin.Stop()

		if err != nil {
			return err
		}

		if len(warnings) > 0 {
			fmt.Println()
			for _, w := range warnings {
				fmt.Printf("  %s %s\n", yellow("⚠"), dim(w))
			}
		}

		if _, err := profile.Save(p); err != nil {
			printErr("\n  %s Failed to save profile: %v\n", red("✗"), err)
			return err
		}

		if old != nil {
			fmt.Println()
			printUpdateDiff(old, p)
		}

		fmt.Printf("\n  %s Profile %s updated\n\n", green("✓"), bold("'"+name+"'"))
		return nil
	},
}

func printUpdateDiff(old, updated *profile.Profile) {
	var lines []string
	for _, s := range profileSections {
		from := len(s.Items(old))
		to := len(s.Items(updated))
		if from == to {
			continue
		}
		diff := to - from
		var diffStr string
		if diff > 0 {
			diffStr = green(fmt.Sprintf("+%d", diff))
		} else {
			diffStr = red(fmt.Sprintf("%d", diff))
		}
		lines = append(lines, fmt.Sprintf("  %s %s: %d → %d (%s)",
			dim("·"), s.Label, from, to, diffStr))
	}

	if len(lines) == 0 {
		fmt.Printf("  %s No changes detected\n", dim("·"))
		return
	}
	fmt.Println(strings.Join(lines, "\n"))
}
