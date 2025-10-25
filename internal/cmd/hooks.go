package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/klauern/blues-traveler/internal/config"
	"github.com/klauern/blues-traveler/internal/core"
	"github.com/klauern/blues-traveler/internal/platform"
	"github.com/klauern/blues-traveler/internal/platform/claude"
	"github.com/klauern/blues-traveler/internal/platform/cursor"
	"github.com/urfave/cli/v3"
	yaml "gopkg.in/yaml.v3"
)

// ClaudeCodeEvent represents a Claude Code hook event type with metadata
type ClaudeCodeEvent struct {
	Type               EventType
	Name               string
	Description        string
	SupportedByCCHooks bool
}

// EventType represents a Claude Code hook event
type EventType string

// NewHooksCommand creates the main hooks command with all subcommands
func NewHooksCommand(getPlugin func(string) (interface {
	Run() error
	Description() string
}, bool), isPluginEnabled func(string) bool, pluginKeys func() []string, isValidEventType func(string) bool, validEventTypes func() []string, allEvents func() []ClaudeCodeEvent,
) *cli.Command {
	return &cli.Command{
		Name:        "hooks",
		Usage:       "Manage and run hook plugins",
		Description: `Manage hook plugins including listing, running, installing, and uninstalling hooks.`,
		Commands: []*cli.Command{
			newHooksListCommand(getPlugin, pluginKeys, allEvents),
			newHooksRunCommand(getPlugin, isPluginEnabled, pluginKeys),
			newHooksInstallCommand(getPlugin, pluginKeys, isValidEventType, validEventTypes),
			newHooksUninstallCommand(),
			newHooksCustomCommand(isValidEventType, validEventTypes),
		},
	}
}

// newHooksListCommand creates the consolidated list command
func newHooksListCommand(getPlugin func(string) (interface {
	Run() error
	Description() string
}, bool), pluginKeys func() []string, allEvents func() []ClaudeCodeEvent,
) *cli.Command {
	return &cli.Command{
		Name:        "list",
		Usage:       "List available hooks, installed hooks, or events",
		Description: `List available hook plugins, installed hooks from settings, or available Claude Code events.`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "installed",
				Aliases: []string{"i"},
				Value:   false,
				Usage:   "Show installed hooks from settings",
			},
			&cli.BoolFlag{
				Name:    "events",
				Aliases: []string{"e"},
				Value:   false,
				Usage:   "Show available Claude Code hook events",
			},
			&cli.BoolFlag{
				Name:    "global",
				Aliases: []string{"g"},
				Value:   false,
				Usage:   "Show global settings (~/.claude/settings.json) when using --installed",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			installed := cmd.Bool("installed")
			events := cmd.Bool("events")
			global := cmd.Bool("global")

			if installed {
				return listInstalledHooks(global)
			}

			if events {
				return listEvents(allEvents)
			}

			// Default: list available hooks
			return listAvailableHooks(getPlugin, pluginKeys)
		},
	}
}

// newHooksRunCommand creates the run command
func newHooksRunCommand(getPlugin func(string) (interface {
	Run() error
	Description() string
}, bool), isPluginEnabled func(string) bool, pluginKeys func() []string,
) *cli.Command {
	return &cli.Command{
		Name:        "run",
		Usage:       "Run a specific hook plugin",
		ArgsUsage:   "[plugin-key]",
		Description: `Run a specific hook plugin. Executes only that hook's handlers (no unified pipeline).`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "cursor-mode",
				Value: false,
				Usage: "Run in Cursor mode (JSON I/O over stdin/stdout)",
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
				return fmt.Errorf("exactly one argument required: [plugin-key]")
			}
			key := args[0]

			// Check for Cursor mode
			cursorMode := cmd.Bool("cursor-mode")
			if cursorMode {
				return runHookCursorMode(key, getPlugin, isPluginEnabled)
			}

			// Validate plugin exists early
			p, exists := getPlugin(key)
			if !exists {
				return fmt.Errorf("plugin '%s' not found.\nAvailable plugins: %s", key, strings.Join(pluginKeys(), ", "))
			}

			// Enablement check before side effects
			if !isPluginEnabled(key) {
				fmt.Printf("Plugin '%s' is disabled via settings. Nothing to do.\n", key)
				return nil
			}

			// Logging flags
			logEnabled := cmd.Bool("log")
			logFormat := cmd.String("log-format")
			if logFormat == "" {
				logFormat = config.LoggingFormatJSONL
			}
			if logEnabled && !config.IsValidLoggingFormat(logFormat) {
				return fmt.Errorf("invalid --log-format '%s'. Valid: jsonl, pretty", logFormat)
			}
			if logEnabled {
				logConfig := config.GetLogRotationConfigFromFile(false)
				if logConfig.MaxAge == 0 && logConfig.MaxSize == 0 {
					logConfig = config.GetLogRotationConfigFromFile(true)
				}

				logPath := config.GetLogPath(key)
				rotatingLogger := config.SetupLogRotation(logPath, logConfig)
				if rotatingLogger != nil {
					core.SetGlobalLoggingConfig(true, ".claude/hooks", logFormat)
					fmt.Printf("Logging enabled with rotation - output will be written to %s\n", logPath)
					fmt.Printf("Log rotation: max %d days, %dMB per file, %d backups\n",
						logConfig.MaxAge, logConfig.MaxSize, logConfig.MaxBackups)
					if err := config.CleanupOldLogs(filepath.Dir(logPath), logConfig.MaxAge); err != nil {
						fmt.Printf("Warning: Failed to cleanup old logs: %v\n", err)
					}
				} else {
					core.SetGlobalLoggingConfig(true, ".claude/hooks", logFormat)
					fmt.Printf("Logging enabled - output will be written to %s\n", logPath)
				}
			}

			fmt.Printf("Running hook '%s'...\n", key)
			if err := p.Run(); err != nil {
				return fmt.Errorf("hook '%s' failed: %v", key, err)
			}
			return nil
		},
	}
}

