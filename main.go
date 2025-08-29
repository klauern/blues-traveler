package main

import (
	"context"
	"fmt"
	"os"

	"github.com/klauern/blues-traveler/internal/cmd"
	"github.com/klauern/blues-traveler/internal/compat"
	"github.com/klauern/blues-traveler/internal/core"
	_ "github.com/klauern/blues-traveler/internal/hooks" // Import for init() registration
	"github.com/urfave/cli/v3"
)

func main() {
	// Create wrapper functions for compatibility
	getPluginWrapper := func(key string) (interface {
		Run() error
		Description() string
	}, bool,
	) {
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

	// Create the root command with urfave/cli v3
	app := &cli.Command{
		Name:  "blues-traveler",
		Usage: "Claude Code hook runner and manager - 'The hook brings you back'",
		Description: `A CLI tool that runs Claude Code hooks directly and manages hook installations.
Like the classic Blues Traveler song, our hooks will bring you back to clean, secure, and well-formatted code.`,
		Commands: []*cli.Command{
			cmd.NewListCmd(getPluginWrapper, compat.PluginKeys),
			cmd.NewRunCmd(getPluginWrapper, compat.IsPluginEnabled, compat.PluginKeys),
			cmd.NewInstallCmd(getPluginWrapper, compat.PluginKeys, core.IsValidEventType, core.ValidEventTypes),
			cmd.NewUninstallCmd(),
			cmd.NewListInstalledCmd(),
			cmd.NewListEventsCmd(eventsWrapper),
			cmd.NewGenerateCmd(),
			cmd.NewConfigLogCmd(),
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}
