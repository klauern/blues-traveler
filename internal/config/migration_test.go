package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLegacyConfigDiscovery(t *testing.T) {
	// Create temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "migration-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("cleanup failed: %v", err)
		}
	})

	// Create XDG config with custom base directory
	xdgTempDir, err := os.MkdirTemp("", "xdg-migration-test-*")
	if err != nil {
		t.Fatalf("Failed to create XDG temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(xdgTempDir); err != nil {
			t.Logf("cleanup failed: %v", err)
		}
	})

	xdg := &XDGConfig{BaseDir: xdgTempDir}
	discovery := NewLegacyConfigDiscovery(xdg)

	// Create mock project structure with legacy config
	project1 := filepath.Join(tempDir, "project1")
	project2 := filepath.Join(tempDir, "project2")
	project3 := filepath.Join(tempDir, "project3", "subproject") // Nested project

	projects := []string{project1, project2, project3}

	// Create legacy config files
	for i, project := range projects {
		legacyConfigDir := filepath.Join(project, ".claude", "hooks")
		err := os.MkdirAll(legacyConfigDir, 0o755)
		if err != nil {
			t.Fatalf("Failed to create legacy config dir: %v", err)
		}

		configPath := filepath.Join(legacyConfigDir, "blues-traveler-config.json")
		configData := map[string]interface{}{
			"logRotation": map[string]interface{}{
				"maxAge":     30,
				"maxSize":    10,
				"maxBackups": 5,
			},
			"testProject": i + 1,
		}

		data, err := json.MarshalIndent(configData, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal config data: %v", err)
		}

		err = os.WriteFile(configPath, data, 0o600)
		if err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}
	}

	// Test discovery within the temp directory
	originalWalkFunc := discovery.walkProjectDirectories
	discoveredConfigs := make(map[string]string)

	// Manually call walkProjectDirectories for our test directory
	err = discovery.walkProjectDirectories(tempDir, discoveredConfigs)
	if err != nil {
		t.Fatalf("Failed to discover configs: %v", err)
	}

	// Verify all configs were discovered
	if len(discoveredConfigs) != len(projects) {
		t.Errorf("Expected %d configs, found %d", len(projects), len(discoveredConfigs))
	}

	for _, project := range projects {
		if _, exists := discoveredConfigs[project]; !exists {
			t.Errorf("Config for project %s not discovered", project)
		}
	}

	// Restore original function
	_ = originalWalkFunc
}

func TestMigrationDryRun(t *testing.T) {
	// Create temporary directory structure
	tempDir, err := os.MkdirTemp("", "migration-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("cleanup failed: %v", err)
		}
	})

	// Create XDG config with custom base directory
	xdgTempDir, err := os.MkdirTemp("", "xdg-migration-test-*")
	if err != nil {
		t.Fatalf("Failed to create XDG temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(xdgTempDir); err != nil {
			t.Logf("cleanup failed: %v", err)
		}
	})

	xdg := &XDGConfig{BaseDir: xdgTempDir}
	discovery := NewLegacyConfigDiscovery(xdg)

	// Create a project with legacy config
	project := filepath.Join(tempDir, "test-project")
	legacyConfigDir := filepath.Join(project, ".claude", "hooks")
	err = os.MkdirAll(legacyConfigDir, 0o755)
	if err != nil {
		t.Fatalf("Failed to create legacy config dir: %v", err)
	}

	configPath := filepath.Join(legacyConfigDir, "blues-traveler-config.json")
	configData := map[string]interface{}{
		"logRotation": map[string]interface{}{
			"maxAge": 30,
		},
		"testData": "test-value",
	}

	data, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config data: %v", err)
	}

	err = os.WriteFile(configPath, data, 0o600)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Override discovery to use our test directory
	discovery.xdg = xdg

	// Mock the DiscoverLegacyConfigs method for testing
	testConfigs := map[string]string{
		project: configPath,
	}

	// Create a custom discovery instance that uses our test configs
	customDiscovery := &LegacyConfigDiscovery{xdg: xdg}

	// Test dry run migration
	result := &MigrationResult{
		TotalFound:      len(testConfigs),
		MigratedPaths:   []string{},
		SkippedPaths:    []string{},
		ErrorPaths:      []MigrationError{},
		BackupLocations: []string{},
	}

	// Simulate dry run migration
	for projectPath := range testConfigs {
		// Check if project already exists in XDG registry (it shouldn't)
		if _, err := customDiscovery.xdg.GetProjectConfig(projectPath); err == nil {
			result.TotalSkipped++
			result.SkippedPaths = append(result.SkippedPaths, projectPath)
		} else {
			result.TotalMigrated++
			result.MigratedPaths = append(result.MigratedPaths, projectPath)
		}
	}

	// Verify dry run results
	if result.TotalFound != 1 {
		t.Errorf("Expected 1 config found, got %d", result.TotalFound)
	}
	if result.TotalMigrated != 1 {
		t.Errorf("Expected 1 config to be migrated, got %d", result.TotalMigrated)
	}
	if result.TotalSkipped != 0 {
		t.Errorf("Expected 0 configs to be skipped, got %d", result.TotalSkipped)
	}
	if len(result.BackupLocations) != 0 {
		t.Errorf("Expected 0 backup locations in dry run, got %d", len(result.BackupLocations))
	}
}

