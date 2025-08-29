package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/klauern/klauer-hooks/internal/config"
	"github.com/spf13/cobra"
)

func NewInstallCmd(getPlugin func(string) (interface {
	Run() error
	Description() string
}, bool), pluginKeys func() []string, isValidEventType func(string) bool, validEventTypes func() []string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install [hook-type] [options]",
		Short: "Install a hook type into Claude Code settings",
		Long: `Install a hook type into your Claude Code settings.json file.
This will automatically configure the hook to run for the specified events.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			hookType := args[0]

			// Validate plugin exists
			if _, exists := getPlugin(hookType); !exists {
				fmt.Fprintf(os.Stderr, "Error: Plugin '%s' not found.\n", hookType)
				fmt.Fprintf(os.Stderr, "Available plugins: %s\n", strings.Join(pluginKeys(), ", "))
				os.Exit(1)
			}

			// Get flags
			global, _ := cmd.Flags().GetBool("global")
			event, _ := cmd.Flags().GetString("event")
			matcher, _ := cmd.Flags().GetString("matcher")
			timeoutFlag, _ := cmd.Flags().GetInt("timeout")
			logEnabled, _ := cmd.Flags().GetBool("log")
			logFormat, _ := cmd.Flags().GetString("log-format")
			if logFormat == "" {
				logFormat = config.LoggingFormatJSONL
			}
			if logEnabled && !config.IsValidLoggingFormat(logFormat) {
				fmt.Fprintf(os.Stderr, "Error: Invalid --log-format '%s'. Valid: jsonl, pretty\n", logFormat)
				os.Exit(1)
			}

			// Validate event
			if !isValidEventType(event) {
				fmt.Fprintf(os.Stderr, "Error: Invalid event '%s'.\n", event)
				fmt.Fprintf(os.Stderr, "Valid events: %s\n", strings.Join(validEventTypes(), ", "))
				fmt.Fprintf(os.Stderr, "Use 'hooks list-events' to see all available events with descriptions.\n")
				os.Exit(1)
			}

			// Get path to this executable
			execPath, err := os.Executable()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to get executable path: %v\n", err)
				os.Exit(1)
			}

			// Create command: hooks run <type>
			hookCommand := fmt.Sprintf("%s run %s", execPath, hookType)
			if logEnabled {
				hookCommand += " --log"
				if logFormat != config.LoggingFormatJSONL {
					hookCommand += fmt.Sprintf(" --log-format %s", logFormat)
				}
			}

			// Get settings path
			settingsPath, err := config.GetSettingsPath(global)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting settings path: %v\n", err)
				os.Exit(1)
			}

			// Load existing settings
			settings, err := config.LoadSettings(settingsPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading settings: %v\n", err)
				os.Exit(1)
			}

			// Add hook to settings
			var timeout *int
			if timeoutFlag > 0 {
				timeout = &timeoutFlag
			}
			result := config.AddHookToSettings(settings, event, matcher, hookCommand, timeout)

			// Check for duplicates or replacements
			isDuplicateNoChange := false
			if result.WasDuplicate {
				if strings.Contains(result.DuplicateInfo, "Replaced existing") {
					fmt.Printf("🔄 %s\n", result.DuplicateInfo)
				} else {
					fmt.Printf("⚠️  Hook already installed: %s\n", result.DuplicateInfo)
					fmt.Printf("No changes made. The hook is already configured for this event.\n")
					isDuplicateNoChange = true
				}
			}

			// Save settings (only if not a duplicate with no changes)
			if !isDuplicateNoChange {
				if err := config.SaveSettings(settingsPath, settings); err != nil {
					fmt.Fprintf(os.Stderr, "Error saving settings: %v\n", err)
					os.Exit(1)
				}
			}

			scope := "project"
			if global {
				scope = "global"
			}

			// Only show installation success message if not a duplicate
			if !isDuplicateNoChange {
				fmt.Printf("✅ Successfully installed %s hook in %s settings\n", hookType, scope)
				fmt.Printf("   Event: %s\n", event)
				fmt.Printf("   Matcher: %s\n", matcher)
				fmt.Printf("   Command: %s\n", hookCommand)
				fmt.Printf("   Settings: %s\n", settingsPath)
				fmt.Println()
			}

			// Post-install actions for specific plugins (run even for duplicates)
			if hookType == "fetch-blocker" {
				createSampleBlockedUrlsFile(global)
			}

			// Only show the activation message if not a duplicate
			if !isDuplicateNoChange {
				fmt.Println("The hook will be active in new Claude Code sessions.")
				fmt.Println("Use 'claude /hooks' to verify the configuration.")
			}
		},
	}

	// Add flags for install command
	cmd.Flags().BoolP("global", "g", false, "Install to global settings (~/.claude/settings.json)")
	cmd.Flags().StringP("event", "e", "PreToolUse", "Hook event (PreToolUse, PostToolUse, UserPromptSubmit, etc.)")
	cmd.Flags().StringP("matcher", "m", "*", "Tool matcher pattern (* for all tools)")
	cmd.Flags().IntP("timeout", "t", 0, "Command timeout in seconds (0 for no timeout)")
	cmd.Flags().BoolP("log", "l", false, "Enable detailed logging to .claude/hooks/<plugin-key>.log")
	cmd.Flags().String("log-format", "jsonl", "Log output format: jsonl or pretty (default jsonl)")

	return cmd
}

func NewUninstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall [hook-type|all]",
		Short: "Remove a hook type from Claude Code settings",
		Long:  `Remove a hook type from your Claude Code settings.json file. Use 'all' to remove all klauer-hooks.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			hookType := args[0]
			global, _ := cmd.Flags().GetBool("global")

			// Handle 'all' case
			if hookType == "all" {
				uninstallAllKlauerHooks(global)
				return
			}

			// Get path to this executable
			execPath, err := os.Executable()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to get executable path: %v\n", err)
				os.Exit(1)
			}

			// Create command pattern to match: hooks run <type>
			hookCommand := fmt.Sprintf("%s run %s", execPath, hookType)

			// Get settings path
			settingsPath, err := config.GetSettingsPath(global)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting settings path: %v\n", err)
				os.Exit(1)
			}

			// Load existing settings
			settings, err := config.LoadSettings(settingsPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading settings: %v\n", err)
				os.Exit(1)
			}

			// Remove hook from settings
			removed := config.RemoveHookFromSettings(settings, hookCommand)

			if !removed {
				fmt.Printf("Hook type '%s' was not found in settings.\n", hookType)
				os.Exit(1)
			}

			// Save settings
			if err := config.SaveSettings(settingsPath, settings); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving settings: %v\n", err)
				os.Exit(1)
			}

			scope := "project"
			if global {
				scope = "global"
			}

			fmt.Printf("✅ Successfully removed %s hook from %s settings\n", hookType, scope)
			fmt.Printf("   Command: %s\n", hookCommand)
			fmt.Printf("   Settings: %s\n", settingsPath)
		},
	}

	cmd.Flags().BoolP("global", "g", false, "Remove from global settings (~/.claude/settings.json)")
	return cmd
}

