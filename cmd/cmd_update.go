package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kasperbasse/skel/internal/profile"
	"github.com/kasperbasse/skel/internal/scanner"
)

var updateCmd = &cobra.Command{
	Use:   "update [profile-name]",
	Short: "Re-scan your Mac and update an existing profile",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runUpdate,
}

func runUpdate(_ *cobra.Command, args []string) error {
	name := selectProfileName(args)
	old, _ := profile.Load(name) // best-effort

	printCommandHeader("update", fmt.Sprintf("Re-scanning and updating %s", bold("'"+name+"'")), randomMessage(updateStartMsgs))

	spin := newSpinner("Re-scanning your environment...")
	spin.Start()
	p, warnings, err := scanner.Run(name)
	spin.Stop()
	if err != nil {
		return enhanceError(err)
	}

	printWarnings(warnings)

	if _, err := profile.Save(p); err != nil {
		return enhanceError(fmt.Errorf("saving profile: %w", err))
	}

	if old != nil {
		printUpdateDiff(old, p)
	}

	fmt.Printf("  %s Profile %s updated\n\n", iconCheck(), bold("'"+name+"'"))
	return nil
}

func printUpdateDiff(old, updated *profile.Profile) {
	var lines []string

	for _, s := range profileSections {
		from := len(s.Items(old))
		to := len(s.Items(updated))
		if from == to {
			continue
		}
		diff := to - from
		var diffStr string
		if diff > 0 {
			diffStr = green(fmt.Sprintf("+%d", diff))
		} else {
			diffStr = red(fmt.Sprintf("%d", diff))
		}
		lines = append(lines, fmt.Sprintf("  %s %-24s %d → %d  %s",
			dim("·"), s.Label, from, to, diffStr))
	}

	for _, v := range versionFields {
		fromVer := shortVer(v.Value(old))
		toVer := shortVer(v.Value(updated))
		if fromVer == toVer {
			continue
		}
		switch {
		case fromVer == "none":
			lines = append(lines, fmt.Sprintf("  %s %-24s %s", dim("·"), v.Label, green(toVer)))
		case toVer == "none":
			lines = append(lines, fmt.Sprintf("  %s %-24s %s", dim("·"), v.Label, red("removed")))
		default:
			lines = append(lines, fmt.Sprintf("  %s %-24s %s → %s", dim("·"), v.Label, dim(fromVer), cyan(toVer)))
		}
	}

	if len(lines) > 0 {
		fmt.Println()
		fmt.Println(strings.Join(lines, "\n"))
	}
}

func shortVer(s string) string {
	if s == "" {
		return "none"
	}
	if strings.HasPrefix(s, "go version go") {
		parts := strings.Fields(strings.TrimPrefix(s, "go version go"))
		if len(parts) > 0 {
			return parts[0]
		}
	}
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return s
	}
	f := strings.Trim(fields[0], `"'()`)
	if len(f) > 0 && ((f[0] >= '0' && f[0] <= '9') || f[0] == 'v') {
		return f
	}
	if len(fields) > 1 {
		return strings.Trim(fields[1], `"'()`)
	}
	return f
}
