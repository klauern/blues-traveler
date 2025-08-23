package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/klauern/klauer-hooks/internal/generator"
	"github.com/klauern/klauer-hooks/internal/hooks"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available hook plugins",
	Long:  `List all registered hook plugins that can be run.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available hook plugins:")
		fmt.Println()
		for _, key := range PluginKeys() {
			p, _ := GetPlugin(key)
			fmt.Printf("  %s - %s\n", key, p.Description())
		}
		fmt.Println()
		fmt.Println("Use 'hooks run <key>' to run a plugin.")
		fmt.Println("Use 'hooks install <key>' to install a plugin in Claude Code settings.")
	},
}

var runCmd = &cobra.Command{
	Use:   "run [plugin-key]",
	Short: "Run a specific hook plugin",
	Long:  `Run a specific hook plugin directly. This is typically called by Claude Code.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		// Get logging flag
		logEnabled, _ := cmd.Flags().GetBool("log")
		if logEnabled {
			// Get log rotation config from our own config file (project first, then global)
			logConfig := getLogRotationConfigFromFile(false)
			if logConfig.MaxAge == 0 && logConfig.MaxSize == 0 {
				// If project config is empty, try global config
				logConfig = getLogRotationConfigFromFile(true)
			}

			// Setup log rotation
			logPath := GetLogPath(key)
			rotatingLogger := SetupLogRotation(logPath, logConfig)
			if rotatingLogger != nil {
				// Set global logging configuration with rotating logger
				hooks.SetGlobalLoggingConfig(true, ".claude/hooks")
				fmt.Printf("Logging enabled with rotation - output will be written to %s\n", logPath)
				fmt.Printf("Log rotation: max %d days, %dMB per file, %d backups\n",
					logConfig.MaxAge, logConfig.MaxSize, logConfig.MaxBackups)

				// Clean up old logs
				if err := CleanupOldLogs(filepath.Dir(logPath), logConfig.MaxAge); err != nil {
					fmt.Printf("Warning: Failed to cleanup old logs: %v\n", err)
				}
			} else {
				// Fallback to standard logging
				hooks.SetGlobalLoggingConfig(true, ".claude/hooks")
				fmt.Printf("Logging enabled - output will be written to %s\n", logPath)
			}
		}

		plugin, exists := GetPlugin(key)
		if !exists {
			fmt.Fprintf(os.Stderr, "Error: Plugin '%s' not found.\n", key)
			fmt.Fprintf(os.Stderr, "Available plugins: %s\n", strings.Join(PluginKeys(), ", "))
			os.Exit(1)
		}
		fmt.Printf("Starting %s...\n", plugin.Name())
		if err := plugin.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running plugin: %v\n", err)
			os.Exit(1)
		}
	},
}

