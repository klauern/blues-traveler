package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestSettingsPrecedence tests that project settings properly override global settings
// in all scenarios specified in blues-traveler-3 acceptance criteria
func TestSettingsPrecedence(t *testing.T) {
	// Create temporary directories for testing
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "project")
	homeDir := filepath.Join(tempDir, "home")

	if err := os.MkdirAll(filepath.Join(projectDir, ".claude"), 0755); err != nil {
		t.Fatalf("failed to create project .claude dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(homeDir, ".claude"), 0755); err != nil {
		t.Fatalf("failed to create home .claude dir: %v", err)
	}

	// Save original working directory and HOME
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	originalHome := os.Getenv("HOME")

	// Change to project directory and set HOME
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("failed to change to project dir: %v", err)
	}
	os.Setenv("HOME", homeDir)

	// Restore original state after test
	t.Cleanup(func() {
		os.Chdir(originalWd)
		os.Setenv("HOME", originalHome)
	})

	tests := []struct {
		name            string
		projectEnabled  *bool // nil means not set
		globalEnabled   *bool // nil means not set
		expectedEnabled bool
		description     string
	}{
		{
			name:            "project enabled + global disabled = enabled",
			projectEnabled:  boolPtr(true),
			globalEnabled:   boolPtr(false),
			expectedEnabled: true,
			description:     "Project explicitly enabled should override global disabled",
		},
		{
			name:            "project disabled + global enabled = disabled",
			projectEnabled:  boolPtr(false),
			globalEnabled:   boolPtr(true),
			expectedEnabled: false,
			description:     "Project explicitly disabled should override global enabled",
		},
		{
			name:            "project nil + global enabled = enabled",
			projectEnabled:  nil,
			globalEnabled:   boolPtr(true),
			expectedEnabled: true,
			description:     "Project nil should fall back to global enabled",
		},
		{
			name:            "project nil + global disabled = disabled",
			projectEnabled:  nil,
			globalEnabled:   boolPtr(false),
			expectedEnabled: false,
			description:     "Project nil should fall back to global disabled",
		},
		{
			name:            "project enabled + global nil = enabled",
			projectEnabled:  boolPtr(true),
			globalEnabled:   nil,
			expectedEnabled: true,
			description:     "Project enabled should override global nil (default)",
		},
		{
			name:            "project disabled + global nil = disabled",
			projectEnabled:  boolPtr(false),
			globalEnabled:   nil,
			expectedEnabled: false,
			description:     "Project disabled should override global nil (default)",
		},
		{
			name:            "project nil + global nil = enabled (default)",
			projectEnabled:  nil,
			globalEnabled:   nil,
			expectedEnabled: true,
			description:     "Both nil should default to enabled",
		},
		{
			name:            "project enabled + global enabled = enabled",
			projectEnabled:  boolPtr(true),
			globalEnabled:   boolPtr(true),
			expectedEnabled: true,
			description:     "Both enabled should result in enabled",
		},
		{
			name:            "project disabled + global disabled = disabled",
			projectEnabled:  boolPtr(false),
			globalEnabled:   boolPtr(false),
			expectedEnabled: false,
			description:     "Both disabled should result in disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create project settings
			projectSettings := &Settings{
				Plugins: make(map[string]PluginConfig),
			}
			if tt.projectEnabled != nil {
				projectSettings.Plugins["test-plugin"] = PluginConfig{
					Enabled: tt.projectEnabled,
				}
			}

			// Create global settings
			globalSettings := &Settings{
				Plugins: make(map[string]PluginConfig),
			}
			if tt.globalEnabled != nil {
				globalSettings.Plugins["test-plugin"] = PluginConfig{
					Enabled: tt.globalEnabled,
				}
			}

			// Write settings to files
			projectPath := filepath.Join(projectDir, ".claude", "settings.json")
			globalPath := filepath.Join(homeDir, ".claude", "settings.json")

			// Only write project settings if there's something to write
			if tt.projectEnabled != nil {
				if err := writeSettings(projectPath, projectSettings); err != nil {
					t.Fatalf("failed to write project settings: %v", err)
				}
			} else {
				// Create empty settings file if project is nil
				emptySettings := &Settings{Plugins: make(map[string]PluginConfig)}
				if err := writeSettings(projectPath, emptySettings); err != nil {
					t.Fatalf("failed to write empty project settings: %v", err)
				}
			}

			// Only write global settings if there's something to write
			if tt.globalEnabled != nil {
				if err := writeSettings(globalPath, globalSettings); err != nil {
					t.Fatalf("failed to write global settings: %v", err)
				}
			} else {
				// Create empty settings file if global is nil
				emptySettings := &Settings{Plugins: make(map[string]PluginConfig)}
				if err := writeSettings(globalPath, emptySettings); err != nil {
					t.Fatalf("failed to write empty global settings: %v", err)
				}
			}

			// Test the precedence
			got := IsPluginEnabled("test-plugin")
			if got != tt.expectedEnabled {
				t.Errorf("%s: IsPluginEnabled() = %v, want %v", tt.description, got, tt.expectedEnabled)
			}
		})
	}
}

