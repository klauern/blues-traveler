package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNewEnhancedConfigLoader(t *testing.T) {
	loader := NewEnhancedConfigLoader(LoadXDGFirst)
	if loader.strategy != LoadXDGFirst {
		t.Errorf("Expected strategy LoadXDGFirst, got %v", loader.strategy)
	}
	if loader.xdg == nil {
		t.Error("XDG config should not be nil")
	}
}

func TestLoadConfigWithFallback(t *testing.T) {
	t.Run("XDGOnly", testLoadXDGOnlyStrategy)
	t.Run("LegacyOnly", testLoadLegacyOnlyStrategy)
	t.Run("XDGFirstPreference", testXDGFirstPreference)
	t.Run("XDGFirstFallback", testXDGFirstFallback)
}

func testLoadXDGOnlyStrategy(t *testing.T) {
	tempDir := t.TempDir()
	xdgTempDir := t.TempDir()
	project := filepath.Join(tempDir, "test-project")

	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}

	loader := NewEnhancedConfigLoader(LoadXDGOnly)
	loader.xdg = &XDGConfig{BaseDir: xdgTempDir}

	// Should error when no XDG config exists
	_, _, err := loader.LoadConfigWithFallback(project)
	if err == nil {
		t.Error("Expected error when loading XDG-only with no XDG config")
	}

	// Create XDG config
	xdgConfigData := map[string]interface{}{
		"logRotation": map[string]interface{}{"maxAge": 45},
		"source":      "xdg",
	}
	if err := loader.xdg.SaveProjectConfig(project, xdgConfigData, "json"); err != nil {
		t.Fatalf("Failed to save XDG config: %v", err)
	}

	// Should load successfully now
	config, configPath, err := loader.LoadConfigWithFallback(project)
	if err != nil {
		t.Fatalf("Failed to load XDG config: %v", err)
	}
	if config.Other["source"] != "xdg" {
		t.Errorf("Expected source=xdg, got %v", config.Other["source"])
	}
	if !filepath.IsAbs(configPath) {
		t.Errorf("Config path should be absolute: %s", configPath)
	}
}

func testLoadLegacyOnlyStrategy(t *testing.T) {
	tempDir := t.TempDir()
	project := filepath.Join(tempDir, "test-project")

	legacyConfigDir := filepath.Join(project, ".claude", "hooks")
	if err := os.MkdirAll(legacyConfigDir, 0o755); err != nil {
		t.Fatalf("Failed to create legacy config dir: %v", err)
	}

	legacyConfigPath := filepath.Join(legacyConfigDir, "blues-traveler-config.json")
	legacyConfigData := map[string]interface{}{
		"logRotation": map[string]interface{}{"maxAge": 30},
		"source":      "legacy",
	}
	data, err := json.MarshalIndent(legacyConfigData, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal legacy config: %v", err)
	}
	if err := os.WriteFile(legacyConfigPath, data, 0o600); err != nil {
		t.Fatalf("Failed to write legacy config: %v", err)
	}

	legacyLoader := NewEnhancedConfigLoader(LoadLegacyOnly)
	legacyConfig, legacyPath, err := legacyLoader.LoadConfigWithFallback(project)
	if err != nil {
		t.Fatalf("Failed to load legacy config: %v", err)
	}
	if legacyConfig.Other["source"] != "legacy" {
		t.Errorf("Expected source=legacy, got %v", legacyConfig.Other["source"])
	}
	if legacyPath != legacyConfigPath {
		t.Errorf("Expected path %s, got %s", legacyConfigPath, legacyPath)
	}
}

func testXDGFirstPreference(t *testing.T) {
	tempDir := t.TempDir()
	xdgTempDir := t.TempDir()
	project := filepath.Join(tempDir, "test-project")

	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}

	// Create both XDG and legacy configs
	setupBothConfigs(t, project, xdgTempDir)

	xdgFirstLoader := NewEnhancedConfigLoader(LoadXDGFirst)
	xdgFirstLoader.xdg = &XDGConfig{BaseDir: xdgTempDir}

	firstConfig, _, err := xdgFirstLoader.LoadConfigWithFallback(project)
	if err != nil {
		t.Fatalf("Failed to load config with XDGFirst: %v", err)
	}
	if firstConfig.Other["source"] != "xdg" {
		t.Errorf("XDGFirst should prefer XDG config, got source=%v", firstConfig.Other["source"])
	}
}

