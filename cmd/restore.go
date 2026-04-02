package cmd

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/kasperbasse/skel/cmd/tui"
	"github.com/kasperbasse/skel/internal/profile"
	"github.com/kasperbasse/skel/internal/restore"
)

var (
	dryRun  bool
	onlyStr string
)

var validSections = allRestoreKeys()

var restoreCmd = &cobra.Command{
	Use:   "restore [profile-name]",
	Short: "Restore a saved profile on this Mac",
	Long: fmt.Sprintf(
		"Restore a saved profile. Use --only to limit sections: %s",
		strings.Join(validSections, ", "),
	),
	Args: requireArgs("restore <profile-name>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := profile.Load(args[0])
		if err != nil {
			return err
		}

		opts, err := parseOnlyFlag(onlyStr)
		if err != nil {
			return err
		}

		fmt.Printf("\n  %s Restoring profile %s\n", cyan("🚀"), bold("'"+p.Name+"'"))
		fmt.Printf("    %s\n\n", dim(fmt.Sprintf("Saved %s from %s", p.CreatedAt.Format("Jan 02 2006"), p.Machine)))
		fmt.Printf("  %s\n\n", dim(randomMessage(restoreStartMsgs)))

		if dryRun {
			fmt.Printf("  %s Dry run - nothing will be installed\n\n", yellow("⚠"))
			printDryRun(p, opts)
			return nil
		}

		if IsInteractive() {
			// Show section picker if no --only flag was provided.
			if onlyStr == "" {
				selectItems := buildSelectItems(p)
				if len(selectItems) > 0 {
					sm := tui.NewSelectRestoreModel(selectItems)
					selectProg := tea.NewProgram(sm)
					result, err := selectProg.Run()
					if err != nil {
						return fmt.Errorf("selection failed: %w", err)
					}
					final := result.(tui.SelectRestoreModel)
					if !final.Confirmed() {
						fmt.Printf("  %s Canceled.\n\n", dim("-"))
						return nil
					}
					opts = &restore.Options{Sections: final.SelectedKeys()}
				}
			}

			m := tui.NewRestoreModel(p, opts, randomMessage(restoreStartMsgs))
			prog := tea.NewProgram(m)
			if _, err := prog.Run(); err != nil {
				return fmt.Errorf("restore failed: %w", err)
			}
		} else {
			// Non-interactive fallback
			var failed []restore.Result

			restore.Run(p, opts, func(r restore.Result) {
				progress := dim(fmt.Sprintf("[%d/%d]", r.Index, r.Total))
				if r.Success {
					if r.Message == "already installed" {
						fmt.Printf("  %s %s %s  %s\n", progress, green("✓"), r.Step, dim("already installed"))
					} else {
						fmt.Printf("  %s %s %s\n", progress, green("✓"), r.Step)
					}
				} else {
					fmt.Printf("  %s %s %s  %s\n", progress, red("✗"), r.Step, dim(r.Message))
					failed = append(failed, r)
				}
			})

			fmt.Println()
			if len(failed) == 0 {
				fmt.Printf("  %s %s\n\n", green("🎉"), randomMessage(restoreCompleteMsgs))
			} else {
				fmt.Printf("  %s Done with %s. Check the output above.\n\n",
					yellow("⚠"), red(fmt.Sprintf("%d errors", len(failed))))
			}
		}

		return nil
	},
}

func init() {
	restoreCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview what would be restored without making changes")
	restoreCmd.Flags().StringVar(&onlyStr, "only", "", "Restore only specific sections (comma-separated: homebrew,shell,git,editors,configs,languages,mas)")
}

func parseOnlyFlag(s string) (*restore.Options, error) {
	if s == "" {
		return &restore.Options{}, nil
	}

	valid := make(map[string]bool)
	for _, v := range validSections {
		valid[v] = true
	}

	sections := make(map[string]bool)
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(strings.ToLower(part))
		if part == "" {
			continue
		}
		if !valid[part] {
			return nil, fmt.Errorf("unknown section %q, valid: %s", part, strings.Join(validSections, ", "))
		}
		sections[part] = true
	}

	return &restore.Options{Sections: sections}, nil
}

// buildSelectItems creates checklist items from scanGroups for the given profile.
// Only includes sections that have restorable data.
func buildSelectItems(p *profile.Profile) []tui.SelectItem {
	var items []tui.SelectItem
	for _, g := range scanGroups {
		if len(g.RestoreKeys) == 0 {
			continue
		}
		summary := ""
		if g.ScanSummary != nil {
			summary = g.ScanSummary(p)
		}
		if summary == "" {
			continue // no data for this section
		}
		items = append(items, tui.SelectItem{
			Icon:     g.Icon,
			Label:    g.Label,
			Keys:     g.RestoreKeys,
			Summary:  summary,
			Selected: true,
		})
	}
	return items
}

func printDryRun(p *profile.Profile, opts *restore.Options) {
	for _, g := range scanGroups {
		if g.DryRun == nil {
			continue
		}
		g.DryRun(p, opts)
	}
	fmt.Println()
}
