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
	RunE: runRestore,
}

func runRestore(_ *cobra.Command, args []string) error {
	name := SelectProfileName(args)

	p, err := loadProfileOrDefault(name)
	if err != nil {
		return err
	}

	opts, err := parseOnlyFlag(onlyStr)
	if err != nil {
		return enhanceError(err)
	}

	return enhanceError(executeRestore(p, opts, dryRun))
}

// loadProfileOrDefault loads a profile by name, or shows first-run guidance if no profiles exist.
func loadProfileOrDefault(name string) (*profile.Profile, error) {
	p, err := profile.Load(name)
	if err == nil {
		return p, nil
	}

	// Profile not found - check if this is a new user
	all, listErr := profile.ListAll()
	if listErr == nil && len(all) == 0 {
		printFirstRun()
		return nil, errSilentExit
	}

	return nil, enhanceError(err)
}

// executeRestore orchestrates the restore flow: header → validation → checks → execution.
func executeRestore(p *profile.Profile, opts *restore.Options, dryRunMode bool) error {
	printRestoreHeader(p)

	if !hasRestorableData(p, opts) {
		fmt.Printf("  %s Nothing to restore from this profile.\n\n", iconDash())
		return nil
	}

	if err := checkToolRequirements(p, opts, dryRunMode); err != nil {
		return err
	}

	if dryRunMode {
		printDryRunPreview(p, opts)
		return nil
	}

	return runRestoreExecution(p, opts)
}

// printRestoreHeader shows the banner and metadata.
func printRestoreHeader(p *profile.Profile) {
	fmt.Printf("\n  %s Restoring profile %s\n", cyan(headlineIcon("restore")), bold("'"+p.Name+"'"))
	fmt.Printf("  %s\n", dividerStyle.Render(dividerLine))
	fmt.Printf("  %s  %s\n\n", dim(fmt.Sprintf("Saved %s from %s", p.CreatedAt.Format(dateFormat), p.Machine)), dim(randomMessage(restoreStartMsgs)))
}

// printDryRunPreview shows what would be restored without making changes.
func printDryRunPreview(p *profile.Profile, opts *restore.Options) {
	fmt.Printf("  %s Dry run - nothing will be installed\n\n", iconWarn())
	printDryRun(p, opts)
}

// runRestoreExecution handles either interactive or non-interactive restore.
func runRestoreExecution(p *profile.Profile, opts *restore.Options) error {
	fmt.Printf("\n  %s\n", dividerStyle.Render(dividerLine))

	if IsInteractive() {
		updatedOpts, proceed, err := selectRestoreOptions(p, opts)
		if err != nil {
			return err
		}
		if !proceed {
			return nil
		}
		return runInteractiveRestore(p, updatedOpts)
	}

	runNonInteractiveRestore(p, opts)
	return nil
}

// selectRestoreOptions prompts the user to select which sections to restore.
// Returns (options, shouldProceed, error).
func selectRestoreOptions(p *profile.Profile, opts *restore.Options) (*restore.Options, bool, error) {
	// If --only was specified, skip interactive selection and use those sections
	if onlyStr != "" {
		return opts, true, nil
	}

	selectItems := buildSelectRestoreChecklistItems(p)
	if len(selectItems) == 0 {
		return opts, true, nil
	}

	// Run TUI selection
	model := tui.NewSelectRestoreModel(selectItems)
	program := tea.NewProgram(model)
	result, err := program.Run()
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
		printRestoreStepResult(r)
		if !r.Success {
			failed = append(failed, r)
		}
	})

	printRestoreCompletion(p, failed)
}

// printRestoreStepResult prints a single restore step's outcome.
func printRestoreStepResult(r restore.Result) {
	progress := dim(fmt.Sprintf("[%d/%d]", r.Index, r.Total))
	if r.Success {
		if r.Message == restore.MsgAlreadyInstalled {
			fmt.Printf("  %s %s %s  %s\n", progress, iconCheck(), r.Step, dim(restore.MsgAlreadyInstalled))
		} else {
			fmt.Printf("  %s %s %s\n", progress, iconCheck(), r.Step)
		}
	} else {
		fmt.Printf("  %s %s %s  %s\n", progress, iconCross(), r.Step, dim(r.Message))
	}
}