func testXDGFirstFallback(t *testing.T) {
	tempDir := t.TempDir()
	xdgTempDir := t.TempDir()
	project := filepath.Join(tempDir, "test-project")

	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}

	// Create only legacy config
	legacyConfigPath := setupLegacyConfig(t, project)

	xdgFirstLoader := NewEnhancedConfigLoader(LoadXDGFirst)
	xdgFirstLoader.xdg = &XDGConfig{BaseDir: xdgTempDir}

	fallbackConfig, fallbackPath, err := xdgFirstLoader.LoadConfigWithFallback(project)
	if err != nil {
		t.Fatalf("Failed to load config with fallback: %v", err)
	}
	if fallbackConfig.Other["source"] != "legacy" {
		t.Errorf("Should fallback to legacy config, got source=%v", fallbackConfig.Other["source"])
	}
	if fallbackPath != legacyConfigPath {
		t.Errorf("Expected fallback path %s, got %s", legacyConfigPath, fallbackPath)
	}
}

func setupBothConfigs(t *testing.T, project, xdgTempDir string) {
	t.Helper()

	// Setup XDG config
	loader := NewEnhancedConfigLoader(LoadXDGOnly)
	loader.xdg = &XDGConfig{BaseDir: xdgTempDir}
	xdgConfigData := map[string]interface{}{
		"logRotation": map[string]interface{}{"maxAge": 45},
		"source":      "xdg",
	}
	if err := loader.xdg.SaveProjectConfig(project, xdgConfigData, "json"); err != nil {
		t.Fatalf("Failed to save XDG config: %v", err)
	}

	// Setup legacy config
	setupLegacyConfig(t, project)
}

func setupLegacyConfig(t *testing.T, project string) string {
	t.Helper()

	legacyConfigDir := filepath.Join(project, ".claude", "hooks")
	if err := os.MkdirAll(legacyConfigDir, 0o755); err != nil {
		t.Fatalf("Failed to create legacy config dir: %v", err)
	}

	legacyConfigPath := filepath.Join(legacyConfigDir, "blues-traveler-config.json")
	legacyConfigData := map[string]interface{}{
		"logRotation": map[string]interface{}{"maxAge": 30},
		"source":      "legacy",
	}
	data, err := json.MarshalIndent(legacyConfigData, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal legacy config: %v", err)
	}
	if err := os.WriteFile(legacyConfigPath, data, 0o600); err != nil {
		t.Fatalf("Failed to write legacy config: %v", err)
	}

	return legacyConfigPath
}

func TestLoadGlobalConfigWithFallback(t *testing.T) {
	// Create temporary directories
	xdgTempDir, err := os.MkdirTemp("", "xdg-global-test-*")
	if err != nil {
		t.Fatalf("Failed to create XDG temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(xdgTempDir); err != nil {
			t.Logf("cleanup failed: %v", err)
		}
	})

	legacyTempDir, err := os.MkdirTemp("", "legacy-global-test-*")
	if err != nil {
		t.Fatalf("Failed to create legacy temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(legacyTempDir); err != nil {
			t.Logf("cleanup failed: %v", err)
		}
	})

	// Test LoadXDGOnly strategy
	loader := NewEnhancedConfigLoader(LoadXDGOnly)
	loader.xdg = &XDGConfig{BaseDir: xdgTempDir}

	// Create XDG global config
	globalXDGData := map[string]interface{}{
		"logRotation": map[string]interface{}{
			"maxAge": 60,
		},
		"globalSource": "xdg",
	}
	err = loader.xdg.SaveGlobalConfig(globalXDGData, "json")
	if err != nil {
		t.Fatalf("Failed to save XDG global config: %v", err)
	}

	// Test loading XDG global config
	config, configPath, err := loader.LoadGlobalConfigWithFallback()
	if err != nil {
		t.Fatalf("Failed to load XDG global config: %v", err)
	}
	if config.Other["globalSource"] != "xdg" {
		t.Errorf("Expected globalSource=xdg, got %v", config.Other["globalSource"])
	}

	expectedPath := loader.xdg.GetGlobalConfigPath("json")
	if configPath != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, configPath)
	}
}

