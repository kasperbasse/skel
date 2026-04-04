package cmd

import (
	"fmt"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/kasperbasse/skel/cmd/tui"
	"github.com/kasperbasse/skel/internal/profile"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		profiles, err := profile.ListAll()
		if err != nil {
			return err
		}

		// Sort profiles ASC by CreatedAt
		slices.SortFunc(profiles, func(a, b *profile.Profile) int {
			return b.CreatedAt.Compare(a.CreatedAt) // newest first
		})

		if len(profiles) == 0 {
			printFirstRun()
			return nil
		}

		if IsInteractive() {
			m := tui.NewListModel(profiles)
			prog := tea.NewProgram(m)
			finalModel, err := prog.Run()
			if err != nil {
				return fmt.Errorf("list failed: %w", err)
			}

			result, ok := finalModel.(tui.ListModel)
			if !ok {
				return fmt.Errorf("unexpected model type from list")
			}
			switch result.Action() {
			case tui.ListActionShow:
				return showCmd.RunE(showCmd, []string{result.Chosen()})
			case tui.ListActionDelete:
				for _, name := range result.Deleted() {
					fmt.Printf("  %s Deleted %s\n", green("✓"), bold("'"+name+"'"))
				}
				fmt.Println()
			default:
				return nil
			}
			return nil
		}

		// Non-interactive fallback
		fmt.Printf("\n  %s Saved profiles\n", cyan("📦"))
		fmt.Printf("  %s\n\n", dividerStyle.Render("────────────────────────────────────────────"))

		for _, p := range profiles {
			parts := profileSummaryParts(p)
			if len(parts) > 5 {
				parts = parts[:5]
			}
			fmt.Printf("  %s %s  %s\n", green("▸"), bold(p.Name), dim(timeAgo(p.CreatedAt)))
			fmt.Printf("  %s\n\n", strings.Join(parts, dim(" · ")))
		}

		printNextSteps(
			nextStep("skel show <name>", "to review a profile"),
			nextStep("skel restore <name>", "to apply a profile"),
		)
		return nil
	},
}
