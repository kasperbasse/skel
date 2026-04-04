package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kasperbasse/skel/cmd/tui"
	"github.com/kasperbasse/skel/internal/profile"
	"github.com/kasperbasse/skel/internal/scanner"
)

var importCmd = &cobra.Command{
	Use:   "import [file]",
	Short: "Import a profile from an exported JSON file",
	Args:  requireArgs("import <file>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("\n  %s Importing profile...\n", cyan(headlineIcon("import")))
		fmt.Printf("  %s\n\n", dividerStyle.Render("────────────────────────────────────────────"))

		p, err := loadImportedProfile(args[0])
		if err != nil {
			return err
		}

		if _, err := profile.Save(p); err != nil {
			return err
		}

		fmt.Printf("\n  %s Imported profile %s\n", iconCheck(), bold("'"+p.Name+"'"))
		fmt.Printf("  %s\n", dividerStyle.Render("────────────────────────────────────────────"))
		fmt.Printf("  %s\n\n", dim(fmt.Sprintf("%s formulas · %s casks · originally saved from %s", num(len(p.Homebrew.Formulas)), num(len(p.Homebrew.Casks)), p.Machine)))

		printImportSecurityNotice(p)

		printNextSteps(
			nextStep("skel show "+p.Name, "to review the profile"),
			nextStep("skel restore "+p.Name, "to apply this setup"),
		)

		return nil
	},
}

func loadImportedProfile(path string) (*profile.Profile, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}
	if fi.Size() > profile.MaxImportSize {
		return nil, fmt.Errorf("profile file too large (%d bytes, max %d)", fi.Size(), profile.MaxImportSize)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}

	var p profile.Profile
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("invalid profile file: %w", err)
	}

	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("unsafe profile: %w", err)
	}

	return &p, nil
}

func printImportSecurityNotice(p *profile.Profile) {
	warnings := collectImportWarnings(p)
	if len(warnings) == 0 {
		return
	}

	warningText := fmt.Sprintf("This profile contains: %s\n\n", strings.Join(warnings, ", ")) +
		"Review the profile with " + cyan("skel show "+p.Name) + " before restoring.\n" +
		"These files execute code when your shell starts or git runs."
	printWarningBox("Security Notice", warningText)
}

var deleteCmd = &cobra.Command{
	Use:   "delete [profile-name]",
	Short: "Delete a saved profile",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := "default"
		if len(args) > 0 {
			name = args[0]
		}

		_, err := profile.Load(name)
		if err != nil {
			return enhanceError(err)
		}

		fmt.Printf("\n  %s Deleting profile %s\n", cyan(headlineIcon("delete")), bold("'"+name+"'"))
		fmt.Printf("  %s\n\n", dividerStyle.Render("────────────────────────────────────────────"))

		ok, err := tui.Confirm(fmt.Sprintf("Are you sure you want to delete %q?", name))
		if err != nil {
			return err
		}

		if !ok {
			fmt.Printf("  %s Delete canceled - Profile kept safe\n\n", iconDash())
			return nil
		}

		if err := profile.Delete(name); err != nil {
			return err
		}

		fmt.Printf("  %s Profile deleted\n", iconCheck())
		return nil
	},
}

var updateCmd = &cobra.Command{
	Use:   "update [profile-name]",
	Short: "Re-scan your Mac and update an existing profile",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := "default"
		if len(args) > 0 {
			name = args[0]
		}

		old, _ := profile.Load(name) // best-effort, nil if it doesn't exist yet

		fmt.Printf("\n  %s Updating profile %s...\n", cyan(headlineIcon("update")), bold("'"+name+"'"))
		fmt.Printf("  %s\n", dividerStyle.Render("────────────────────────────────────────────"))

		spin := NewSpinner("Re-scanning your environment...")
		spin.Start()

		p, warnings, err := scanner.Run(name)
		spin.Stop()

		if err != nil {
			return err
		}

		if len(warnings) > 0 {
			fmt.Println()
			for _, w := range warnings {
				fmt.Printf("  %s %s\n", iconWarn(), dim(w))
			}
		}

		if _, err := profile.Save(p); err != nil {
			printErr("  %s Failed to save profile: %v\n", iconCross(), err)
			return err
		}

		if old != nil {
			printUpdateDiff(old, p)
		}

		fmt.Printf("  %s Profile %s updated\n", iconCheck(), bold("'"+name+"'"))
		return nil
	},
}

func printUpdateDiff(old, updated *profile.Profile) {
	var lines []string

	// List section diffs — skip if count unchanged.
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

	// Version string diffs.
	for _, v := range versionFields {
		fromVer := shortVer(v.Value(old))
		toVer := shortVer(v.Value(updated))
		if fromVer == toVer {
			continue
		}
		switch {
		case fromVer == "none":
			lines = append(lines, fmt.Sprintf("  %s %-24s %s",
				dim("·"), v.Label, green(toVer)))
		case toVer == "none":
			lines = append(lines, fmt.Sprintf("  %s %-24s %s",
				dim("·"), v.Label, red("removed")))
		default:
			lines = append(lines, fmt.Sprintf("  %s %-24s %s → %s",
				dim("·"), v.Label, dim(fromVer), cyan(toVer)))
		}
	}

	if len(lines) > 0 {
		fmt.Println()
		fmt.Println(strings.Join(lines, "\n"))
	}
	// No output when nothing changed — the \n before the success message is enough.
}

// shortVer extracts a clean version token from a raw version string.
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
	if len(f) > 0 && (f[0] >= '0' && f[0] <= '9' || f[0] == 'v') {
		return f
	}
	if len(fields) > 1 {
		return strings.Trim(fields[1], `"'()`)
	}
	return f
}
