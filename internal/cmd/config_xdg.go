package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/klauern/blues-traveler/internal/config"
	"github.com/klauern/blues-traveler/internal/constants"
	"github.com/urfave/cli/v3"
)

// NewConfigCmd creates the main config command with subcommands
func NewConfigCmd() *cli.Command {
	return &cli.Command{
		Name:        "config",
		Usage:       "Manage XDG-compliant configuration files",
		Description: `Manage blues-traveler configuration files using the XDG Base Directory Specification.`,
		Commands: []*cli.Command{
			NewConfigMigrateCmd(),
			NewConfigListCmd(),
			NewConfigEditCmd(),
			NewConfigCleanCmd(),
			NewConfigStatusCmd(),
			NewConfigLogCmd(),
		},
	}
}

// NewConfigMigrateCmd creates the config migrate subcommand
func NewConfigMigrateCmd() *cli.Command {
	return &cli.Command{
		Name:  "migrate",
		Usage: "Migrate existing configuration files to XDG structure",
		Description: `Discover and migrate existing .claude/hooks/blues-traveler-config.json files to XDG-compliant locations.

By default, only searches the current directory for legacy configurations.
Use --all to search across common project directories (~/dev, ~/projects, etc.).`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "dry-run",
				Aliases: []string{"n"},
				Value:   false,
				Usage:   "Show what would be migrated without making changes",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Value:   false,
				Usage:   "Show detailed migration information",
			},
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Value:   false,
				Usage:   "Search all common project directories instead of just current directory",
			},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			dryRun := cmd.Bool("dry-run")
			verbose := cmd.Bool("verbose")
			all := cmd.Bool("all")

			xdg := config.NewXDGConfig()
			discovery := config.NewLegacyConfigDiscovery(xdg)
			discovery.SetVerbose(verbose)

			printMigrateSearchInfo(xdg, verbose, all)

			// Discover and migrate
			configs, err := discovery.DiscoverLegacyConfigsWithScope(all)
			if err != nil {
				return fmt.Errorf("discovery failed: %w", err)
			}

			if len(configs) == 0 {
				printNoConfigsFound(all)
				return nil
			}

			printFoundConfigs(configs, verbose)

			result, err := discovery.MigrateConfigs(configs, dryRun)
			if err != nil {
				return fmt.Errorf("migration failed: %w", err)
			}

			// Display results
			fmt.Print(config.FormatMigrationResult(result, dryRun))
			printMigrationSummary(result, dryRun, xdg.GetConfigDir())

			return nil
		},
	}
}

// printEmptyProjectsMessage prints message when no projects are found
func printEmptyProjectsMessage(pathsOnly bool) {
	if !pathsOnly {
		fmt.Println("No projects found in XDG configuration registry.")
		fmt.Printf("Run 'blues-traveler config migrate' to migrate existing configurations.\n")
	}
}

// printPathsOnly outputs just the project paths
func printPathsOnly(projects []string) {
	for _, project := range projects {
		fmt.Println(project)
	}
}

// printProjectSummary outputs the project summary header
func printProjectSummary(projects []string) {
	fmt.Printf("Found %d project(s) in XDG configuration registry:\n\n", len(projects))
}

// printProjectBasic outputs basic project information
func printProjectBasic(project string) {
	fmt.Printf("Project: %s\n", project)
}

// printProjectVerbose outputs detailed project information
func printProjectVerbose(xdg *config.XDGConfig, project string) {
	projectConfig, err := xdg.GetProjectConfig(project)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}

	configPath := filepath.Join(xdg.GetConfigDir(), projectConfig.ConfigFile)
	fmt.Printf("  Config File: %s\n", configPath)
	fmt.Printf("  Format: %s\n", projectConfig.ConfigFormat)
	fmt.Printf("  Last Modified: %s\n", projectConfig.LastModified)

	// Check if config file exists
	if _, err := os.Stat(configPath); err != nil {
		fmt.Printf("  Status: Missing (config file not found)\n")
	} else {
		fmt.Printf("  Status: OK\n")
	}
}

