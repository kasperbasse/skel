package cmd

import (
	"fmt"
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
	CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	SilenceUsage:      true,
	Use:               "skel",
	Short:             "📦 Save and restore your Mac developer setup",
	Version: fmt.Sprintf("%s (commit: %s) built: %s",
		version.Version,
		version.Commit,
		version.Date,
	),
}

func printBanner() {

	banner := `
   ____  _  _______ _      
  / ___|| |/ / ____| |     
  \___ \| ' /|  _| | |     
   ___) | . \| |___| |___  
  |____/|_|\_\_____|_____|`

	fmt.Println(logoStyle.Render(banner))
	fmt.Printf("\n  %s\n\n", boneStyle.Render(bold("Save your Mac dev setup. Restore it anywhere. In minutes.")))

	fmt.Printf("  %s\n", bold("Commands:"))
	cmds := []struct{ name, desc string }{
		{"scan", "Scan your Mac and save a setup profile"},
		{"restore", "Restore a saved profile on this Mac"},
		{"list", "List all saved profiles"},
		{"show", "Show the contents of a profile"},
		{"drift", "Detect what's changed since last scan"},
		{"diff", "Compare two profiles"},
		{"update", "Re-scan and update an existing profile"},
		{"export", "Export a profile to a shareable JSON file"},
		{"import", "Import a profile from a JSON file"},
		{"delete", "Delete a saved profile"},
		{"clone", "Clone a profile from a GitHub Gist"},
		{"publish", "Publish a profile as a GitHub Gist"},
		{"brewfile", "Import and export Brewfiles"},
	}
	for _, c := range cmds {
		// Pad after the colored name to align descriptions
		padding := 10 - len(c.name)
		fmt.Printf("    %s%*s%s\n", green(c.name), padding, "", c.desc)
	}

	fmt.Printf("\n  %s", boneStyle.Faint(true).Render("Version: "+version.Version))
	fmt.Println()
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
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(restoreCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(driftCmd)
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
