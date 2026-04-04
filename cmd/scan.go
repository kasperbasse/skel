package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		name := "default"
		if len(args) > 0 {
			name = args[0]
		}

		// Overwrite warning
		if !forceOverwrite && profile.Exists(name) {
			existing, err := profile.Load(name)
			if err == nil {
				fmt.Printf("\n  %s Profile %s already exists (saved %s)\n\n", yellow("⚠"), bold("'"+name+"'"), existing.CreatedAt.Format("Jan 02 2006 15:04"))
				fmt.Printf("  Overwrite? [y/N] ")
				reader := bufio.NewReader(os.Stdin)
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(strings.ToLower(answer))
				if answer != "y" && answer != "yes" {
					fmt.Printf("  %s Canceled.\n\n", dim("-"))
					return nil
				}
			}
		}

		startMsg := randomMessage(scanStartMsgs)
		fmt.Printf("\n  %s %s\n", cyan("🔍"), startMsg)
		fmt.Printf("  %s\n\n", dividerStyle.Render("────────────────────────────────────────────"))

		var p *profile.Profile
		var warnings []string

		if IsInteractive() {
			m := tui.NewScanModel(name, startMsg)
			prog := tea.NewProgram(m)
			finalModel, err := prog.Run()
			if err != nil {
				return fmt.Errorf("scan failed: %w", err)
			}
			scanModel, ok := finalModel.(tui.ScanModel)
			if !ok {
				return fmt.Errorf("unexpected model type from scan")
			}
			result := scanModel.Result()
			if result == nil {
				return fmt.Errorf("scan was interrupted")
			}
			if result.Err != nil {
				return result.Err
			}
			p = result.Profile
			warnings = result.Warnings
		} else {
			spin := NewSpinner("Gathering your environment...")
			spin.Start()
			var err error
			p, warnings, err = scanner.Run(name)
			spin.Stop()
			if err != nil {
				return err
			}
		}

		if len(warnings) > 0 {
			fmt.Println()
			for _, w := range warnings {
				fmt.Printf("  %s %s\n", yellow("⚠"), dim(w))
			}
		}

		for _, g := range scanGroups {
			if summary := g.ScanSummary(p); summary != "" {
				printRow(green("✓"), g.Label, summary)
			}
		}

		size, err := profile.Save(p)
		if err != nil {
			printErr("\n  %s Failed to save profile: %v\n", red("✗"), err)
			return err
		}

		if size > 5*1024*1024 {
			fmt.Printf("\n  %s Profile is %dMB - consider trimming large configs\n",
				yellow("⚠"), size/(1024*1024))
		}

		fmt.Printf("\n  %s %s %s\n", green("✓"), randomMessage(scanCompleteMsgs), dim(fmt.Sprintf("(%d items captured)", profileItemCount(p))))
		fmt.Printf("  %s\n\n", dividerStyle.Render("────────────────────────────────────────────"))
		printNextSteps(
			nextStep("skel show "+name, "to review details"),
			nextStep("skel restore "+name, "to apply this setup"),
		)
		return nil
	},
}

func init() {
	scanCmd.Flags().BoolVar(&forceOverwrite, "force", false, "Overwrite existing profile without confirmation")
}