// printVerboseHint prints hint about using verbose flag
func printVerboseHint() {
	fmt.Printf("Use --verbose flag for detailed information.\n")
}

// NewConfigListCmd creates the config list subcommand
func NewConfigListCmd() *cli.Command {
	return &cli.Command{
		Name:        "list",
		Usage:       "List all tracked project configurations",
		Description: `Show all projects registered in the XDG configuration system.`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Value:   false,
				Usage:   "Show detailed information about each project",
			},
			&cli.BoolFlag{
				Name:  "paths-only",
				Value: false,
				Usage: "Show only project paths (useful for scripting)",
			},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			verbose := cmd.Bool("verbose")
			pathsOnly := cmd.Bool("paths-only")

			xdg := config.NewXDGConfig()
			projects, err := xdg.ListProjects()
			if err != nil {
				return fmt.Errorf("failed to list projects: %w", err)
			}

			if len(projects) == 0 {
				printEmptyProjectsMessage(pathsOnly)
				return nil
			}

			if pathsOnly {
				printPathsOnly(projects)
				return nil
			}

			printProjectSummary(projects)

			for _, project := range projects {
				printProjectBasic(project)

				if verbose {
					printProjectVerbose(xdg, project)
				}

				fmt.Println()
			}

			if !verbose {
				printVerboseHint()
			}

			return nil
		},
	}
}

// NewConfigEditCmd creates the config edit subcommand
func NewConfigEditCmd() *cli.Command {
	return &cli.Command{
		Name:        "edit",
		Usage:       "Edit configuration files",
		Description: `Open configuration files in your default editor.`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "global",
				Aliases: []string{"g"},
				Value:   false,
				Usage:   "Edit global configuration instead of project configuration",
			},
			&cli.StringFlag{
				Name:    "project",
				Aliases: []string{"p"},
				Value:   "",
				Usage:   "Edit configuration for specific project path",
			},
			&cli.StringFlag{
				Name:    "editor",
				Aliases: []string{"e"},
				Value:   "",
				Usage:   "Override default editor (uses $EDITOR environment variable by default)",
			},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			xdg := config.NewXDGConfig()

			configPath, err := determineConfigPath(cmd, xdg)
			if err != nil {
				return err
			}

			editor, err := selectEditor(cmd.String("editor"))
			if err != nil {
				return err
			}

			return launchEditor(editor, configPath)
		},
	}
}

// determineConfigPath determines which config file to edit based on flags
func determineConfigPath(cmd *cli.Command, xdg *config.XDGConfig) (string, error) {
	if cmd.Bool("global") {
		return ensureGlobalConfigExists(xdg)
	}
	return ensureProjectConfigExists(cmd.String("project"), xdg)
}

// ensureGlobalConfigExists ensures the global config file exists and returns its path
func ensureGlobalConfigExists(xdg *config.XDGConfig) (string, error) {
	configPath := xdg.GetGlobalConfigPath("json")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := xdg.SaveGlobalConfig(make(map[string]interface{}), "json"); err != nil {
			return "", fmt.Errorf("failed to create global config: %w", err)
		}
		fmt.Printf("Created new global configuration file: %s\n", configPath)
	}

	return configPath, nil
}

// ensureProjectConfigExists ensures the project config file exists and returns its path
func ensureProjectConfigExists(projectPath string, xdg *config.XDGConfig) (string, error) {
	absProjectPath, err := resolveProjectPath(projectPath)
	if err != nil {
		return "", err
	}

	projectConfig, err := xdg.GetProjectConfig(absProjectPath)
	if err != nil {
		// Project not registered, create new config
		if err := createNewProjectConfig(absProjectPath, xdg); err != nil {
			return "", err
		}
		projectConfig, _ = xdg.GetProjectConfig(absProjectPath)
	}

	return filepath.Join(xdg.GetConfigDir(), projectConfig.ConfigFile), nil
}

