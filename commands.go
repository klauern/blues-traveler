package main

import (
	"fmt"
	"os"
	"strings"

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

func getHookTypeNames() []string {
	// Backwards-compatible helper now delegates to plugin registry.
	return PluginKeys()
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

		// Validate event
		validEvents := []string{"PreToolUse", "PostToolUse", "UserPromptSubmit", "Notification", "Stop", "SubagentStop", "PreCompact", "SessionStart"}
		eventValid := false
		for _, validEvent := range validEvents {
			if event == validEvent {
				eventValid = true
				break
			}
		}
		if !eventValid {
			fmt.Fprintf(os.Stderr, "Error: Invalid event '%s'.\n", event)
			fmt.Fprintf(os.Stderr, "Valid events: %s\n", strings.Join(validEvents, ", "))
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

		// Add hook to settings
		var timeout *int
		if timeoutFlag > 0 {
			timeout = &timeoutFlag
		}
		addHookToSettings(settings, event, matcher, hookCommand, timeout)

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
	Use:   "uninstall [hook-type]",
	Short: "Remove a hook type from Claude Code settings",
	Long:  `Remove a hook type from your Claude Code settings.json file.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hookType := args[0]
		global, _ := cmd.Flags().GetBool("global")

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
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
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

		fmt.Printf("Installed hooks (%s settings):\n", scope)
		fmt.Printf("Settings file: %s\n\n", settingsPath)

		if isHooksConfigEmpty(settings.Hooks) {
			fmt.Println("No hooks are currently installed.")
			return
		}

		printHookMatchers("PreToolUse", settings.Hooks.PreToolUse)
		printHookMatchers("PostToolUse", settings.Hooks.PostToolUse)
		printHookMatchers("UserPromptSubmit", settings.Hooks.UserPromptSubmit)
		printHookMatchers("Notification", settings.Hooks.Notification)
		printHookMatchers("Stop", settings.Hooks.Stop)
		printHookMatchers("SubagentStop", settings.Hooks.SubagentStop)
		printHookMatchers("PreCompact", settings.Hooks.PreCompact)
		printHookMatchers("SessionStart", settings.Hooks.SessionStart)
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
