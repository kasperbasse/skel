package cmd

import (
	"fmt"

	"github.com/kasperbasse/skel/internal/profile"
)

// SelectProfileName extracts the profile name from command arguments.
// Defaults to "default" if no argument provided.
func SelectProfileName(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return "default"
}

// LoadAnyProfile loads a profile with first-run fallback.
// Prints first-run message if no profiles exist.
func LoadAnyProfile(name string) (*profile.Profile, error) {
	p, err := profile.Load(name)
	if err != nil {
		all, listErr := profile.ListAll()
		if listErr == nil && len(all) == 0 {
			printFirstRun()
			return nil, fmt.Errorf("no profiles found")
		}
		return nil, enhanceError(err)
	}
	return p, nil
}

// loadProfileOrFail loads a profile and returns a contextual error on failure.
func loadProfileOrFail(name string) (*profile.Profile, error) {
	p, err := profile.Load(name)
	if err != nil {
		return nil, enhanceError(fmt.Errorf("loading profile '%s': %w", name, err))
	}
	return p, nil
}
