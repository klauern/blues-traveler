package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/klauern/blues-traveler/internal/config"
	"github.com/klauern/blues-traveler/internal/constants"
	"github.com/klauern/blues-traveler/internal/core"
	"github.com/urfave/cli/v3"
)

// installFlags holds the parsed command flags
type installFlags struct {
	global     bool
	event      string
	matcher    string
	timeout    int
	logEnabled bool
	logFormat  string
}

// parseInstallFlags extracts and validates flags from the command
func parseInstallFlags(cmd *cli.Command) (installFlags, error) {
	flags := installFlags{
		global:     cmd.Bool("global"),
		event:      cmd.String("event"),
		matcher:    cmd.String("matcher"),
		timeout:    cmd.Int("timeout"),
		logEnabled: cmd.Bool("log"),
		logFormat:  cmd.String("log-format"),
	}

	if flags.logFormat == "" {
		flags.logFormat = config.LoggingFormatJSONL
	}

	if flags.logEnabled && !config.IsValidLoggingFormat(flags.logFormat) {
		return flags, fmt.Errorf("invalid --log-format '%s'. Valid: jsonl, pretty", flags.logFormat)
	}

	return flags, nil
}

// buildInstallHookCommand constructs the hook command string for install
func buildInstallHookCommand(hookType string, flags installFlags) (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	// Quote the path when it contains spaces to handle paths like "/Program Files/app"
	// Commands are executed via bash -lc, so unquoted paths with spaces break
	if strings.ContainsRune(execPath, ' ') {
		execPath = `"` + execPath + `"`
	}

	hookCommand := fmt.Sprintf("%s hooks run %s", execPath, hookType)
	if flags.logEnabled {
		hookCommand += " --log"
		if flags.logFormat != config.LoggingFormatJSONL {
			hookCommand += fmt.Sprintf(" --log-format %s", flags.logFormat)
		}
	}

	return hookCommand, nil
}

// handleDuplicateHookResult processes duplicate detection results
func handleDuplicateHookResult(result config.MergeResult) bool {
	if !result.WasDuplicate {
		return false
	}

	if strings.Contains(result.DuplicateInfo, "Replaced existing") {
		fmt.Printf("🔄 %s\n", result.DuplicateInfo)
		return false
	}

	fmt.Printf("⚠️  Hook already installed: %s\n", result.DuplicateInfo)
	fmt.Printf("No changes made. The hook is already configured for this event.\n")
	return true
}

// printHookInstallSuccess displays success message
func printHookInstallSuccess(hookType, scope, event, matcher, hookCommand, settingsPath string) {
	fmt.Printf("✅ Successfully installed %s hook in %s settings\n", hookType, scope)
	fmt.Printf("   Event: %s\n", event)
	fmt.Printf("   Matcher: %s\n", matcher)
	fmt.Printf("   Command: %s\n", hookCommand)
	fmt.Printf("   Settings: %s\n", settingsPath)
	fmt.Println()
}

// resolveAndValidateEvent resolves event alias and validates it
func resolveAndValidateEvent(event string, isValidEventType func(string) bool, validEventTypes func() []string) (string, error) {
	resolvedEvent := core.ResolveEventAlias(event)
	if resolvedEvent == "" {
		resolvedEvent = event // Already canonical (or unknown)
	}

	if !isValidEventType(resolvedEvent) {
		return "", fmt.Errorf("invalid event '%s'.\nValid events: %s\nUse 'hooks list --events' to see all available events with descriptions", event, strings.Join(validEventTypes(), ", "))
	}

	return resolvedEvent, nil
}

// loadAndValidateSettings loads settings from the specified path
func loadAndValidateSettings(settingsPath string) (*config.Settings, error) {
	settings, err := config.LoadSettings(settingsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load settings from %s: %w\n  Suggestion: Verify the settings file format is valid JSON", settingsPath, err)
	}
	return settings, nil
}

// saveSettingsIfNeeded saves settings only if not a duplicate
func saveSettingsIfNeeded(settingsPath string, settings *config.Settings, isDuplicateNoChange bool) error {
	if !isDuplicateNoChange {
		if err := config.SaveSettings(settingsPath, settings); err != nil {
			return fmt.Errorf("failed to save settings to %s: %w\n  Suggestion: Check file permissions and disk space", settingsPath, err)
		}
	}
	return nil
}

// performPostInstallActions runs post-install actions for specific hook types
func performPostInstallActions(hookType string, global bool) {
	if hookType == "fetch-blocker" {
		createSampleBlockedUrlsFile(global)
	}
}

// showInstallSuccessMessages shows success messages if not a duplicate
func showInstallSuccessMessages(hookType string, scope string, flags installFlags, hookCommand string, settingsPath string, isDuplicateNoChange bool) {
	if !isDuplicateNoChange {
		printHookInstallSuccess(hookType, scope, flags.event, flags.matcher, hookCommand, settingsPath)
		fmt.Println("The hook will be active in new Claude Code sessions.")
		fmt.Println("Use 'claude /hooks' to verify the configuration.")
	}
}

