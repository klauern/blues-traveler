package platform

import (
	"fmt"
	"os"
	"strings"
)

// DefaultDetector implements platform auto-detection
type DefaultDetector struct{}

// NewDetector creates a new platform detector
func NewDetector() Detector {
	return &DefaultDetector{}
}

// Detect attempts to detect the current platform type
func (d *DefaultDetector) DetectType() (Type, error) {
	// 1. Check environment variable override
	if platformEnv := os.Getenv("BLUES_TRAVELER_PLATFORM"); platformEnv != "" {
		return TypeFromString(platformEnv)
	}

	// 2. Check for .cursor directory in current working directory
	if _, err := os.Stat(".cursor"); err == nil {
		return Cursor, nil
	}

	// 3. Check for .claude directory in current working directory
	if _, err := os.Stat(".claude"); err == nil {
		return ClaudeCode, nil
	}

	// 4. Check for Cursor config in home directory
	home, err := os.UserHomeDir()
	if err == nil {
		cursorConfig := home + "/.cursor/hooks.json"
		if _, err := os.Stat(cursorConfig); err == nil {
			return Cursor, nil
		}
	}

	// 5. Default to Claude Code for backward compatibility
	return ClaudeCode, nil
}

// Detect attempts to detect the current platform (legacy interface)
func (d *DefaultDetector) Detect() (Platform, error) {
	_, err := d.DetectType()
	if err != nil {
		return nil, err
	}
	// Note: Caller must instantiate the platform to avoid import cycles
	// This method exists for interface compatibility but should not be used
	return nil, fmt.Errorf("use DetectType() instead to avoid import cycles")
}

// TypeFromString converts a string to a platform Type
func TypeFromString(s string) (Type, error) {
	switch strings.ToLower(s) {
	case "cursor":
		return Cursor, nil
	case "claudecode", "claude", "claude-code":
		return ClaudeCode, nil
	default:
		return "", fmt.Errorf("unknown platform: %s (valid: cursor, claudecode)", s)
	}
}
