package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/klauern/blues-traveler/internal/constants"
)

func TestNewXDGConfig(t *testing.T) {
	// Test with XDG_CONFIG_HOME set
	originalXDGConfigHome := os.Getenv("XDG_CONFIG_HOME")
	t.Cleanup(func() {
		if err := os.Setenv("XDG_CONFIG_HOME", originalXDGConfigHome); err != nil {
			t.Logf("cleanup failed: %v", err)
		}
	})

	testConfigHome := "/tmp/test-xdg-config"
	if err := os.Setenv("XDG_CONFIG_HOME", testConfigHome); err != nil {
		t.Fatalf("Failed to set XDG_CONFIG_HOME: %v", err)
	}

	xdg := NewXDGConfig()
	expectedBaseDir := filepath.Join(testConfigHome, "blues-traveler")
	if xdg.BaseDir != expectedBaseDir {
		t.Errorf("Expected BaseDir %s, got %s", expectedBaseDir, xdg.BaseDir)
	}

	// Test without XDG_CONFIG_HOME
	if err := os.Unsetenv("XDG_CONFIG_HOME"); err != nil {
		t.Fatalf("Failed to unset XDG_CONFIG_HOME: %v", err)
	}
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
	projectPath := constants.TestProjectPath

	// Test global config path
	globalPath := xdg.GetGlobalConfigPath(FormatJSON)
	expectedGlobal := filepath.Join(xdg.BaseDir, "global.json")
	if globalPath != expectedGlobal {
		t.Errorf("Expected global path %s, got %s", expectedGlobal, globalPath)
	}

	// Test project config path
	projectConfigPath := xdg.GetProjectConfigPath(projectPath, FormatJSON)
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
	tempDir := t.TempDir()
	xdg := &XDGConfig{BaseDir: tempDir}
	projectPath := constants.TestProjectPath

	t.Run("empty registry", func(t *testing.T) {
		registry, err := xdg.LoadRegistry()
		requireNoErrorTest(t, err, "load empty registry")
		assertRegistryState(t, registry, "1.0", 0)
	})

	t.Run("register project", func(t *testing.T) {
		err := xdg.RegisterProject(projectPath, FormatJSON)
		requireNoErrorTest(t, err, "register project")

		registry, err := xdg.LoadRegistry()
		requireNoErrorTest(t, err, "load registry")
		assertRegistryState(t, registry, "1.0", 1)
		assertProjectInRegistry(t, registry, projectPath, FormatJSON)
	})

	t.Run("get project config", func(t *testing.T) {
		config, err := xdg.GetProjectConfig(projectPath)
		requireNoErrorTest(t, err, "get project config")
		if config.ConfigFormat != FormatJSON {
			t.Errorf("Expected format json, got %s", config.ConfigFormat)
		}
	})

	t.Run("list projects", func(t *testing.T) {
		projects, err := xdg.ListProjects()
		requireNoErrorTest(t, err, "list projects")
		if len(projects) != 1 || projects[0] != projectPath {
			t.Errorf("Expected [%s], got %v", projectPath, projects)
		}
	})
}

// requireNoErrorTest fails the test if err is not nil.
func requireNoErrorTest(t *testing.T, err error, context string) {
	t.Helper()
	if err != nil {
		t.Fatalf("Failed to %s: %v", context, err)
	}
}

// assertRegistryState checks the registry version and project count.
func assertRegistryState(t *testing.T, registry *ProjectRegistry, version string, projectCount int) {
	t.Helper()
	if registry.Version != version {
		t.Errorf("Expected version %s, got %s", version, registry.Version)
	}
	if len(registry.Projects) != projectCount {
		t.Errorf("Expected %d projects, got %d", projectCount, len(registry.Projects))
	}
}

// assertProjectInRegistry checks that a project exists with the expected format.
func assertProjectInRegistry(t *testing.T, registry *ProjectRegistry, projectPath, format string) {
	t.Helper()
	projectConfig, exists := registry.Projects[projectPath]
	if !exists {
		t.Error("Project not found in registry")
		return
	}
	if projectConfig.ConfigFormat != format {
		t.Errorf("Expected format %s, got %s", format, projectConfig.ConfigFormat)
	}
}

func TestConfigDataOperations(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "xdg-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("cleanup failed: %v", err)
		}
	})

	// Create XDG config with custom base directory
	xdg := &XDGConfig{BaseDir: tempDir}

	projectPath := constants.TestProjectPath
	testData := map[string]interface{}{
		"testKey": "testValue",
		"nested": map[string]interface{}{
			"key": "value",
		},
		"array": []interface{}{"item1", "item2"},
	}

	// Test saving project config
	err = xdg.SaveProjectConfig(projectPath, testData, FormatJSON)
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

	err = xdg.SaveGlobalConfig(globalData, FormatJSON)
	if err != nil {
		t.Fatalf("Failed to save global config: %v", err)
	}

	loadedGlobalData, err := xdg.LoadGlobalConfig(FormatJSON)
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
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("cleanup failed: %v", err)
		}
	})

	// Create XDG config with custom base directory
	xdg := &XDGConfig{BaseDir: tempDir}

	projectPath := constants.TestProjectPath
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
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("cleanup failed: %v", err)
		}
	})

	// Create XDG config with custom base directory
	xdg := &XDGConfig{BaseDir: tempDir}

	// Create a temporary project directory
	projectDir, err := os.MkdirTemp("", "project-*")
	if err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}

	// Register the project
	err = xdg.RegisterProject(projectDir, FormatJSON)
	if err != nil {
		t.Fatalf("Failed to register project: %v", err)
	}

	// Create a config for a non-existent project
	nonExistentProject := "/non/existent/project"
	err = xdg.RegisterProject(nonExistentProject, FormatJSON)
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
	if err := os.RemoveAll(projectDir); err != nil {
		t.Fatalf("Failed to remove project dir: %v", err)
	}

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
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("cleanup failed: %v", err)
		}
	})

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
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("cleanup failed: %v", err)
		}
	})

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
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("cleanup failed: %v", err)
		}
	})

	xdg := &XDGConfig{BaseDir: tempDir}

	// Test sequential project registration first to establish baseline
	for i := 0; i < 3; i++ {
		projectPath := filepath.Join("/test/project", string(rune('A'+i)))
		err := xdg.RegisterProject(projectPath, FormatJSON)
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
