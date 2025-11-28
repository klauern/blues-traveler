package platform

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultDetector_DetectType(t *testing.T) {
	detector := NewDetector()

	tests := []struct {
		name         string
		setupFunc    func(t *testing.T) (cleanup func())
		expectedType Type
	}{
		{
			name: "env override to cursor",
			setupFunc: func(t *testing.T) func() {
				os.Setenv("BLUES_TRAVELER_PLATFORM", "cursor")
				return func() { os.Unsetenv("BLUES_TRAVELER_PLATFORM") }
			},
			expectedType: Cursor,
		},
		{
			name: "env override to claude",
			setupFunc: func(t *testing.T) func() {
				os.Setenv("BLUES_TRAVELER_PLATFORM", "claude")
				return func() { os.Unsetenv("BLUES_TRAVELER_PLATFORM") }
			},
			expectedType: ClaudeCode,
		},
		{
			name: "env override to claudecode",
			setupFunc: func(t *testing.T) func() {
				os.Setenv("BLUES_TRAVELER_PLATFORM", "claudecode")
				return func() { os.Unsetenv("BLUES_TRAVELER_PLATFORM") }
			},
			expectedType: ClaudeCode,
		},
		{
			name: "env override to claude-code",
			setupFunc: func(t *testing.T) func() {
				os.Setenv("BLUES_TRAVELER_PLATFORM", "claude-code")
				return func() { os.Unsetenv("BLUES_TRAVELER_PLATFORM") }
			},
			expectedType: ClaudeCode,
		},
		{
			name: "env override with mixed case",
			setupFunc: func(t *testing.T) func() {
				os.Setenv("BLUES_TRAVELER_PLATFORM", "CuRsOr")
				return func() { os.Unsetenv("BLUES_TRAVELER_PLATFORM") }
			},
			expectedType: Cursor,
		},
		{
			name: ".cursor directory exists",
			setupFunc: func(t *testing.T) func() {
				tmpDir := t.TempDir()
				cursorDir := filepath.Join(tmpDir, ".cursor")
				if err := os.Mkdir(cursorDir, 0o755); err != nil {
					t.Fatalf("Failed to create .cursor dir: %v", err)
				}
				oldCwd, _ := os.Getwd()
				os.Chdir(tmpDir)
				return func() { os.Chdir(oldCwd) }
			},
			expectedType: Cursor,
		},
		{
			name: ".claude directory exists",
			setupFunc: func(t *testing.T) func() {
				tmpDir := t.TempDir()
				claudeDir := filepath.Join(tmpDir, ".claude")
				if err := os.Mkdir(claudeDir, 0o755); err != nil {
					t.Fatalf("Failed to create .claude dir: %v", err)
				}
				oldCwd, _ := os.Getwd()
				os.Chdir(tmpDir)
				return func() { os.Chdir(oldCwd) }
			},
			expectedType: ClaudeCode,
		},
		{
			name: ".cursor takes precedence over .claude",
			setupFunc: func(t *testing.T) func() {
				tmpDir := t.TempDir()
				cursorDir := filepath.Join(tmpDir, ".cursor")
				claudeDir := filepath.Join(tmpDir, ".claude")
				if err := os.Mkdir(cursorDir, 0o755); err != nil {
					t.Fatalf("Failed to create .cursor dir: %v", err)
				}
				if err := os.Mkdir(claudeDir, 0o755); err != nil {
					t.Fatalf("Failed to create .claude dir: %v", err)
				}
				oldCwd, _ := os.Getwd()
				os.Chdir(tmpDir)
				return func() { os.Chdir(oldCwd) }
			},
			expectedType: Cursor,
		},
		{
			name: "no markers defaults to claude",
			setupFunc: func(t *testing.T) func() {
				tmpDir := t.TempDir()
				oldCwd, _ := os.Getwd()
				os.Chdir(tmpDir)
				return func() { os.Chdir(oldCwd) }
			},
			expectedType: ClaudeCode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setupFunc(t)
			defer cleanup()

			result, err := detector.DetectType()
			if err != nil {
				t.Fatalf("DetectType() error = %v", err)
			}

			if result != tt.expectedType {
				t.Errorf("DetectType() = %v, want %v", result, tt.expectedType)
			}
		})
	}
}

func TestTypeFromString(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedType Type
		expectError  bool
	}{
		{
			name:         "cursor lowercase",
			input:        "cursor",
			expectedType: Cursor,
			expectError:  false,
		},
		{
			name:         "cursor uppercase",
			input:        "CURSOR",
			expectedType: Cursor,
			expectError:  false,
		},
		{
			name:         "cursor mixed case",
			input:        "CuRsOr",
			expectedType: Cursor,
			expectError:  false,
		},
		{
			name:         "claudecode",
			input:        "claudecode",
			expectedType: ClaudeCode,
			expectError:  false,
		},
		{
			name:         "claude",
			input:        "claude",
			expectedType: ClaudeCode,
			expectError:  false,
		},
		{
			name:         "claude-code",
			input:        "claude-code",
			expectedType: ClaudeCode,
			expectError:  false,
		},
		{
			name:         "CLAUDECODE uppercase",
			input:        "CLAUDECODE",
			expectedType: ClaudeCode,
			expectError:  false,
		},
		{
			name:         "invalid platform",
			input:        "vscode",
			expectedType: "",
			expectError:  true,
		},
		{
			name:         "empty string",
			input:        "",
			expectedType: "",
			expectError:  true,
		},
		{
			name:         "random string",
			input:        "foobar",
			expectedType: "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := TypeFromString(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("TypeFromString(%q) expected error but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("TypeFromString(%q) unexpected error: %v", tt.input, err)
				}
			}

			if result != tt.expectedType {
				t.Errorf("TypeFromString(%q) = %v, want %v", tt.input, result, tt.expectedType)
			}
		})
	}
}

func TestDefaultDetector_Detect(t *testing.T) {
	detector := NewDetector()

	// This method should return an error as per the implementation
	platform, err := detector.Detect()

	if err == nil {
		t.Error("Detect() should return an error (legacy interface)")
	}

	if platform != nil {
		t.Errorf("Detect() should return nil platform, got %v", platform)
	}
}

func TestNewDetector(t *testing.T) {
	detector := NewDetector()

	if detector == nil {
		t.Error("NewDetector() returned nil")
	}

	// Verify it implements the Detector interface
	var _ Detector = detector
}
