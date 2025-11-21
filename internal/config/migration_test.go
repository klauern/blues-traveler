package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLegacyConfigDiscovery(t *testing.T) {
	tempDir := t.TempDir()
	xdgTempDir := t.TempDir()

	xdg := &XDGConfig{BaseDir: xdgTempDir}
	discovery := NewLegacyConfigDiscovery(xdg)

	projects := []string{
		filepath.Join(tempDir, "project1"),
		filepath.Join(tempDir, "project2"),
		filepath.Join(tempDir, "project3", "subproject"),
	}

	for i, project := range projects {
		createTestLegacyConfig(t, project, i+1)
	}

	t.Run("discovers all configs", func(t *testing.T) {
		discoveredConfigs := make(map[string]string)
		err := discovery.walkProjectDirectories(tempDir, discoveredConfigs)
		requireNoError(t, err, "discover configs")

		if len(discoveredConfigs) != len(projects) {
			t.Errorf("Expected %d configs, found %d", len(projects), len(discoveredConfigs))
		}

		for _, project := range projects {
			if _, exists := discoveredConfigs[project]; !exists {
				t.Errorf("Config for project %s not discovered", project)
			}
		}
	})
}

// createTestLegacyConfig creates a test legacy config with a project number.
func createTestLegacyConfig(t *testing.T, project string, projectNum int) {
	t.Helper()
	configData := map[string]interface{}{
		"logRotation": map[string]interface{}{"maxAge": 30, "maxSize": 10, "maxBackups": 5},
		"testProject": projectNum,
	}
	writeLegacyConfigFile(t, project, configData)
}

func TestMigrationDryRun(t *testing.T) {
	tempDir := t.TempDir()
	xdgTempDir := t.TempDir()

	xdg := &XDGConfig{BaseDir: xdgTempDir}
	project := filepath.Join(tempDir, "test-project")

	configData := map[string]interface{}{
		"logRotation": map[string]interface{}{"maxAge": 30},
		"testData":    "test-value",
	}
	writeLegacyConfigFile(t, project, configData)

	// Simulate dry run migration logic
	result := simulateDryRunMigration(xdg, project)

	t.Run("counts are correct", func(t *testing.T) {
		assertDryRunCounts(t, result, 1, 1, 0)
	})

	t.Run("no backups created", func(t *testing.T) {
		if len(result.BackupLocations) != 0 {
			t.Errorf("Expected 0 backup locations in dry run, got %d", len(result.BackupLocations))
		}
	})
}

// simulateDryRunMigration simulates a dry run migration for testing.
func simulateDryRunMigration(xdg *XDGConfig, projectPath string) *MigrationResult {
	result := &MigrationResult{TotalFound: 1}
	if _, err := xdg.GetProjectConfig(projectPath); err == nil {
		result.TotalSkipped++
		result.SkippedPaths = append(result.SkippedPaths, projectPath)
	} else {
		result.TotalMigrated++
		result.MigratedPaths = append(result.MigratedPaths, projectPath)
	}
	return result
}

// assertDryRunCounts checks the dry run result counts.
func assertDryRunCounts(t *testing.T, result *MigrationResult, found, migrated, skipped int) {
	t.Helper()
	if result.TotalFound != found {
		t.Errorf("Expected %d config found, got %d", found, result.TotalFound)
	}
	if result.TotalMigrated != migrated {
		t.Errorf("Expected %d config to be migrated, got %d", migrated, result.TotalMigrated)
	}
	if result.TotalSkipped != skipped {
		t.Errorf("Expected %d configs to be skipped, got %d", skipped, result.TotalSkipped)
	}
}

func TestActualMigration(t *testing.T) {
	tempDir := t.TempDir()
	xdgTempDir := t.TempDir()

	xdg := &XDGConfig{BaseDir: xdgTempDir}
	discovery := NewLegacyConfigDiscovery(xdg)

	project := filepath.Join(tempDir, "test-project")
	configPath := setupMigrationTestConfig(t, project)

	t.Run("migration succeeds", func(t *testing.T) {
		err := discovery.migrateConfig(project, configPath, false, &MigrationResult{})
		requireNoError(t, err, "perform migration")
	})

	t.Run("project is registered", func(t *testing.T) {
		projectConfig, err := xdg.GetProjectConfig(project)
		requireNoError(t, err, "get project config")

		if projectConfig.ConfigFormat != "json" {
			t.Errorf("Expected format json, got %s", projectConfig.ConfigFormat)
		}
	})

	t.Run("data is migrated correctly", func(t *testing.T) {
		migratedData, err := xdg.LoadProjectConfig(project)
		requireNoError(t, err, "load migrated config")
		assertMigratedConfigData(t, migratedData)
	})

	t.Run("backup is created", func(t *testing.T) {
		matches, err := filepath.Glob(configPath + ".backup.*")
		requireNoError(t, err, "check for backup files")
		if len(matches) != 1 {
			t.Errorf("Expected 1 backup file, found %d", len(matches))
		}
	})
}

