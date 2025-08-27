package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/klauern/klauer-hooks/internal/config"
	"github.com/klauern/klauer-hooks/internal/core"
	"github.com/spf13/cobra"
)

func NewRunCmd(getPlugin func(string) (interface {
	Run() error
	Description() string
}, bool), isPluginEnabled func(string) bool, pluginKeys func() []string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [plugin-key]",
		Short: "Run a specific hook plugin",
		Long:  `Run a specific hook plugin. Executes only that hook's handlers (no unified pipeline).`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			key := args[0]

			// Validate plugin exists early
			p, exists := getPlugin(key)
			if !exists {
				fmt.Fprintf(os.Stderr, "Error: Plugin '%s' not found.\n", key)
				fmt.Fprintf(os.Stderr, "Available plugins: %s\n", strings.Join(pluginKeys(), ", "))
				os.Exit(1)
			}

			// Enablement check before side effects
			if !isPluginEnabled(key) {
				fmt.Printf("Plugin '%s' is disabled via settings. Nothing to do.\n", key)
				return
			}

			// Logging flags
			logEnabled, _ := cmd.Flags().GetBool("log")
			logFormat, _ := cmd.Flags().GetString("log-format")
			if logFormat == "" {
				logFormat = config.LoggingFormatJSONL
			}
			if logEnabled && !config.IsValidLoggingFormat(logFormat) {
				fmt.Fprintf(os.Stderr, "Error: Invalid --log-format '%s'. Valid: jsonl, pretty\n", logFormat)
				os.Exit(1)
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
				fmt.Fprintf(os.Stderr, "Hook '%s' failed: %v\n", key, err)
				os.Exit(1)
			}
		},
	}

	// Add flags for run command
	cmd.Flags().BoolP("log", "l", false, "Enable detailed logging to .claude/hooks/<plugin-key>.log")
	cmd.Flags().String("log-format", "jsonl", "Log output format: jsonl or pretty (default jsonl)")

	return cmd
}

// Plugin interface for compatibility
type Plugin interface {
	Run() error
}
