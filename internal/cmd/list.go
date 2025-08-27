package cmd

import (
	"fmt"
	"os"

	"github.com/klauern/klauer-hooks/internal/config"
	"github.com/spf13/cobra"
)

func NewListCmd(getPlugin func(string) (interface {
	Run() error
	Description() string
}, bool), pluginKeys func() []string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available hook plugins",
		Long:  `List all registered hook plugins that can be run.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Available hook plugins:")
			fmt.Println()
			for _, key := range pluginKeys() {
				p, _ := getPlugin(key)
				fmt.Printf("  %s - %s\n", key, p.Description())
			}
			fmt.Println()
			fmt.Println("Use 'hooks run <key>' to run a plugin.")
			fmt.Println("Use 'hooks install <key>' to install a plugin in Claude Code settings.")
		},
	}
}

func NewListInstalledCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-installed",
		Short: "List installed hooks from settings",
		Long:  `List all hooks currently configured in Claude Code settings.`,
		Run: func(cmd *cobra.Command, args []string) {
			global, _ := cmd.Flags().GetBool("global")

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

			scope := "project"
			if global {
				scope = "global"
			}

			fmt.Printf("Installed hooks (%s settings):\n", scope)
			fmt.Printf("Settings file: %s\n\n", settingsPath)

			if config.IsHooksConfigEmpty(settings.Hooks) {
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

	cmd.Flags().BoolP("global", "g", false, "Show global settings (~/.claude/settings.json)")
	return cmd
}

func NewListEventsCmd(allEvents func() []ClaudeCodeEvent) *cobra.Command {
	return &cobra.Command{
		Use:   "list-events",
		Short: "List all available Claude Code hook events",
		Long:  `List all Claude Code hook events that can be configured in settings.json, including their descriptions and when they trigger.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Available Claude Code Hook Events:")
			fmt.Println()

			events := allEvents()
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
}

func printHookMatchers(eventName string, matchers []config.HookMatcher) {
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

// ClaudeCodeEvent represents a Claude Code hook event type with metadata
type ClaudeCodeEvent struct {
	Type               EventType
	Name               string
	Description        string
	SupportedByCCHooks bool
}

// EventType represents a Claude Code hook event
type EventType string
