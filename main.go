package main

import (
	"fmt"
	"os"

	"github.com/klauern/blues-traveler/internal/cmd"
	"github.com/klauern/blues-traveler/internal/compat"
	"github.com/klauern/blues-traveler/internal/core"
	_ "github.com/klauern/blues-traveler/internal/hooks" // Import for init() registration
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "blues-traveler",
	Short: "Claude Code hook runner and manager - 'The hook brings you back'",
	Long: `A CLI tool that runs Claude Code hooks directly and manages hook installations.
Like the classic Blues Traveler song, our hooks will bring you back to clean, secure, and well-formatted code.`,
}

func init() {
	// Create wrapper functions for compatibility
	getPluginWrapper := func(key string) (interface {
		Run() error
		Description() string
	}, bool) {
		p, exists := compat.GetPlugin(key)
		if !exists {
			return nil, false
		}
		return p, true
	}

	eventsWrapper := func() []cmd.ClaudeCodeEvent {
		events := core.AllClaudeCodeEvents()
		result := make([]cmd.ClaudeCodeEvent, len(events))
		for i, e := range events {
			result[i] = cmd.ClaudeCodeEvent{
				Type:               cmd.EventType(e.Type),
				Name:               e.Name,
				Description:        e.Description,
				SupportedByCCHooks: e.SupportedByCCHooks,
			}
		}
		return result
	}

	// Create command instances with dependency injection
	rootCmd.AddCommand(cmd.NewListCmd(getPluginWrapper, compat.PluginKeys))
	rootCmd.AddCommand(cmd.NewRunCmd(getPluginWrapper, compat.IsPluginEnabled, compat.PluginKeys))
	rootCmd.AddCommand(cmd.NewInstallCmd(getPluginWrapper, compat.PluginKeys, core.IsValidEventType, core.ValidEventTypes))
	rootCmd.AddCommand(cmd.NewUninstallCmd())
	rootCmd.AddCommand(cmd.NewListInstalledCmd())
	rootCmd.AddCommand(cmd.NewListEventsCmd(eventsWrapper))
	rootCmd.AddCommand(cmd.NewGenerateCmd())
	rootCmd.AddCommand(cmd.NewConfigLogCmd())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}