// newHooksInstallCommand creates the install command
func newHooksInstallCommand(getPlugin func(string) (interface {
	Run() error
	Description() string
}, bool), pluginKeys func() []string, isValidEventType func(string) bool, validEventTypes func() []string,
) *cli.Command {
	return &cli.Command{
		Name:      "install",
		Usage:     "Install a hook type into IDE settings (Claude Code or Cursor)",
		ArgsUsage: "[hook-type]",
		Description: `Install a hook type into your IDE settings.
For Claude Code: Updates .claude/settings.json
For Cursor: Generates wrapper script and updates ~/.cursor/hooks.json`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "platform",
				Aliases: []string{"p"},
				Value:   "",
				Usage:   "Target platform: claudecode, cursor (auto-detect if not specified)",
			},
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
			plugin, exists := getPlugin(hookType)
			if !exists {
				return fmt.Errorf("plugin '%s' not found.\nAvailable plugins: %s", hookType, strings.Join(pluginKeys(), ", "))
			}

			// Get flags
			platformFlag := cmd.String("platform")
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

			// Detect platform
			var platformType platform.Type
			if platformFlag != "" {
				var err error
				platformType, err = platform.TypeFromString(platformFlag)
				if err != nil {
					return fmt.Errorf("invalid platform: %w", err)
				}
			} else {
				detector := platform.NewDetector()
				detectedType, err := detector.DetectType()
				if err != nil {
					return fmt.Errorf("failed to detect platform: %w", err)
				}
				platformType = detectedType
			}

			// Create platform instance
			p := newPlatformFromType(platformType)

			// Route to platform-specific installation
			switch platformType {
			case platform.Cursor:
				return installHookCursor(p, hookType, plugin, event, matcher)
			case platform.ClaudeCode:
				return installHookClaudeCode(hookType, plugin, global, event, matcher, timeoutFlag, logEnabled, logFormat, isValidEventType)
			default:
				return fmt.Errorf("unsupported platform: %s", platformType)
			}
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
				uninstallAllBluesTravelerHooks(global, cmd.Bool("yes"))
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
				return fmt.Errorf("error getting settings path: %v", err)
			}

			// Load existing settings
			settings, err := config.LoadSettings(settingsPath)
			if err != nil {
				return fmt.Errorf("error loading settings: %v", err)
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

// newHooksCustomCommand creates the custom hooks command group
func newHooksCustomCommand(isValidEventType func(string) bool, validEventTypes func() []string) *cli.Command {
	return &cli.Command{
		Name:        "custom",
		Usage:       "Manage custom hooks from hooks.yml",
		Description: `Manage custom hooks defined in .claude/hooks.yml configuration files.`,
		Commands: []*cli.Command{
			newHooksCustomInstallCommand(isValidEventType, validEventTypes),
			newHooksCustomListCommand(),
			newHooksCustomSyncCommand(),
			newHooksCustomInitCommand(),
			newHooksCustomValidateCommand(),
			newHooksCustomShowCommand(),
			newHooksCustomBlockedCommand(),
		},
	}
}

// Helper functions for the consolidated list command
func listAvailableHooks(getPlugin func(string) (interface {
	Run() error
	Description() string
}, bool), pluginKeys func() []string,
) error {
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
		fmt.Println("Custom config hooks (from config):")
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
		fmt.Println("No custom hooks found. Create .claude/hooks.yml and use 'blues-traveler hooks custom install --list'.")
		fmt.Println()
	}

	fmt.Println("Use 'blues-traveler hooks run <key>' to run a hook.")
	fmt.Println("Use 'blues-traveler hooks install <key>' to install a built-in hook.")
	fmt.Println("Use 'blues-traveler hooks custom install <group>' to install a group from hooks.yml.")
	return nil
}

// listInstalledHooks displays hooks currently installed in Claude Code settings
func listInstalledHooks(global bool) error {
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
}

// listEvents displays all available hook events and their descriptions
func listEvents(allEvents func() []ClaudeCodeEvent) error {
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
	fmt.Println("Use 'blues-traveler hooks install <plugin-key> --event <event-name>' to install a hook for a specific event.")
	fmt.Println("Use 'blues-traveler hooks list --installed' to see currently configured hooks.")
	return nil
}

