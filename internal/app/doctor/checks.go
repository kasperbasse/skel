package doctor

import (
	"fmt"

	"github.com/kasperbasse/skel/internal/profile"
	internalui "github.com/kasperbasse/skel/internal/ui"
)

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
func buildChecksWith(p *profile.Profile, resolve ToolResolver, exists ToolExists) []Check {
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

func BuildChecks(p *profile.Profile) []Check {
	return buildChecksWith(p, ToolDoctorInfo, CommandExists)
}

func PrintCheck(c Check) {
	if c.OK {
		fmt.Printf("  %s  %s\n", internalui.IconCheck(), c.Label)
	} else {
		fmt.Printf("  %s  %s\n", internalui.IconCross(), internalui.Bold(c.Label))
		fmt.Printf("       %s  %s\n", internalui.Dim("→"), internalui.Dim(c.Fix))
	}
}

// RunChecks prints all checks for a profile and returns the number of issues found and if It's empty or not.
func RunChecks(p *profile.Profile) (issues int, empty bool) {
	checks := BuildChecks(p)
	if len(checks) == 0 {
		return 0, true
	}
	for _, c := range checks {
		PrintCheck(c)
		if !c.OK {
			issues++
		}
	}
	return issues, false
}
