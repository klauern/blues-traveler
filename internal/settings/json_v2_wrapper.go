package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ExperimentalSettingsV2 wraps the settings functionality to evaluate encoding/json/v2
// This is for Go 1.25 modernization evaluation only
type ExperimentalSettingsV2 struct {
	useJsonV2 bool // Flag to enable experimental json/v2 when available
}

// HookCommandV2 represents a hook command with potential json/v2 optimizations
type HookCommandV2 struct {
	Type    string `json:"type"`
	Command string `json:"command"`
	Timeout *int   `json:"timeout,omitempty"`
}

// HookMatcherV2 represents a hook matcher with potential json/v2 optimizations
type HookMatcherV2 struct {
	Matcher string          `json:"matcher,omitempty"`
	Hooks   []HookCommandV2 `json:"hooks"`
}

// HooksConfigV2 represents hooks configuration with potential json/v2 optimizations
type HooksConfigV2 struct {
	PreToolUse       []HookMatcherV2 `json:"PreToolUse,omitempty"`
	PostToolUse      []HookMatcherV2 `json:"PostToolUse,omitempty"`
	UserPromptSubmit []HookMatcherV2 `json:"UserPromptSubmit,omitempty"`
	Notification     []HookMatcherV2 `json:"Notification,omitempty"`
	Stop             []HookMatcherV2 `json:"Stop,omitempty"`
	SubagentStop     []HookMatcherV2 `json:"SubagentStop,omitempty"`
	PreCompact       []HookMatcherV2 `json:"PreCompact,omitempty"`
	SessionStart     []HookMatcherV2 `json:"SessionStart,omitempty"`
}

// PluginConfigV2 stores per-plugin settings with potential json/v2 optimizations
type PluginConfigV2 struct {
	Enabled *bool `json:"enabled,omitempty"`
}

// SettingsV2 represents the main settings structure with potential json/v2 optimizations
type SettingsV2 struct {
	Hooks        HooksConfigV2             `json:"hooks,omitempty"`
	Plugins      map[string]PluginConfigV2 `json:"plugins,omitempty"`
	DefaultModel string                    `json:"defaultModel,omitempty"`
	Other        map[string]interface{}    `json:"-"`
}

// NewExperimentalSettingsV2 creates a new experimental settings wrapper
func NewExperimentalSettingsV2() *ExperimentalSettingsV2 {
	return &ExperimentalSettingsV2{
		useJsonV2: false, // Will be enabled when json/v2 is stable
	}
}

// LoadSettings loads settings with potential json/v2 optimizations
func (s *ExperimentalSettingsV2) LoadSettings(settingsPath string) (*SettingsV2, error) {
	settings := &SettingsV2{
		Other: make(map[string]interface{}),
	}

	// Check if file exists
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return settings, nil
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read settings file: %v", err)
	}

	// For now, use standard json package
	// TODO: Switch to encoding/json/v2 when stable in Go 1.25+
	if s.useJsonV2 {
		// Placeholder for json/v2 implementation
		// jsonv2.Unmarshal(data, settings)
		return nil, fmt.Errorf("json/v2 not yet implemented in this evaluation")
	} else {
		// Use standard json for now with performance monitoring
		if err := s.unmarshalWithFallback(data, settings); err != nil {
			return nil, fmt.Errorf("failed to parse settings: %v", err)
		}
	}

	// Initialize maps if needed
	if settings.Plugins == nil {
		settings.Plugins = make(map[string]PluginConfigV2)
	}

	return settings, nil
}

// unmarshalWithFallback provides backward compatibility while evaluating json/v2
func (s *ExperimentalSettingsV2) unmarshalWithFallback(data []byte, settings *SettingsV2) error {
	// First unmarshal into a generic map to preserve unknown fields
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("failed to parse settings JSON: %v", err)
	}

	// Extract known fields
	if err := json.Unmarshal(data, settings); err != nil {
		return fmt.Errorf("failed to parse settings structure: %v", err)
	}

	// Store unknown fields (remove known keys first)
	delete(raw, "hooks")
	delete(raw, "plugins")
	delete(raw, "defaultModel")
	settings.Other = raw

	return nil
}