// resolveProjectPath resolves the project path to an absolute path
func resolveProjectPath(projectPath string) (string, error) {
	if projectPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
		projectPath = cwd
	}

	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	return absPath, nil
}

// createNewProjectConfig creates a new project configuration
func createNewProjectConfig(projectPath string, xdg *config.XDGConfig) error {
	defaultConfig := make(map[string]interface{})
	if err := xdg.SaveProjectConfig(projectPath, defaultConfig, "json"); err != nil {
		return fmt.Errorf("failed to create project config: %w", err)
	}
	fmt.Printf("Created new project configuration for: %s\n", projectPath)
	return nil
}

// selectEditor determines which editor to use based on flag, environment, or common editors
func selectEditor(editorFlag string) (string, error) {
	if editorFlag != "" {
		return editorFlag, nil
	}

	if envEditor := os.Getenv("EDITOR"); envEditor != "" {
		return envEditor, nil
	}

	if commonEditor := findCommonEditor(); commonEditor != "" {
		return commonEditor, nil
	}

	return "", fmt.Errorf("no editor found. Set $EDITOR environment variable or use --editor flag")
}

// findCommonEditor searches for commonly available editors
func findCommonEditor() string {
	editors := []string{"code", "vim", "nano", "gedit"}
	for _, editor := range editors {
		if _, err := exec.LookPath(editor); err == nil {
			return editor
		}
	}
	return ""
}

// launchEditor launches the specified editor with the config file
func launchEditor(editor, configPath string) error {
	fmt.Printf("Opening %s with %s...\n", configPath, editor)

	// Parse editor string to support commands with arguments (e.g., "code -w")
	parts := strings.Fields(editor)
	if len(parts) == 0 {
		return fmt.Errorf("invalid editor command")
	}

	cmdName := parts[0]
	args := make([]string, 0, len(parts))
	args = append(args, parts[1:]...)
	args = append(args, configPath)
	cmd := exec.Command(cmdName, args...) // #nosec G204 - editor is from controlled sources: user flag, $EDITOR env var, or predefined safe list
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// findOrphanedProjects finds projects that no longer exist on filesystem
func findOrphanedProjects(xdg *config.XDGConfig) ([]string, error) {
	projects, err := xdg.ListProjects()
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	var orphaned []string
	for _, project := range projects {
		if _, err := os.Stat(project); os.IsNotExist(err) {
			orphaned = append(orphaned, project)
		}
	}
	return orphaned, nil
}

// printOrphanedProjects prints the list of orphaned projects
func printOrphanedProjects(orphaned []string, action string) {
	fmt.Printf("Found %d orphaned configuration(s):\n", len(orphaned))
	for _, project := range orphaned {
		fmt.Printf("  - %s\n", project)
	}

	if len(orphaned) > 0 {
		fmt.Printf("\nTo remove these configurations, run: blues-traveler config clean\n")
	} else {
		fmt.Printf("No orphaned configurations found.\n")
	}
}

// performCleanup performs the actual cleanup of orphaned configurations
func performCleanup(xdg *config.XDGConfig) error {
	orphaned, err := xdg.CleanupOrphanedConfigs()
	if err != nil {
		return fmt.Errorf("cleanup failed: %w", err)
	}

	fmt.Printf("Cleaned up %d orphaned configuration(s):\n", len(orphaned))
	for _, project := range orphaned {
		fmt.Printf("  - %s\n", project)
	}

	if len(orphaned) == 0 {
		fmt.Printf("No orphaned configurations found.\n")
	}

	return nil
}

// NewConfigCleanCmd creates the config clean subcommand
func NewConfigCleanCmd() *cli.Command {
	return &cli.Command{
		Name:        "clean",
		Usage:       "Remove configuration files for deleted projects",
		Description: `Remove configuration files for projects that no longer exist on the filesystem.`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "dry-run",
				Aliases: []string{"n"},
				Value:   false,
				Usage:   "Show what would be cleaned without making changes",
			},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			dryRun := cmd.Bool("dry-run")
			xdg := config.NewXDGConfig()

			if dryRun {
				fmt.Println("Dry run: checking for orphaned configurations...")
				orphaned, err := findOrphanedProjects(xdg)
				if err != nil {
					return err
				}
				printOrphanedProjects(orphaned, "would be removed")
			} else {
				fmt.Println("Cleaning up orphaned configurations...")
				return performCleanup(xdg)
			}

			return nil
		},
	}
}