func TestSaveConfigWithXDG(t *testing.T) {
	// Create temporary directory
	xdgTempDir, err := os.MkdirTemp("", "xdg-save-test-*")
	if err != nil {
		t.Fatalf("Failed to create XDG temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(xdgTempDir); err != nil {
			t.Logf("cleanup failed: %v", err)
		}
	})

	tempDir, err := os.MkdirTemp("", "save-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("cleanup failed: %v", err)
		}
	})

	loader := NewEnhancedConfigLoader(LoadXDGFirst)
	loader.xdg = &XDGConfig{BaseDir: xdgTempDir}

	project := filepath.Join(tempDir, "test-project")

	// Create test config
	testConfig := &LogConfig{
		LogRotation: LogRotationConfig{
			MaxAge:     45,
			MaxSize:    20,
			MaxBackups: 7,
		},
		Other: map[string]interface{}{
			"testKey": "testValue",
		},
	}

	// Test saving project config
	err = loader.SaveConfigWithXDG(project, testConfig, "json")
	if err != nil {
		t.Fatalf("Failed to save config with XDG: %v", err)
	}

	// Verify config was saved and registered
	if !loader.IsProjectRegistered(project) {
		t.Error("Project should be registered after saving")
	}

	// Load and verify config
	loadedConfig, _, err := loader.LoadConfigWithFallback(project)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loadedConfig.LogRotation.MaxAge != 45 {
		t.Errorf("Expected MaxAge=45, got %d", loadedConfig.LogRotation.MaxAge)
	}
	if loadedConfig.Other["testKey"] != "testValue" {
		t.Errorf("Expected testKey=testValue, got %v", loadedConfig.Other["testKey"])
	}

	// Test saving global config
	err = loader.SaveGlobalConfigWithXDG(testConfig, "json")
	if err != nil {
		t.Fatalf("Failed to save global config with XDG: %v", err)
	}

	// Load and verify global config
	loadedGlobalConfig, _, err := loader.LoadGlobalConfigWithFallback()
	if err != nil {
		t.Fatalf("Failed to load saved global config: %v", err)
	}

	if loadedGlobalConfig.LogRotation.MaxSize != 20 {
		t.Errorf("Expected MaxSize=20, got %d", loadedGlobalConfig.LogRotation.MaxSize)
	}
}

func TestGetConfigPath(t *testing.T) {
	// Create temporary directories
	xdgTempDir, err := os.MkdirTemp("", "xdg-path-test-*")
	if err != nil {
		t.Fatalf("Failed to create XDG temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(xdgTempDir); err != nil {
			t.Logf("cleanup failed: %v", err)
		}
	})

	tempDir, err := os.MkdirTemp("", "path-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("cleanup failed: %v", err)
		}
	})

	project := filepath.Join(tempDir, "test-project")

	// Test with LoadLegacyOnly strategy
	legacyLoader := NewEnhancedConfigLoader(LoadLegacyOnly)

	// Test legacy project path
	legacyProjectPath, err := legacyLoader.GetConfigPath(project, false)
	if err != nil {
		t.Fatalf("Failed to get legacy project path: %v", err)
	}
	expectedLegacyPath := GetLegacyConfigPath(project)
	if legacyProjectPath != expectedLegacyPath {
		t.Errorf("Expected legacy path %s, got %s", expectedLegacyPath, legacyProjectPath)
	}

	// Test XDG paths
	xdgLoader := NewEnhancedConfigLoader(LoadXDGFirst)
	xdgLoader.xdg = &XDGConfig{BaseDir: xdgTempDir}

	// Test unregistered project (should return potential XDG path)
	xdgProjectPath, err := xdgLoader.GetConfigPath(project, false)
	if err != nil {
		t.Fatalf("Failed to get XDG project path: %v", err)
	}
	expectedXDGPath := xdgLoader.xdg.GetProjectConfigPath(project, "json")
	if xdgProjectPath != expectedXDGPath {
		t.Errorf("Expected XDG path %s, got %s", expectedXDGPath, xdgProjectPath)
	}

	// Register project and test again
	testData := map[string]interface{}{"test": "data"}
	err = xdgLoader.xdg.SaveProjectConfig(project, testData, "json")
	if err != nil {
		t.Fatalf("Failed to save XDG config: %v", err)
	}

	registeredProjectPath, err := xdgLoader.GetConfigPath(project, false)
	if err != nil {
		t.Fatalf("Failed to get registered project path: %v", err)
	}

	projectConfig, _ := xdgLoader.xdg.GetProjectConfig(project)
	expectedRegisteredPath := filepath.Join(xdgLoader.xdg.GetConfigDir(), projectConfig.ConfigFile)
	if registeredProjectPath != expectedRegisteredPath {
		t.Errorf("Expected registered path %s, got %s", expectedRegisteredPath, registeredProjectPath)
	}

	// Test global paths
	globalPath, err := xdgLoader.GetConfigPath("", true)
	if err != nil {
		t.Fatalf("Failed to get global path: %v", err)
	}
	expectedGlobalPath := xdgLoader.xdg.GetGlobalConfigPath("json")
	if globalPath != expectedGlobalPath {
		t.Errorf("Expected global path %s, got %s", expectedGlobalPath, globalPath)
	}
}