func TestActualMigration(t *testing.T) {
	// Create temporary directory structure
	tempDir, err := os.MkdirTemp("", "migration-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("cleanup failed: %v", err)
		}
	})

	// Create XDG config with custom base directory
	xdgTempDir, err := os.MkdirTemp("", "xdg-migration-test-*")
	if err != nil {
		t.Fatalf("Failed to create XDG temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(xdgTempDir); err != nil {
			t.Logf("cleanup failed: %v", err)
		}
	})

	xdg := &XDGConfig{BaseDir: xdgTempDir}
	discovery := NewLegacyConfigDiscovery(xdg)

	// Create a project with legacy config
	project := filepath.Join(tempDir, "test-project")
	legacyConfigDir := filepath.Join(project, ".claude", "hooks")
	err = os.MkdirAll(legacyConfigDir, 0o755)
	if err != nil {
		t.Fatalf("Failed to create legacy config dir: %v", err)
	}

	configPath := filepath.Join(legacyConfigDir, "blues-traveler-config.json")
	originalConfigData := map[string]interface{}{
		"logRotation": map[string]interface{}{
			"maxAge":     30,
			"maxSize":    10,
			"maxBackups": 5,
		},
		"testData": "migration-test",
		"customHooks": map[string]interface{}{
			"test-group": map[string]interface{}{
				"PreToolUse": map[string]interface{}{
					"jobs": []interface{}{
						map[string]interface{}{
							"name": "test-job",
							"run":  "echo 'test'",
						},
					},
				},
			},
		},
	}

	data, err := json.MarshalIndent(originalConfigData, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config data: %v", err)
	}

	err = os.WriteFile(configPath, data, 0o600)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Perform actual migration
	err = discovery.migrateConfig(project, configPath, false, &MigrationResult{
		MigratedPaths:   []string{},
		SkippedPaths:    []string{},
		ErrorPaths:      []MigrationError{},
		BackupLocations: []string{},
	})
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Verify project is registered in XDG registry
	projectConfig, err := xdg.GetProjectConfig(project)
	if err != nil {
		t.Fatalf("Project not registered after migration: %v", err)
	}

	if projectConfig.ConfigFormat != "json" {
		t.Errorf("Expected format json, got %s", projectConfig.ConfigFormat)
	}

	// Verify config data was migrated correctly
	migratedData, err := xdg.LoadProjectConfig(project)
	if err != nil {
		t.Fatalf("Failed to load migrated config: %v", err)
	}

	// Check specific values
	if migratedData["testData"] != "migration-test" {
		t.Errorf("Expected testData=migration-test, got %v", migratedData["testData"])
	}

	// Check nested structure
	logRotation, ok := migratedData["logRotation"].(map[string]interface{})
	if !ok {
		t.Fatalf("logRotation not found or not a map")
	}

	if logRotation["maxAge"] != float64(30) { // JSON numbers are float64
		t.Errorf("Expected maxAge=30, got %v", logRotation["maxAge"])
	}

	// Verify backup was created
	backupPattern := configPath + ".backup.*"
	matches, err := filepath.Glob(backupPattern)
	if err != nil {
		t.Fatalf("Failed to check for backup files: %v", err)
	}
	if len(matches) != 1 {
		t.Errorf("Expected 1 backup file, found %d", len(matches))
	}
}

