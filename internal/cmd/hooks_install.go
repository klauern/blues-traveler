package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/klauern/blues-traveler/internal/config"
	"github.com/urfave/cli/v3"
)

// newHooksInstallCommand creates the install command
func newHooksInstallCommand(getPlugin func(string) (interface {
	Run() error
	Description() string
}, bool), pluginKeys func() []string, isValidEventType func(string) bool, validEventTypes func() []string,
) *cli.Command {
	return &cli.Command{
		Name:      "install",
		Usage:     "Install a hook type into Claude Code settings",
		ArgsUsage: "[hook-type]",
		Description: `Install a hook type into your Claude Code settings.json file.
This will automatically configure the hook to run for the specified events.`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "global",
				Aliases: []string{"g"},
				Value:   false,
				Usage:   "Install to global settings (~/.claude/settings.json)",
			},
			&cli.StringFlag{
				Name:    "event",
				Aliases: []string{"e"},
				Value:   "PreToolUse",
				Usage:   "Hook event (PreToolUse, PostToolUse, UserPromptSubmit, etc.)",
			},
			&cli.StringFlag{
				Name:    "matcher",
				Aliases: []string{"m"},
				Value:   "*",
				Usage:   "Tool matcher pattern (* for all tools)",
			},
			&cli.IntFlag{
				Name:    "timeout",
				Aliases: []string{"t"},
				Value:   0,
				Usage:   "Command timeout in seconds (0 for no timeout)",
			},
			&cli.BoolFlag{
				Name:    "log",
				Aliases: []string{"l"},
				Value:   false,
				Usage:   "Enable detailed logging to .claude/hooks/<plugin-key>.log",
			},
			&cli.StringFlag{
				Name:  "log-format",
				Value: "jsonl",
				Usage: "Log output format: jsonl or pretty (default jsonl)",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			args := cmd.Args().Slice()
			if len(args) != 1 {
				return fmt.Errorf("exactly one argument required: [hook-type]")
			}
			hookType := args[0]

			// Validate plugin exists
			if _, exists := getPlugin(hookType); !exists {
				return fmt.Errorf("plugin '%s' not found.\nAvailable plugins: %s", hookType, strings.Join(pluginKeys(), ", "))
			}

			// Get flags
			global := cmd.Bool("global")
			event := cmd.String("event")
			matcher := cmd.String("matcher")
			timeoutFlag := cmd.Int("timeout")
			logEnabled := cmd.Bool("log")
			logFormat := cmd.String("log-format")
			if logFormat == "" {
				logFormat = config.LoggingFormatJSONL
			}
			if logEnabled && !config.IsValidLoggingFormat(logFormat) {
				return fmt.Errorf("invalid --log-format '%s'. Valid: jsonl, pretty", logFormat)
			}

			// Validate event
			if !isValidEventType(event) {
				return fmt.Errorf("invalid event '%s'.\nValid events: %s\nUse 'hooks list --events' to see all available events with descriptions", event, strings.Join(validEventTypes(), ", "))
			}

			// Get path to this executable
			execPath, err := os.Executable()
			if err != nil {
				return fmt.Errorf("failed to get executable path: %v", err)
			}

			// Create command: blues-traveler hooks run <type>
			hookCommand := fmt.Sprintf("%s hooks run %s", execPath, hookType)
			if logEnabled {
				hookCommand += " --log"
				if logFormat != config.LoggingFormatJSONL {
					hookCommand += fmt.Sprintf(" --log-format %s", logFormat)
				}
			}

			// Get settings path
			settingsPath, err := config.GetSettingsPath(global)
			if err != nil {
				scope := ScopeProject
				if global {
					scope = ScopeGlobal
				}
				return fmt.Errorf("failed to locate %s settings path: %w\n  Suggestion: Run 'blues-traveler hooks init' to initialize the project", scope, err)
			}

			// Load existing settings
			settings, err := config.LoadSettings(settingsPath)
			if err != nil {
				return fmt.Errorf("failed to load settings from %s: %w\n  Suggestion: Verify the settings file format is valid YAML/JSON", settingsPath, err)
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
					return fmt.Errorf("failed to save settings to %s: %w\n  Suggestion: Check file permissions and disk space", settingsPath, err)
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
			return nil
		},
	}
}

