package cmd

import (
    "context"
    "fmt"
    "sort"
    "strings"

    btconfig "github.com/klauern/blues-traveler/internal/config"
    "github.com/klauern/blues-traveler/internal/config"
    "github.com/urfave/cli/v3"
)

func NewListCmd(getPlugin func(string) (interface {
	Run() error
	Description() string
}, bool), pluginKeys func() []string,
) *cli.Command {
	return &cli.Command{
		Name:        "list",
		Usage:       "List available hook plugins",
		Description: `List all registered hook plugins that can be run.`,
        Action: func(ctx context.Context, cmd *cli.Command) error {
            keys := pluginKeys()
            var builtin, custom []string
            groupSet := map[string]struct{}{}
            for _, k := range keys {
                if strings.HasPrefix(k, "config:") {
                    custom = append(custom, k)
                    parts := strings.SplitN(k, ":", 3)
                    if len(parts) == 3 {
                        groupSet[parts[1]] = struct{}{}
                    }
                } else {
                    builtin = append(builtin, k)
                }
            }
            sort.Strings(builtin)
            sort.Strings(custom)

            fmt.Println("Available hook plugins:")
            fmt.Println()
            if len(builtin) > 0 {
                fmt.Println("Built-in hooks:")
                for _, key := range builtin {
                    p, _ := getPlugin(key)
                    fmt.Printf("  %s - %s\n", key, p.Description())
                }
                fmt.Println()
            }
            if len(custom) > 0 {
                fmt.Println("Custom config hooks (from hooks.yml):")
                for _, key := range custom {
                    p, _ := getPlugin(key)
                    fmt.Printf("  %s - %s\n", key, p.Description())
                }
                // Show groups summary
                groups := make([]string, 0, len(groupSet))
                for g := range groupSet {
                    groups = append(groups, g)
                }
                sort.Strings(groups)
                if len(groups) > 0 {
                    fmt.Printf("\nGroups: %s\n", strings.Join(groups, ", "))
                }
                fmt.Println()
            } else {
                // Suggest how to create custom hooks if none present
                if cfg, err := btconfig.LoadHooksConfig(); err == nil && cfg != nil {
                    // loaded but no hooks registered likely means no groups/jobs defined
                }
                fmt.Println("No custom hooks found. Create .claude/hooks.yml and use 'blues-traveler install custom --list'.")
                fmt.Println()
            }

            fmt.Println("Use 'blues-traveler run <key>' to run a hook.")
            fmt.Println("Use 'blues-traveler install <key>' to install a built-in hook.")
            fmt.Println("Use 'blues-traveler install custom <group>' to install a group from hooks.yml.")
            return nil
        },
    }
}

func NewListInstalledCmd() *cli.Command {
	return &cli.Command{
		Name:        "list-installed",
		Usage:       "List installed hooks from settings",
		Description: `List all hooks currently configured in Claude Code settings.`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "global",
				Aliases: []string{"g"},
				Value:   false,
				Usage:   "Show global settings (~/.claude/settings.json)",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			global := cmd.Bool("global")

			// Get settings path
			settingsPath, err := config.GetSettingsPath(global)
			if err != nil {
				return fmt.Errorf("error getting settings path: %v", err)
			}

			// Load existing settings
			settings, err := config.LoadSettings(settingsPath)
			if err != nil {
				return fmt.Errorf("error loading settings: %v", err)
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
			return nil
		},
	}
}

func NewListEventsCmd(allEvents func() []ClaudeCodeEvent) *cli.Command {
	return &cli.Command{
		Name:        "list-events",
		Usage:       "List all available Claude Code hook events",
		Description: `List all Claude Code hook events that can be configured in settings.json, including their descriptions and when they trigger.`,
		Action: func(ctx context.Context, cmd *cli.Command) error {
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
			fmt.Println("✓ Events marked with checkmark can be handled by blues-traveler plugins")
			fmt.Println("⚠ Events marked with warning require custom hook implementations")
			fmt.Println()
			fmt.Println("Use 'blues-traveler install <plugin-key> --event <event-name>' to install a hook for a specific event.")
			fmt.Println("Use 'blues-traveler list-installed' to see currently configured hooks.")
			return nil
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
	fmt.Printf("  blues-traveler uninstall debug%s\n", globalFlag)
	fmt.Printf("  blues-traveler uninstall security%s\n", globalFlag)
	fmt.Printf("  blues-traveler uninstall audit%s\n\n", globalFlag)

	fmt.Printf("Remove ALL blues-traveler hooks (preserves other hooks):\n")
	fmt.Printf("  blues-traveler uninstall all%s\n\n", globalFlag)

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
