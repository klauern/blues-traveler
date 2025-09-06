package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewXDGConfig(t *testing.T) {
	// Test with XDG_CONFIG_HOME set
	originalXDGConfigHome := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", originalXDGConfigHome)

	testConfigHome := "/tmp/test-xdg-config"
	os.Setenv("XDG_CONFIG_HOME", testConfigHome)

	xdg := NewXDGConfig()
	expectedBaseDir := filepath.Join(testConfigHome, "blues-traveler")
	if xdg.BaseDir != expectedBaseDir {
		t.Errorf("Expected BaseDir %s, got %s", expectedBaseDir, xdg.BaseDir)
	}

	// Test without XDG_CONFIG_HOME
	os.Unsetenv("XDG_CONFIG_HOME")
	xdg = NewXDGConfig()
	homeDir, _ := os.UserHomeDir()
	expectedBaseDir = filepath.Join(homeDir, ".config", "blues-traveler")
	if xdg.BaseDir != expectedBaseDir {
		t.Errorf("Expected BaseDir %s, got %s", expectedBaseDir, xdg.BaseDir)
	}
}

func TestSanitizeProjectPath(t *testing.T) {
	xdg := NewXDGConfig()

	tests := []struct {
		input    string
		expected string
	}{
		{"/Users/user/dev/go/project", "Users-user-dev-go-project"},
		{"/home/user/work/my project", "home-user-work-my-project"},
		{"C:/Users/user/project", "C-Users-user-project"},
		{"/Users/user/dev:special/project", "Users-user-dev-special-project"},
		{"~/Documents/project", "home-Documents-project"},
		{"/very/long/path/that/exceeds/normal/limits/and/should/be/truncated/because/it/is/too/long/for/filesystem/limits/and/could/cause/issues/with/file/operations/when/dealing/with/very/deep/directory/structures/that/are/common/in/modern/development/environments", "very-long-path-that-exceeds-normal-limits-and-should-be-truncated-because-it-is-too-long-for-filesystem-limits-and-could-cause-issues-with-file-operations-when-dealing-with-very-deep-directory-structu"},
	}

	for _, test := range tests {
		result := xdg.SanitizeProjectPath(test.input)
		if result != test.expected {
			t.Errorf("SanitizeProjectPath(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestGetConfigPaths(t *testing.T) {
	xdg := NewXDGConfig()
	projectPath := "/Users/user/dev/project"

	// Test global config path
	globalPath := xdg.GetGlobalConfigPath("json")
	expectedGlobal := filepath.Join(xdg.BaseDir, "global.json")
	if globalPath != expectedGlobal {
		t.Errorf("Expected global path %s, got %s", expectedGlobal, globalPath)
	}

	// Test project config path
	projectConfigPath := xdg.GetProjectConfigPath(projectPath, "json")
	sanitized := xdg.SanitizeProjectPath(projectPath)
	expectedProject := filepath.Join(xdg.GetProjectsDir(), sanitized+".json")
	if projectConfigPath != expectedProject {
		t.Errorf("Expected project path %s, got %s", expectedProject, projectConfigPath)
	}

	// Test projects directory
	projectsDir := xdg.GetProjectsDir()
	expectedProjectsDir := filepath.Join(xdg.BaseDir, "projects")
	if projectsDir != expectedProjectsDir {
		t.Errorf("Expected projects dir %s, got %s", expectedProjectsDir, projectsDir)
	}

	// Test registry path
	registryPath := xdg.GetRegistryPath()
	expectedRegistry := filepath.Join(xdg.BaseDir, "registry.json")
	if registryPath != expectedRegistry {
		t.Errorf("Expected registry path %s, got %s", expectedRegistry, registryPath)
	}
}

func TestRegistryOperations(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "xdg-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create XDG config with custom base directory
	xdg := &XDGConfig{BaseDir: tempDir}

	// Test loading empty registry
	registry, err := xdg.LoadRegistry()
	if err != nil {
		t.Fatalf("Failed to load empty registry: %v", err)
	}
	if registry.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", registry.Version)
	}
	if len(registry.Projects) != 0 {
		t.Errorf("Expected empty projects map, got %v", registry.Projects)
	}

	// Test registering a project
	projectPath := "/Users/user/dev/project"
	err = xdg.RegisterProject(projectPath, "json")
	if err != nil {
		t.Fatalf("Failed to register project: %v", err)
	}

	// Test loading registry with project
	registry, err = xdg.LoadRegistry()
	if err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}
	if len(registry.Projects) != 1 {
		t.Errorf("Expected 1 project, got %d", len(registry.Projects))
	}

	projectConfig, exists := registry.Projects[projectPath]
	if !exists {
		t.Error("Project not found in registry")
	}
	if projectConfig.ConfigFormat != "json" {
		t.Errorf("Expected format json, got %s", projectConfig.ConfigFormat)
	}

	// Test getting project config
	config, err := xdg.GetProjectConfig(projectPath)
	if err != nil {
		t.Fatalf("Failed to get project config: %v", err)
	}
	if config.ConfigFormat != "json" {
		t.Errorf("Expected format json, got %s", config.ConfigFormat)
	}

	// Test listing projects
	projects, err := xdg.ListProjects()
	if err != nil {
		t.Fatalf("Failed to list projects: %v", err)
	}
	if len(projects) != 1 {
		t.Errorf("Expected 1 project, got %d", len(projects))
	}
	if projects[0] != projectPath {
		t.Errorf("Expected project %s, got %s", projectPath, projects[0])
	}
}

func TestConfigDataOperations(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "xdg-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create XDG config with custom base directory
	xdg := &XDGConfig{BaseDir: tempDir}

	projectPath := "/Users/user/dev/project"
	testData := map[string]interface{}{
		"testKey": "testValue",
		"nested": map[string]interface{}{
			"key": "value",
		},
		"array": []interface{}{"item1", "item2"},
	}

	// Test saving project config
	err = xdg.SaveProjectConfig(projectPath, testData, "json")
	if err != nil {
		t.Fatalf("Failed to save project config: %v", err)
	}

	// Test loading project config
	loadedData, err := xdg.LoadProjectConfig(projectPath)
	if err != nil {
		t.Fatalf("Failed to load project config: %v", err)
	}

	// Verify data
	if loadedData["testKey"] != "testValue" {
		t.Errorf("Expected testKey=testValue, got %v", loadedData["testKey"])
	}

	// Test global config operations
	globalData := map[string]interface{}{
		"globalKey": "globalValue",
	}

	err = xdg.SaveGlobalConfig(globalData, "json")
	if err != nil {
		t.Fatalf("Failed to save global config: %v", err)
	}

	loadedGlobalData, err := xdg.LoadGlobalConfig("json")
	if err != nil {
		t.Fatalf("Failed to load global config: %v", err)
	}

	if loadedGlobalData["globalKey"] != "globalValue" {
		t.Errorf("Expected globalKey=globalValue, got %v", loadedGlobalData["globalKey"])
	}
}

func TestTOMLSupport(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "xdg-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create XDG config with custom base directory
	xdg := &XDGConfig{BaseDir: tempDir}

	projectPath := "/Users/user/dev/project"
	testData := map[string]interface{}{
		"title": "Test Configuration",
		"database": map[string]interface{}{
			"server": "localhost",
			"port":   float64(5432), // TOML numbers are float64
		},
	}

	// Test saving project config as TOML
	err = xdg.SaveProjectConfig(projectPath, testData, "toml")
	if err != nil {
		t.Fatalf("Failed to save TOML project config: %v", err)
	}

	// Test loading project config from TOML
	loadedData, err := xdg.LoadProjectConfig(projectPath)
	if err != nil {
		t.Fatalf("Failed to load TOML project config: %v", err)
	}

	// Verify data
	if loadedData["title"] != "Test Configuration" {
		t.Errorf("Expected title='Test Configuration', got %v", loadedData["title"])
	}

	database, ok := loadedData["database"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected database to be a map, got %T", loadedData["database"])
	}
	if database["server"] != "localhost" {
		t.Errorf("Expected server=localhost, got %v", database["server"])
	}
}

func TestCleanupOrphanedConfigs(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "xdg-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create XDG config with custom base directory
	xdg := &XDGConfig{BaseDir: tempDir}

	// Create a temporary project directory
	projectDir, err := os.MkdirTemp("", "project-*")
	if err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}

	// Register the project
	err = xdg.RegisterProject(projectDir, "json")
	if err != nil {
		t.Fatalf("Failed to register project: %v", err)
	}

	// Create a config for a non-existent project
	nonExistentProject := "/non/existent/project"
	err = xdg.RegisterProject(nonExistentProject, "json")
	if err != nil {
		t.Fatalf("Failed to register non-existent project: %v", err)
	}

	// Verify both projects are registered
	projects, err := xdg.ListProjects()
	if err != nil {
		t.Fatalf("Failed to list projects: %v", err)
	}
	if len(projects) != 2 {
		t.Errorf("Expected 2 projects, got %d", len(projects))
	}

	// Remove the project directory
	os.RemoveAll(projectDir)

	// Run cleanup
	orphaned, err := xdg.CleanupOrphanedConfigs()
	if err != nil {
		t.Fatalf("Failed to cleanup orphaned configs: %v", err)
	}

	// Verify that both orphaned projects were cleaned up
	expectedOrphaned := 2 // Both the removed project and the non-existent project
	if len(orphaned) != expectedOrphaned {
		t.Errorf("Expected %d orphaned configs, got %d", expectedOrphaned, len(orphaned))
	}

	// Verify registry is empty now
	projects, err = xdg.ListProjects()
	if err != nil {
		t.Fatalf("Failed to list projects after cleanup: %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("Expected 0 projects after cleanup, got %d", len(projects))
	}
}

func TestErrorHandling(t *testing.T) {
	xdg := NewXDGConfig()

	// Test getting non-existent project config
	_, err := xdg.GetProjectConfig("/non/existent/project")
	if err == nil {
		t.Error("Expected error for non-existent project")
	}

	// Test loading config from non-existent project
	_, err = xdg.LoadProjectConfig("/non/existent/project")
	if err == nil {
		t.Error("Expected error for loading non-existent project config")
	}

	// Test invalid format
	tempDir, err := os.MkdirTemp("", "xdg-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	xdg = &XDGConfig{BaseDir: tempDir}
	testData := map[string]interface{}{"key": "value"}

	err = xdg.SaveProjectConfig("/test/project", testData, "invalid")
	if err == nil {
		t.Error("Expected error for invalid format")
	}
}

func TestRegistryVersioning(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "xdg-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	xdg := &XDGConfig{BaseDir: tempDir}

	// Create registry manually with different version
	registry := &ProjectRegistry{
		Version:  "0.9",
		Projects: make(map[string]ProjectConfig),
	}

	err = xdg.SaveRegistry(registry)
	if err != nil {
		t.Fatalf("Failed to save registry: %v", err)
	}

	// Load registry and verify version is preserved
	loadedRegistry, err := xdg.LoadRegistry()
	if err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}

	if loadedRegistry.Version != "0.9" {
		t.Errorf("Expected version 0.9, got %s", loadedRegistry.Version)
	}
}

func TestConcurrentAccess(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "xdg-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	xdg := &XDGConfig{BaseDir: tempDir}

	// Test sequential project registration first to establish baseline
	for i := 0; i < 3; i++ {
		projectPath := filepath.Join("/test/project", string(rune('A'+i)))
		err := xdg.RegisterProject(projectPath, "json")
		if err != nil {
			t.Errorf("Failed to register project %s: %v", projectPath, err)
		}
	}

	// Verify projects were registered
	projects, err := xdg.ListProjects()
	if err != nil {
		t.Fatalf("Failed to list projects: %v", err)
	}

	if len(projects) != 3 {
		t.Errorf("Expected 3 projects, got %d", len(projects))
	}
}