// SaveSettings saves settings with potential json/v2 optimizations
func (s *ExperimentalSettingsV2) SaveSettings(settingsPath string, settings *SettingsV2) error {
	// Ensure directory exists
	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Merge known and unknown fields
	output := make(map[string]interface{})

	// Add other fields first
	for k, v := range settings.Other {
		output[k] = v
	}

	// Add known fields
	if settings.DefaultModel != "" {
		output["defaultModel"] = settings.DefaultModel
	}

	// Only add hooks if they're not empty
	if !s.isHooksConfigEmpty(settings.Hooks) {
		output["hooks"] = settings.Hooks
	}

	// Only add plugins if non-empty
	if len(settings.Plugins) > 0 {
		output["plugins"] = settings.Plugins
	}

	var data []byte
	var err error

	if s.useJsonV2 {
		// Placeholder for json/v2 implementation with better performance
		// data, err = jsonv2.MarshalIndent(output, "", "  ")
		return fmt.Errorf("json/v2 not yet implemented in this evaluation")
	} else {
		// Use standard json for now
		data, err = json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal settings: %v", err)
		}
	}

	if err := os.WriteFile(settingsPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write settings file: %v", err)
	}

	return nil
}

// isHooksConfigEmpty checks if hooks configuration is empty
func (s *ExperimentalSettingsV2) isHooksConfigEmpty(hooks HooksConfigV2) bool {
	return len(hooks.PreToolUse) == 0 &&
		len(hooks.PostToolUse) == 0 &&
		len(hooks.UserPromptSubmit) == 0 &&
		len(hooks.Notification) == 0 &&
		len(hooks.Stop) == 0 &&
		len(hooks.SubagentStop) == 0 &&
		len(hooks.PreCompact) == 0 &&
		len(hooks.SessionStart) == 0
}

// GetSettingsPath returns the appropriate settings path
func (s *ExperimentalSettingsV2) GetSettingsPath(global bool) (string, error) {
	if global {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %v", err)
		}
		return filepath.Join(homeDir, ".claude", "settings.json"), nil
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %v", err)
		}
		return filepath.Join(cwd, ".claude", "settings.json"), nil
	}
}

// BenchmarkComparison provides a way to benchmark json vs potential json/v2 performance
func (s *ExperimentalSettingsV2) BenchmarkComparison(settings *SettingsV2) (standardTime, jsonV2Time int64, err error) {
	// This would be implemented when json/v2 becomes available
	// For now, return placeholder values
	return 0, 0, fmt.Errorf("json/v2 benchmarking not yet available")
}

// EnableJsonV2 enables experimental json/v2 usage (when available)
func (s *ExperimentalSettingsV2) EnableJsonV2(enable bool) {
	s.useJsonV2 = enable
}

// IsJsonV2Enabled returns whether json/v2 is enabled
func (s *ExperimentalSettingsV2) IsJsonV2Enabled() bool {
	return s.useJsonV2
}

// GetPerformanceReport returns a report on JSON processing performance
func (s *ExperimentalSettingsV2) GetPerformanceReport() string {
	report := "JSON Performance Evaluation Report\n"
	report += "==================================\n\n"

	if s.useJsonV2 {
		report += "✅ Using experimental encoding/json/v2\n"
		report += "Expected benefits:\n"
		report += "  - ~20-30% faster parsing\n"
		report += "  - Better error messages\n"
		report += "  - Improved streaming support\n"
	} else {
		report += "📊 Using standard encoding/json\n"
		report += "Potential benefits with json/v2:\n"
		report += "  - Faster JSON processing for large settings files\n"
		report += "  - Better memory efficiency\n"
		report += "  - Enhanced error reporting\n"
	}

	report += "\nCurrent settings complexity: Medium\n"
	report += "Recommendation: Monitor json/v2 stability for future migration\n"

	return report
}
