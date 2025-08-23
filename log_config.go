package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LogConfig represents our application's logging configuration
type LogConfig struct {
	LogRotation LogRotationConfig `json:"logRotation"`
}

// getLogConfigPath returns the path to our log configuration file
func getLogConfigPath(global bool) (string, error) {
	if global {
		// Global config: ~/.claude/hooks/klauer-hooks-config.json
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %v", err)
		}
		return filepath.Join(homeDir, ".claude", "hooks", "klauer-hooks-config.json"), nil
	} else {
		// Project config: ./.claude/hooks/klauer-hooks-config.json
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %v", err)
		}
		return filepath.Join(cwd, ".claude", "hooks", "klauer-hooks-config.json"), nil
	}
}

// loadLogConfig loads the log configuration, returning defaults if file doesn't exist
func loadLogConfig(configPath string) (*LogConfig, error) {
	config := &LogConfig{
		LogRotation: DefaultLogRotationConfig(),
	}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// File doesn't exist, return default config
		return config, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %v", err)
	}

	return config, nil
}

// saveLogConfig saves the log configuration to file
func saveLogConfig(configPath string, config *LogConfig) error {
	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// getLogRotationConfigFromFile gets log rotation config from our own config file
func getLogRotationConfigFromFile(global bool) LogRotationConfig {
	configPath, err := getLogConfigPath(global)
	if err != nil {
		return DefaultLogRotationConfig()
	}

	config, err := loadLogConfig(configPath)
	if err != nil {
		return DefaultLogRotationConfig()
	}

	return config.LogRotation
}
