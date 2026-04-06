package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kasperbasse/skel/cmd/tui"
	"github.com/kasperbasse/skel/internal/profile"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [profile-name]",
	Short: "Delete a saved profile",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runDelete,
}

func runDelete(_ *cobra.Command, args []string) error {
	name := SelectProfileName(args)

	if !profile.Exists(name) {
		return enhanceError(fmt.Errorf("profile '%s' not found", name))
	}

	PrintCommandHeader("delete", fmt.Sprintf("Deleting %s", bold("'"+name+"'")), randomMessage(deleteStartMsgs))

	ok, err := tui.Confirm(fmt.Sprintf("Are you sure you want to delete %q?", name))
	if err != nil {
		return enhanceError(err)
	}
	if !ok {
		fmt.Printf("\n  %s Delete canceled - Profile kept safe\n\n", iconDash())
		return nil
	}

	if err := profile.Delete(name); err != nil {
		return enhanceError(fmt.Errorf("deleting profile: %w", err))
	}

	fmt.Printf("  %s Profile deleted\n\n", iconCheck())
	return nil
}
