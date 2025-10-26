package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/klauern/blues-traveler/internal/config"
	"github.com/urfave/cli/v3"
	yaml "gopkg.in/yaml.v3"
)

// newHooksCustomCommand creates the custom hooks command group
func newHooksCustomCommand(isValidEventType func(string) bool, validEventTypes func() []string) *cli.Command {
	return &cli.Command{
		Name:        "custom",
		Usage:       "Manage custom hooks from hooks.yml",
		Description: `Manage custom hooks defined in .claude/hooks.yml configuration files.`,
		Commands: []*cli.Command{
			newHooksCustomInstallCommand(isValidEventType, validEventTypes),
			newHooksCustomListCommand(),
			newHooksCustomSyncCommand(isValidEventType, validEventTypes),
			newHooksCustomInitCommand(),
			newHooksCustomValidateCommand(),
			newHooksCustomShowCommand(),
			newHooksCustomBlockedCommand(),
		},
	}
}

// newHooksCustomListCommand creates the list command for custom hooks
func newHooksCustomListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List available custom hook groups",
		Action: func(_ context.Context, _ *cli.Command) error {
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

// newHooksCustomValidateCommand creates the validate command for custom hooks
func newHooksCustomValidateCommand() *cli.Command {
	return &cli.Command{
		Name:  "validate",
		Usage: "Validate hooks.yml syntax",
		Action: func(_ context.Context, _ *cli.Command) error {
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

// newHooksCustomShowCommand creates the show command for custom hooks
func newHooksCustomShowCommand() *cli.Command {
	return &cli.Command{
		Name:  "show",
		Usage: "Display the effective custom hooks configuration",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "format", Aliases: []string{"f"}, Value: "yaml", Usage: "Output format: yaml or json"},
			&cli.BoolFlag{Name: "global", Aliases: []string{"g"}, Usage: "Prefer global config when showing embedded sections"},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
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