// TestSettingsPrecedenceNoFiles tests the default behavior when no settings files exist
func TestSettingsPrecedenceNoFiles(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "project")
	homeDir := filepath.Join(tempDir, "home")

	// Create directories but no .claude subdirectories
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatalf("failed to create home dir: %v", err)
	}

	// Save original working directory and HOME
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	originalHome := os.Getenv("HOME")

	// Change to project directory and set HOME
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("failed to change to project dir: %v", err)
	}
	os.Setenv("HOME", homeDir)

	// Restore original state after test
	t.Cleanup(func() {
		os.Chdir(originalWd)
		os.Setenv("HOME", originalHome)
	})

	// Test that default is enabled when no files exist
	got := IsPluginEnabled("any-plugin")
	if !got {
		t.Errorf("IsPluginEnabled() with no settings files = false, want true (default enabled)")
	}
}

// TestSettingsPrecedenceProjectOnly tests behavior when only project settings exist
func TestSettingsPrecedenceProjectOnly(t *testing.T) {
	// Create temporary directories
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "project")
	homeDir := filepath.Join(tempDir, "home")

	if err := os.MkdirAll(filepath.Join(projectDir, ".claude"), 0755); err != nil {
		t.Fatalf("failed to create project .claude dir: %v", err)
	}
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatalf("failed to create home dir: %v", err)
	}

	// Save original state
	originalWd, _ := os.Getwd()
	originalHome := os.Getenv("HOME")
	defer func() {
		os.Chdir(originalWd)
		os.Setenv("HOME", originalHome)
	}()

	// Change to project directory
	os.Chdir(projectDir)
	os.Setenv("HOME", homeDir)

	tests := []struct {
		name            string
		projectEnabled  *bool
		expectedEnabled bool
	}{
		{
			name:            "project enabled only",
			projectEnabled:  boolPtr(true),
			expectedEnabled: true,
		},
		{
			name:            "project disabled only",
			projectEnabled:  boolPtr(false),
			expectedEnabled: false,
		},
		{
			name:            "project nil only",
			projectEnabled:  nil,
			expectedEnabled: true, // Default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create project settings
			projectSettings := &Settings{
				Plugins: make(map[string]PluginConfig),
			}
			if tt.projectEnabled != nil {
				projectSettings.Plugins["test-plugin"] = PluginConfig{
					Enabled: tt.projectEnabled,
				}
			}

			// Write project settings
			projectPath := filepath.Join(projectDir, ".claude", "settings.json")
			if err := writeSettings(projectPath, projectSettings); err != nil {
				t.Fatalf("failed to write project settings: %v", err)
			}

			// Test
			got := IsPluginEnabled("test-plugin")
			if got != tt.expectedEnabled {
				t.Errorf("IsPluginEnabled() = %v, want %v", got, tt.expectedEnabled)
			}
		})
	}
}

