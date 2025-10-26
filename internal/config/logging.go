package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/klauern/blues-traveler/internal/constants"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LogRotationConfig holds configuration for log rotation
type LogRotationConfig struct {
	MaxAge     int  // Maximum number of days to retain log files
	MaxSize    int  // Maximum size in megabytes before rotation
	MaxBackups int  // Maximum number of backup files to retain
	Compress   bool // Whether to compress rotated files
}

// DefaultLogRotationConfig returns sensible defaults for log rotation
func DefaultLogRotationConfig() LogRotationConfig {
	return LogRotationConfig{
		MaxAge:     30,   // 30 days default retention
		MaxSize:    10,   // 10MB per file
		MaxBackups: 5,    // Keep 5 backup files
		Compress:   true, // Compress old files
	}
}

// LogConfig represents our application's logging configuration
type LogConfig struct {
	LogRotation LogRotationConfig      `json:"logRotation"`
	CustomHooks CustomHooksConfig      `json:"customHooks,omitempty"`
	BlockedURLs []BlockedURL           `json:"blockedUrls,omitempty"`
	Other       map[string]interface{} `json:"-"`
}

// BlockedURL represents a blocked URL prefix + optional suggestion
type BlockedURL struct {
	Prefix     string `json:"prefix"`
	Suggestion string `json:"suggestion,omitempty"`
}

// GetLogConfigPath returns the path to our log configuration file
func GetLogConfigPath(global bool) (string, error) {
	if global {
		// Global config: ~/.claude/hooks/blues-traveler-config.json
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %v", err)
		}
		return constants.GetConfigPath(homeDir), nil
	}
	// Project config: ./.claude/hooks/blues-traveler-config.json
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %v", err)
	}
	return constants.GetConfigPath(cwd), nil
}

// LoadLogConfig loads the log configuration, returning defaults if file doesn't exist
func LoadLogConfig(configPath string) (*LogConfig, error) {
	config := &LogConfig{LogRotation: DefaultLogRotationConfig(), Other: map[string]interface{}{}}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// File doesn't exist, return default config
		return config, nil
	}

	data, err := os.ReadFile(configPath) // #nosec G304 - controlled config paths
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	// Preserve unknown fields
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %v", err)
	}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %v", err)
	}
	// Remove known
	delete(raw, "logRotation")
	delete(raw, "customHooks")
	delete(raw, "blockedUrls")
	config.Other = raw

	return config, nil
}

// SaveLogConfig saves the log configuration to file
func SaveLogConfig(configPath string, config *LogConfig) error {
	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Merge known and unknown
	out := map[string]interface{}{}
	for k, v := range config.Other {
		out[k] = v
	}
	out["logRotation"] = config.LogRotation
	if len(config.CustomHooks) > 0 {
		out["customHooks"] = config.CustomHooks
	}
	if len(config.BlockedURLs) > 0 {
		out["blockedUrls"] = config.BlockedURLs
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// GetLogRotationConfigFromFile gets log rotation config from our own config file
func GetLogRotationConfigFromFile(global bool) LogRotationConfig {
	configPath, err := GetLogConfigPath(global)
	if err != nil {
		return DefaultLogRotationConfig()
	}

	config, err := LoadLogConfig(configPath)
	if err != nil {
		return DefaultLogRotationConfig()
	}

	return config.LogRotation
}

// SetupLogRotation configures log rotation for a given log file path
func SetupLogRotation(logPath string, config LogRotationConfig) *lumberjack.Logger {
	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(logPath), 0o750); err != nil {
		log.Printf("Failed to create log directory: %v", err)
		return nil
	}

	logger := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
		LocalTime:  true, // Use local time for timestamps
	}

	return logger
}

// CleanupOldLogs manually removes log files older than the specified number of days
// This provides additional cleanup beyond lumberjack's built-in MaxAge
func CleanupOldLogs(logDir string, maxAgeDays int) error {
	if maxAgeDays <= 0 {
		return nil // No cleanup if maxAge is 0 or negative
	}

	cutoff := time.Now().AddDate(0, 0, -maxAgeDays)

	err := filepath.Walk(logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only consider .log files and compressed log files
		if filepath.Ext(path) == ".log" || filepath.Ext(path) == ".gz" {
			if info.ModTime().Before(cutoff) {
				if err := os.Remove(path); err != nil {
					log.Printf("Failed to remove old log file %s: %v", path, err)
				} else {
					log.Printf("Removed old log file: %s", path)
				}
			}
		}

		return nil
	})

	return err
}

// GetLogPath returns the standard log path for a given plugin key
func GetLogPath(pluginKey string) string {
	return filepath.Join(".claude", "hooks", fmt.Sprintf("%s.log", pluginKey))
}

// Logging format constants
const (
	LoggingFormatJSONL  = "jsonl"
	LoggingFormatPretty = "pretty"
)

// IsValidLoggingFormat returns true if the provided format is supported.
func IsValidLoggingFormat(f string) bool {
	return f == LoggingFormatJSONL || f == LoggingFormatPretty
}