// Custom command implementations
func newHooksCustomInstallCommand(isValidEventType func(string) bool, validEventTypes func() []string) *cli.Command {
	return &cli.Command{
		Name:  "install",
		Usage: "Install hooks from a named group defined in hooks.yml",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "global", Aliases: []string{"g"}, Usage: "Install to global settings"},
			&cli.StringFlag{Name: "event", Aliases: []string{"e"}, Usage: "Filter to a single event"},
			&cli.StringFlag{Name: "matcher", Aliases: []string{"m"}, Value: "*", Usage: "Default tool matcher for events (e.g., '*')"},
			&cli.StringFlag{Name: "post-matcher", Value: "Edit,Write", Usage: "Matcher for PostToolUse when not overridden"},
			&cli.BoolFlag{Name: "list", Usage: "List available groups"},
			&cli.IntFlag{Name: "timeout", Aliases: []string{"t"}, Usage: "Override timeout in seconds for installed commands"},
			&cli.BoolFlag{Name: "init", Usage: "If group not found, create a sample group stub in hooks.yml"},
			&cli.BoolFlag{Name: "prune", Usage: "Remove previously installed commands for this group before installing"},
		},
		ArgsUsage: "<group-name>",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			list := cmd.Bool("list")
			if list {
				cfg, err := config.LoadHooksConfig()
				if err != nil {
					return fmt.Errorf("failed to load hooks config: %v", err)
				}
				groups := config.ListHookGroups(cfg)
				if len(groups) == 0 {
					fmt.Println("No custom hook groups found. Create .claude/hooks.yml to define groups.")
					return nil
				}
				fmt.Println("Available custom hook groups:")
				for _, g := range groups {
					fmt.Printf("- %s\n", g)
				}
				return nil
			}

			args := cmd.Args().Slice()
			if len(args) != 1 {
				return fmt.Errorf("exactly one argument required: <group-name>")
			}
			groupName := args[0]
			global := cmd.Bool("global")
			defaultMatcher := cmd.String("matcher")
			postMatcher := cmd.String("post-matcher")
			eventFilter := strings.TrimSpace(cmd.String("event"))
			timeoutOverride := cmd.Int("timeout")

			if eventFilter != "" && !isValidEventType(eventFilter) {
				return fmt.Errorf("invalid --event '%s'. Valid events: %s", eventFilter, strings.Join(validEventTypes(), ", "))
			}

			cfg, err := config.LoadHooksConfig()
			if err != nil {
				return fmt.Errorf("failed to load hooks config: %v", err)
			}
			if cfg == nil || (*cfg)[groupName] == nil {
				if cmd.Bool("init") {
					// Create a stub group embedded in main config
					sample := fmt.Sprintf(`%s:
  PreToolUse:
    jobs:
      - name: example-check
        run: echo "TOOL=${TOOL_NAME} FILES=${FILES_CHANGED}"
        glob: ["*"]
  PostToolUse:
    jobs:
      - name: example-post
        run: echo "Post ${EVENT_NAME} on ${TOOL_NAME}"
        glob: ["*"]
`, groupName)
					// Merge sample into main config under customHooks
					if _, err := config.WriteSampleHooksConfig(global, sample, false); err != nil {
						return fmt.Errorf("write hooks sample: %v", err)
					}
					// Reload after creating stub (embedded)
					cfg, err = config.LoadHooksConfig()
					if err != nil {
						return fmt.Errorf("reload hooks config: %v", err)
					}
					if cfg == nil || (*cfg)[groupName] == nil {
						return fmt.Errorf("failed to create group '%s' in hooks.yml", groupName)
					}
				} else {
					return fmt.Errorf("group '%s' not found in hooks config (use --init to stub one)", groupName)
				}
			}

			// Resolve settings path and load
			settingsPath, err := config.GetSettingsPath(global)
			if err != nil {
				return fmt.Errorf("error getting settings path: %v", err)
			}
			settings, err := config.LoadSettings(settingsPath)
			if err != nil {
				return fmt.Errorf("error loading settings: %v", err)
			}

			// Optionally prune previously installed entries for this group
			if cmd.Bool("prune") {
				removed := config.RemoveConfigGroupFromSettings(settings, groupName, eventFilter)
				if removed > 0 {
					fmt.Printf("Pruned %d entries for group '%s'%s\n", removed, groupName, func() string {
						if eventFilter != "" {
							return " (event: " + eventFilter + ")"
						}
						return ""
					}())
				}
			}

			// Helper to choose matcher based on event type
			pickMatcher := func(event string) string {
				if event == "PostToolUse" {
					return postMatcher
				}
				return defaultMatcher
			}

			// Build commands per event
			group := (*cfg)[groupName]
			installed := 0
			for eventName, ev := range group {
				if eventFilter != "" && eventFilter != eventName {
					continue
				}
				for _, job := range ev.Jobs {
					if job.Name == "" {
						continue
					}
					execPath, err := os.Executable()
					if err != nil {
						return fmt.Errorf("failed to get executable path: %v", err)
					}
					hookCommand := fmt.Sprintf("%s hooks run config:%s:%s", execPath, groupName, job.Name)

					// Timeout selection: CLI override > job.Timeout > none
					var timeout *int
					if timeoutOverride > 0 {
						timeout = &timeoutOverride
					} else if job.Timeout > 0 {
						t := job.Timeout
						timeout = &t
					}

					// Use event-specific matcher for settings (Edit,Write,*, etc.).
					// File globs are evaluated at runtime inside the hook.
					matcher := pickMatcher(eventName)
					res := config.AddHookToSettings(settings, eventName, matcher, hookCommand, timeout)
					_ = res
					installed++
				}
			}

			// Save once after all additions
			if err := config.SaveSettings(settingsPath, settings); err != nil {
				return fmt.Errorf("error saving settings: %v", err)
			}

			scope := "project"
			if global {
				scope = "global"
			}
			fmt.Printf("✅ Installed custom group '%s' to %s settings (%d entries)\n", groupName, scope, installed)
			fmt.Printf("   Settings: %s\n", settingsPath)
			return nil
		},
	}
}

// newHooksCustomListCommand creates the custom hooks list command
func newHooksCustomListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List available custom hook groups",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			cfg, err := config.LoadHooksConfig()
			if err != nil {
				return fmt.Errorf("load error: %v", err)
			}
			groups := config.ListHookGroups(cfg)
			if len(groups) == 0 {
				fmt.Println("No custom hook groups found")
				return nil
			}
			for _, g := range groups {
				fmt.Println(g)
			}
			return nil
		},
	}
}

