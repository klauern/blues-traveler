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

func NewRunCmd(getPlugin func(string) (interface {
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
		Action: func(ctx context.Context, cmd *cli.Command) error {
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

// Plugin interface for compatibility
type Plugin interface {
	Run() error
}
