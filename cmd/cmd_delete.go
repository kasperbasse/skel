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
		return fmt.Errorf("profile %q not found", name)
	}

	PrintCommandHeader("delete", fmt.Sprintf("Deleting %s", bold("'"+name+"'")))

	ok, err := tui.Confirm(fmt.Sprintf("Are you sure you want to delete %q?", name))
	if err != nil {
		return err
	}
	if !ok {
		fmt.Printf("  %s Delete canceled - Profile kept safe\n\n", iconDash())
		return nil
	}

	if err := profile.Delete(name); err != nil {
		return fmt.Errorf("deleting profile: %w", err)
	}

	fmt.Printf("  %s Profile deleted\n\n", iconCheck())
	return nil
}