// newHooksCustomSyncCommand creates the custom hooks sync command
func newHooksCustomSyncCommand() *cli.Command {
	return &cli.Command{
		Name:      "sync",
		Usage:     "Sync custom hooks from hooks.yml into Claude settings",
		ArgsUsage: "[group]",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "global", Aliases: []string{"g"}, Usage: "Sync to global settings (~/.claude/settings.json)"},
			&cli.BoolFlag{Name: "dry-run", Aliases: []string{"n"}, Usage: "Show intended changes without writing"},
			&cli.StringFlag{Name: "event", Aliases: []string{"e"}, Usage: "Restrict sync to a single event (e.g., PreToolUse, PostToolUse)"},
			&cli.StringFlag{Name: "matcher", Aliases: []string{"m"}, Value: "*", Usage: "Default tool matcher for events (e.g., '*')"},
			&cli.StringFlag{Name: "post-matcher", Value: "Edit,Write", Usage: "Matcher for PostToolUse when not overridden"},
			&cli.IntFlag{Name: "timeout", Aliases: []string{"t"}, Usage: "Override timeout in seconds for installed commands"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			args := cmd.Args().Slice()
			var groupFilter string
			if len(args) > 0 {
				if len(args) > 1 {
					return fmt.Errorf("at most one [group] argument is allowed")
				}
				groupFilter = args[0]
			}

			useGlobal := cmd.Bool("global")
			dry := cmd.Bool("dry-run")
			eventFilter := strings.TrimSpace(cmd.String("event"))
			defaultMatcher := cmd.String("matcher")
			postMatcher := cmd.String("post-matcher")
			timeoutOverride := cmd.Int("timeout")

			// Load config (embedded + legacy merge)
			hooksCfg, err := config.LoadHooksConfig()
			if err != nil {
				return fmt.Errorf("load hooks config: %v", err)
			}

			// Load settings
			settingsPath, err := config.GetSettingsPath(useGlobal)
			if err != nil {
				return err
			}
			settings, err := config.LoadSettings(settingsPath)
			if err != nil {
				return err
			}

			// Resolve a stable blues-traveler path for settings entries:
			// prefer PATH lookup, then local ./blues-traveler, then current executable.
			execPath := func() string {
				if p, err := os.Executable(); err == nil {
					return p
				}
				return "blues-traveler"
			}()

			// Helper to choose matcher
			pickMatcher := func(event string) string {
				if event == "PostToolUse" {
					return postMatcher
				}
				return defaultMatcher
			}

			changed := 0

			// Step 1: Clean up stale groups that no longer exist in config
			existingGroups := config.GetConfigGroupsInSettings(settings)
			configGroups := make(map[string]bool)
			if hooksCfg != nil {
				for groupName := range *hooksCfg {
					configGroups[groupName] = true
				}
			}

			// Remove groups that exist in settings but not in current config
			for existingGroup := range existingGroups {
				// If we have a group filter, only process that specific group
				if groupFilter != "" && existingGroup != groupFilter {
					continue
				}
				// If the group doesn't exist in current config, remove it
				if !configGroups[existingGroup] {
					removed := config.RemoveConfigGroupFromSettings(settings, existingGroup, eventFilter)
					if removed > 0 {
						fmt.Printf("Cleaned up %d stale entries for removed group '%s'%s\n", removed, existingGroup, func() string {
							if eventFilter != "" {
								return " (event: " + eventFilter + ")"
							}
							return ""
						}())
						changed += removed
					}
				}
			}

			// Step 2: Iterate current config groups and sync them
			if hooksCfg != nil {
				for groupName, group := range *hooksCfg {
					if groupFilter != "" && groupName != groupFilter {
						continue
					}

					// Prune existing settings for this group (optionally event-filtered)
					removed := config.RemoveConfigGroupFromSettings(settings, groupName, eventFilter)
					if removed > 0 {
						fmt.Printf("Pruned %d entries for group '%s'%s\n", removed, groupName, func() string {
							if eventFilter != "" {
								return " (event: " + eventFilter + ")"
							}
							return ""
						}())
					}

					// Add current definitions
					for eventName, ev := range group {
						if eventFilter != "" && eventFilter != eventName {
							continue
						}
						for _, job := range ev.Jobs {
							if job.Name == "" {
								continue
							}
							// Build command to run this job
							hookCommand := fmt.Sprintf("%s hooks run config:%s:%s", execPath, groupName, job.Name)
							// Timeout preference: CLI override > job.Timeout
							var timeout *int
							if timeoutOverride > 0 {
								timeout = &timeoutOverride
							} else if job.Timeout > 0 {
								t := job.Timeout
								timeout = &t
							}
							// Matcher
							matcher := pickMatcher(eventName)
							// Add to settings
							res := config.AddHookToSettings(settings, eventName, matcher, hookCommand, timeout)
							_ = res
							changed++
							if dry {
								fmt.Printf("Would add: [%s] matcher=%q command=%q\n", eventName, matcher, hookCommand)
							}
						}
					}
				}
			}

			if changed == 0 {
				fmt.Println("No changes detected.")
				return nil
			}

			if dry {
				fmt.Println("Dry run; not writing settings.")
				return nil
			}

			if err := config.SaveSettings(settingsPath, settings); err != nil {
				return err
			}
			scope := "project"
			if useGlobal {
				scope = "global"
			}
			fmt.Printf("Synced %d entries into %s settings: %s\n", changed, scope, settingsPath)
			return nil
		},
	}
}

