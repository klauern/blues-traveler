package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/klauern/blues-traveler/internal/config"
	"github.com/klauern/blues-traveler/internal/core"
	"github.com/urfave/cli/v3"
)

// newHooksCustomSyncCommand creates the sync command for custom hooks
func newHooksCustomSyncCommand(isValidEventType func(string) bool, validEventTypes func() []string) *cli.Command {
	return &cli.Command{
		Name:      "sync",
		Usage:     "Sync custom hooks from hooks.yml into Claude settings",
		ArgsUsage: "[group]",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "global", Aliases: []string{"g"}, Usage: "Sync to global settings (~/.claude/settings.json)"},
			&cli.BoolFlag{Name: "dry-run", Aliases: []string{"n"}, Usage: "Show intended changes without writing"},
			&cli.StringFlag{Name: "event", Aliases: []string{"e"}, Usage: "Restrict sync to a single event (e.g., PreToolUse, PostToolUse)"},
			&cli.StringFlag{Name: "matcher", Aliases: []string{"m"}, Value: "*", Usage: "Default tool matcher for events (e.g., '*')"},
			&cli.StringFlag{Name: "post-matcher", Value: "Edit,Write", Usage: "Matcher for PostToolUse when not overridden"},
			&cli.IntFlag{Name: "timeout", Aliases: []string{"t"}, Usage: "Override timeout in seconds for installed commands"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			opts, err := parseSyncOptions(cmd, isValidEventType, validEventTypes)
			if err != nil {
				return err
			}

			hooksCfg, settings, settingsPath, err := loadSyncDependencies(opts.useGlobal)
			if err != nil {
				return err
			}

			changed := performSync(settings, hooksCfg, opts)

			return finalizeSyncOperation(settingsPath, settings, changed, opts)
		},
	}
}

// parseSyncOptions extracts and validates command line options
func parseSyncOptions(cmd *cli.Command, isValidEventType func(string) bool, validEventTypes func() []string) (syncOptions, error) {
	args := cmd.Args().Slice()
	var groupFilter string
	if len(args) > 0 {
		if len(args) > 1 {
			return syncOptions{}, fmt.Errorf("at most one [group] argument is allowed")
		}
		groupFilter = args[0]
	}

	execPath := resolveExecutablePath()
	eventFilter := strings.TrimSpace(cmd.String("event"))

	// Validate event filter if provided (accepts Cursor aliases)
	if eventFilter != "" && !isValidEventType(eventFilter) {
		return syncOptions{}, fmt.Errorf("invalid event '%s'.\nValid events: %s\nUse 'hooks list --events' to see all available events with descriptions", eventFilter, strings.Join(validEventTypes(), ", "))
	}

	// Resolve Cursor alias to canonical event name
	if eventFilter != "" {
		resolvedEvent := core.ResolveEventAlias(eventFilter)
		if resolvedEvent != "" {
			eventFilter = resolvedEvent
		}
	}

	return syncOptions{
		useGlobal:       cmd.Bool("global"),
		dryRun:          cmd.Bool("dry-run"),
		eventFilter:     eventFilter,
		groupFilter:     groupFilter,
		defaultMatcher:  cmd.String("matcher"),
		postMatcher:     cmd.String("post-matcher"),
		timeoutOverride: cmd.Int("timeout"),
		execPath:        execPath,
	}, nil
}

// resolveExecutablePath returns a stable blues-traveler path for settings entries
func resolveExecutablePath() string {
	if p, err := os.Executable(); err == nil {
		return p
	}
	return "blues-traveler"
}

// loadSyncDependencies loads hooks config and settings
func loadSyncDependencies(useGlobal bool) (*config.CustomHooksConfig, *config.Settings, string, error) {
	hooksCfg, err := config.LoadHooksConfig()
	if err != nil {
		return nil, nil, "", fmt.Errorf("load hooks config: %v", err)
	}

	settingsPath, err := config.GetSettingsPath(useGlobal)
	if err != nil {
		return nil, nil, "", err
	}

	settings, err := config.LoadSettings(settingsPath)
	if err != nil {
		return nil, nil, "", err
	}

	return hooksCfg, settings, settingsPath, nil
}

// performSync executes the sync operation
func performSync(settings *config.Settings, hooksCfg *config.CustomHooksConfig, opts syncOptions) int {
	changed := 0

	// Step 1: Clean up stale groups
	existingGroups := config.GetConfigGroupsInSettings(settings)
	configGroups := buildConfigGroupsMap(hooksCfg)
	changed += cleanupStaleGroups(settings, existingGroups, configGroups, opts)

	// Step 2: Sync current config groups
	changed += syncConfigGroups(settings, hooksCfg, opts)

	return changed
}

// syncConfigGroups syncs all config groups to settings
func syncConfigGroups(settings *config.Settings, hooksCfg *config.CustomHooksConfig, opts syncOptions) int {
	changed := 0
	if hooksCfg == nil {
		return changed
	}

	for groupName, group := range *hooksCfg {
		if shouldSkipGroup(groupName, opts.groupFilter) {
			continue
		}

		// Prune existing settings for this group
		removed := config.RemoveConfigGroupFromSettings(settings, groupName, opts.eventFilter)
		if removed > 0 {
			printPrunedMessage(removed, groupName, opts.eventFilter)
		}

		// Add current definitions
		changed += syncGroupToSettings(settings, groupName, group, opts)
	}

	return changed
}

// finalizeSyncOperation handles final output and saving
func finalizeSyncOperation(settingsPath string, settings *config.Settings, changed int, opts syncOptions) error {
	if changed == 0 {
		fmt.Println("No changes detected.")
		return nil
	}

	if opts.dryRun {
		fmt.Println("Dry run; not writing settings.")
		return nil
	}

	if err := config.SaveSettings(settingsPath, settings); err != nil {
		return err
	}

	scope := "project"
	if opts.useGlobal {
		scope = "global"
	}
	fmt.Printf("Synced %d entries into %s settings: %s\n", changed, scope, settingsPath)
	return nil
}
