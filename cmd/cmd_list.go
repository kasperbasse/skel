package cmd

import (
	"fmt"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/kasperbasse/skel/cmd/tui"
	"github.com/kasperbasse/skel/internal/profile"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved profiles",
	RunE:  runList,
}

var runListProgram = func(m tui.ListModel) (tea.Model, error) {
	return tea.NewProgram(m).Run()
}

// runList displays all profiles, sorted by recency.
func runList(_ *cobra.Command, _ []string) error {
	profiles, err := profile.ListAll()
	if err != nil {
		return enhanceError(fmt.Errorf("listing profiles: %w", err))
	}

	if len(profiles) == 0 {
		printFirstRun()
		return nil
	}

	// Sort newest first
	slices.SortFunc(profiles, func(a, b *profile.Profile) int {
		return b.CreatedAt.Compare(a.CreatedAt)
	})

	if IsInteractive() {
		return runListInteractive(profiles)
	}
	return runListNonInteractive(profiles)
}

// runListInteractive displays profiles in interactive mode.
func runListInteractive(profiles []*profile.Profile) error {
	m := tui.NewListModel(profiles)
	finalModel, err := runListProgram(m)
	if err != nil {
		return enhanceError(fmt.Errorf("list failed: %w", err))
	}

	result, ok := finalModel.(tui.ListModel)
	if !ok {
		return enhanceError(fmt.Errorf("unexpected model type from list"))
	}

	return handleListAction(result)
}

// handleListAction processes the user's list action (show/delete/cancel).
func handleListAction(result tui.ListModel) error {
	switch result.Action() {
	case tui.ListActionShow:
		return showCmd.RunE(showCmd, []string{result.Chosen()})

	case tui.ListActionDelete:
		for _, name := range result.Deleted() {
			fmt.Printf("  %s Deleted %s\n", iconCheck(), bold("'"+name+"'"))
		}
		fmt.Println()

	default:
		// User canceled
	}

	return nil
}

// runListNonInteractive displays profiles in non-interactive table mode.
func runListNonInteractive(profiles []*profile.Profile) error {
	printCommandHeader("list", fmt.Sprintf("Profiles (%s)", cyan(fmt.Sprintf("%d", len(profiles)))))
	fmt.Printf("  %s\n\n", dim("Overview: profile · status · modified · machine"))

	printProfilesTable(profiles)
	fmt.Println()

	printNextSteps(
		nextStep("skel show <name>", "to review a profile"),
		nextStep("skel restore <name>", "to apply a profile"),
	)
	return nil
}
