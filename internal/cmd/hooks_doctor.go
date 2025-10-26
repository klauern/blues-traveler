package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/klauern/blues-traveler/internal/config"
	"github.com/urfave/cli/v3"
)

// NewDoctorCommand creates the doctor command for diagnosing hook installation
func NewDoctorCommand() *cli.Command {
	return &cli.Command{
		Name:        "doctor",
		Usage:       "Diagnose hooks installation and configuration",
		Description: `Check the health of your hooks installation, showing what's configured, where, and any potential issues.`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Value:   false,
				Usage:   "Show detailed configuration information",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			verbose := cmd.Bool("verbose")
			return runDoctorCheck(verbose)
		},
	}
}

// runDoctorCheck performs the diagnosis of the hooks system
func runDoctorCheck(verbose bool) error {
	fmt.Println("🔍 Blues Traveler Hooks Doctor")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Println()

	// Check project settings
	fmt.Println("📁 Project Settings")
	fmt.Println(strings.Repeat("-", 52))
	checkProjectSettings(verbose)
	fmt.Println()

	// Check global settings
	fmt.Println("🌍 Global Settings")
	fmt.Println(strings.Repeat("-", 52))
	checkGlobalSettings(verbose)
	fmt.Println()

	// Check custom hooks configuration
	fmt.Println("⚙️  Custom Hooks Configuration")
	fmt.Println(strings.Repeat("-", 52))
	checkCustomHooksConfig(verbose)
	fmt.Println()

	// Summary and recommendations
	fmt.Println("📋 Summary")
	fmt.Println(strings.Repeat("-", 52))
	printSummary()

	return nil
}

// checkProjectSettings checks project-level hook settings
func checkProjectSettings(verbose bool) {
	checkSettings(false, verbose, "project", "hooks install <plugin>")
}

// checkGlobalSettings checks global-level hook settings
func checkGlobalSettings(verbose bool) {
	checkSettings(true, verbose, "global", "hooks install <plugin> --global")
}

// checkSettings is a helper that checks hook settings for project or global scope
func checkSettings(isGlobal bool, verbose bool, scope string, installCmd string) {
	settingsPath, err := config.GetSettingsPath(isGlobal)
	if err != nil {
		fmt.Printf("⚠️  Error getting %s settings path: %v\n", scope, err)
		return
	}

	fmt.Printf("Location: %s\n", settingsPath)

	// Check if file exists first, since LoadSettings returns empty settings for missing files
	if _, err := os.Stat(settingsPath); err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Status: ✗ No %s settings file found\n", scope)
			fmt.Printf("        Use '%s' to create %s settings\n", installCmd, scope)
		} else {
			fmt.Printf("Status: ⚠️  Error checking settings file: %v\n", err)
		}
		return
	}

	settings, err := config.LoadSettings(settingsPath)
	if err != nil {
		fmt.Printf("Status: ⚠️  Error loading settings: %v\n", err)
		return
	}

	if config.IsHooksConfigEmpty(settings.Hooks) {
		fmt.Println("Status: ✓ Settings file exists, but no hooks installed")
	} else {
		fmt.Println("Status: ✓ Hooks configured")
		printHooksSummary(settings.Hooks, verbose)
	}
}

// checkCustomHooksConfig checks custom hooks configuration files
func checkCustomHooksConfig(verbose bool) {
	cfg, err := config.LoadHooksConfig()
	if err != nil {
		fmt.Printf("⚠️  Error loading hooks config: %v\n", err)
		return
	}

	// Get candidate paths to show where we looked
	candidates, err := getCandidateConfigPaths()
	if err != nil {
		fmt.Printf("⚠️  Error getting config paths: %v\n", err)
		return
	}

	// Check which files exist
	var foundFiles []string
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			foundFiles = append(foundFiles, path)
		}
	}

	if len(foundFiles) == 0 {
		fmt.Println("Status: ✗ No custom hooks configuration files found")
		fmt.Println()
		fmt.Println("Searched locations:")
		fmt.Println("  Project: .claude/hooks/hooks.yml")
		fmt.Println("           .claude/hooks.yml")
		fmt.Println("           .claude/hooks/*.yml")
		fmt.Println("  Global:  ~/.claude/hooks/hooks.yml")
		fmt.Println("           ~/.claude/hooks.yml")
		fmt.Println("           ~/.claude/hooks/*.yml")
		fmt.Println()
		fmt.Println("To create: Use 'hooks custom init <group-name>' to get started")
		return
	}

	fmt.Printf("Status: ✓ Found %d configuration file(s)\n", len(foundFiles))
	fmt.Println()

	if verbose {
		fmt.Println("Configuration files (in merge order):")
		for _, f := range foundFiles {
			scope := "project"
			home, _ := os.UserHomeDir()
			if home != "" {
				globalPrefix := filepath.Join(home, ".claude")
				if strings.HasPrefix(f, globalPrefix) {
					scope = "global"
				}
			}
			fmt.Printf("  • %s (%s)\n", f, scope)
		}
		fmt.Println()
	}

	// Validate and show groups
	if cfg == nil || len(*cfg) == 0 {
		fmt.Println("⚠️  Configuration files exist but no groups defined")
		return
	}

	groups := config.ListHookGroups(cfg)
	fmt.Printf("Groups: %d defined\n", len(groups))

	if verbose {
		for _, groupName := range groups {
			group := (*cfg)[groupName]
			eventCount := len(group)
			jobCount := 0
			for _, ev := range group {
				jobCount += len(ev.Jobs)
			}
			fmt.Printf("  • %s (%d events, %d jobs)\n", groupName, eventCount, jobCount)
		}
		fmt.Println()
	}

	// Validate configuration
	if err := config.ValidateHooksConfig(cfg); err != nil {
		fmt.Printf("⚠️  Configuration validation failed: %v\n", err)
	} else {
		fmt.Println("✓ Configuration is valid")
	}
}

