package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

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

// SetupLogRotation configures log rotation for a given log file path
func SetupLogRotation(logPath string, config LogRotationConfig) *lumberjack.Logger {
	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
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