// newHooksCustomInitCommand creates the custom hooks init command
func newHooksCustomInitCommand() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Create a sample hooks configuration file",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "global", Aliases: []string{"g"}, Usage: "Create in ~/.claude"},
			&cli.BoolFlag{Name: "overwrite", Usage: "Overwrite existing file if present"},
			&cli.StringFlag{Name: "group", Aliases: []string{"G"}, Value: "example", Usage: "Group name for this config"},
			&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Usage: "Filename for per-group config (writes .claude/hooks/<name>.yml)"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			global := cmd.Bool("global")
			overwrite := cmd.Bool("overwrite")
			group := cmd.String("group")
			fileName := cmd.String("name")

			var sample string
			if global {
				// Minimal global config - no example hooks to avoid accidental installation
				sample = fmt.Sprintf(`# Global hooks configuration for group '%s'
# This is your personal global configuration. Add real hooks here.
# See README.md in this directory for documentation and examples.
%s:
  # Add your custom hooks here
  # Example structure:
  # PreToolUse:
  #   jobs:
  #     - name: my-security-check
  #       run: ./my-script.sh
  #       glob: ["*.go"]
`, group, group)
			} else {
				// Project config with comprehensive examples for learning
				sample = fmt.Sprintf(`# Sample hooks configuration for group '%s'
%s:
  PreToolUse:
    jobs:
      - name: pre-sample
        run: echo "PreToolUse TOOL=${TOOL_NAME}"
        glob: ["*"]
  PostToolUse:
    jobs:
      - name: post-format-sample
        # Demonstrates file-based action with TOOL_OUTPUT_FILE for Edit/Write
        run: ruff format --fix ${TOOL_OUTPUT_FILE}
        only: ${TOOL_NAME} == "Edit" || ${TOOL_NAME} == "Write"
        glob: ["*.py"]
      - name: post-regex-sample
        # Demonstrates regex filtering on FILES_CHANGED
        run: echo "Matched regex on ${FILES_CHANGED}"
        only: ${FILES_CHANGED} regex ".*\\.py$"
  UserPromptSubmit:
    jobs:
      - name: user-prompt-sample
        run: echo "UserPrompt ${USER_PROMPT}"
  Notification:
    jobs:
      - name: notification-sample
        run: echo "Notification EVENT=${EVENT_NAME}"
  Stop:
    jobs:
      - name: stop-sample
        run: echo "Stop EVENT=${EVENT_NAME}"
  SubagentStop:
    jobs:
      - name: subagent-stop-sample
        run: echo "SubagentStop EVENT=${EVENT_NAME}"
  PreCompact:
    jobs:
      - name: precompact-sample
        run: echo "PreCompact EVENT=${EVENT_NAME}"
  SessionStart:
    jobs:
      - name: session-start-sample
        run: echo "SessionStart EVENT=${EVENT_NAME}"
  SessionEnd:
    jobs:
      - name: session-end-sample
        run: echo "SessionEnd EVENT=${EVENT_NAME}"
`, group, group)
			}

			// If --name provided, create .claude/hooks/<name>.yml; else .claude/hooks.yml
			var path string
			if fileName != "" {
				dir, err := config.EnsureClaudeDir(global)
				if err != nil {
					return err
				}
				hooksDir := filepath.Join(dir, "hooks")
				if err := os.MkdirAll(hooksDir, 0o750); err != nil {
					return err
				}
				// sanitize minimal: ensure .yml extension
				base := fileName
				if !strings.HasSuffix(strings.ToLower(base), ".yml") && !strings.HasSuffix(strings.ToLower(base), ".yaml") {
					base = base + ".yml"
				}
				target := filepath.Join(hooksDir, base)
				if !overwrite {
					if _, err := os.Stat(target); err == nil {
						fmt.Printf("File already exists: %s (use --overwrite to replace)\n", target)
						return nil
					}
				}
				if err := os.WriteFile(target, []byte(sample), 0o600); err != nil {
					return err
				}
				path = target
			} else {
				if global {
					// For global configs, create minimal config directly
					configPath, err := config.GetLogConfigPath(global)
					if err != nil {
						return err
					}

					// Load existing config or create default
					logCfg, err := config.LoadLogConfig(configPath)
					if err != nil {
						return err
					}

					// Check for existing config without overwrite
					if !overwrite && logCfg.CustomHooks != nil && len(logCfg.CustomHooks) > 0 {
						fmt.Printf("File already exists: %s (use --overwrite to replace)\n", configPath)
						return nil
					}

					// Create minimal hooks structure (empty)
					logCfg.CustomHooks = config.CustomHooksConfig{}

					// Ensure directory exists
					if err := os.MkdirAll(filepath.Dir(configPath), 0o750); err != nil {
						return err
					}

					// Save the minimal config
					if err := config.SaveLogConfig(configPath, logCfg); err != nil {
						return err
					}
					path = configPath
				} else {
					// For project configs, use existing sample logic
					var werr error
					path, werr = config.WriteSampleHooksConfig(global, sample, overwrite)
					if werr != nil {
						return werr
					}
				}
			}

			fmt.Printf("Created sample hooks config at %s\n", path)
			return nil
		},
	}
}

// newHooksCustomValidateCommand creates the custom hooks validate command
func newHooksCustomValidateCommand() *cli.Command {
	return &cli.Command{
		Name:  "validate",
		Usage: "Validate hooks.yml syntax",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			cfg, err := config.LoadHooksConfig()
			if err != nil {
				return fmt.Errorf("load error: %v", err)
			}
			if err := config.ValidateHooksConfig(cfg); err != nil {
				return fmt.Errorf("invalid hooks config: %v", err)
			}
			fmt.Println("hooks config is valid")
			return nil
		},
	}
}

// newHooksCustomShowCommand creates the custom hooks show command
func newHooksCustomShowCommand() *cli.Command {
	return &cli.Command{
		Name:  "show",
		Usage: "Display the effective custom hooks configuration",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "format", Aliases: []string{"f"}, Value: "yaml", Usage: "Output format: yaml or json"},
			&cli.BoolFlag{Name: "global", Aliases: []string{"g"}, Usage: "Prefer global config when showing embedded sections"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			// Load merged hooks config (project over global, including embedded and legacy)
			hooksCfg, err := config.LoadHooksConfig()
			if err != nil {
				return fmt.Errorf("load hooks config: %v", err)
			}

			// Load embedded blocked URLs for display (prefer project unless --global)
			useGlobal := cmd.Bool("global")
			cfgPath, err := config.GetLogConfigPath(useGlobal)
			if err != nil {
				return fmt.Errorf("get config path: %v", err)
			}
			logCfg, err := config.LoadLogConfig(cfgPath)
			if err != nil {
				return fmt.Errorf("load main config: %v", err)
			}

			// Build output view
			out := map[string]interface{}{
				"customHooks": hooksCfg,
			}
			if len(logCfg.BlockedURLs) > 0 {
				out["blockedUrls"] = logCfg.BlockedURLs
			}

			switch strings.ToLower(cmd.String("format")) {
			case "json":
				b, err := json.MarshalIndent(out, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(b))
			default:
				b, err := yaml.Marshal(out)
				if err != nil {
					return err
				}
				fmt.Print(string(b))
			}
			return nil
		},
	}
}