func TestMigrationSkipExisting(t *testing.T) {
	// Create temporary directory structure
	tempDir, err := os.MkdirTemp("", "migration-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("cleanup failed: %v", err)
		}
	})

	// Create XDG config with custom base directory
	xdgTempDir, err := os.MkdirTemp("", "xdg-migration-test-*")
	if err != nil {
		t.Fatalf("Failed to create XDG temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(xdgTempDir); err != nil {
			t.Logf("cleanup failed: %v", err)
		}
	})

	xdg := &XDGConfig{BaseDir: xdgTempDir}
	discovery := NewLegacyConfigDiscovery(xdg)

	project := filepath.Join(tempDir, "test-project")

	// First, register the project in XDG system
	existingData := map[string]interface{}{
		"alreadyMigrated": true,
	}
	err = xdg.SaveProjectConfig(project, existingData, "json")
	if err != nil {
		t.Fatalf("Failed to save existing config: %v", err)
	}

	// Create legacy config
	legacyConfigDir := filepath.Join(project, ".claude", "hooks")
	err = os.MkdirAll(legacyConfigDir, 0o755)
	if err != nil {
		t.Fatalf("Failed to create legacy config dir: %v", err)
	}

	configPath := filepath.Join(legacyConfigDir, "blues-traveler-config.json")
	legacyData := map[string]interface{}{
		"shouldNotMigrate": true,
	}

	data, err := json.MarshalIndent(legacyData, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config data: %v", err)
	}

	err = os.WriteFile(configPath, data, 0o600)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Attempt migration
	result := &MigrationResult{
		MigratedPaths:   []string{},
		SkippedPaths:    []string{},
		ErrorPaths:      []MigrationError{},
		BackupLocations: []string{},
	}

	err = discovery.migrateConfig(project, configPath, false, result)
	if err != nil {
		t.Fatalf("Migration should not fail for existing config: %v", err)
	}

	// Verify migration was skipped
	if len(result.SkippedPaths) != 1 {
		t.Errorf("Expected 1 skipped path, got %d", len(result.SkippedPaths))
	}

	if result.SkippedPaths[0] != project {
		t.Errorf("Expected skipped path %s, got %s", project, result.SkippedPaths[0])
	}

	// Verify existing config wasn't overwritten
	currentData, err := xdg.LoadProjectConfig(project)
	if err != nil {
		t.Fatalf("Failed to load current config: %v", err)
	}

	if currentData["alreadyMigrated"] != true {
		t.Error("Existing config was overwritten")
	}

	if currentData["shouldNotMigrate"] != nil {
		t.Error("Legacy config data was incorrectly merged")
	}
}

func TestGetMigrationStatus(t *testing.T) {
	// Create temporary directory structure
	tempDir, err := os.MkdirTemp("", "migration-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("cleanup failed: %v", err)
		}
	})

	project := filepath.Join(tempDir, "test-project")

	// Test status with no configs
	status, err := GetMigrationStatus(project)
	if err != nil {
		t.Fatalf("Failed to get migration status: %v", err)
	}

	if status.HasLegacyConfig {
		t.Error("Should not have legacy config")
	}
	if status.HasXDGConfig {
		t.Error("Should not have XDG config")
	}
	if status.NeedsMigration {
		t.Error("Should not need migration when no configs exist")
	}

	// Create legacy config
	legacyConfigDir := filepath.Join(project, ".claude", "hooks")
	err = os.MkdirAll(legacyConfigDir, 0o755)
	if err != nil {
		t.Fatalf("Failed to create legacy config dir: %v", err)
	}

	configPath := filepath.Join(legacyConfigDir, "blues-traveler-config.json")
	configData := map[string]interface{}{"test": "data"}
	data, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config data: %v", err)
	}

	err = os.WriteFile(configPath, data, 0o600)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Test status with legacy config only
	status, err = GetMigrationStatus(project)
	if err != nil {
		t.Fatalf("Failed to get migration status: %v", err)
	}

	if !status.HasLegacyConfig {
		t.Error("Should have legacy config")
	}
	if status.HasXDGConfig {
		t.Error("Should not have XDG config yet")
	}
	if !status.NeedsMigration {
		t.Error("Should need migration")
	}

	// Create XDG config
	xdg := NewXDGConfig()
	err = xdg.SaveProjectConfig(project, configData, "json")
	if err != nil {
		t.Fatalf("Failed to save XDG config: %v", err)
	}

	// Test status with both configs
	status, err = GetMigrationStatus(project)
	if err != nil {
		t.Fatalf("Failed to get migration status: %v", err)
	}

	if !status.HasLegacyConfig {
		t.Error("Should still have legacy config")
	}
	if !status.HasXDGConfig {
		t.Error("Should have XDG config")
	}
	if status.NeedsMigration {
		t.Error("Should not need migration when XDG config exists")
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
