package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestJsonV2Evaluation evaluates the experimental JSON v2 wrapper
func TestJsonV2Evaluation(t *testing.T) {
	t.Attr("category", "performance")
	t.Attr("component", "json-evaluation")

	// Create a temporary settings file for testing
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.json")

	// Create test settings data
	testSettings := &SettingsV2{
		DefaultModel: "claude-3-5-sonnet",
		Plugins: map[string]PluginConfigV2{
			"security": {Enabled: boolPtr(true)},
			"format":   {Enabled: boolPtr(true)},
			"debug":    {Enabled: boolPtr(false)},
		},
		Hooks: HooksConfigV2{
			PreToolUse: []HookMatcherV2{
				{
					Matcher: "*",
					Hooks: []HookCommandV2{
						{Type: "command", Command: "./hooks run security"},
					},
				},
			},
		},
	}

	wrapper := NewExperimentalSettingsV2()

	// Test saving settings
	err := wrapper.SaveSettings(settingsPath, testSettings)
	if err != nil {
		t.Fatalf("Failed to save settings: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		t.Fatal("Settings file was not created")
	}

	// Test loading settings
	loadedSettings, err := wrapper.LoadSettings(settingsPath)
	if err != nil {
		t.Fatalf("Failed to load settings: %v", err)
	}

	// Verify settings were loaded correctly
	if loadedSettings.DefaultModel != testSettings.DefaultModel {
		t.Errorf("DefaultModel mismatch: got %s, want %s",
			loadedSettings.DefaultModel, testSettings.DefaultModel)
	}

	if len(loadedSettings.Plugins) != len(testSettings.Plugins) {
		t.Errorf("Plugins count mismatch: got %d, want %d",
			len(loadedSettings.Plugins), len(testSettings.Plugins))
	}

	// Test performance report
	report := wrapper.GetPerformanceReport()
	if len(report) == 0 {
		t.Error("Performance report should not be empty")
	}

	t.Logf("Performance Report:\n%s", report)
}

// BenchmarkStandardJson benchmarks standard JSON performance
func BenchmarkStandardJson(b *testing.B) {
	testData := createLargeSettingsData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, err := json.Marshal(testData)
		if err != nil {
			b.Fatal(err)
		}

		var result SettingsV2
		err = json.Unmarshal(data, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSettingsOperations benchmarks complete settings operations
func BenchmarkSettingsOperations(b *testing.B) {
	tmpDir := b.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.json")
	wrapper := NewExperimentalSettingsV2()
	testSettings := createLargeSettingsData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Save
		err := wrapper.SaveSettings(settingsPath, testSettings)
		if err != nil {
			b.Fatal(err)
		}

		// Load
		_, err = wrapper.LoadSettings(settingsPath)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// TestJsonV2PerformanceComparison tests performance comparison when available
func TestJsonV2PerformanceComparison(t *testing.T) {
	t.Attr("category", "performance")
	t.Attr("component", "json-comparison")

	wrapper := NewExperimentalSettingsV2()
	testSettings := createLargeSettingsData()

	// Test with standard JSON
	start := time.Now()
	data, err := json.Marshal(testSettings)
	if err != nil {
		t.Fatal(err)
	}

	var result SettingsV2
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatal(err)
	}
	standardDuration := time.Since(start)

	t.Logf("Standard JSON duration: %v", standardDuration)
	t.Logf("JSON data size: %d bytes", len(data))

	// Test benchmark comparison (placeholder for when json/v2 is available)
	standardTime, jsonV2Time, err := wrapper.BenchmarkComparison(testSettings)
	if err == nil {
		improvement := float64(standardTime-jsonV2Time) / float64(standardTime) * 100
		t.Logf("JSON v2 performance improvement: %.2f%%", improvement)
	} else {
		t.Logf("JSON v2 benchmark not available: %v", err)
	}
}

// TestJsonV2Migration tests migration compatibility
func TestJsonV2Migration(t *testing.T) {
	t.Attr("category", "migration")
	t.Attr("component", "json-v2")

	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.json")

	// Create settings with standard JSON
	testSettings := &SettingsV2{
		DefaultModel: "claude-3-5-sonnet",
		Other: map[string]interface{}{
			"customField": "customValue",
			"experimentalFeatures": map[string]interface{}{
				"jsonV2": true,
			},
		},
	}

	wrapper := NewExperimentalSettingsV2()

	// Save with standard JSON
	err := wrapper.SaveSettings(settingsPath, testSettings)
	if err != nil {
		t.Fatalf("Failed to save settings: %v", err)
	}

	// Enable JSON v2 (when available)
	wrapper.EnableJsonV2(true)

	// Try to load with JSON v2 enabled (should fallback gracefully)
	_, err = wrapper.LoadSettings(settingsPath)
	if err != nil && err.Error() != "json/v2 not yet implemented in this evaluation" {
		t.Fatalf("Unexpected error: %v", err)
	}

	// If we get the expected "not implemented" error, that's correct for evaluation
	if err != nil && err.Error() == "json/v2 not yet implemented in this evaluation" {
		t.Log("JSON v2 correctly identified as not yet implemented")

		// Test fallback to standard JSON
		wrapper.EnableJsonV2(false)
		fallbackSettings, err := wrapper.LoadSettings(settingsPath)
		if err != nil {
			t.Fatalf("Fallback to standard JSON failed: %v", err)
		}

		// Verify custom fields are preserved
		if fallbackSettings.Other["customField"] != "customValue" {
			t.Error("Custom fields not preserved during migration")
		}
	}
}

// createLargeSettingsData creates a large settings structure for performance testing
func createLargeSettingsData() *SettingsV2 {
	settings := &SettingsV2{
		DefaultModel: "claude-3-5-sonnet",
		Plugins:      make(map[string]PluginConfigV2),
		Other:        make(map[string]interface{}),
	}

	// Add many plugins
	for i := 0; i < 100; i++ {
		pluginName := fmt.Sprintf("plugin-%03d", i)
		settings.Plugins[pluginName] = PluginConfigV2{
			Enabled: boolPtr(i%2 == 0),
		}
	}

	// Add many hook matchers
	for i := 0; i < 50; i++ {
		matcher := HookMatcherV2{
			Matcher: fmt.Sprintf("pattern-%d", i),
			Hooks: []HookCommandV2{
				{
					Type:    "command",
					Command: fmt.Sprintf("./hooks run test-%d", i),
					Timeout: intPtr(30),
				},
			},
		}
		settings.Hooks.PreToolUse = append(settings.Hooks.PreToolUse, matcher)
	}

	// Add other fields
	for i := 0; i < 20; i++ {
		settings.Other[fmt.Sprintf("customField%d", i)] = fmt.Sprintf("value%d", i)
	}

	return settings
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int) *int {
	return &i
}