// newHooksCustomBlockedCommand creates the custom hooks blocked command
func newHooksCustomBlockedCommand() *cli.Command {
	return &cli.Command{
		Name:  "blocked",
		Usage: "Manage blocked URL prefixes used by fetch-blocker",
		Commands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List blocked URL prefixes",
				Flags: []cli.Flag{&cli.BoolFlag{Name: "global", Aliases: []string{"g"}}},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					global := cmd.Bool("global")
					path, err := config.GetLogConfigPath(global)
					if err != nil {
						return err
					}
					lc, err := config.LoadLogConfig(path)
					if err != nil {
						return err
					}
					scope := "project"
					if global {
						scope = "global"
					}
					fmt.Printf("Blocked URLs (%s config: %s):\n", scope, path)
					if len(lc.BlockedURLs) == 0 {
						fmt.Println("(none)")
						return nil
					}
					for _, b := range lc.BlockedURLs {
						if b.Suggestion != "" {
							fmt.Printf("- %s | %s\n", b.Prefix, b.Suggestion)
						} else {
							fmt.Printf("- %s\n", b.Prefix)
						}
					}
					return nil
				},
			},
			{
				Name:  "add",
				Usage: "Add a blocked URL prefix",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "global", Aliases: []string{"g"}},
					&cli.StringFlag{Name: "suggestion", Aliases: []string{"s"}},
				},
				ArgsUsage: "<prefix>",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					args := cmd.Args().Slice()
					if len(args) != 1 {
						return fmt.Errorf("exactly one argument required: <prefix>")
					}
					prefix := strings.TrimSpace(args[0])
					if prefix == "" {
						return fmt.Errorf("prefix cannot be empty")
					}
					global := cmd.Bool("global")
					suggestion := cmd.String("suggestion")
					path, err := config.GetLogConfigPath(global)
					if err != nil {
						return err
					}
					lc, err := config.LoadLogConfig(path)
					if err != nil {
						return err
					}
					// Check duplicate
					for _, b := range lc.BlockedURLs {
						if b.Prefix == prefix {
							fmt.Println("Prefix already present; no change.")
							return nil
						}
					}
					lc.BlockedURLs = append(lc.BlockedURLs, config.BlockedURL{Prefix: prefix, Suggestion: suggestion})
					if err := config.SaveLogConfig(path, lc); err != nil {
						return err
					}
					fmt.Printf("Added blocked prefix to %s: %s\n", path, prefix)
					return nil
				},
			},
			{
				Name:      "remove",
				Usage:     "Remove a blocked URL prefix",
				Flags:     []cli.Flag{&cli.BoolFlag{Name: "global", Aliases: []string{"g"}}},
				ArgsUsage: "<prefix>",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					args := cmd.Args().Slice()
					if len(args) != 1 {
						return fmt.Errorf("exactly one argument required: <prefix>")
					}
					prefix := strings.TrimSpace(args[0])
					global := cmd.Bool("global")
					path, err := config.GetLogConfigPath(global)
					if err != nil {
						return err
					}
					lc, err := config.LoadLogConfig(path)
					if err != nil {
						return err
					}
					filtered := lc.BlockedURLs[:0]
					removed := false
					for _, b := range lc.BlockedURLs {
						if b.Prefix == prefix {
							removed = true
							continue
						}
						filtered = append(filtered, b)
					}
					if !removed {
						fmt.Println("Prefix not found; no change.")
						return nil
					}
					lc.BlockedURLs = filtered
					if err := config.SaveLogConfig(path, lc); err != nil {
						return err
					}
					fmt.Printf("Removed blocked prefix from %s: %s\n", path, prefix)
					return nil
				},
			},
			{
				Name:  "clear",
				Usage: "Clear all blocked URL prefixes",
				Flags: []cli.Flag{&cli.BoolFlag{Name: "global", Aliases: []string{"g"}}},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					global := cmd.Bool("global")
					path, err := config.GetLogConfigPath(global)
					if err != nil {
						return err
					}
					lc, err := config.LoadLogConfig(path)
					if err != nil {
						return err
					}
					if len(lc.BlockedURLs) == 0 {
						fmt.Println("Blocked URLs already empty; no change.")
						return nil
					}
					lc.BlockedURLs = nil
					if err := config.SaveLogConfig(path, lc); err != nil {
						return err
					}
					fmt.Printf("Cleared blocked URLs in %s\n", path)
					return nil
				},
			},
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

// printHookMatchers displays the matchers for a specific event in a formatted list
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

// printUninstallExamples shows example commands for uninstalling hooks
func printUninstallExamples(global bool) {
	scope := "project"
	globalFlag := ""
	if global {
		scope = "global"
		globalFlag = " --global"
	}

	fmt.Printf("📝 How to remove hooks:\n\n")

	fmt.Printf("Remove a specific hook type (removes ALL instances):\n")
	fmt.Printf("  blues-traveler hooks uninstall debug%s\n", globalFlag)
	fmt.Printf("  blues-traveler hooks uninstall security%s\n", globalFlag)
	fmt.Printf("  blues-traveler hooks uninstall audit%s\n\n", globalFlag)

	fmt.Printf("Remove ALL blues-traveler hooks (preserves other hooks):\n")
	fmt.Printf("  blues-traveler hooks uninstall all%s\n\n", globalFlag)

	fmt.Printf("Remove ALL hooks from %s settings:\n", scope)
	if global {
		fmt.Printf("  rm ~/.claude/settings.json\n")
		fmt.Printf("  # (or edit the file manually to remove specific events)\n\n")
	} else {
		fmt.Printf("  rm .claude/settings.json\n")
		fmt.Printf("  # (or edit the file manually to remove specific events)\n\n")
	}

	fmt.Printf("View this list again:\n")
	fmt.Printf("  %s hooks list --installed%s\n\n", "blues-traveler", globalFlag)

	fmt.Printf("💡 Note: The 'uninstall' command removes ALL instances of a hook type\n")
	fmt.Printf("   from ALL events (PreToolUse, PostToolUse, etc.)\n")
}

