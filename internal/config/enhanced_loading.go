package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ConfigLoaderStrategy defines different strategies for loading configuration
type ConfigLoaderStrategy int

const (
	// LoadXDGFirst tries XDG paths first, then falls back to legacy
	LoadXDGFirst ConfigLoaderStrategy = iota
	// LoadLegacyOnly only loads from legacy paths
	LoadLegacyOnly
	// LoadXDGOnly only loads from XDG paths
	LoadXDGOnly
)

// EnhancedConfigLoader provides configuration loading with XDG support and fallback
type EnhancedConfigLoader struct {
	xdg      *XDGConfig
	strategy ConfigLoaderStrategy
}

// NewEnhancedConfigLoader creates a new enhanced configuration loader
func NewEnhancedConfigLoader(strategy ConfigLoaderStrategy) *EnhancedConfigLoader {
	return &EnhancedConfigLoader{
		xdg:      NewXDGConfig(),
		strategy: strategy,
	}
}

// LoadConfigWithFallback loads configuration using the specified strategy
func (e *EnhancedConfigLoader) LoadConfigWithFallback(projectPath string) (*LogConfig, string, error) {
	absProjectPath, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	switch e.strategy {
	case LoadXDGFirst:
		return e.loadXDGFirstWithFallback(absProjectPath)
	case LoadXDGOnly:
		return e.loadXDGOnly(absProjectPath)
	case LoadLegacyOnly:
		return e.loadLegacyOnly(absProjectPath)
	default:
		return e.loadXDGFirstWithFallback(absProjectPath)
	}
}

// LoadGlobalConfigWithFallback loads global configuration using the specified strategy
func (e *EnhancedConfigLoader) LoadGlobalConfigWithFallback() (*LogConfig, string, error) {
	switch e.strategy {
	case LoadXDGFirst:
		return e.loadGlobalXDGFirstWithFallback()
	case LoadXDGOnly:
		return e.loadGlobalXDGOnly()
	case LoadLegacyOnly:
		return e.loadGlobalLegacyOnly()
	default:
		return e.loadGlobalXDGFirstWithFallback()
	}
}

// loadXDGFirstWithFallback tries XDG config first, then falls back to legacy
func (e *EnhancedConfigLoader) loadXDGFirstWithFallback(projectPath string) (*LogConfig, string, error) {
	// Try XDG config first
	if config, configPath, err := e.tryLoadXDGConfig(projectPath); err == nil {
		return config, configPath, nil
	}

	// Fall back to legacy config
	return e.loadLegacyOnly(projectPath)
}

// loadGlobalXDGFirstWithFallback tries global XDG config first, then falls back to legacy
func (e *EnhancedConfigLoader) loadGlobalXDGFirstWithFallback() (*LogConfig, string, error) {
	// Try XDG global config first
	if config, configPath, err := e.tryLoadGlobalXDGConfig(); err == nil {
		return config, configPath, nil
	}

	// Fall back to legacy global config
	return e.loadGlobalLegacyOnly()
}

// tryLoadXDGConfig attempts to load project configuration from XDG location
func (e *EnhancedConfigLoader) tryLoadXDGConfig(projectPath string) (*LogConfig, string, error) {
	// Check if project is registered in XDG registry
	projectConfig, err := e.xdg.GetProjectConfig(projectPath)
	if err != nil {
		return nil, "", fmt.Errorf("project not in XDG registry: %w", err)
	}

	configPath := filepath.Join(e.xdg.GetConfigDir(), projectConfig.ConfigFile)

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, "", fmt.Errorf("XDG config file does not exist: %s", configPath)
	}

	// Load XDG config data
	configData, err := e.xdg.LoadProjectConfig(projectPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load XDG project config: %w", err)
	}

	// Convert to LogConfig
	config, err := e.convertToLogConfig(configData)
	if err != nil {
		return nil, "", fmt.Errorf("failed to convert XDG config: %w", err)
	}

	return config, configPath, nil
}

// tryLoadGlobalXDGConfig attempts to load global configuration from XDG location
func (e *EnhancedConfigLoader) tryLoadGlobalXDGConfig() (*LogConfig, string, error) {
	configPath := e.xdg.GetGlobalConfigPath("json")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, "", fmt.Errorf("XDG global config file does not exist: %s", configPath)
	}

	// Load XDG global config data
	configData, err := e.xdg.LoadGlobalConfig("json")
	if err != nil {
		return nil, "", fmt.Errorf("failed to load XDG global config: %w", err)
	}

	// Convert to LogConfig
	config, err := e.convertToLogConfig(configData)
	if err != nil {
		return nil, "", fmt.Errorf("failed to convert XDG global config: %w", err)
	}

	return config, configPath, nil
}

// loadXDGOnly loads configuration only from XDG locations
func (e *EnhancedConfigLoader) loadXDGOnly(projectPath string) (*LogConfig, string, error) {
	return e.tryLoadXDGConfig(projectPath)
}

