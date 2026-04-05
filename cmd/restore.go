package cmd

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/kasperbasse/skel/cmd/tui"
	appdoctor "github.com/kasperbasse/skel/internal/app/doctor"
	"github.com/kasperbasse/skel/internal/profile"
	"github.com/kasperbasse/skel/internal/restore"
)

var (
	dryRun  bool
	onlyStr string
)

var validSections = allRestoreKeys()

var restoreCmd = &cobra.Command{
	Use:   "restore [profile-name]",
	Short: "Restore a saved profile on this Mac",
	Long: fmt.Sprintf(
		"Restore a saved profile. Use --only to limit sections: %s",
		strings.Join(validSections, ", "),
	),
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := "default"
		if len(args) > 0 {
			name = args[0]
		}

		p, err := profile.Load(name)
		if err != nil {
			all, listErr := profile.ListAll()
			if listErr == nil && len(all) == 0 {
				printFirstRun()
				return nil
			}
			return enhanceError(err)
		}

		opts, err := parseOnlyFlag(onlyStr)
		if err != nil {
			return enhanceError(err)
		}

		// Restoring profile
		fmt.Printf("\n  %s Restoring profile %s\n", cyan(headlineIcon("restore")), bold("'"+p.Name+"'"))
		fmt.Printf("  %s\n", dividerStyle.Render("────────────────────────────────────────────"))
		fmt.Printf("  %s · %s\n\n", dim(fmt.Sprintf("Saved %s from %s", p.CreatedAt.Format("Jan 02 2006"), p.Machine)), dim(randomMessage(restoreStartMsgs)))

		// Checking with Doctor if everything looks good to continue with restore
		checks := appdoctor.BuildChecks(p)
		if len(checks) == 0 {
			fmt.Printf("  %s Nothing to restore from this profile.\n\n", iconDash())
			return nil
		}

		fmt.Printf("  %s Checking requirements\n\n", iconDot())

		if dryRun {
			fmt.Printf("  %s Dry run - nothing will be installed\n\n", iconWarn())
			printDryRun(p, opts)
			return nil
		}

		issues,_ := appdoctor.RunChecks(p)
		if issues > 0 {
			fmt.Printf("\n  %s %s not available. Restore paused.\n\n",
				iconWarn(),
				bold(fmt.Sprintf("%d required tool%s", issues, pluralS(issues))),
			)
			printNextSteps(
				nextStep("skel doctor "+name, "to verify requirements"),
				nextStep("skel restore "+name, "to retry after fixing"),
			)
			return nil
		}

		fmt.Printf("\n  %s\n", dividerStyle.Render("────────────────────────────────────────────"))

		if IsInteractive() {
			updatedOpts, proceed, err := selectRestoreOptions(p, opts)
			if err != nil {
				return err
			}
			if !proceed {
				return nil
			}
			if err := runInteractiveRestore(p, updatedOpts); err != nil {
				return err
			}
		} else {
			runNonInteractiveRestore(p, opts)
		}

		return nil
	},
}

func selectRestoreOptions(p *profile.Profile, opts *restore.Options) (*restore.Options, bool, error) {
	if onlyStr != "" {
		return opts, true, nil
	}

	selectItems := buildSelectItems(p)
	if len(selectItems) == 0 {
		return opts, true, nil
	}

	sm := tui.NewSelectRestoreModel(selectItems)
	selectProg := tea.NewProgram(sm)
	result, err := selectProg.Run()
	if err != nil {
		return nil, false, fmt.Errorf("selection failed: %w", err)
	}
	final := result.(tui.SelectRestoreModel)
	if !final.Confirmed() {
		fmt.Printf("  %s Canceled.\n\n", iconDash())
		return nil, false, nil
	}

	return &restore.Options{Sections: final.SelectedKeys()}, true, nil
}

func runInteractiveRestore(p *profile.Profile, opts *restore.Options) error {
	m := tui.NewRestoreModel(p, opts, randomMessage(restoreStartMsgs))
	prog := tea.NewProgram(m)
	if _, err := prog.Run(); err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}
	return nil
}

func runNonInteractiveRestore(p *profile.Profile, opts *restore.Options) {
	var failed []restore.Result

	restore.Run(p, opts, func(r restore.Result) {
		progress := dim(fmt.Sprintf("[%d/%d]", r.Index, r.Total))
		if r.Success {
			if r.Message == "already installed" {
				fmt.Printf("  %s %s %s  %s\n", progress, iconCheck(), r.Step, dim("already installed"))
			} else {
				fmt.Printf("  %s %s %s\n", progress, iconCheck(), r.Step)
			}
		} else {
			fmt.Printf("  %s %s %s  %s\n", progress, iconCross(), r.Step, dim(r.Message))
			failed = append(failed, r)
		}
	})

	fmt.Println()
	if len(failed) == 0 {
		fmt.Printf("  %s %s\n", iconCheck(), randomMessage(restoreCompleteMsgs))
		printNextSteps(
			nextStep("Restart your shell", "to apply all changes"),
		)
		return
	}

	fmt.Printf("  %s Done with %s. Check the output above.\n",
		iconWarn(), red(fmt.Sprintf("%d errors", len(failed))))
	printNextSteps(
		nextStep("skel restore "+p.Name, "to retry"),
	)
}

func init() {
	restoreCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview what would be restored without making changes")
	restoreCmd.Flags().StringVar(&onlyStr, "only", "", "Restore only specific sections (comma-separated: homebrew,shell,git,editors,configs,languages,mas)")
}

func parseOnlyFlag(s string) (*restore.Options, error) {
	if s == "" {
		return &restore.Options{}, nil
	}

	valid := make(map[string]bool)
	for _, v := range validSections {
		valid[v] = true
	}

	sections := make(map[string]bool)
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(strings.ToLower(part))
		if part == "" {
			continue
		}
		if !valid[part] {
			return nil, fmt.Errorf("unknown section %q, valid: %s", part, strings.Join(validSections, ", "))
		}
		sections[part] = true
	}

	return &restore.Options{Sections: sections}, nil
}

// buildSelectItems creates checklist items from scanGroups for the given profile.
// Only includes sections that have restorable data.
func buildSelectItems(p *profile.Profile) []tui.SelectItem {
	var items []tui.SelectItem
	for _, g := range scanGroups {
		if len(g.RestoreKeys) == 0 {
			continue
		}
		summary := ""
		if g.ScanSummary != nil {
			summary = g.ScanSummary(p)
		}
		if summary == "" {
			continue // no data for this section
		}
		items = append(items, tui.SelectItem{
			Icon:     g.Icon,
			Label:    g.Label,
			Keys:     g.RestoreKeys,
			Summary:  summary,
			Selected: true,
		})
	}
	return items
}

func printDryRun(p *profile.Profile, opts *restore.Options) {
	for _, g := range scanGroups {
		if g.DryRun == nil {
			continue
		}
		g.DryRun(p, opts)
	}
	fmt.Println()
}