func TestIsProjectRegistered(t *testing.T) {
	// Create temporary directories
	xdgTempDir, err := os.MkdirTemp("", "xdg-registered-test-*")
	if err != nil {
		t.Fatalf("Failed to create XDG temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(xdgTempDir); err != nil {
			t.Logf("cleanup failed: %v", err)
		}
	})

	tempDir, err := os.MkdirTemp("", "registered-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("cleanup failed: %v", err)
		}
	})

	loader := NewEnhancedConfigLoader(LoadXDGFirst)
	loader.xdg = &XDGConfig{BaseDir: xdgTempDir}

	project := filepath.Join(tempDir, "test-project")

	// Test unregistered project
	if loader.IsProjectRegistered(project) {
		t.Error("Project should not be registered initially")
	}

	// Register project
	testData := map[string]interface{}{"test": "data"}
	err = loader.xdg.SaveProjectConfig(project, testData, "json")
	if err != nil {
		t.Fatalf("Failed to save XDG config: %v", err)
	}

	// Test registered project
	if !loader.IsProjectRegistered(project) {
		t.Error("Project should be registered after saving config")
	}
}

func TestConvertToLogConfig(t *testing.T) {
	loader := NewEnhancedConfigLoader(LoadXDGFirst)

	// Test converting valid config data
	configData := map[string]interface{}{
		"logRotation": map[string]interface{}{
			"maxAge":     float64(30),
			"maxSize":    float64(10),
			"maxBackups": float64(5),
			"compress":   true,
		},
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
		"unknownField": "shouldBePreserved",
	}

	config, err := loader.convertToLogConfig(configData)
	if err != nil {
		t.Fatalf("Failed to convert to LogConfig: %v", err)
	}

	// Check log rotation
	if config.LogRotation.MaxAge != 30 {
		t.Errorf("Expected MaxAge=30, got %d", config.LogRotation.MaxAge)
	}
	if config.LogRotation.MaxSize != 10 {
		t.Errorf("Expected MaxSize=10, got %d", config.LogRotation.MaxSize)
	}
	if config.LogRotation.MaxBackups != 5 {
		t.Errorf("Expected MaxBackups=5, got %d", config.LogRotation.MaxBackups)
	}
	if !config.LogRotation.Compress {
		t.Error("Expected Compress=true")
	}

	// Check custom hooks
	if len(config.CustomHooks) == 0 {
		t.Error("CustomHooks should not be empty")
	}

	// Check unknown field preservation
	if config.Other["unknownField"] != "shouldBePreserved" {
		t.Errorf("Unknown field not preserved: %v", config.Other["unknownField"])
	}
}

func TestConvertFromLogConfig(t *testing.T) {
	loader := NewEnhancedConfigLoader(LoadXDGFirst)

	// Create test LogConfig
	config := &LogConfig{
		LogRotation: LogRotationConfig{
			MaxAge:     45,
			MaxSize:    15,
			MaxBackups: 8,
			Compress:   false,
		},
		CustomHooks: CustomHooksConfig{
			"test-group": HookGroup{
				"PreToolUse": &EventConfig{
					Jobs: []HookJob{
						{
							Name: "test-job",
							Run:  "echo 'converted'",
						},
					},
				},
			},
		},
		Other: map[string]interface{}{
			"preservedField": "preservedValue",
		},
	}

	configData, err := loader.convertFromLogConfig(config)
	if err != nil {
		t.Fatalf("Failed to convert from LogConfig: %v", err)
	}

	// Check log rotation
	logRotation, ok := configData["logRotation"].(map[string]interface{})
	if !ok {
		t.Fatalf("logRotation should be a map, got %T: %+v", configData["logRotation"], configData["logRotation"])
	}

	if logRotation["MaxAge"] != float64(45) { // JSON marshal/unmarshal converts to float64
		t.Errorf("Expected MaxAge=45, got %v (type %T)", logRotation["MaxAge"], logRotation["MaxAge"])
	}

	// Check preserved field
	if configData["preservedField"] != "preservedValue" {
		t.Errorf("Expected preservedField=preservedValue, got %v", configData["preservedField"])
	}

	// Check custom hooks
	if configData["customHooks"] == nil {
		t.Error("customHooks should be present")
	}
}

func TestGetXDGConfig(t *testing.T) {
	loader := NewEnhancedConfigLoader(LoadXDGFirst)
	xdg := loader.GetXDGConfig()

	if xdg == nil {
		t.Error("GetXDGConfig should return non-nil XDG config")
	}

	if xdg != loader.xdg {
		t.Error("GetXDGConfig should return the same instance as loader.xdg")
	}
}
