package cmd

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/kasperbasse/skel/internal/version"
)

var (
	// Using a subtle gray for the "bones" and a bright white/cyan for the main text
	boneStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	logoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("81")).Bold(true)
)

var rootCmd = &cobra.Command{
	SilenceUsage:  true,
	SilenceErrors: true,
	Use:           "skel",
	Short:         "💀 Save and restore your Mac developer setup",
	Version: fmt.Sprintf("%s (commit: %s) built: %s",
		version.Version,
		version.Commit,
		version.Date,
	),
}

var dividerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("237"))

// dividerLine is the standard horizontal rule rendered throughout the UI.
const dividerLine = "────────────────────────────────────────────"

func printBanner() {

	banner := `
   ____  _  _______ _
  / ___|| |/ / ____| |
  \___ \| ' /|  _| | |
   ___) | . \| |___| |___
  |____/|_|\_\_____|_____|`

	fmt.Println(logoStyle.Render(banner))
	fmt.Printf("\n  %s\n", boneStyle.Render(bold("Save your Mac dev setup. Restore it anywhere. In minutes.")))
	fmt.Printf("  %s\n", dividerStyle.Render(dividerLine))

	type cmdEntry struct{ name, desc string }
	type cmdGroup struct {
		title string
		cmds  []cmdEntry
	}
	groups := []cmdGroup{
		{"Profile", []cmdEntry{
			{"scan", "Scan your Mac and save a setup profile"},
			{"restore", "Restore a saved profile on this Mac"},
			{"list", "List all saved profiles"},
			{"show", "Show the contents of a profile"},
			{"update", "Re-scan and update an existing profile"},
			{"delete", "Delete a saved profile"},
		}},
		{"Inspect", []cmdEntry{
			{"status", "One-line summary of a profile"},
			{"drift", "Detect what's changed since last scan"},
			{"diff", "Compare two profiles"},
			{"doctor", "Check that a profile can be restored here"},
		}},
		{"Share", []cmdEntry{
			{"export", "Export a profile to a shareable JSON file"},
			{"import", "Import a profile from a JSON file"},
			{"clone", "Clone a profile from a GitHub Gist"},
			{"publish", "Publish a profile as a GitHub Gist"},
			{"brewfile", "Import and export Brewfiles"},
		}},
	}
	for _, g := range groups {
		fmt.Println()
		fmt.Printf("  %s\n", boneStyle.Render(bold(g.title)))
		for _, c := range g.cmds {
			cmdLabel := fmt.Sprintf("%-10s", c.name)
			fmt.Printf("    %s  %s\n", green(cmdLabel), c.desc)
		}
	}
}

const cursorShow = "\033[?25h"

func Execute() {
	// Always restore cursor visibility on exit. Covers panics, signals, and
	// normal exits where the spinner may have hidden the cursor.
	defer func() { _, _ = fmt.Fprint(os.Stdout, cursorShow) }()

	// Restore cursor if the process is interrupted or terminated.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		_, _ = fmt.Fprint(os.Stdout, cursorShow)
		os.Exit(130)
	}()

	if err := rootCmd.Execute(); err != nil {
		printCLIError(rootCmd.ErrOrStderr(), err)
		os.Exit(1)
	}
}

func printCLIError(w io.Writer, err error) {
	if err == nil {
		return
	}
	_, _ = fmt.Fprintf(w, "\nError: %v\n", err)
}

func init() {
	rootCmd.SetHelpTemplate(prettyHelpTemplate)
	rootCmd.SetUsageTemplate(prettyHelpTemplate)

	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(restoreCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(driftCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(brewfileCmd)
	rootCmd.AddCommand(cloneCmd)
	rootCmd.AddCommand(publishCmd)

	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})

	defaultHelp := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		if cmd == rootCmd {
			printBanner()
		} else {
			defaultHelp(cmd, args)
		}
	})

	rootCmd.Run = func(_ *cobra.Command, _ []string) {
		printBanner()
	}
}