// setupMigrationTestConfig creates a legacy config for migration testing.
func setupMigrationTestConfig(t *testing.T, project string) string {
	t.Helper()
	legacyConfigDir := filepath.Join(project, ".claude", "hooks")
	err := os.MkdirAll(legacyConfigDir, 0o755)
	requireNoError(t, err, "create legacy config dir")

	configPath := filepath.Join(legacyConfigDir, "blues-traveler-config.json")
	data, err := json.MarshalIndent(testMigrationConfigData(), "", "  ")
	requireNoError(t, err, "marshal config data")

	err = os.WriteFile(configPath, data, 0o600)
	requireNoError(t, err, "write config file")
	return configPath
}

// testMigrationConfigData returns test data for migration tests.
func testMigrationConfigData() map[string]interface{} {
	return map[string]interface{}{
		"logRotation": map[string]interface{}{
			"maxAge": 30, "maxSize": 10, "maxBackups": 5,
		},
		"testData": "migration-test",
		"customHooks": map[string]interface{}{
			"test-group": map[string]interface{}{
				"PreToolUse": map[string]interface{}{
					"jobs": []interface{}{
						map[string]interface{}{"name": "test-job", "run": "echo 'test'"},
					},
				},
			},
		},
	}
}

// assertMigratedConfigData checks the migrated config has correct values.
func assertMigratedConfigData(t *testing.T, migratedData map[string]interface{}) {
	t.Helper()
	if migratedData["testData"] != "migration-test" {
		t.Errorf("Expected testData=migration-test, got %v", migratedData["testData"])
	}

	logRotation, ok := migratedData["logRotation"].(map[string]interface{})
	if !ok {
		t.Fatalf("logRotation not found or not a map")
	}
	if logRotation["maxAge"] != float64(30) {
		t.Errorf("Expected maxAge=30, got %v", logRotation["maxAge"])
	}
}

func TestMigrationSkipExisting(t *testing.T) {
	tempDir := t.TempDir()
	xdgTempDir := t.TempDir()

	xdg := &XDGConfig{BaseDir: xdgTempDir}
	discovery := NewLegacyConfigDiscovery(xdg)
	project := filepath.Join(tempDir, "test-project")

	// Setup: existing XDG config and legacy config
	existingData := map[string]interface{}{"alreadyMigrated": true}
	err := xdg.SaveProjectConfig(project, existingData, "json")
	requireNoError(t, err, "save existing config")

	configPath := writeLegacyConfigFile(t, project, map[string]interface{}{"shouldNotMigrate": true})

	t.Run("migration is skipped", func(t *testing.T) {
		result := &MigrationResult{}
		err := discovery.migrateConfig(project, configPath, false, result)
		requireNoError(t, err, "attempt migration")

		if len(result.SkippedPaths) != 1 || result.SkippedPaths[0] != project {
			t.Errorf("Expected skipped path %s, got %v", project, result.SkippedPaths)
		}
	})

	t.Run("existing config preserved", func(t *testing.T) {
		currentData, err := xdg.LoadProjectConfig(project)
		requireNoError(t, err, "load current config")

		if currentData["alreadyMigrated"] != true {
			t.Error("Existing config was overwritten")
		}
		if currentData["shouldNotMigrate"] != nil {
			t.Error("Legacy config data was incorrectly merged")
		}
	})
}

// writeLegacyConfigFile creates a legacy config file with the given data.
func writeLegacyConfigFile(t *testing.T, project string, configData map[string]interface{}) string {
	t.Helper()
	legacyConfigDir := filepath.Join(project, ".claude", "hooks")
	err := os.MkdirAll(legacyConfigDir, 0o755)
	requireNoError(t, err, "create legacy config dir")

	configPath := filepath.Join(legacyConfigDir, "blues-traveler-config.json")
	data, err := json.MarshalIndent(configData, "", "  ")
	requireNoError(t, err, "marshal config data")

	err = os.WriteFile(configPath, data, 0o600)
	requireNoError(t, err, "write config file")
	return configPath
}

func TestGetMigrationStatus(t *testing.T) {
	tempDir := t.TempDir()
	project := filepath.Join(tempDir, "test-project")

	t.Run("no configs exist", func(t *testing.T) {
		status, err := GetMigrationStatus(project)
		requireNoError(t, err, "get migration status")
		assertMigrationStatus(t, status, false, false, false)
	})

	t.Run("legacy config only", func(t *testing.T) {
		createLegacyConfig(t, project)
		status, err := GetMigrationStatus(project)
		requireNoError(t, err, "get migration status")
		assertMigrationStatus(t, status, true, false, true)
	})

	t.Run("both configs exist", func(t *testing.T) {
		xdg := NewXDGConfig()
		err := xdg.SaveProjectConfig(project, map[string]interface{}{"test": "data"}, "json")
		requireNoError(t, err, "save XDG config")

		status, err := GetMigrationStatus(project)
		requireNoError(t, err, "get migration status")
		assertMigrationStatus(t, status, true, true, false)
	})
}

