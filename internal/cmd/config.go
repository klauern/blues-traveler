package cmd

import (
	"context"
	"fmt"

	"github.com/klauern/blues-traveler/internal/config"
	"github.com/urfave/cli/v3"
)

func NewConfigLogCmd() *cli.Command {
	return &cli.Command{
		Name:        "config-log",
		Usage:       "Configure log rotation settings",
		Description: `Configure log rotation settings including maximum age, file size, and backup count.`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "global",
				Aliases: []string{"g"},
				Value:   false,
				Usage:   "Configure global settings (~/.claude/settings.json)",
			},
			&cli.IntFlag{
				Name:    "max-age",
				Aliases: []string{"a"},
				Value:   0,
				Usage:   "Maximum age in days to retain log files (default: 30)",
			},
			&cli.IntFlag{
				Name:    "max-size",
				Aliases: []string{"s"},
				Value:   0,
				Usage:   "Maximum size in MB per log file before rotation (default: 10)",
			},
			&cli.IntFlag{
				Name:    "max-backups",
				Aliases: []string{"b"},
				Value:   0,
				Usage:   "Maximum number of backup files to retain (default: 5)",
			},
			&cli.BoolFlag{
				Name:    "compress",
				Aliases: []string{"c"},
				Value:   false,
				Usage:   "Compress rotated log files",
			},
			&cli.BoolFlag{
				Name:  "show",
				Value: false,
				Usage: "Show current log rotation settings",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			global := cmd.Bool("global")
			maxAge := cmd.Int("max-age")
			maxSize := cmd.Int("max-size")
			maxBackups := cmd.Int("max-backups")
			compress := cmd.Bool("compress")
			show := cmd.Bool("show")

			configPath, err := config.GetLogConfigPath(global)
			if err != nil {
				return fmt.Errorf("error getting config path: %v", err)
			}

			logConfig, err := config.LoadLogConfig(configPath)
			if err != nil {
				return fmt.Errorf("error loading config: %v", err)
			}

			if show {
				// Show current log rotation settings
				scope := "project"
				if global {
					scope = "global"
				}
				fmt.Printf("Current log rotation settings (%s: %s):\n", scope, configPath)
				fmt.Printf("  Max Age: %d days\n", logConfig.LogRotation.MaxAge)
				fmt.Printf("  Max Size: %d MB\n", logConfig.LogRotation.MaxSize)
				fmt.Printf("  Max Backups: %d files\n", logConfig.LogRotation.MaxBackups)
				fmt.Printf("  Compress: %t\n", logConfig.LogRotation.Compress)
				return nil
			}

			// Only update non-zero values
			if maxAge > 0 {
				logConfig.LogRotation.MaxAge = maxAge
			}
			if maxSize > 0 {
				logConfig.LogRotation.MaxSize = maxSize
			}
			if maxBackups > 0 {
				logConfig.LogRotation.MaxBackups = maxBackups
			}
			// Note: urfave/cli v3 doesn't have Changed() method, so we check compress directly
			// This means compress will be set to false if not explicitly provided
			if compress {
				logConfig.LogRotation.Compress = compress
			}

			if err := config.SaveLogConfig(configPath, logConfig); err != nil {
				return fmt.Errorf("error saving config: %v", err)
			}

			scope := "project"
			if global {
				scope = "global"
			}
			fmt.Printf("Log rotation configuration updated in %s config (%s):\n", scope, configPath)
			fmt.Printf("  Max Age: %d days\n", logConfig.LogRotation.MaxAge)
			fmt.Printf("  Max Size: %d MB\n", logConfig.LogRotation.MaxSize)
			fmt.Printf("  Max Backups: %d files\n", logConfig.LogRotation.MaxBackups)
			fmt.Printf("  Compress: %t\n", logConfig.LogRotation.Compress)
			return nil
		},
	}
}