func uninstallAllKlauerHooks(global bool) {
	// Get settings path
	settingsPath, err := config.GetSettingsPath(global)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Load existing settings
	settings, err := config.LoadSettings(settingsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading settings: %v\n", err)
		os.Exit(1)
	}

	scope := "project"
	if global {
		scope = "global"
	}

	// Count klauer-hooks before removal
	totalHooksBefore := config.CountKlauerHooksInSettings(settings)

	if totalHooksBefore == 0 {
		fmt.Printf("No klauer-hooks found in %s settings.\n", scope)
		return
	}

	// Show what will be removed
	fmt.Printf("Found %d klauer-hooks in %s settings:\n\n", totalHooksBefore, scope)
	config.PrintKlauerHooksToRemove(settings)

	// Confirmation prompt
	fmt.Printf("\nThis will remove ALL klauer-hooks from %s settings.\n", scope)
	fmt.Printf("Other hooks (not from klauer-hooks) will be preserved.\n")
	fmt.Printf("Continue? (y/N): ")

	var response string
	_, _ = fmt.Scanln(&response)
	if response != "y" && response != "Y" && response != "yes" {
		fmt.Println("Operation cancelled.")
		return
	}

	// Remove all klauer-hooks
	removed := config.RemoveAllKlauerHooksFromSettings(settings)

	if removed == 0 {
		fmt.Printf("No klauer-hooks were found to remove.\n")
		return
	}

	// Save settings
	if err := config.SaveSettings(settingsPath, settings); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving settings: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Successfully removed %d klauer-hooks from %s settings\n", removed, scope)
	fmt.Printf("   Settings: %s\n", settingsPath)

	globalFlag := ""
	if global {
		globalFlag = " --global"
	}
	fmt.Printf("\nUse 'hooks list-installed%s' to verify the changes.\n", globalFlag)
}

// createSampleBlockedUrlsFile creates a sample blocked-urls.txt file for the fetch-blocker hook
func createSampleBlockedUrlsFile(global bool) {
	// Determine the target directory
	var targetDir string
	var scope string

	if global {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("⚠️  Could not create sample blocked-urls.txt: %v\n", err)
			return
		}
		targetDir = filepath.Join(homeDir, ".claude")
		scope = "global"
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf("⚠️  Could not create sample blocked-urls.txt: %v\n", err)
			return
		}
		targetDir = filepath.Join(cwd, ".claude")
		scope = "project"
	}

	blockedUrlsPath := filepath.Join(targetDir, "blocked-urls.txt")

	// Check if file already exists
	if _, err := os.Stat(blockedUrlsPath); err == nil {
		fmt.Printf("📄 Sample blocked-urls.txt already exists: %s\n", blockedUrlsPath)
		return
	}

	// Ensure the .claude directory exists
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		fmt.Printf("⚠️  Could not create .claude directory: %v\n", err)
		return
	}

	// Create the sample file content
	sampleContent := `# Blocked URL prefixes for fetch-blocker hook
# Format: prefix|suggestion (suggestion is optional)
# Lines starting with # are comments

# Private GitHub repos (use gh CLI instead)
https://github.com/*/*/private/*|Use 'gh api' or 'gh repo view' instead for private repositories
https://api.github.com/repos/*/*/contents/*|Use 'gh api' for authenticated GitHub API access

# Internal/VPN-only domains
# https://company.internal.com/*|This domain requires VPN access
# https://admin.internal/*|Admin panel requires authentication

# Auth-required API endpoints
# https://api.company.com/private/*|Private API endpoints require authentication tokens

# Example: Zendesk secure backup domain (from your use case)
https://atlantis-foundation-secure.backups.zendesk-dev.com/*|This domain requires VPN access and authentication

# Add your own blocked prefixes here...
# Format examples:
# https://exact-domain.com/path|Alternative suggestion
# https://example.com/*|Wildcard blocks all paths under domain
# *.internal.company.com/*|Wildcard subdomain pattern
`

	// Write the sample file
	if err := os.WriteFile(blockedUrlsPath, []byte(sampleContent), 0644); err != nil {
		fmt.Printf("⚠️  Could not create sample blocked-urls.txt: %v\n", err)
		return
	}

	fmt.Printf("📄 Created sample blocked-urls.txt (%s): %s\n", scope, blockedUrlsPath)
	fmt.Printf("   Edit this file to add your own blocked URL prefixes.\n")
}
