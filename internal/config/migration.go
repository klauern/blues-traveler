package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
)

// MigrationResult represents the result of a configuration migration
type MigrationResult struct {
	TotalFound      int
	TotalMigrated   int
	TotalSkipped    int
	TotalErrors     int
	MigratedPaths   []string
	SkippedPaths    []string
	ErrorPaths      []MigrationError
	BackupLocations []string
}

// MigrationError represents an error that occurred during migration
type MigrationError struct {
	Path  string
	Error string
}

// LegacyConfigDiscovery discovers existing blues-traveler configuration files
type LegacyConfigDiscovery struct {
	xdg     *XDGConfig
	verbose bool
}

// NewLegacyConfigDiscovery creates a new legacy config discovery instance
func NewLegacyConfigDiscovery(xdg *XDGConfig) *LegacyConfigDiscovery {
	return &LegacyConfigDiscovery{xdg: xdg, verbose: false}
}

// SetVerbose enables or disables verbose logging during discovery
func (d *LegacyConfigDiscovery) SetVerbose(verbose bool) {
	d.verbose = verbose
}

// DiscoverLegacyConfigs searches for existing blues-traveler-config.json files
func (d *LegacyConfigDiscovery) DiscoverLegacyConfigs() (map[string]string, error) {
	return d.DiscoverLegacyConfigsWithScope(false)
}

// DiscoverLegacyConfigsWithScope searches for configs with optional global scope
func (d *LegacyConfigDiscovery) DiscoverLegacyConfigsWithScope(globalSearch bool) (map[string]string, error) {
	configs := make(map[string]string)

	var searchPaths []string

	if globalSearch {
		// Global search: look in common project locations
		searchPaths = []string{
			// Current directory and parent directories
			".",
			"..",
			"../..",
		}

		// Add user's home directory common project locations
		homeDir, err := os.UserHomeDir()
		if err == nil {
			searchPaths = append(searchPaths,
				filepath.Join(homeDir, "dev"),
				filepath.Join(homeDir, "projects"),
				filepath.Join(homeDir, "work"),
				filepath.Join(homeDir, "src"),
			)
		}
	} else {
		// Local search: only check current directory
		searchPaths = []string{"."}
	}

	var bar *progressbar.ProgressBar
	if !d.verbose {
		bar = progressbar.NewOptions(len(searchPaths),
			progressbar.OptionSetDescription("Searching for configs..."),
			progressbar.OptionSetWidth(40),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionSetRenderBlankState(true),
		)
	}

	for i, searchPath := range searchPaths {
		if d.verbose {
			fmt.Printf("Searching in: %s\n", searchPath)
		} else {
			_ = bar.Set(i)
		}

		if err := d.walkProjectDirectories(searchPath, configs); err != nil {
			if d.verbose {
				fmt.Printf("  Warning: could not search %s: %v\n", searchPath, err)
			}
			// Continue searching other paths on error
			continue
		}

		if d.verbose && len(configs) > 0 {
			fmt.Printf("  Found %d configuration(s) so far\n", len(configs))
		}
	}

	if !d.verbose {
		_ = bar.Finish()
		fmt.Printf("Found %d legacy configuration file(s)\n", len(configs))
	}

	return configs, nil
}

// walkProjectDirectories recursively searches for .claude/hooks/blues-traveler-config.json files
func (d *LegacyConfigDiscovery) walkProjectDirectories(basePath string, configs map[string]string) error {
	return filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors and continue
		}

		// Skip hidden directories except .claude
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != ".claude" {
			return filepath.SkipDir
		}

		// Look for blues-traveler-config.json in .claude/hooks/ directories
		if info.Name() == "blues-traveler-config.json" {
			// Check if this is in a .claude/hooks directory
			dir := filepath.Dir(path)
			if filepath.Base(dir) == "hooks" && filepath.Base(filepath.Dir(dir)) == ".claude" {
				// Get the project root (directory containing .claude)
				projectRoot := filepath.Dir(filepath.Dir(dir))
				absProjectRoot, err := filepath.Abs(projectRoot)
				if err != nil {
					return nil // Skip this config
				}

				configs[absProjectRoot] = path
				if d.verbose {
					fmt.Printf("  Found config: %s\n", path)
				}
			}
		}

		return nil
	})
}

// MigrateConfigs migrates provided configurations to XDG structure
func (d *LegacyConfigDiscovery) MigrateConfigs(configs map[string]string, dryRun bool) (*MigrationResult, error) {
	result := &MigrationResult{
		TotalFound:      len(configs),
		MigratedPaths:   []string{},
		SkippedPaths:    []string{},
		ErrorPaths:      []MigrationError{},
		BackupLocations: []string{},
	}

	keys := d.sortedKeys(configs)

	var bar *progressbar.ProgressBar
	if !d.verbose {
		action := "Migrating"
		if dryRun {
			action = "Checking"
		}
		bar = progressbar.NewOptions(len(keys),
			progressbar.OptionSetDescription(action+" configs..."),
			progressbar.OptionSetWidth(40),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionSetRenderBlankState(true),
		)
	}

	for i, projectPath := range keys {
		configPath := configs[projectPath]

		if d.verbose {
			fmt.Printf("Processing %d/%d: %s\n", i+1, len(keys), projectPath)
		} else {
			_ = bar.Set(i)
		}

		if err := d.migrateConfig(projectPath, configPath, dryRun, result); err != nil {
			result.TotalErrors++
			result.ErrorPaths = append(result.ErrorPaths, MigrationError{
				Path:  projectPath,
				Error: err.Error(),
			})
			if d.verbose {
				fmt.Printf("  Error: %v\n", err)
			}
		} else if d.verbose {
			if dryRun {
				fmt.Printf("  Would migrate successfully\n")
			} else {
				fmt.Printf("  Migrated successfully\n")
			}
		}
	}

	if !d.verbose {
		_ = bar.Finish()
	}

	return result, nil
}