// installHookAction performs the hook installation
func installHookAction(hookType string, flags installFlags, isValidEventType func(string) bool, validEventTypes func() []string) error {
	// Resolve and validate event
	resolvedEvent, err := resolveAndValidateEvent(flags.event, isValidEventType, validEventTypes)
	if err != nil {
		return err
	}
	flags.event = resolvedEvent

	// Build hook command
	hookCommand, err := buildInstallHookCommand(hookType, flags)
	if err != nil {
		return err
	}

	// Get settings path
	settingsPath, err := config.GetSettingsPath(flags.global)
	if err != nil {
		scope := ScopeProject
		if flags.global {
			scope = ScopeGlobal
		}
		return fmt.Errorf("failed to locate %s settings path: %w\n  Suggestion: Run 'blues-traveler hooks init' to initialize the project", scope, err)
	}

	// Load existing settings
	settings, err := loadAndValidateSettings(settingsPath)
	if err != nil {
		return err
	}

	// Add hook to settings
	var timeout *int
	if flags.timeout > 0 {
		timeout = &flags.timeout
	}
	result := config.AddHookToSettings(settings, flags.event, flags.matcher, hookCommand, timeout)

	// Check for duplicates or replacements
	isDuplicateNoChange := handleDuplicateHookResult(result)

	// Save settings (only if not a duplicate with no changes)
	if err := saveSettingsIfNeeded(settingsPath, settings, isDuplicateNoChange); err != nil {
		return err
	}

	scope := ScopeProject
	if flags.global {
		scope = ScopeGlobal
	}

	// Show success messages
	showInstallSuccessMessages(hookType, scope, flags, hookCommand, settingsPath, isDuplicateNoChange)

	// Post-install actions for specific plugins (run even for duplicates)
	performPostInstallActions(hookType, flags.global)

	return nil
}

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
		Action: func(_ context.Context, cmd *cli.Command) error {
			args := cmd.Args().Slice()
			if len(args) != 1 {
				return fmt.Errorf("exactly one argument required: [hook-type]")
			}
			hookType := args[0]

			// Validate plugin exists
			if _, exists := getPlugin(hookType); !exists {
				return fmt.Errorf("plugin '%s' not found.\nAvailable plugins: %s", hookType, strings.Join(pluginKeys(), ", "))
			}

			// Parse and validate flags
			flags, err := parseInstallFlags(cmd)
			if err != nil {
				return err
			}

			return installHookAction(hookType, flags, isValidEventType, validEventTypes)
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
		Action: func(_ context.Context, cmd *cli.Command) error {
			args := cmd.Args().Slice()
			if len(args) != 1 {
				return fmt.Errorf("exactly one argument required: [hook-type|all]")
			}
			hookType := args[0]
			global := cmd.Bool("global")

			// Handle 'all' case
			if hookType == "all" {
				return uninstallAllKlauerHooks(global, cmd.Bool("yes"))
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
				return fmt.Errorf("failed to load settings from %s: %w\n  Suggestion: Verify the settings file format is valid JSON", settingsPath, err)
			}

			// Remove hook from settings using pattern matching
			// This handles hooks installed with flags (--log, --format) or different executable paths
			removed := config.RemoveHookTypeFromSettings(settings, hookType)

			if !removed {
				return fmt.Errorf("hook type '%s' was not found in settings", hookType)
			}

			// Save settings
			if err := config.SaveSettings(settingsPath, settings); err != nil {
				return fmt.Errorf("error saving settings: %w", err)
			}

			scope := constants.ScopeProject
			if global {
				scope = constants.ScopeGlobal
			}

			fmt.Printf("✅ Successfully removed all '%s' hooks from %s settings\n", hookType, scope)
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
		scope = constants.ScopeGlobal
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf("⚠️  Could not create sample blocked-urls.txt: %v\n", err)
			return
		}
		targetDir = filepath.Join(cwd, ".claude")
		scope = constants.ScopeProject
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
func uninstallAllKlauerHooks(global bool, skipConfirmation bool) error {
	// Get settings path
	settingsPath, err := config.GetSettingsPath(global)
	if err != nil {
		return fmt.Errorf("failed to get settings path: %w", err)
	}

	// Load existing settings
	settings, err := config.LoadSettings(settingsPath)
	if err != nil {
		return fmt.Errorf("failed to load settings from %s: %w", settingsPath, err)
	}

	scope := constants.ScopeProject
	if global {
		scope = constants.ScopeGlobal
	}

	// Count blues-traveler hooks before removal
	totalHooksBefore := config.CountBluesTravelerInSettings(settings)

	if totalHooksBefore == 0 {
		fmt.Printf("No blues-traveler hooks found in %s settings.\n", scope)
		return nil
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
			return nil
		}
	}

	// Remove all blues-traveler hooks
	removed := config.RemoveAllBluesTravelerFromSettings(settings)

	if removed == 0 {
		fmt.Printf("No blues-traveler hooks were found to remove.\n")
		return nil
	}

	// Save settings
	if err := config.SaveSettings(settingsPath, settings); err != nil {
		return fmt.Errorf("failed to save settings to %s: %w", settingsPath, err)
	}

	fmt.Printf("✅ Successfully removed %d blues-traveler hooks from %s settings\n", removed, scope)
	fmt.Printf("   Settings: %s\n", settingsPath)

	globalFlag := ""
	if global {
		globalFlag = " --global"
	}
	fmt.Printf("\nUse 'blues-traveler hooks list --installed%s' to verify the changes.\n", globalFlag)
	return nil
}