// NewConfigStatusCmd creates the config status subcommand
func NewConfigStatusCmd() *cli.Command {
	return &cli.Command{
		Name:        "status",
		Usage:       "Show configuration status for current or specified project",
		Description: `Display information about configuration files and migration status.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "project",
				Aliases: []string{"p"},
				Value:   "",
				Usage:   "Check status for specific project path (defaults to current directory)",
			},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			projectPath := cmd.String("project")

			// Use current directory if no project specified
			if projectPath == "" {
				var err error
				projectPath, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get current directory: %w", err)
				}
			}

			status, err := config.GetMigrationStatus(projectPath)
			if err != nil {
				return fmt.Errorf("failed to get migration status: %w", err)
			}

			fmt.Printf("Configuration Status for: %s\n\n", status.ProjectPath)

			// Legacy configuration
			fmt.Printf("Legacy Configuration (.claude/hooks/):\n")
			if status.HasLegacyConfig {
				fmt.Printf("  ✓ Found: %s\n", status.LegacyConfigPath)
			} else {
				fmt.Printf("  ✗ Not found: %s\n", status.LegacyConfigPath)
			}

			// XDG configuration
			fmt.Printf("\nXDG Configuration (~/.config/blues-traveler/):\n")
			if status.HasXDGConfig {
				fmt.Printf("  ✓ Found: %s\n", status.XDGConfigPath)
			} else {
				fmt.Printf("  ✗ Not found\n")
			}

			// Migration status
			fmt.Printf("\nMigration Status:\n")
			switch {
			case status.NeedsMigration:
				fmt.Printf("  ⚠ Migration needed\n")
				fmt.Printf("  Run: blues-traveler config migrate\n")
			case status.HasXDGConfig:
				fmt.Printf("  ✓ Already migrated to XDG\n")
			case !status.HasLegacyConfig:
				fmt.Printf("  ✓ No configuration found (will use defaults)\n")
			}

			// Recommendations
			fmt.Printf("\nRecommendations:\n")
			switch {
			case status.NeedsMigration:
				fmt.Printf("  • Run 'blues-traveler config migrate' to migrate to XDG structure\n")
			case status.HasXDGConfig:
				fmt.Printf("  • Use 'blues-traveler config edit' to modify configuration\n")
			default:
				fmt.Printf("  • Use 'blues-traveler config edit' to create a new configuration\n")
			}

			return nil
		},
	}
}

// NewConfigLogCmd creates the config log subcommand
//
//nolint:gocognit // CLI command with multiple validation steps and comprehensive help output
func NewConfigLogCmd() *cli.Command {
	return &cli.Command{
		Name:        "log",
		Usage:       "Configure log rotation settings",
		Description: `Configure log rotation settings including maximum age, file size, and backup count.`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "global",
				Aliases: []string{"g"},
				Value:   false,
				Usage:   "Configure global settings",
			},
			&cli.IntFlag{
				Name:    "max-age",
				Aliases: []string{"a"},
				Value:   0,
				Usage:   "Maximum age in days to retain log files (default: 30)",
			},
			&cli.IntFlag{
				Name:    "max-size",
				Aliases: []string{"s"},
				Value:   0,
				Usage:   "Maximum size in MB per log file before rotation (default: 10)",
			},
			&cli.IntFlag{
				Name:    "max-backups",
				Aliases: []string{"b"},
				Value:   0,
				Usage:   "Maximum number of backup files to retain (default: 5)",
			},
			&cli.BoolFlag{
				Name:    "compress",
				Aliases: []string{"c"},
				Value:   false,
				Usage:   "Compress rotated log files",
			},
			&cli.BoolFlag{
				Name:  "show",
				Value: false,
				Usage: "Show current log rotation settings",
			},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			global := cmd.Bool("global")
			maxAge := cmd.Int("max-age")
			maxSize := cmd.Int("max-size")
			maxBackups := cmd.Int("max-backups")
			compress := cmd.Bool("compress")
			show := cmd.Bool("show")

			configPath, err := config.GetLogConfigPath(global)
			if err != nil {
				scope := constants.ScopeProject
				if global {
					scope = constants.ScopeGlobal
				}
				return fmt.Errorf("failed to locate %s config path: %w\n  Suggestion: Ensure your project is properly initialized with 'blues-traveler hooks init'", scope, err)
			}

			logConfig, err := config.LoadLogConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config from %s: %w\n  Suggestion: Check if the config file is valid JSON format", configPath, err)
			}

			if show {
				// Show current log rotation settings
				scope := constants.ScopeProject
				if global {
					scope = constants.ScopeGlobal
				}
				fmt.Printf("Current log rotation settings (%s: %s):\n", scope, configPath)
				fmt.Printf("  Max Age: %d days\n", logConfig.LogRotation.MaxAge)
				fmt.Printf("  Max Size: %d MB\n", logConfig.LogRotation.MaxSize)
				fmt.Printf("  Max Backups: %d files\n", logConfig.LogRotation.MaxBackups)
				fmt.Printf("  Compress: %t\n", logConfig.LogRotation.Compress)
				return nil
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
			// Respect explicit presence of the flag (true or false)
			if cmd.IsSet("compress") {
				logConfig.LogRotation.Compress = compress
			}

			if err := config.SaveLogConfig(configPath, logConfig); err != nil {
				return fmt.Errorf("failed to save config to %s: %w\n  Suggestion: Check file permissions and ensure the directory is writable", configPath, err)
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
			return nil
		},
	}
}

// Helper functions for migrate command

// printMigrateSearchInfo displays information about the migration search scope
func printMigrateSearchInfo(xdg *config.XDGConfig, verbose, all bool) {
	if !verbose {
		return
	}
	fmt.Printf("XDG config directory: %s\n", xdg.GetConfigDir())
	if all {
		fmt.Printf("Searching globally across common project directories\n")
	} else {
		fmt.Printf("Searching only in current directory\n")
	}
}

// printNoConfigsFound displays message when no legacy configs are found
func printNoConfigsFound(all bool) {
	if all {
		fmt.Printf("No legacy configurations found to migrate.\n")
	} else {
		fmt.Printf("No legacy configuration found in current directory.\n")
		fmt.Printf("Use --all flag to search across common project directories.\n")
	}
}

// printFoundConfigs displays discovered configuration files
func printFoundConfigs(configs map[string]string, verbose bool) {
	fmt.Printf("Found %d legacy configuration file(s)\n", len(configs))

	if verbose && len(configs) > 0 {
		fmt.Printf("\nDiscovered configurations:\n")
		for projectPath, configPath := range configs {
			fmt.Printf("  - %s\n    → %s\n", projectPath, configPath)
		}
		fmt.Println()
	}
}

// printMigrationSummary displays summary after migration
func printMigrationSummary(result *config.MigrationResult, dryRun bool, configDir string) {
	if !dryRun && result.TotalMigrated > 0 {
		fmt.Printf("\nMigration completed successfully!\n")
		fmt.Printf("Configurations are now available in: %s\n", configDir)
	} else if dryRun && result.TotalFound > 0 {
		fmt.Printf("\nTo perform the actual migration, run: blues-traveler config migrate\n")
	}
}