var installCmd = &cobra.Command{
	Use:   "install [hook-type] [options]",
	Short: "Install a hook type into Claude Code settings",
	Long: `Install a hook type into your Claude Code settings.json file.
This will automatically configure the hook to run for the specified events.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hookType := args[0]

		// Validate plugin exists
		if _, exists := GetPlugin(hookType); !exists {
			fmt.Fprintf(os.Stderr, "Error: Plugin '%s' not found.\n", hookType)
			fmt.Fprintf(os.Stderr, "Available plugins: %s\n", strings.Join(PluginKeys(), ", "))
			os.Exit(1)
		}

		// Get flags
		global, _ := cmd.Flags().GetBool("global")
		event, _ := cmd.Flags().GetString("event")
		matcher, _ := cmd.Flags().GetString("matcher")
		timeoutFlag, _ := cmd.Flags().GetInt("timeout")
		logEnabled, _ := cmd.Flags().GetBool("log")

		// Validate event
		if !IsValidEventType(event) {
			fmt.Fprintf(os.Stderr, "Error: Invalid event '%s'.\n", event)
			fmt.Fprintf(os.Stderr, "Valid events: %s\n", strings.Join(ValidEventTypes(), ", "))
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
		}

		// Get settings path
		settingsPath, err := getSettingsPath(global)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting settings path: %v\n", err)
			os.Exit(1)
		}

		// Load existing settings
		settings, err := loadSettings(settingsPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading settings: %v\n", err)
			os.Exit(1)
		}

		// Add hook to settings
		var timeout *int
		if timeoutFlag > 0 {
			timeout = &timeoutFlag
		}
		result := addHookToSettings(settings, event, matcher, hookCommand, timeout)

		// Check for duplicates or replacements
		if result.WasDuplicate {
			if strings.Contains(result.DuplicateInfo, "Replaced existing") {
				fmt.Printf("🔄 %s\n", result.DuplicateInfo)
			} else {
				fmt.Printf("⚠️  Hook already installed: %s\n", result.DuplicateInfo)
				fmt.Printf("No changes made. The hook is already configured for this event.\n")
				return
			}
		}

		// Save settings
		if err := saveSettings(settingsPath, settings); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving settings: %v\n", err)
			os.Exit(1)
		}

		scope := "project"
		if global {
			scope = "global"
		}

		fmt.Printf("✅ Successfully installed %s hook in %s settings\n", hookType, scope)
		fmt.Printf("   Event: %s\n", event)
		fmt.Printf("   Matcher: %s\n", matcher)
		fmt.Printf("   Command: %s\n", hookCommand)
		fmt.Printf("   Settings: %s\n", settingsPath)
		fmt.Println()
		fmt.Println("The hook will be active in new Claude Code sessions.")
		fmt.Println("Use 'claude /hooks' to verify the configuration.")
	},
}

var uninstallCmd = &cobra.Command{
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
		settingsPath, err := getSettingsPath(global)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting settings path: %v\n", err)
			os.Exit(1)
		}

		// Load existing settings
		settings, err := loadSettings(settingsPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading settings: %v\n", err)
			os.Exit(1)
		}

		// Remove hook from settings
		removed := removeHookFromSettings(settings, hookCommand)

		if !removed {
			fmt.Printf("Hook type '%s' was not found in settings.\n", hookType)
			os.Exit(1)
		}

		// Save settings
		if err := saveSettings(settingsPath, settings); err != nil {
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

var listHooksCmd = &cobra.Command{
	Use:   "list-installed",
	Short: "List installed hooks from settings",
	Long:  `List all hooks currently configured in Claude Code settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		global, _ := cmd.Flags().GetBool("global")

		// Get settings path
		settingsPath, err := getSettingsPath(global)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting settings path: %v\n", err)
			os.Exit(1)
		}

		// Load existing settings
		settings, err := loadSettings(settingsPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading settings: %v\n", err)
			os.Exit(1)
		}

		scope := "project"
		if global {
			scope = "global"
		}

		fmt.Printf("Installed hooks (%s settings):\n", scope)
		fmt.Printf("Settings file: %s\n\n", settingsPath)

		if isHooksConfigEmpty(settings.Hooks) {
			fmt.Println("No hooks are currently installed.")
		} else {
			printHookMatchers("PreToolUse", settings.Hooks.PreToolUse)
			printHookMatchers("PostToolUse", settings.Hooks.PostToolUse)
			printHookMatchers("UserPromptSubmit", settings.Hooks.UserPromptSubmit)
			printHookMatchers("Notification", settings.Hooks.Notification)
			printHookMatchers("Stop", settings.Hooks.Stop)
			printHookMatchers("SubagentStop", settings.Hooks.SubagentStop)
			printHookMatchers("PreCompact", settings.Hooks.PreCompact)
			printHookMatchers("SessionStart", settings.Hooks.SessionStart)
			printHookMatchers("SessionEnd", settings.Hooks.SessionEnd)
		}

		// Add examples section
		printUninstallExamples(global)
	},
}

