package main

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/klauern/blues-traveler/internal/cmd"
	"github.com/klauern/blues-traveler/internal/compat"
	"github.com/klauern/blues-traveler/internal/core"
	_ "github.com/klauern/blues-traveler/internal/hooks" // Import for init() registration
	"github.com/urfave/cli/v3"
)

// Version information injected at build time
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
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

	// Create version info
	versionInfo := cmd.VersionInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
		GoVer:   fmt.Sprintf("%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH),
	}

	// Create the root command with urfave/cli v3
	app := &cli.Command{
		Name:  "blues-traveler",
		Usage: "Claude Code hook runner and manager - 'The hook brings you back'",
		Description: `A CLI tool that runs Claude Code hooks directly and manages hook installations.
Like the classic Blues Traveler song, our hooks will bring you back to clean, secure, and well-formatted code.`,
		Commands: []*cli.Command{
			cmd.NewHooksCommand(getPluginWrapper, compat.IsPluginEnabled, compat.PluginKeys, core.IsValidEventType, core.ValidEventTypes, eventsWrapper),
			cmd.NewDoctorCommand(),
			cmd.NewConfigCmd(),
			cmd.NewGenerateCmd(),
			cmd.NewVersionCmd(versionInfo),
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}
