package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kasperbasse/skel/internal/profile"
)

var importCmd = &cobra.Command{
	Use:   "import [file]",
	Short: "Import a profile from an exported JSON file",
	Args:  requireArgs("import <file>"),
	RunE:  runImport,
}

func runImport(_ *cobra.Command, args []string) error {
	printCommandHeader("import", "Importing profile...", randomMessage(importStartMsgs))

	p, err := loadImportedProfile(args[0])
	if err != nil {
		return enhanceError(err)
	}

	if _, err := profile.Save(p); err != nil {
		return enhanceError(err)
	}

	printImportSummary(p)
	printImportSecurityNotice(p)
	printNextSteps(
		nextStep("skel show "+p.Name, "to review the profile"),
		nextStep("skel restore "+p.Name, "to apply this setup"),
	)

	return nil
}

func loadImportedProfile(path string) (*profile.Profile, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}
	if fi.Size() > profile.MaxImportSize {
		return nil, fmt.Errorf("profile file too large (%d bytes, max %d)", fi.Size(), profile.MaxImportSize)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}

	var p profile.Profile
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("invalid profile file: %w", err)
	}

	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("unsafe profile: %w", err)
	}

	return &p, nil
}

func printImportSummary(p *profile.Profile) {
	fmt.Printf("\n  %s Imported profile %s\n", iconCheck(), bold("'"+p.Name+"'"))
	fmt.Printf("  %s\n", dividerStyle.Render(dividerLine))
	fmt.Printf("  %s\n\n",
		dim(fmt.Sprintf("%s · %s · originally saved from %s",
			countLabel(len(p.Homebrew.Formulas), "formula", "formulas"),
			countLabel(len(p.Homebrew.Casks), "cask", "casks"),
			p.Machine,
		)),
	)
}

func printImportSecurityNotice(p *profile.Profile) {
	warnings := collectImportWarnings(p)
	if len(warnings) == 0 {
		return
	}

	warningText := fmt.Sprintf("This profile contains: %s\n\n", strings.Join(warnings, ", ")) +
		"Review the profile with " + cyan("skel show "+p.Name) + " before restoring.\n" +
		"These files execute code when your shell starts or git runs."
	printWarningBox("Security Notice", warningText)
}