// loadGlobalXDGOnly loads global configuration only from XDG locations
func (e *EnhancedConfigLoader) loadGlobalXDGOnly() (*LogConfig, string, error) {
	return e.tryLoadGlobalXDGConfig()
}

// loadLegacyOnly loads configuration only from legacy locations
func (e *EnhancedConfigLoader) loadLegacyOnly(projectPath string) (*LogConfig, string, error) {
	legacyPath := GetLegacyConfigPath(projectPath)
	config, err := LoadLogConfig(legacyPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load legacy config: %w", err)
	}

	return config, legacyPath, nil
}

// loadGlobalLegacyOnly loads global configuration only from legacy locations
func (e *EnhancedConfigLoader) loadGlobalLegacyOnly() (*LogConfig, string, error) {
	legacyPath, err := GetLogConfigPath(true)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get legacy global config path: %w", err)
	}

	config, err := LoadLogConfig(legacyPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load legacy global config: %w", err)
	}

	return config, legacyPath, nil
}

// convertToLogConfig converts a generic config map to LogConfig structure
func (e *EnhancedConfigLoader) convertToLogConfig(configData map[string]interface{}) (*LogConfig, error) {
	// Marshal and unmarshal to convert the map to LogConfig struct
	data, err := json.Marshal(configData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config data: %w", err)
	}

	config := &LogConfig{
		LogRotation: DefaultLogRotationConfig(),
		Other:       make(map[string]interface{}),
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config data: %w", err)
	}

	// Preserve unknown fields
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal raw config data: %w", err)
	}

	// Remove known fields from raw data
	delete(raw, "logRotation")
	delete(raw, "customHooks")
	delete(raw, "blockedUrls")
	config.Other = raw

	return config, nil
}

// SaveConfigWithXDG saves configuration to XDG location and registers the project
func (e *EnhancedConfigLoader) SaveConfigWithXDG(projectPath string, config *LogConfig, format string) error {
	absProjectPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Convert LogConfig to generic map
	configData, err := e.convertFromLogConfig(config)
	if err != nil {
		return fmt.Errorf("failed to convert LogConfig: %w", err)
	}

	// Save to XDG location
	if err := e.xdg.SaveProjectConfig(absProjectPath, configData, format); err != nil {
		return fmt.Errorf("failed to save XDG project config: %w", err)
	}

	return nil
}

// SaveGlobalConfigWithXDG saves global configuration to XDG location
func (e *EnhancedConfigLoader) SaveGlobalConfigWithXDG(config *LogConfig, format string) error {
	// Convert LogConfig to generic map
	configData, err := e.convertFromLogConfig(config)
	if err != nil {
		return fmt.Errorf("failed to convert LogConfig: %w", err)
	}

	// Save to XDG location
	if err := e.xdg.SaveGlobalConfig(configData, format); err != nil {
		return fmt.Errorf("failed to save XDG global config: %w", err)
	}

	return nil
}

// convertFromLogConfig converts LogConfig to a generic config map
func (e *EnhancedConfigLoader) convertFromLogConfig(config *LogConfig) (map[string]interface{}, error) {
	// Marshal and unmarshal to convert LogConfig to map
	data, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal LogConfig: %w", err)
	}

	var configData map[string]interface{}
	if err := json.Unmarshal(data, &configData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	// Add back the "Other" fields
	for k, v := range config.Other {
		configData[k] = v
	}

	return configData, nil
}

// GetConfigPath returns the configuration path based on the strategy
func (e *EnhancedConfigLoader) GetConfigPath(projectPath string, global bool) (string, error) {
	if global {
		if e.strategy == LoadLegacyOnly {
			return GetLogConfigPath(true)
		}
		return e.xdg.GetGlobalConfigPath("json"), nil
	}

	absProjectPath, err := filepath.Abs(projectPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	if e.strategy == LoadLegacyOnly {
		return GetLegacyConfigPath(absProjectPath), nil
	}

	// For XDG paths, we need to check if the project is registered
	if projectConfig, err := e.xdg.GetProjectConfig(absProjectPath); err == nil {
		return filepath.Join(e.xdg.GetConfigDir(), projectConfig.ConfigFile), nil
	}

	// If not registered, return the potential XDG path
	return e.xdg.GetProjectConfigPath(absProjectPath, "json"), nil
}

// IsProjectRegistered checks if a project is registered in the XDG system
func (e *EnhancedConfigLoader) IsProjectRegistered(projectPath string) bool {
	absProjectPath, err := filepath.Abs(projectPath)
	if err != nil {
		return false
	}

	_, err = e.xdg.GetProjectConfig(absProjectPath)
	return err == nil
}

// GetXDGConfig returns the underlying XDG configuration manager
func (e *EnhancedConfigLoader) GetXDGConfig() *XDGConfig {
	return e.xdg
}
