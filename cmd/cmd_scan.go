package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/kasperbasse/skel/cmd/tui"
	"github.com/kasperbasse/skel/internal/profile"
	"github.com/kasperbasse/skel/internal/scanner"
)

var forceOverwrite bool

var scanCmd = &cobra.Command{
	Use:   "scan [profile-name]",
	Short: "Scan your Mac and save a setup profile",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runScan,
}

// runScan captures the current machine state and saves it as a profile.
func runScan(_ *cobra.Command, args []string) error {
	name := selectProfileName(args)

	// Check if overwriting existing profile
	if !forceOverwrite {
		ok, err := confirmOverwrite(name)
		if err != nil {
			return enhanceError(err)
		}
		if !ok {
			fmt.Printf("  %s Canceled.\n\n", iconDash())
			return nil
		}
	}

	startMsg := randomMessage(scanStartMsgs)
	printCommandHeader("scan", startMsg)
	fmt.Println()

	// Perform scan
	p, warnings, err := performScan(name)
	if err != nil {
		return enhanceError(err)
	}

	// Display warnings if any
	printWarnings(warnings)

	// Display captured sections
	for _, g := range scanGroups {
		if summary := g.ScanSummary(p); summary != "" {
			printRow(g.Label, summary)
		}
	}

	// Save profile
	return saveScanResult(name, p)
}

// performScan captures the current machine state.
func performScan(profileName string) (*profile.Profile, []string, error) {
	if IsInteractive() {
		return performScanInteractive(profileName)
	}
	return performScanNonInteractive(profileName)
}

// performScanInteractive uses TUI to show progress while scanning.
func performScanInteractive(profileName string) (*profile.Profile, []string, error) {
	m := tui.NewScanModel(profileName, "Gathering your environment...")
	prog := tea.NewProgram(m)
	finalModel, err := prog.Run()
	if err != nil {
		return nil, nil, fmt.Errorf("scan failed: %w", err)
	}

	scanModel, ok := finalModel.(tui.ScanModel)
	if !ok {
		return nil, nil, fmt.Errorf("unexpected model type from scan")
	}

	result := scanModel.Result()
	if result == nil {
		return nil, nil, fmt.Errorf("scan was interrupted")
	}

	return result.Profile, result.Warnings, result.Err
}

// performScanNonInteractive scans without UI feedback.
func performScanNonInteractive(profileName string) (*profile.Profile, []string, error) {
	spin := newSpinner("Gathering your environment...")
	spin.Start()

	p, warnings, err := scanner.Run(profileName)

	spin.Stop()

	return p, warnings, err
}

// saveScanResult saves the scanned profile and reports results.
func saveScanResult(profileName string, p *profile.Profile) error {
	size, err := profile.Save(p)
	if err != nil {
		err = enhanceError(err)
		printError(err)
		return err
	}

	if size > 5*1024*1024 {
		fmt.Printf("  %s Profile is %dMB - consider trimming large configs\n\n",
			iconWarn(), size/(1024*1024))
	}

	itemCount := profileItemCount(p)
	fmt.Printf("  %s %s %s\n",
		iconCheck(),
		randomMessage(scanCompleteMsgs),
		dim(fmt.Sprintf("(%d items captured)", itemCount)),
	)
	fmt.Printf("  %s\n\n", dividerStyle.Render(dividerLine))

	printNextSteps(
		nextStep("skel show "+profileName, "to review details"),
		nextStep("skel restore "+profileName, "to apply this setup"),
	)

	return nil
}

func init() {
	scanCmd.Flags().BoolVar(&forceOverwrite, "force", false, "Overwrite existing profile without confirmation")
}