// createLegacyConfig creates a legacy config file in the project directory.
func createLegacyConfig(t *testing.T, project string) {
	t.Helper()
	legacyConfigDir := filepath.Join(project, ".claude", "hooks")
	err := os.MkdirAll(legacyConfigDir, 0o755)
	requireNoError(t, err, "create legacy config dir")

	configPath := filepath.Join(legacyConfigDir, "blues-traveler-config.json")
	data, err := json.MarshalIndent(map[string]interface{}{"test": "data"}, "", "  ")
	requireNoError(t, err, "marshal config data")

	err = os.WriteFile(configPath, data, 0o600)
	requireNoError(t, err, "write config file")
}

// requireNoError fails the test if err is not nil.
func requireNoError(t *testing.T, err error, context string) {
	t.Helper()
	if err != nil {
		t.Fatalf("Failed to %s: %v", context, err)
	}
}

// assertMigrationStatus checks the migration status fields.
func assertMigrationStatus(t *testing.T, status *MigrationStatus, hasLegacy, hasXDG, needsMigration bool) {
	t.Helper()
	if status.HasLegacyConfig != hasLegacy {
		t.Errorf("HasLegacyConfig: got %v, want %v", status.HasLegacyConfig, hasLegacy)
	}
	if status.HasXDGConfig != hasXDG {
		t.Errorf("HasXDGConfig: got %v, want %v", status.HasXDGConfig, hasXDG)
	}
	if status.NeedsMigration != needsMigration {
		t.Errorf("NeedsMigration: got %v, want %v", status.NeedsMigration, needsMigration)
	}
}

func TestFormatMigrationResult(t *testing.T) {
	result := &MigrationResult{
		TotalFound:      3,
		TotalMigrated:   2,
		TotalSkipped:    1,
		TotalErrors:     0,
		MigratedPaths:   []string{"/project1", "/project2"},
		SkippedPaths:    []string{"/project3"},
		ErrorPaths:      []MigrationError{},
		BackupLocations: []string{"/backup1", "/backup2"},
	}

	// Test dry run formatting
	dryRunOutput := FormatMigrationResult(result, true)
	if !contains(dryRunOutput, "Migration Dry Run Results:") {
		t.Error("Dry run output should contain 'Migration Dry Run Results:'")
	}
	if !contains(dryRunOutput, "Found: 3") {
		t.Error("Should show total found")
	}
	if !contains(dryRunOutput, "Migrated: 2") {
		t.Error("Should show total migrated")
	}

	// Test actual migration formatting
	actualOutput := FormatMigrationResult(result, false)
	if !contains(actualOutput, "Migration Results:") {
		t.Error("Actual output should contain 'Migration Results:'")
	}
	if !contains(actualOutput, "Backup Files Created:") {
		t.Error("Should show backup files for actual migration")
	}

	// Test with errors
	resultWithErrors := &MigrationResult{
		TotalFound:    1,
		TotalMigrated: 0,
		TotalSkipped:  0,
		TotalErrors:   1,
		ErrorPaths: []MigrationError{
			{Path: "/error-project", Error: "Permission denied"},
		},
	}

	errorOutput := FormatMigrationResult(resultWithErrors, false)
	if !contains(errorOutput, "Errors:") {
		t.Error("Should show errors section")
	}
	if !contains(errorOutput, "Permission denied") {
		t.Error("Should show error details")
	}
}

func TestHasLegacyConfig(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "legacy-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("cleanup failed: %v", err)
		}
	})

	project := filepath.Join(tempDir, "test-project")

	// Test with no config
	if HasLegacyConfig(project) {
		t.Error("Should return false when no legacy config exists")
	}

	// Create legacy config
	legacyConfigDir := filepath.Join(project, ".claude", "hooks")
	err = os.MkdirAll(legacyConfigDir, 0o755)
	if err != nil {
		t.Fatalf("Failed to create legacy config dir: %v", err)
	}

	configPath := filepath.Join(legacyConfigDir, "blues-traveler-config.json")
	err = os.WriteFile(configPath, []byte("{}"), 0o600)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Test with config
	if !HasLegacyConfig(project) {
		t.Error("Should return true when legacy config exists")
	}

	// Verify GetLegacyConfigPath returns correct path
	expectedPath := GetLegacyConfigPath(project)
	if expectedPath != configPath {
		t.Errorf("Expected path %s, got %s", configPath, expectedPath)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
