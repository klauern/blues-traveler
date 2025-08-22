package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "hooks",
	Short: "Claude Code hook runner and manager",
	Long:  `A CLI tool that runs Claude Code hooks directly and manages hook installations.`,
}

func init() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(listHooksCmd)

	// Add flags for install command
	installCmd.Flags().BoolP("global", "g", false, "Install to global settings (~/.claude/settings.json)")
	installCmd.Flags().StringP("event", "e", "PreToolUse", "Hook event (PreToolUse, PostToolUse, UserPromptSubmit, etc.)")
	installCmd.Flags().StringP("matcher", "m", "*", "Tool matcher pattern (* for all tools)")
	installCmd.Flags().IntP("timeout", "t", 0, "Command timeout in seconds (0 for no timeout)")

	// Add flags for uninstall command
	uninstallCmd.Flags().BoolP("global", "g", false, "Remove from global settings (~/.claude/settings.json)")

	// Add flags for list-installed command
	listHooksCmd.Flags().BoolP("global", "g", false, "Show global settings (~/.claude/settings.json)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
