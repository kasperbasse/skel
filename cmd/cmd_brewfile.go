package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kasperbasse/skel/internal/brewfile"
	"github.com/kasperbasse/skel/internal/profile"
)

var brewfileOutputFile string
var brewfileImportName string

var brewfileCmd = &cobra.Command{
	Use:   "brewfile",
	Short: "Import and export Brewfiles",
	RunE:  runBrewfileHelp,
}

func runBrewfileHelp(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}

var brewfileExportCmd = &cobra.Command{
	Use:   "export [profile-name]",
	Short: "Export a profile's Homebrew packages as a Brewfile",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runBrewfileExport,
}

func runBrewfileExport(_ *cobra.Command, args []string) error {
	name := SelectProfileName(args)

	PrintCommandHeader("brewfile-export", fmt.Sprintf("Exporting Brewfile from %s", bold("'"+name+"'")))

	p, err := LoadAnyProfile(name)
	if err != nil {
		return err
	}

	content := brewfile.Generate(p.Homebrew)
	if content == "" {
		fmt.Printf("  %s Profile %s has no Homebrew packages to export.\n\n",
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
	fmt.Printf("\n  %s Exported to %s %s\n", iconCheck(), bold(output), dim(fmt.Sprintf("(%d entries)", total)))
	fmt.Printf("  %s\n", dividerStyle.Render(dividerLine))
	fmt.Printf("  %s\n\n", dim("Compatible with 'brew bundle install'"))
	printNextSteps(
		nextStep("brew bundle install", "to install all packages"),
	)

	return nil
}

var brewfileImportCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import a Brewfile into a profile",
	Args:  requireArgs("brewfile import <file>"),
	RunE:  runBrewfileImport,
}

func runBrewfileImport(_ *cobra.Command, args []string) error {
	PrintCommandHeader("brewfile-import", "Importing Brewfile...")

	data, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("reading %s: %w", args[0], err)
	}

	h, err := brewfile.Parse(string(data))
	if err != nil {
		return fmt.Errorf("parsing brewfile: %w", err)
	}

	name := brewfileImportName
	if name == "" {
		name = defaultBrewfileProfileName(args[0])
	}

	p := &profile.Profile{
		Name:     name,
		Homebrew: h,
	}

	if _, err := profile.Save(p); err != nil {
		return fmt.Errorf("saving profile: %w", err)
	}

	printBrewfileImportSummary(p)
	printNextSteps(
		nextStep("skel show "+p.Name, "to review the imported profile"),
		nextStep("skel restore "+p.Name, "to apply this setup"),
	)

	return nil
}

func defaultBrewfileProfileName(path string) string {
	base := filepath.Base(path)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	if name == "" || name == "." {
		return "brewfile-import"
	}
	return name
}

func printBrewfileImportSummary(p *profile.Profile) {
	total := len(p.Homebrew.Taps) + len(p.Homebrew.Formulas) + len(p.Homebrew.Casks) + len(p.Homebrew.MasApps)
	fmt.Printf("\n  %s Imported Brewfile into profile %s %s\n", iconCheck(), bold("'"+p.Name+"'"), dim(fmt.Sprintf("(%d entries)", total)))
	fmt.Printf("  %s\n", dividerStyle.Render(dividerLine))
	fmt.Printf("  %s\n\n", dim("Homebrew taps, formulas, casks, and mas apps were imported"))
}

func init() {
	brewfileCmd.AddCommand(brewfileImportCmd)
	brewfileCmd.AddCommand(brewfileExportCmd)
	brewfileExportCmd.Flags().StringVar(&brewfileOutputFile, "output", "Brewfile", "Output file")
	brewfileImportCmd.Flags().StringVar(&brewfileImportName, "name", "", "Profile name to save the imported Brewfile as")
}