// uninstallAllBluesTravelerHooks removes all blues-traveler hooks from settings
func uninstallAllBluesTravelerHooks(global bool, skipConfirmation bool) {
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

// newPlatformFromType creates a Platform instance from a Type
// This helper exists in cmd to avoid import cycles in the platform package
func newPlatformFromType(t platform.Type) platform.Platform {
	switch t {
	case platform.Cursor:
		return cursor.New()
	case platform.ClaudeCode:
		return claude.New()
	default:
		return claude.New()
	}
}

// runHookCursorMode runs a hook in Cursor mode (JSON I/O over stdin/stdout)
func runHookCursorMode(key string, getPlugin func(string) (interface {
	Run() error
	Description() string
}, bool), isPluginEnabled func(string) bool,
) error {
	// Read JSON input from stdin
	var input cursor.HookInput
	decoder := json.NewDecoder(os.Stdin)

	if err := decoder.Decode(&input); err != nil {
		// Output error as JSON response
		output := cursor.HookOutput{
			Permission:  "deny",
			UserMessage: fmt.Sprintf("Failed to parse hook input: %v", err),
		}
		_ = json.NewEncoder(os.Stdout).Encode(output)
		os.Exit(3)
	}

	// Validate plugin exists
	p, exists := getPlugin(key)
	if !exists {
		output := cursor.HookOutput{
			Permission:  "deny",
			UserMessage: fmt.Sprintf("Hook '%s' not found", key),
		}
		_ = json.NewEncoder(os.Stdout).Encode(output)
		os.Exit(3)
	}

	// Check if plugin is enabled
	if !isPluginEnabled(key) {
		// Disabled hooks should allow (not interfere)
		output := cursor.HookOutput{
			Permission: "allow",
		}
		_ = json.NewEncoder(os.Stdout).Encode(output)
		return nil
	}

	// Set up environment variables from JSON input
	setupCursorEnvironment(input)

	// Execute the hook in Cursor-specific mode
	// We call the hook directly but need to handle the response ourselves
	// since we can't use the cchooks Runner (it reads stdin which we've consumed)
	output := executeCursorHook(p, input)

	// Write JSON response to stdout
	encoder := json.NewEncoder(os.Stdout)
	if err := encoder.Encode(output); err != nil {
		// Can't write response, exit with error
		os.Exit(3)
	}

	// Exit with code 3 if denied
	if output.Permission == "deny" {
		os.Exit(3)
	}

	return nil
}

// installHookClaudeCode installs a hook for Claude Code platform
func installHookClaudeCode(hookType string, plugin interface {
	Run() error
	Description() string
}, global bool, event string, matcher string, timeoutFlag int, logEnabled bool, logFormat string, isValidEventType func(string) bool,
) error {
	// Validate event
	if !isValidEventType(event) {
		return fmt.Errorf("invalid event '%s' for Claude Code", event)
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
		return fmt.Errorf("error getting settings path: %v", err)
	}

	// Load existing settings
	settings, err := config.LoadSettings(settingsPath)
	if err != nil {
		return fmt.Errorf("error loading settings: %v", err)
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
			return fmt.Errorf("error saving settings: %v", err)
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
}

// installHookCursor installs a hook for Cursor platform
func installHookCursor(p platform.Platform, hookType string, plugin interface {
	Run() error
	Description() string
}, event string, matcher string,
) error {
	// Validate event is supported by Cursor
	genericEvent := core.EventType(event)
	if !p.SupportsEvent(genericEvent) {
		return fmt.Errorf("event '%s' is not supported by Cursor platform", event)
	}

	// Map generic event to Cursor-specific events
	cursorEvents := p.MapEventFromGeneric(genericEvent)
	if len(cursorEvents) == 0 {
		return fmt.Errorf("event '%s' cannot be mapped to Cursor events", event)
	}

	// Get path to this executable
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}

	// Load Cursor config
	cursorPlatform, ok := p.(*cursor.CursorPlatform)
	if !ok {
		return fmt.Errorf("platform is not Cursor")
	}

	cfg, err := cursorPlatform.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load Cursor config: %w", err)
	}

	installedCount := 0
	installedEvents := []string{}

	// Build the command to register in Cursor hooks.json
	// Format: /path/to/blues-traveler hooks run <hook> --cursor-mode
	// Note: Matcher filtering happens inside blues-traveler (passed via stdin JSON matching logic)
	command := fmt.Sprintf("%s hooks run %s --cursor-mode", execPath, hookType)

	// Install hook for each mapped Cursor event
	for _, cursorEvent := range cursorEvents {
		// Check if hook already exists in config
		if !cfg.HasHook(cursorEvent, command) {
			// Add hook to Cursor config
			cfg.AddHook(cursorEvent, command)
			installedCount++
			installedEvents = append(installedEvents, cursorEvent)
		}
	}

	// Save Cursor config if changes were made
	if installedCount > 0 {
		if err := cursorPlatform.SaveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save Cursor config: %w", err)
		}

		configPath, _ := cursorPlatform.ConfigPath()
		fmt.Printf("✅ Successfully installed %s hook for Cursor\n", hookType)
		fmt.Printf("   Generic Event: %s\n", event)
		fmt.Printf("   Cursor Events: %s\n", strings.Join(installedEvents, ", "))
		fmt.Printf("   Command: %s\n", command)
		fmt.Printf("   Config: %s\n", configPath)
		fmt.Println()
		fmt.Println("The hook will be active in new Cursor sessions.")
	} else {
		fmt.Printf("⚠️  Hook '%s' is already installed for all mapped Cursor events\n", hookType)
		fmt.Println("No changes made.")
	}

	return nil
}

