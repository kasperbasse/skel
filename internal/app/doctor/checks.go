package doctor

import "github.com/kasperbasse/skel/internal/profile"

// Check is a rendered doctor check row.
type Check struct {
	Label string
	OK    bool
	Fix   string
}

// ToolResolver maps a tool command to display metadata.
type ToolResolver func(command string) (label, validatorCmd, fix string, ok bool)

// ToolExists checks whether a tool command is available.
type ToolExists func(command string) bool

// BuildChecks builds doctor checks from a profile and dependency callbacks.
func BuildChecks(p *profile.Profile, resolve ToolResolver, exists ToolExists) []Check {
	if p == nil {
		return nil
	}
	if resolve == nil || exists == nil {
		return nil
	}

	tools := RequiredTools(p)
	checks := make([]Check, 0, len(tools))
	for _, cmd := range tools {
		label, validatorCmd, fix, ok := resolve(cmd)
		if !ok {
			label = cmd
			validatorCmd = cmd
			fix = "Install and ensure it's in your PATH"
		}
		checks = append(checks, Check{Label: label, OK: exists(validatorCmd), Fix: fix})
	}
	return checks
}