// sortedKeys returns sorted keys from the configs map for consistent processing order
func (d *LegacyConfigDiscovery) sortedKeys(configs map[string]string) []string {
	keys := make([]string, 0, len(configs))
	for k := range configs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// migrateConfig migrates a single configuration file
func (d *LegacyConfigDiscovery) migrateConfig(projectPath, configPath string, dryRun bool, result *MigrationResult) error {
	// Check if project already exists in XDG registry
	if _, err := d.xdg.GetProjectConfig(projectPath); err == nil {
		result.TotalSkipped++
		result.SkippedPaths = append(result.SkippedPaths, projectPath)
		return nil
	}

	// Read the legacy config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read legacy config: %w", err)
	}

	// Parse the legacy config
	var legacyConfig map[string]interface{}
	if err := json.Unmarshal(data, &legacyConfig); err != nil {
		return fmt.Errorf("failed to parse legacy config JSON: %w", err)
	}

	if dryRun {
		result.TotalMigrated++
		result.MigratedPaths = append(result.MigratedPaths, projectPath)
		return nil
	}

	// Create backup of original config
	backupPath := configPath + ".backup." + time.Now().Format("20060102-150405")
	if err := copyFile(configPath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}
	result.BackupLocations = append(result.BackupLocations, backupPath)

	// Save to XDG location
	if err := d.xdg.SaveProjectConfig(projectPath, legacyConfig, "json"); err != nil {
		return fmt.Errorf("failed to save XDG config: %w", err)
	}

	result.TotalMigrated++
	result.MigratedPaths = append(result.MigratedPaths, projectPath)

	return nil
}

// copyFile creates a copy of a file
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0o600)
}

// GetLegacyConfigPath returns the legacy config path for a project
func GetLegacyConfigPath(projectPath string) string {
	return filepath.Join(projectPath, ".claude", "hooks", "blues-traveler-config.json")
}

// HasLegacyConfig checks if a project has a legacy configuration file
func HasLegacyConfig(projectPath string) bool {
	legacyPath := GetLegacyConfigPath(projectPath)
	_, err := os.Stat(legacyPath)
	return err == nil
}

// MigrationStatus represents the migration status of a project
type MigrationStatus struct {
	ProjectPath      string
	HasLegacyConfig  bool
	HasXDGConfig     bool
	NeedsMigration   bool
	XDGConfigPath    string
	LegacyConfigPath string
}

// GetMigrationStatus checks the migration status of a specific project
func GetMigrationStatus(projectPath string) (*MigrationStatus, error) {
	xdg := NewXDGConfig()

	absProjectPath, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	status := &MigrationStatus{
		ProjectPath:      absProjectPath,
		LegacyConfigPath: GetLegacyConfigPath(absProjectPath),
		HasLegacyConfig:  HasLegacyConfig(absProjectPath),
	}

	// Check if XDG config exists
	_, err = xdg.GetProjectConfig(absProjectPath)
	status.HasXDGConfig = err == nil

	if status.HasXDGConfig {
		config, _ := xdg.GetProjectConfig(absProjectPath)
		status.XDGConfigPath = filepath.Join(xdg.GetConfigDir(), config.ConfigFile)
	}

	// Migration is needed if legacy config exists but XDG config doesn't
	status.NeedsMigration = status.HasLegacyConfig && !status.HasXDGConfig

	return status, nil
}

// FormatMigrationResult formats the migration result as a human-readable string
func FormatMigrationResult(result *MigrationResult, dryRun bool) string {
	var sb strings.Builder

	if dryRun {
		sb.WriteString("Migration Dry Run Results:\n")
	} else {
		sb.WriteString("Migration Results:\n")
	}

	sb.WriteString(fmt.Sprintf("  Found: %d legacy configurations\n", result.TotalFound))
	sb.WriteString(fmt.Sprintf("  Migrated: %d\n", result.TotalMigrated))
	sb.WriteString(fmt.Sprintf("  Skipped: %d (already migrated)\n", result.TotalSkipped))
	sb.WriteString(fmt.Sprintf("  Errors: %d\n", result.TotalErrors))

	if len(result.MigratedPaths) > 0 {
		sb.WriteString("\nMigrated Projects:\n")
		for _, path := range result.MigratedPaths {
			sb.WriteString(fmt.Sprintf("  - %s\n", path))
		}
	}

	if len(result.SkippedPaths) > 0 {
		sb.WriteString("\nSkipped Projects (already migrated):\n")
		for _, path := range result.SkippedPaths {
			sb.WriteString(fmt.Sprintf("  - %s\n", path))
		}
	}

	if len(result.ErrorPaths) > 0 {
		sb.WriteString("\nErrors:\n")
		for _, errPath := range result.ErrorPaths {
			sb.WriteString(fmt.Sprintf("  - %s: %s\n", errPath.Path, errPath.Error))
		}
	}

	if !dryRun && len(result.BackupLocations) > 0 {
		sb.WriteString("\nBackup Files Created:\n")
		for _, backup := range result.BackupLocations {
			sb.WriteString(fmt.Sprintf("  - %s\n", backup))
		}
	}

	return sb.String()
}