// newHooksUninstallCommand creates the uninstall command
func newHooksUninstallCommand() *cli.Command {
	return &cli.Command{
		Name:        "uninstall",
		Usage:       "Remove a hook type from Claude Code settings",
		ArgsUsage:   "[hook-type|all]",
		Description: `Remove a hook type from your Claude Code settings.json file. Use 'all' to remove all blues-traveler hooks.`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "global",
				Aliases: []string{"g"},
				Value:   false,
				Usage:   "Remove from global settings (~/.claude/settings.json)",
			},
			&cli.BoolFlag{
				Name:    "yes",
				Aliases: []string{"y"},
				Value:   false,
				Usage:   "Skip interactive confirmation for 'uninstall all'",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			args := cmd.Args().Slice()
			if len(args) != 1 {
				return fmt.Errorf("exactly one argument required: [hook-type|all]")
			}
			hookType := args[0]
			global := cmd.Bool("global")

			// Handle 'all' case
			if hookType == "all" {
				uninstallAllKlauerHooks(global, cmd.Bool("yes"))
				return nil
			}

			// Get path to this executable
			execPath, err := os.Executable()
			if err != nil {
				return fmt.Errorf("failed to get executable path: %v", err)
			}

			// Create command pattern to match: blues-traveler hooks run <type>
			hookCommand := fmt.Sprintf("%s hooks run %s", execPath, hookType)

			// Get settings path
			settingsPath, err := config.GetSettingsPath(global)
			if err != nil {
				scope := ScopeProject
				if global {
					scope = ScopeGlobal
				}
				return fmt.Errorf("failed to locate %s settings path: %w\n  Suggestion: Run 'blues-traveler hooks init' to initialize the project", scope, err)
			}

			// Load existing settings
			settings, err := config.LoadSettings(settingsPath)
			if err != nil {
				return fmt.Errorf("failed to load settings from %s: %w\n  Suggestion: Verify the settings file format is valid YAML/JSON", settingsPath, err)
			}

			// Remove hook from settings
			removed := config.RemoveHookFromSettings(settings, hookCommand)

			if !removed {
				return fmt.Errorf("hook type '%s' was not found in settings", hookType)
			}

			// Save settings
			if err := config.SaveSettings(settingsPath, settings); err != nil {
				return fmt.Errorf("error saving settings: %v", err)
			}

			scope := "project"
			if global {
				scope = "global"
			}

			fmt.Printf("✅ Successfully removed %s hook from %s settings\n", hookType, scope)
			fmt.Printf("   Command: %s\n", hookCommand)
			fmt.Printf("   Settings: %s\n", settingsPath)
			return nil
		},
	}
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
	if err := os.MkdirAll(targetDir, 0o750); err != nil {
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

# Example: Company secure backup domain
# https://secure-backups.company-dev.com/*|This domain requires VPN access and authentication

# Add your own blocked prefixes here...
# Format examples:
# https://exact-domain.com/path|Alternative suggestion
# https://example.com/*|Wildcard blocks all paths under domain
# *.internal.company.com/*|Wildcard subdomain pattern
`

	// Write the sample file
	if err := os.WriteFile(blockedUrlsPath, []byte(sampleContent), 0o600); err != nil {
		fmt.Printf("⚠️  Could not create sample blocked-urls.txt: %v\n", err)
		return
	}

	fmt.Printf("📄 Created sample blocked-urls.txt (%s): %s\n", scope, blockedUrlsPath)
	fmt.Printf("   Edit this file to add your own blocked URL prefixes.\n")
}

// uninstallAllKlauerHooks removes all blues-traveler hooks from settings
func uninstallAllKlauerHooks(global bool, skipConfirmation bool) {
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

	// Count blues-traveler hooks before removal
	totalHooksBefore := config.CountBluesTravelerInSettings(settings)

	if totalHooksBefore == 0 {
		fmt.Printf("No blues-traveler hooks found in %s settings.\n", scope)
		return
	}

	// Show what will be removed
	fmt.Printf("Found %d blues-traveler hooks in %s settings:\n\n", totalHooksBefore, scope)
	config.PrintBluesTravelerToRemove(settings)

	// Confirmation prompt
	fmt.Printf("\nThis will remove ALL blues-traveler hooks from %s settings.\n", scope)
	fmt.Printf("Other hooks (not from blues-traveler) will be preserved.\n")

	if !skipConfirmation {
		fmt.Printf("Continue? (y/N): ")
		var response string
		_, _ = fmt.Scanln(&response)
		if response != "y" && response != "Y" && response != "yes" {
			fmt.Println("Operation cancelled.")
			return
		}
	}

	// Remove all blues-traveler hooks
	removed := config.RemoveAllBluesTravelerFromSettings(settings)

	if removed == 0 {
		fmt.Printf("No blues-traveler hooks were found to remove.\n")
		return
	}

	// Save settings
	if err := config.SaveSettings(settingsPath, settings); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving settings: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Successfully removed %d blues-traveler hooks from %s settings\n", removed, scope)
	fmt.Printf("   Settings: %s\n", settingsPath)

	globalFlag := ""
	if global {
		globalFlag = " --global"
	}
	fmt.Printf("\nUse 'hooks list --installed%s' to verify the changes.\n", globalFlag)
}
