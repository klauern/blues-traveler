package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/klauern/blues-traveler/internal/config"
	"github.com/urfave/cli/v3"
)

// newHooksCustomInstallCommand creates the install command for custom hooks
func newHooksCustomInstallCommand(isValidEventType func(string) bool, validEventTypes func() []string) *cli.Command {
	return &cli.Command{
		Name:  "install",
		Usage: "Install hooks from a named group defined in hooks.yml",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "global", Aliases: []string{"g"}, Usage: "Install to global settings"},
			&cli.StringFlag{Name: "event", Aliases: []string{"e"}, Usage: "Filter to a single event"},
			&cli.StringFlag{Name: "matcher", Aliases: []string{"m"}, Value: "*", Usage: "Default tool matcher for events (e.g., '*')"},
			&cli.StringFlag{Name: "post-matcher", Value: "Edit,Write", Usage: "Matcher for PostToolUse when not overridden"},
			&cli.BoolFlag{Name: "list", Usage: "List available groups"},
			&cli.IntFlag{Name: "timeout", Aliases: []string{"t"}, Usage: "Override timeout in seconds for installed commands"},
			&cli.BoolFlag{Name: "init", Usage: "If group not found, create a sample group stub in hooks.yml"},
			&cli.BoolFlag{Name: "prune", Usage: "Remove previously installed commands for this group before installing"},
		},
		ArgsUsage: "<group-name>",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Bool("list") {
				return handleListGroups()
			}

			opts, err := parseInstallOptions(cmd, isValidEventType, validEventTypes)
			if err != nil {
				return err
			}

			cfg, err := loadAndPrepareConfig(opts)
			if err != nil {
				return err
			}

			settings, settingsPath, err := loadSettingsForInstall(opts.useGlobal)
			if err != nil {
				return err
			}

			if opts.prune {
				handlePruneGroup(settings, opts)
			}

			installed := installGroupHooks(settings, (*cfg)[opts.groupName], opts)

			if err := config.SaveSettings(settingsPath, settings); err != nil {
				return fmt.Errorf("error saving settings: %v", err)
			}

			printInstallSuccess(opts.groupName, getScopeName(opts.useGlobal), installed, settingsPath)
			return nil
		},
	}
}

// handleListGroups handles the --list flag
func handleListGroups() error {
	cfg, err := config.LoadHooksConfig()
	if err != nil {
		return fmt.Errorf("failed to load hooks config: %v", err)
	}
	return listCustomHookGroups(cfg)
}

// parseInstallOptions extracts and validates install command options
func parseInstallOptions(cmd *cli.Command, isValidEventType func(string) bool, validEventTypes func() []string) (installOptions, error) {
	args := cmd.Args().Slice()
	if len(args) != 1 {
		return installOptions{}, fmt.Errorf("exactly one argument required: <group-name>")
	}

	eventFilter := strings.TrimSpace(cmd.String("event"))
	if eventFilter != "" && !isValidEventType(eventFilter) {
		return installOptions{}, fmt.Errorf("invalid --event '%s'. Valid events: %s", eventFilter, strings.Join(validEventTypes(), ", "))
	}

	return installOptions{
		groupName:       args[0],
		useGlobal:       cmd.Bool("global"),
		defaultMatcher:  cmd.String("matcher"),
		postMatcher:     cmd.String("post-matcher"),
		eventFilter:     eventFilter,
		timeoutOverride: cmd.Int("timeout"),
		prune:           cmd.Bool("prune"),
	}, nil
}

// loadAndPrepareConfig loads config and optionally creates a stub group
func loadAndPrepareConfig(opts installOptions) (*config.CustomHooksConfig, error) {
	cfg, err := config.LoadHooksConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load hooks config: %v", err)
	}

	return loadOrCreateGroup(cfg, opts.groupName, false, opts.useGlobal)
}

// loadSettingsForInstall loads settings for installation
func loadSettingsForInstall(useGlobal bool) (*config.Settings, string, error) {
	settingsPath, err := config.GetSettingsPath(useGlobal)
	if err != nil {
		return nil, "", fmt.Errorf("error getting settings path: %v", err)
	}

	settings, err := config.LoadSettings(settingsPath)
	if err != nil {
		return nil, "", fmt.Errorf("error loading settings: %v", err)
	}

	return settings, settingsPath, nil
}

// handlePruneGroup prunes previously installed entries for a group
func handlePruneGroup(settings *config.Settings, opts installOptions) {
	removed := config.RemoveConfigGroupFromSettings(settings, opts.groupName, opts.eventFilter)
	if removed > 0 {
		printPrunedMessage(removed, opts.groupName, opts.eventFilter)
	}
}
