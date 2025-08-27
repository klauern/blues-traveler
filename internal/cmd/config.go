package cmd

import (
	"fmt"
	"os"

	"github.com/klauern/klauer-hooks/internal/config"
	"github.com/spf13/cobra"
)

func NewConfigLogCmd() *cobra.Command {
	cmd := &cobra.Command{
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

			configPath, err := config.GetLogConfigPath(global)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting config path: %v\n", err)
				os.Exit(1)
			}

			logConfig, err := config.LoadLogConfig(configPath)
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
				fmt.Printf("  Max Age: %d days\n", logConfig.LogRotation.MaxAge)
				fmt.Printf("  Max Size: %d MB\n", logConfig.LogRotation.MaxSize)
				fmt.Printf("  Max Backups: %d files\n", logConfig.LogRotation.MaxBackups)
				fmt.Printf("  Compress: %t\n", logConfig.LogRotation.Compress)
				return
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
			if cmd.Flags().Changed("compress") {
				logConfig.LogRotation.Compress = compress
			}

			if err := config.SaveLogConfig(configPath, logConfig); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
				os.Exit(1)
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
		},
	}

	// Add flags for config-log command
	cmd.Flags().BoolP("global", "g", false, "Configure global settings (~/.claude/settings.json)")
	cmd.Flags().IntP("max-age", "a", 0, "Maximum age in days to retain log files (default: 30)")
	cmd.Flags().IntP("max-size", "s", 0, "Maximum size in MB per log file before rotation (default: 10)")
	cmd.Flags().IntP("max-backups", "b", 0, "Maximum number of backup files to retain (default: 5)")
	cmd.Flags().BoolP("compress", "c", false, "Compress rotated log files")
	cmd.Flags().BoolP("show", "", false, "Show current log rotation settings")

	return cmd
}