func printHookMatchers(eventName string, matchers []HookMatcher) {
	if len(matchers) == 0 {
		return
	}

	fmt.Printf("%s:\n", eventName)
	for _, matcher := range matchers {
		matcherStr := matcher.Matcher
		if matcherStr == "" {
			matcherStr = "*"
		}
		fmt.Printf("  Matcher: %s\n", matcherStr)
		for _, hook := range matcher.Hooks {
			fmt.Printf("    - %s", hook.Command)
			if hook.Timeout != nil {
				fmt.Printf(" (timeout: %ds)", *hook.Timeout)
			}
			fmt.Println()
		}
	}
	fmt.Println()
}

func printUninstallExamples(global bool) {
	scope := "project"
	globalFlag := ""
	if global {
		scope = "global"
		globalFlag = " --global"
	}

	fmt.Printf("📝 How to remove hooks:\n\n")

	fmt.Printf("Remove a specific hook type (removes ALL instances):\n")
	fmt.Printf("  hooks uninstall debug%s\n", globalFlag)
	fmt.Printf("  hooks uninstall security%s\n", globalFlag)
	fmt.Printf("  hooks uninstall audit%s\n\n", globalFlag)

	fmt.Printf("Remove ALL klauer-hooks (preserves other hooks):\n")
	fmt.Printf("  hooks uninstall all%s\n\n", globalFlag)

	fmt.Printf("Remove ALL hooks from %s settings:\n", scope)
	if global {
		fmt.Printf("  rm ~/.claude/settings.json\n")
		fmt.Printf("  # (or edit the file manually to remove specific events)\n\n")
	} else {
		fmt.Printf("  rm .claude/settings.json\n")
		fmt.Printf("  # (or edit the file manually to remove specific events)\n\n")
	}

	fmt.Printf("View this list again:\n")
	fmt.Printf("  hooks list-installed%s\n\n", globalFlag)

	fmt.Printf("💡 Note: The 'uninstall' command removes ALL instances of a hook type\n")
	fmt.Printf("   from ALL events (PreToolUse, PostToolUse, etc.)\n")
}

var listEventsCmd = &cobra.Command{
	Use:   "list-events",
	Short: "List all available Claude Code hook events",
	Long:  `List all Claude Code hook events that can be configured in settings.json, including their descriptions and when they trigger.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available Claude Code Hook Events:")
		fmt.Println()

		events := AllClaudeCodeEvents()
		ccHooksSupported := 0
		for _, event := range events {
			status := ""
			if event.SupportedByCCHooks {
				status = " ✓ (cchooks library)"
				ccHooksSupported++
			} else {
				status = " ⚠ (Claude Code only)"
			}

			fmt.Printf("  %s%s\n", event.Name, status)
			fmt.Printf("      %s\n", event.Description)
			fmt.Println()
		}

		fmt.Printf("Total: %d events available (%d supported by cchooks library)\n\n", len(events), ccHooksSupported)
		fmt.Println("✓ Events marked with checkmark can be handled by klauer-hooks plugins")
		fmt.Println("⚠ Events marked with warning require custom hook implementations")
		fmt.Println()
		fmt.Println("Use 'hooks install <plugin-key> --event <event-name>' to install a hook for a specific event.")
		fmt.Println("Use 'hooks list-installed' to see currently configured hooks.")
	},
}

var generateCmd = &cobra.Command{
	Use:   "generate [hook-name]",
	Short: "Generate a new hook from template",
	Long: `Generate a new hook file from a template. This creates the hook implementation
and optionally a test file. The hook will need to be registered manually in the registry.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hookName := args[0]

		// Get flags
		description, _ := cmd.Flags().GetString("description")
		hookTypeStr, _ := cmd.Flags().GetString("type")
		includeTest, _ := cmd.Flags().GetBool("test")
		outputDir, _ := cmd.Flags().GetString("output")

		// Validate hook name
		if err := generator.ValidateHookName(hookName); err != nil {
			fmt.Fprintf(os.Stderr, "Error validating hook name: %v\n", err)
			os.Exit(1)
		}

		// Set default description if not provided
		if description == "" {
			description = fmt.Sprintf("Custom %s hook implementation", hookName)
		}

		// Parse hook type
		var hookType generator.HookType
		switch hookTypeStr {
		case "pre", "pre_tool":
			hookType = generator.PreToolHook
		case "post", "post_tool":
			hookType = generator.PostToolHook
		case "both":
			hookType = generator.BothHooks
		default:
			fmt.Fprintf(os.Stderr, "Error: Invalid hook type '%s'. Valid types: pre, post, both\n", hookTypeStr)
			os.Exit(1)
		}

		// Create generator
		gen := generator.NewGenerator(outputDir)

		// Generate hook
		if err := gen.GenerateHook(hookName, description, hookType, includeTest); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating hook: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\n✅ Successfully generated hook '%s'\n", hookName)
	},
}