// printHooksSummary prints a summary of installed hooks
func printHooksSummary(hooks config.HooksConfig, verbose bool) {
	events := countHooksByEvent(hooks)
	totalHooks := 0
	for _, count := range events {
		totalHooks += count
	}

	fmt.Printf("        %d hook(s) installed across %d event type(s)\n", totalHooks, len(events))

	if verbose && totalHooks > 0 {
		fmt.Println()
		fmt.Println("        Event breakdown:")

		// Sort events for consistent display
		eventNames := make([]string, 0, len(events))
		for name := range events {
			eventNames = append(eventNames, name)
		}
		sort.Strings(eventNames)

		for _, name := range eventNames {
			count := events[name]
			fmt.Printf("          • %s: %d hook(s)\n", name, count)
		}
	}
}

// countHooksByEvent counts hooks installed for each event type
func countHooksByEvent(hooks config.HooksConfig) map[string]int {
	events := make(map[string]int)

	if len(hooks.PreToolUse) > 0 {
		count := 0
		for _, m := range hooks.PreToolUse {
			count += len(m.Hooks)
		}
		events["PreToolUse"] = count
	}

	if len(hooks.PostToolUse) > 0 {
		count := 0
		for _, m := range hooks.PostToolUse {
			count += len(m.Hooks)
		}
		events["PostToolUse"] = count
	}

	if len(hooks.UserPromptSubmit) > 0 {
		count := 0
		for _, m := range hooks.UserPromptSubmit {
			count += len(m.Hooks)
		}
		events["UserPromptSubmit"] = count
	}

	if len(hooks.Notification) > 0 {
		count := 0
		for _, m := range hooks.Notification {
			count += len(m.Hooks)
		}
		events["Notification"] = count
	}

	if len(hooks.Stop) > 0 {
		count := 0
		for _, m := range hooks.Stop {
			count += len(m.Hooks)
		}
		events["Stop"] = count
	}

	if len(hooks.SubagentStop) > 0 {
		count := 0
		for _, m := range hooks.SubagentStop {
			count += len(m.Hooks)
		}
		events["SubagentStop"] = count
	}

	if len(hooks.PreCompact) > 0 {
		count := 0
		for _, m := range hooks.PreCompact {
			count += len(m.Hooks)
		}
		events["PreCompact"] = count
	}

	if len(hooks.SessionStart) > 0 {
		count := 0
		for _, m := range hooks.SessionStart {
			count += len(m.Hooks)
		}
		events["SessionStart"] = count
	}

	if len(hooks.SessionEnd) > 0 {
		count := 0
		for _, m := range hooks.SessionEnd {
			count += len(m.Hooks)
		}
		events["SessionEnd"] = count
	}

	return events
}

// printSummary prints overall summary and recommendations
func printSummary() {
	fmt.Println("The hooks system has been checked.")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  • View available plugins: blues-traveler hooks list")
	fmt.Println("  • View installed hooks: blues-traveler hooks list --installed")
	fmt.Println("  • Install a hook: blues-traveler hooks install <plugin-key>")
	fmt.Println("  • Create custom hooks: blues-traveler hooks custom init <group-name>")
	fmt.Println()
	fmt.Println("For more verbose output, run: blues-traveler doctor --verbose")
}

// getCandidateConfigPaths returns all potential hook config file locations
// This is a simplified version that doesn't use the internal candidateConfigPaths
func getCandidateConfigPaths() ([]string, error) {
	var paths []string

	// Project scope
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %v", err)
	}
	proj := filepath.Join(cwd, ".claude")

	// Main hooks config files
	paths = append(paths,
		filepath.Join(proj, "hooks", "hooks.yml"),
		filepath.Join(proj, "hooks", "hooks.yaml"),
		filepath.Join(proj, "hooks.yml"),
		filepath.Join(proj, "hooks.yaml"),
		filepath.Join(proj, "hooks.json"),
	)

	// Global scope
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}
	glob := filepath.Join(home, ".claude")

	paths = append(paths,
		filepath.Join(glob, "hooks", "hooks.yml"),
		filepath.Join(glob, "hooks", "hooks.yaml"),
		filepath.Join(glob, "hooks.yml"),
		filepath.Join(glob, "hooks.yaml"),
		filepath.Join(glob, "hooks.json"),
	)

	// Enumerate all *.yml and *.yaml files in project hooks directory
	projHooksDir := filepath.Join(proj, "hooks")
	if projYmls, err := filepath.Glob(filepath.Join(projHooksDir, "*.yml")); err == nil {
		paths = append(paths, projYmls...)
	}
	if projYamls, err := filepath.Glob(filepath.Join(projHooksDir, "*.yaml")); err == nil {
		paths = append(paths, projYamls...)
	}

	// Enumerate all *.yml and *.yaml files in global hooks directory
	globHooksDir := filepath.Join(glob, "hooks")
	if globYmls, err := filepath.Glob(filepath.Join(globHooksDir, "*.yml")); err == nil {
		paths = append(paths, globYmls...)
	}
	if globYamls, err := filepath.Glob(filepath.Join(globHooksDir, "*.yaml")); err == nil {
		paths = append(paths, globYamls...)
	}

	return paths, nil
}
