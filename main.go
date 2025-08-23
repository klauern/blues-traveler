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
	rootCmd.AddCommand(listEventsCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(configLogCmd)

	// Add flags for install command
	installCmd.Flags().BoolP("global", "g", false, "Install to global settings (~/.claude/settings.json)")
	installCmd.Flags().StringP("event", "e", "PreToolUse", "Hook event (PreToolUse, PostToolUse, UserPromptSubmit, etc.)")
	installCmd.Flags().StringP("matcher", "m", "*", "Tool matcher pattern (* for all tools)")
	installCmd.Flags().IntP("timeout", "t", 0, "Command timeout in seconds (0 for no timeout)")
	installCmd.Flags().BoolP("log", "l", false, "Enable detailed logging to .claude/hooks/<plugin-key>.log")

	// Add flags for uninstall command
	uninstallCmd.Flags().BoolP("global", "g", false, "Remove from global settings (~/.claude/settings.json)")

	// Add flags for run command
	runCmd.Flags().BoolP("log", "l", false, "Enable detailed logging to .claude/hooks/<plugin-key>.log")

	// Add flags for list-installed command
	listHooksCmd.Flags().BoolP("global", "g", false, "Show global settings (~/.claude/settings.json)")

	// Add flags for generate command
	generateCmd.Flags().StringP("description", "d", "", "Description of the hook")
	generateCmd.Flags().StringP("type", "t", "both", "Hook type: pre, post, or both")
	generateCmd.Flags().BoolP("test", "", true, "Generate test file")
	generateCmd.Flags().StringP("output", "o", "", "Output directory (default: internal/hooks)")

	// Add flags for config-log command
	configLogCmd.Flags().BoolP("global", "g", false, "Configure global settings (~/.claude/settings.json)")
	configLogCmd.Flags().IntP("max-age", "a", 0, "Maximum age in days to retain log files (default: 30)")
	configLogCmd.Flags().IntP("max-size", "s", 0, "Maximum size in MB per log file before rotation (default: 10)")
	configLogCmd.Flags().IntP("max-backups", "b", 0, "Maximum number of backup files to retain (default: 5)")
	configLogCmd.Flags().BoolP("compress", "c", false, "Compress rotated log files")
	configLogCmd.Flags().BoolP("show", "", false, "Show current log rotation settings")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
