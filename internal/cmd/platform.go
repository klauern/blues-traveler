package cmd

import (
	"context"
	"fmt"

	"github.com/klauern/blues-traveler/internal/platform"
	"github.com/klauern/blues-traveler/internal/platform/claude"
	"github.com/klauern/blues-traveler/internal/platform/cursor"
	"github.com/urfave/cli/v3"
)

// NewPlatformCmd creates the platform command with subcommands
func NewPlatformCmd() *cli.Command {
	return &cli.Command{
		Name:        "platform",
		Usage:       "Platform detection and information",
		Description: `Detect and display information about the current IDE platform (Claude Code or Cursor).`,
		Commands: []*cli.Command{
			newPlatformDetectCommand(),
			newPlatformInfoCommand(),
		},
	}
}

// newPlatformDetectCommand creates the detect subcommand
func newPlatformDetectCommand() *cli.Command {
	return &cli.Command{
		Name:        "detect",
		Usage:       "Auto-detect the current platform",
		Description: `Automatically detect which IDE platform is in use (Claude Code or Cursor).`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Value:   false,
				Usage:   "Show detection steps and reasoning",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			verbose := cmd.Bool("verbose")

			detector := platform.NewDetector()

			if verbose {
				fmt.Println("Detecting platform...")
				fmt.Println()
			}

			platformType, err := detector.DetectType()
			if err != nil {
				return fmt.Errorf("failed to detect platform: %w", err)
			}

			// Create platform instance to get full info
			var p platform.Platform
			switch platformType {
			case platform.Cursor:
				p = cursor.New()
			case platform.ClaudeCode:
				p = claude.New()
			default:
				return fmt.Errorf("unknown platform type: %s", platformType)
			}

			// Display results
			fmt.Printf("Platform: %s\n", p.Name())
			fmt.Printf("Type: %s\n", p.Type())

			configPath, err := p.ConfigPath()
			if err == nil {
				fmt.Printf("Config Path: %s\n", configPath)
			}

			if verbose {
				fmt.Println()
				fmt.Println("Supported Events:")
				for _, event := range p.AllEvents() {
					fmt.Printf("  - %s: %s\n", event.Name, event.Description)
				}
			}

			return nil
		},
	}
}

// newPlatformInfoCommand creates the info subcommand
func newPlatformInfoCommand() *cli.Command {
	return &cli.Command{
		Name:      "info",
		Usage:     "Show detailed information about a platform",
		ArgsUsage: "[platform]",
		Description: `Show detailed information about a specific platform (claudecode or cursor).
If no platform is specified, shows info for the detected platform.`,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			args := cmd.Args().Slice()

			var p platform.Platform

			if len(args) == 0 {
				// Auto-detect
				detector := platform.NewDetector()
				platformType, err := detector.DetectType()
				if err != nil {
					return fmt.Errorf("failed to detect platform: %w", err)
				}

				switch platformType {
				case platform.Cursor:
					p = cursor.New()
				case platform.ClaudeCode:
					p = claude.New()
				default:
					return fmt.Errorf("unknown platform type: %s", platformType)
				}
			} else {
				// Parse specified platform
				platformType, err := platform.TypeFromString(args[0])
				if err != nil {
					return fmt.Errorf("invalid platform: %w", err)
				}

				switch platformType {
				case platform.Cursor:
					p = cursor.New()
				case platform.ClaudeCode:
					p = claude.New()
				default:
					return fmt.Errorf("unknown platform type: %s", platformType)
				}
			}

			// Display detailed information
			fmt.Printf("Platform: %s\n", p.Name())
			fmt.Printf("Type: %s\n", p.Type())
			fmt.Println()

			configPath, err := p.ConfigPath()
			if err != nil {
				fmt.Printf("Config Path: Error - %v\n", err)
			} else {
				fmt.Printf("Config Path: %s\n", configPath)
			}
			fmt.Println()

			fmt.Println("Supported Events:")
			events := p.AllEvents()
			for _, event := range events {
				fmt.Printf("  • %s\n", event.Name)
				fmt.Printf("    Description: %s\n", event.Description)
				fmt.Printf("    Generic Event: %s\n", event.GenericEvent)
				fmt.Printf("    Requires Stdio: %v\n", event.RequiresStdio)
				fmt.Printf("    Supports Config Filters: %v\n", event.SupportsFilter)
				fmt.Println()
			}

			return nil
		},
	}
}
