package cmd

import (
	"github.com/spf13/cobra"

	"github.com/kasperbasse/skel/internal/profile"
)

// profileNameCompletion returns saved profile names for shell tab-completion.
func profileNameCompletion(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	profiles, err := profile.ListAll()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	names := make([]string, 0, len(profiles))
	for _, p := range profiles {
		names = append(names, p.Name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

// singleProfileCompletion completes only when no profile arg has been given yet.
func singleProfileCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return profileNameCompletion(cmd, args, toComplete)
}

// twoProfileCompletion completes profile names for commands that take two
// profile arguments (e.g. diff).
func twoProfileCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) >= 2 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return profileNameCompletion(cmd, args, toComplete)
}

func init() {
	// Commands that accept exactly one profile name.
	for _, cmd := range []*cobra.Command{
		scanCmd, showCmd, restoreCmd, driftCmd, updateCmd, deleteCmd, publishCmd,
	} {
		cmd.ValidArgsFunction = singleProfileCompletion
	}

	// diff takes two profile names.
	diffCmd.ValidArgsFunction = twoProfileCompletion
}
