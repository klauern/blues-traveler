package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/klauern/blues-traveler/internal/config"
	"github.com/klauern/blues-traveler/internal/core"
	"github.com/urfave/cli/v3"
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
		Action: func(_ context.Context, cmd *cli.Command) error {
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
				return fmt.Errorf("exactly one argument required: [plugin-key]")
			}
			key := args[0]

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
				if err := setupHookLogging(key, logFormat); err != nil {
					return err
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

// setupHookLogging configures logging with rotation for hook execution
func setupHookLogging(hookKey, logFormat string) error {
	logConfig := config.GetLogRotationConfigFromFile(false)
	if logConfig.MaxAge == 0 && logConfig.MaxSize == 0 {
		logConfig = config.GetLogRotationConfigFromFile(true)
	}

	logPath := config.GetLogPath(hookKey)
	rotatingLogger := config.SetupLogRotation(logPath, logConfig)

	core.SetGlobalLoggingConfig(true, ".claude/hooks", logFormat)

	if rotatingLogger != nil {
		fmt.Printf("Logging enabled with rotation - output will be written to %s\n", logPath)
		fmt.Printf("Log rotation: max %d days, %dMB per file, %d backups\n",
			logConfig.MaxAge, logConfig.MaxSize, logConfig.MaxBackups)
		if err := config.CleanupOldLogs(filepath.Dir(logPath), logConfig.MaxAge); err != nil {
			fmt.Printf("Warning: Failed to cleanup old logs: %v\n", err)
		}
	} else {
		fmt.Printf("Logging enabled - output will be written to %s\n", logPath)
	}

	return nil
}
