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
	PrintCommandHeader("import", "Importing profile...")

	p, err := loadImportedProfile(args[0])
	if err != nil {
		return err
	}

	if _, err := profile.Save(p); err != nil {
		return err
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
	fmt.Printf("  %s\n", dividerStyle.Render("────────────────────────────────────────────"))
	fmt.Printf("  %s\n\n",
		dim(fmt.Sprintf("%s formulas · %s casks · originally saved from %s",
			num(len(p.Homebrew.Formulas)), num(len(p.Homebrew.Casks)), p.Machine)),
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