// printRestoreCompletion shows completion status and next steps.
func printRestoreCompletion(p *profile.Profile, failed []restore.Result) {
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
	restoreCmd.Flags().StringVar(&onlyStr, "only", "", "Restore only specific sections (comma-separated: "+strings.Join(validSections, ",")+")")
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

// buildSelectRestoreChecklistItems transforms profile data into interactive selection rows,
// accounting for missing tools that would block restores.
func buildSelectRestoreChecklistItems(p *profile.Profile) []tui.SelectItem {
	missingToolsBySection := appdoctor.BlockedSectionTools(p)

	var items []tui.SelectItem
	for _, g := range scanGroups {
		if len(g.RestoreKeys) == 0 {
			continue
		}

		summary := ""
		if g.ScanSummary != nil {
			summary = g.ScanSummary(p)
		}

		// Skip sections with no data
		if summary == "" {
			continue
		}

		blocked, missingTools := gatherMissingToolsForSection(g, missingToolsBySection)

		items = append(items, tui.SelectItem{
			Icon:         g.Icon,
			Label:        g.Label,
			Keys:         g.RestoreKeys,
			Summary:      summary,
			Selected:     !blocked,
			Blocked:      blocked,
			MissingTools: missingTools,
		})
	}
	return items
}

// gatherMissingToolsForSection collects all missing tools for a section and returns
// whether the section is blocked.
func gatherMissingToolsForSection(g scanGroup, missingBySection map[string][]string) (blocked bool, tools []string) {
	seenTools := make(map[string]struct{})

	for _, key := range g.RestoreKeys {
		toolsForSection := missingBySection[key]
		if len(toolsForSection) > 0 {
			blocked = true
		}

		for _, tool := range toolsForSection {
			if _, seen := seenTools[tool]; seen {
				continue
			}
			seenTools[tool] = struct{}{}
			tools = append(tools, tool)
		}
	}

	return blocked, tools
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

// hasRestorableData checks if there are any sections with restorable data,
// respecting the --only filter if specified.
func hasRestorableData(p *profile.Profile, opts *restore.Options) bool {
	for _, g := range scanGroups {
		if g.ScanSummary == nil || len(g.RestoreKeys) == 0 {
			continue
		}

		if summary := g.ScanSummary(p); summary != "" {
			// If --only filter is active, only count matching sections
			if len(opts.Sections) > 0 {
				for _, key := range g.RestoreKeys {
					if opts.Sections[key] {
						return true
					}
				}
			} else {
				return true
			}
		}
	}
	return false
}

// checkToolRequirements validates that all required tools are available and prints
// the results. This is informational only — it never blocks execution.
func checkToolRequirements(p *profile.Profile, opts *restore.Options, dryRunMode bool) error {
	requiredTools := appdoctor.RequiredToolsForSections(p, func(section string) bool {
		if len(opts.Sections) > 0 {
			return opts.Sections[section]
		}
		return true
	})

	if len(requiredTools) == 0 {
		fmt.Printf("  %s No external tool requirements\n\n", iconDot())
		return nil
	}

	fmt.Printf("  %s Checking requirements\n\n", iconDot())

	issues, _ := appdoctor.RunChecks(requiredTools)
	if issues > 0 && !dryRunMode {
		printMissingToolsWarning(issues, len(opts.Sections) > 0)
	}

	return nil
}

// printMissingToolsWarning shows which tools need to be installed.
func printMissingToolsWarning(count int, scoped bool) {
	pronoun := "it"
	if count > 1 {
		pronoun = "them"
	}
	scope := "all sections"
	if scoped {
		scope = "selected sections"
	}
	fmt.Printf("\n  %s %s missing — install %s to unlock %s\n",
		iconWarn(),
		bold(fmt.Sprintf("%d required tool%s", count, pluralS(count))),
		pronoun,
		scope,
	)
}