var configLogCmd = &cobra.Command{
	Use:   "config-log",
	Short: "Configure log rotation settings",
	Long:  `Configure log rotation settings including maximum age, file size, and backup count.`,
	Run: func(cmd *cobra.Command, args []string) {
		global, _ := cmd.Flags().GetBool("global")
		maxAge, _ := cmd.Flags().GetInt("max-age")
		maxSize, _ := cmd.Flags().GetInt("max-size")
		maxBackups, _ := cmd.Flags().GetInt("max-backups")
		compress, _ := cmd.Flags().GetBool("compress")
		show, _ := cmd.Flags().GetBool("show")

		configPath, err := getLogConfigPath(global)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting config path: %v\n", err)
			os.Exit(1)
		}

		config, err := loadLogConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		if show {
			// Show current log rotation settings
			scope := "project"
			if global {
				scope = "global"
			}
			fmt.Printf("Current log rotation settings (%s: %s):\n", scope, configPath)
			fmt.Printf("  Max Age: %d days\n", config.LogRotation.MaxAge)
			fmt.Printf("  Max Size: %d MB\n", config.LogRotation.MaxSize)
			fmt.Printf("  Max Backups: %d files\n", config.LogRotation.MaxBackups)
			fmt.Printf("  Compress: %t\n", config.LogRotation.Compress)
			return
		}

		// Only update non-zero values
		if maxAge > 0 {
			config.LogRotation.MaxAge = maxAge
		}
		if maxSize > 0 {
			config.LogRotation.MaxSize = maxSize
		}
		if maxBackups > 0 {
			config.LogRotation.MaxBackups = maxBackups
		}
		if cmd.Flags().Changed("compress") {
			config.LogRotation.Compress = compress
		}

		if err := saveLogConfig(configPath, config); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			os.Exit(1)
		}

		scope := "project"
		if global {
			scope = "global"
		}
		fmt.Printf("Log rotation configuration updated in %s config (%s):\n", scope, configPath)
		fmt.Printf("  Max Age: %d days\n", config.LogRotation.MaxAge)
		fmt.Printf("  Max Size: %d MB\n", config.LogRotation.MaxSize)
		fmt.Printf("  Max Backups: %d files\n", config.LogRotation.MaxBackups)
		fmt.Printf("  Compress: %t\n", config.LogRotation.Compress)
	},
}

func uninstallAllKlauerHooks(global bool) {
	// Get settings path
	settingsPath, err := getSettingsPath(global)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Load existing settings
	settings, err := loadSettings(settingsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading settings: %v\n", err)
		os.Exit(1)
	}

	scope := "project"
	if global {
		scope = "global"
	}

	// Count klauer-hooks before removal
	totalHooksBefore := countKlauerHooksInSettings(settings)

	if totalHooksBefore == 0 {
		fmt.Printf("No klauer-hooks found in %s settings.\n", scope)
		return
	}

	// Show what will be removed
	fmt.Printf("Found %d klauer-hooks in %s settings:\n\n", totalHooksBefore, scope)
	printKlauerHooksToRemove(settings)

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
	removed := removeAllKlauerHooksFromSettings(settings)

	if removed == 0 {
		fmt.Printf("No klauer-hooks were found to remove.\n")
		return
	}

	// Save settings
	if err := saveSettings(settingsPath, settings); err != nil {
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
