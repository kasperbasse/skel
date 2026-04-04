package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// requireArgs returns a Cobra args validator that shows a friendly error with usage hint.
func requireArgs(usage string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("missing required argument\n\nUsage: skel %s", usage)
		}
		return nil
	}
}

// requireExactArgs returns a Cobra args validator for exactly n args with a friendly error.
func requireExactArgs(n int, usage string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < n {
			return fmt.Errorf("missing required argument(s)\n\nUsage: skel %s", usage)
		}
		if len(args) > n {
			return fmt.Errorf("too many arguments\n\nUsage: skel %s", usage)
		}
		return nil
	}
}
