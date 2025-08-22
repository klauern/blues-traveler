package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available hook templates",
	Long:  `List all available hook templates that can be created.`,
	Run: func(cmd *cobra.Command, args []string) {
		templates := GetHookTemplates()

		fmt.Println("Available hook templates:")
		fmt.Println()

		for name, template := range templates {
			fmt.Printf("  %s - %s\n", name, template.Description)
		}
		fmt.Println()
		fmt.Println("Use 'hooks create <template>' to create a hook from a template.")
		fmt.Println("Use 'hooks show <template>' to preview a template.")
	},
}

var createCmd = &cobra.Command{
	Use:   "create [template] [name]",
	Short: "Create a new hook from a template",
	Long:  `Create a new hook file from one of the available templates.`,
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		templateName := args[0]

		templates := GetHookTemplates()
		template, exists := templates[templateName]
		if !exists {
			fmt.Fprintf(os.Stderr, "Error: Template '%s' not found.\n", templateName)
			fmt.Fprintf(os.Stderr, "Available templates: %s\n", strings.Join(getTemplateNames(), ", "))
			os.Exit(1)
		}

		// Determine output filename
		var filename string
		if len(args) > 1 {
			filename = args[1]
		} else {
			filename = templateName + "-hook"
		}

		// Ensure .go extension
		if !strings.HasSuffix(filename, ".go") {
			filename += ".go"
		}

		// Check if file already exists
		if _, err := os.Stat(filename); err == nil {
			fmt.Fprintf(os.Stderr, "Error: File '%s' already exists.\n", filename)
			os.Exit(1)
		}

		// Write template to file
		err := os.WriteFile(filename, []byte(template.Code), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Created %s hook: %s\n", template.Name, filename)
		fmt.Printf("Description: %s\n", template.Description)
		fmt.Println()
		fmt.Printf("Next steps:\n")
		fmt.Printf("  1. Review and customize the hook code in %s\n", filename)
		fmt.Printf("  2. Build the hook: go build -o %s %s\n",
			strings.TrimSuffix(filename, ".go"), filename)
		fmt.Printf("  3. Configure Claude Code to use the hook\n")
	},
}

var showCmd = &cobra.Command{
	Use:   "show [template]",
	Short: "Show the code for a template",
	Long:  `Display the source code for a hook template without creating a file.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		templateName := args[0]

		templates := GetHookTemplates()
		template, exists := templates[templateName]
		if !exists {
			fmt.Fprintf(os.Stderr, "Error: Template '%s' not found.\n", templateName)
			fmt.Fprintf(os.Stderr, "Available templates: %s\n", strings.Join(getTemplateNames(), ", "))
			os.Exit(1)
		}

		fmt.Printf("Template: %s\n", template.Name)
		fmt.Printf("Description: %s\n", template.Description)
		fmt.Println()
		fmt.Println("Code:")
		fmt.Println("---")
		fmt.Println(template.Code)
	},
}

func getTemplateNames() []string {
	templates := GetHookTemplates()
	names := make([]string, 0, len(templates))
	for name := range templates {
		names = append(names, name)
	}
	return names
}

// Add build command
var buildCmd = &cobra.Command{
	Use:   "build [file]",
	Short: "Build a hook file into an executable",
	Long:  `Build a Go hook file into an executable that can be used with Claude Code.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sourceFile := args[0]

		// Check if source file exists
		if _, err := os.Stat(sourceFile); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: File '%s' does not exist.\n", sourceFile)
			os.Exit(1)
		}

		// Determine output name
		outputName := strings.TrimSuffix(filepath.Base(sourceFile), ".go")

		// Build the hook
		buildCmd := fmt.Sprintf("go build -o %s %s", outputName, sourceFile)
		fmt.Printf("Building: %s\n", buildCmd)

		// You would typically use exec.Command here, but for simplicity:
		fmt.Printf("Run this command to build your hook:\n")
		fmt.Printf("  %s\n", buildCmd)
		fmt.Println()
		fmt.Printf("Then configure Claude Code to use: ./%s\n", outputName)
	},
}

var installCmd = &cobra.Command{
	Use:   "install [hook-binary] [options]",
	Short: "Install a hook into Claude Code settings",
	Long: `Install a hook binary into your Claude Code settings.json file.
This will automatically configure the hook to run for the specified events.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hookBinary := args[0]

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

		// Check if hook binary exists
		if _, err := os.Stat(hookBinary); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: Hook binary '%s' does not exist.\n", hookBinary)
			os.Exit(1)
		}

		// Convert to absolute path
		absPath, err := filepath.Abs(hookBinary)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to get absolute path: %v\n", err)
			os.Exit(1)
		}

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
		addHookToSettings(settings, event, matcher, absPath, timeout)

		// Save settings
		if err := saveSettings(settingsPath, settings); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving settings: %v\n", err)
			os.Exit(1)
		}

		scope := "project"
		if global {
			scope = "global"
		}

		fmt.Printf("✅ Successfully installed hook in %s settings\n", scope)
		fmt.Printf("   Event: %s\n", event)
		fmt.Printf("   Matcher: %s\n", matcher)
		fmt.Printf("   Command: %s\n", absPath)
		fmt.Printf("   Settings: %s\n", settingsPath)
		fmt.Println()
		fmt.Println("The hook will be active in new Claude Code sessions.")
		fmt.Println("Use 'claude /hooks' to verify the configuration.")
	},
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall [hook-binary]",
	Short: "Remove a hook from Claude Code settings",
	Long:  `Remove a hook binary from your Claude Code settings.json file.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hookBinary := args[0]
		global, _ := cmd.Flags().GetBool("global")

		// Convert to absolute path for comparison
		absPath, err := filepath.Abs(hookBinary)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to get absolute path: %v\n", err)
			os.Exit(1)
		}

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
		removed := removeHookFromSettings(settings, absPath)

		if !removed {
			fmt.Printf("Hook '%s' was not found in settings.\n", absPath)
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

		fmt.Printf("✅ Successfully removed hook from %s settings\n", scope)
		fmt.Printf("   Command: %s\n", absPath)
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