// TestSettingsPrecedenceGlobalOnly tests behavior when only global settings exist
func TestSettingsPrecedenceGlobalOnly(t *testing.T) {
	// Create temporary directories
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "project")
	homeDir := filepath.Join(tempDir, "home")

	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(homeDir, ".claude"), 0755); err != nil {
		t.Fatalf("failed to create home .claude dir: %v", err)
	}

	// Save original state
	originalWd, _ := os.Getwd()
	originalHome := os.Getenv("HOME")
	defer func() {
		os.Chdir(originalWd)
		os.Setenv("HOME", originalHome)
	}()

	// Change to project directory
	os.Chdir(projectDir)
	os.Setenv("HOME", homeDir)

	tests := []struct {
		name            string
		globalEnabled   *bool
		expectedEnabled bool
	}{
		{
			name:            "global enabled only",
			globalEnabled:   boolPtr(true),
			expectedEnabled: true,
		},
		{
			name:            "global disabled only",
			globalEnabled:   boolPtr(false),
			expectedEnabled: false,
		},
		{
			name:            "global nil only",
			globalEnabled:   nil,
			expectedEnabled: true, // Default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create global settings
			globalSettings := &Settings{
				Plugins: make(map[string]PluginConfig),
			}
			if tt.globalEnabled != nil {
				globalSettings.Plugins["test-plugin"] = PluginConfig{
					Enabled: tt.globalEnabled,
				}
			}

			// Write global settings
			globalPath := filepath.Join(homeDir, ".claude", "settings.json")
			if err := writeSettings(globalPath, globalSettings); err != nil {
				t.Fatalf("failed to write global settings: %v", err)
			}

			// Test
			got := IsPluginEnabled("test-plugin")
			if got != tt.expectedEnabled {
				t.Errorf("IsPluginEnabled() = %v, want %v", got, tt.expectedEnabled)
			}
		})
	}
}

// TestSettingsPrecedenceMultiplePlugins tests precedence with multiple plugins
func TestSettingsPrecedenceMultiplePlugins(t *testing.T) {
	// Create temporary directories
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "project")
	homeDir := filepath.Join(tempDir, "home")

	if err := os.MkdirAll(filepath.Join(projectDir, ".claude"), 0755); err != nil {
		t.Fatalf("failed to create project .claude dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(homeDir, ".claude"), 0755); err != nil {
		t.Fatalf("failed to create home .claude dir: %v", err)
	}

	// Save original state
	originalWd, _ := os.Getwd()
	originalHome := os.Getenv("HOME")
	defer func() {
		os.Chdir(originalWd)
		os.Setenv("HOME", originalHome)
	}()

	// Change to project directory
	os.Chdir(projectDir)
	os.Setenv("HOME", homeDir)

	// Create project settings with multiple plugins
	projectSettings := &Settings{
		Plugins: map[string]PluginConfig{
			"plugin1": {Enabled: boolPtr(true)},  // Override global
			"plugin2": {Enabled: boolPtr(false)}, // Override global
			// plugin3 not in project, should use global
		},
	}

	// Create global settings with multiple plugins
	globalSettings := &Settings{
		Plugins: map[string]PluginConfig{
			"plugin1": {Enabled: boolPtr(false)}, // Should be overridden
			"plugin2": {Enabled: boolPtr(true)},  // Should be overridden
			"plugin3": {Enabled: boolPtr(true)},  // Should be used
			"plugin4": {Enabled: boolPtr(false)}, // Should be used
		},
	}

	// Write settings
	projectPath := filepath.Join(projectDir, ".claude", "settings.json")
	globalPath := filepath.Join(homeDir, ".claude", "settings.json")

	if err := writeSettings(projectPath, projectSettings); err != nil {
		t.Fatalf("failed to write project settings: %v", err)
	}
	if err := writeSettings(globalPath, globalSettings); err != nil {
		t.Fatalf("failed to write global settings: %v", err)
	}

	// Test each plugin
	tests := []struct {
		plugin   string
		expected bool
	}{
		{"plugin1", true},  // Project overrides global
		{"plugin2", false}, // Project overrides global
		{"plugin3", true},  // Global used
		{"plugin4", false}, // Global used
		{"plugin5", true},  // Not in either, default enabled
	}

	for _, tt := range tests {
		t.Run(tt.plugin, func(t *testing.T) {
			got := IsPluginEnabled(tt.plugin)
			if got != tt.expected {
				t.Errorf("IsPluginEnabled(%q) = %v, want %v", tt.plugin, got, tt.expected)
			}
		})
	}
}

// Helper functions

func boolPtr(b bool) *bool {
	return &b
}

func writeSettings(path string, settings *Settings) error {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