// executeCursorHook executes a hook in Cursor mode without using the cchooks Runner
// which would try to read from stdin (already consumed)
func executeCursorHook(hook interface {
	Run() error
	Description() string
}, input cursor.HookInput,
) cursor.HookOutput {
	// Transform Cursor JSON to Claude Code format
	claudeJSON, err := transformCursorToClaudeFormat(input)
	if err != nil {
		return cursor.HookOutput{
			Permission:  "deny",
			UserMessage: fmt.Sprintf("Failed to transform input: %v", err),
		}
	}

	// Set up a pipe to feed the transformed JSON to the hook's stdin
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Create a pipe with the transformed JSON input
	r, w, err := os.Pipe()
	if err != nil {
		return cursor.HookOutput{
			Permission:  "deny",
			UserMessage: fmt.Sprintf("Failed to create pipe: %v", err),
		}
	}
	defer func() { _ = r.Close() }() // Ignore close error in defer

	// Replace stdin with our pipe
	os.Stdin = r

	// Write JSON in a goroutine
	go func() {
		defer func() { _ = w.Close() }() // Ignore close error in defer
		_, _ = w.Write(claudeJSON)       // Ignore write error, hook will fail if needed
	}()

	// Execute the hook
	err = hook.Run()
	// Convert error to Cursor response
	if err != nil {
		return cursor.HookOutput{
			Permission:  "deny",
			UserMessage: err.Error(),
		}
	}

	return cursor.HookOutput{
		Permission: "allow",
	}
}

// transformCursorToClaudeFormat converts Cursor JSON to Claude Code format
// that cchooks can understand
func transformCursorToClaudeFormat(input cursor.HookInput) ([]byte, error) {
	// Base event structure
	event := map[string]interface{}{
		"conversation_id": input.ConversationID,
		"generation_id":   input.GenerationID,
		"workspace_roots": input.WorkspaceRoots,
	}

	// Map Cursor event to Claude Code event and add event-specific fields
	switch input.HookEventName {
	case cursor.BeforeShellExecution:
		event["hook_event_name"] = "PreToolUse"
		event["tool_name"] = "Bash"
		event["tool_parameters"] = map[string]interface{}{
			"command": input.Command,
		}
		if input.CWD != "" {
			event["cwd"] = input.CWD
		}

	case cursor.BeforeMCPExecution:
		event["hook_event_name"] = "PreToolUse"
		event["tool_name"] = input.ToolName
		// Parse tool_input if it's JSON
		var toolParams interface{}
		if err := json.Unmarshal([]byte(input.ToolInput), &toolParams); err == nil {
			event["tool_parameters"] = toolParams
		} else {
			event["tool_parameters"] = map[string]interface{}{
				"input": input.ToolInput,
			}
		}
		if input.URL != "" {
			event["mcp_url"] = input.URL
		}

	case cursor.AfterFileEdit:
		event["hook_event_name"] = "PostToolUse"
		event["tool_name"] = "Edit"
		event["tool_parameters"] = map[string]interface{}{
			"file_path": input.FilePath,
			"edits":     input.Edits,
		}

	case cursor.BeforeReadFile:
		event["hook_event_name"] = "PreToolUse"
		event["tool_name"] = "Read"
		event["tool_parameters"] = map[string]interface{}{
			"file_path": input.FilePath,
		}

	case cursor.BeforeSubmitPrompt:
		event["hook_event_name"] = "UserPromptSubmit"
		event["user_prompt"] = input.Prompt
		event["attachments"] = input.Attachments

	case cursor.Stop:
		event["hook_event_name"] = "Stop"
		event["status"] = input.Status

	default:
		return nil, fmt.Errorf("unsupported Cursor event: %s", input.HookEventName)
	}

	return json.Marshal(event)
}

// setupCursorEnvironment maps Cursor JSON input to environment variables
// that blues-traveler hooks expect
func setupCursorEnvironment(input cursor.HookInput) {
	// Common fields (ignore errors as these are best-effort environment setup)
	_ = os.Setenv("CONVERSATION_ID", input.ConversationID)
	_ = os.Setenv("GENERATION_ID", input.GenerationID)
	_ = os.Setenv("EVENT_NAME", input.HookEventName)
	_ = os.Setenv("WORKSPACE_ROOTS", strings.Join(input.WorkspaceRoots, ":"))

	// Event-specific fields (all Setenv calls ignore errors - best effort)
	switch input.HookEventName {
	case cursor.BeforeShellExecution:
		_ = os.Setenv("TOOL_NAME", "Bash")
		_ = os.Setenv("TOOL_ARGS", input.Command)
		_ = os.Setenv("CWD", input.CWD)

	case cursor.BeforeMCPExecution:
		_ = os.Setenv("TOOL_NAME", input.ToolName)
		_ = os.Setenv("TOOL_ARGS", input.ToolInput)
		if input.URL != "" {
			_ = os.Setenv("MCP_URL", input.URL)
		}

	case cursor.AfterFileEdit:
		_ = os.Setenv("TOOL_NAME", "Edit")
		_ = os.Setenv("FILE_PATH", input.FilePath)
		if editsJSON, err := json.Marshal(input.Edits); err == nil {
			_ = os.Setenv("FILE_EDITS", string(editsJSON))
		}

	case cursor.BeforeReadFile:
		_ = os.Setenv("TOOL_NAME", "Read")
		_ = os.Setenv("FILE_PATH", input.FilePath)
		_ = os.Setenv("FILE_CONTENT", input.Content)

	case cursor.BeforeSubmitPrompt:
		_ = os.Setenv("USER_PROMPT", input.Prompt)
		if attachmentsJSON, err := json.Marshal(input.Attachments); err == nil {
			_ = os.Setenv("PROMPT_ATTACHMENTS", string(attachmentsJSON))
		}

	case cursor.Stop:
		_ = os.Setenv("STOP_STATUS", input.Status)
	}
}
