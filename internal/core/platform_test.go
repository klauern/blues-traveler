package core

import (
	"os"
	"testing"
)

func TestDetectPlatform(t *testing.T) {
	tests := []struct {
		name             string
		envVar           string
		cursorEnvs       map[string]string
		expectedPlatform Platform
	}{
		{
			name:             "explicit override to cursor",
			envVar:           "cursor",
			expectedPlatform: PlatformCursor,
		},
		{
			name:             "explicit override to claude",
			envVar:           "claude",
			expectedPlatform: PlatformClaude,
		},
		{
			name:             "explicit override with uppercase",
			envVar:           "CURSOR",
			expectedPlatform: PlatformCursor,
		},
		{
			name:             "explicit override with mixed case",
			envVar:           "Claude",
			expectedPlatform: PlatformClaude,
		},
		{
			name:             "invalid override falls back to detection",
			envVar:           "invalid",
			expectedPlatform: PlatformClaude, // should default
		},
		{
			name: "cursor env var CURSOR_AGENT_ROOT",
			cursorEnvs: map[string]string{
				"CURSOR_AGENT_ROOT": "/some/path",
			},
			expectedPlatform: PlatformCursor,
		},
		{
			name: "cursor env var CURSOR_WORKSPACE_ID",
			cursorEnvs: map[string]string{
				"CURSOR_WORKSPACE_ID": "workspace-123",
			},
			expectedPlatform: PlatformCursor,
		},
		{
			name: "cursor env var CURSOR_SESSION_ID",
			cursorEnvs: map[string]string{
				"CURSOR_SESSION_ID": "session-456",
			},
			expectedPlatform: PlatformCursor,
		},
		{
			name: "cursor env var CURSOR_CLI_HOST",
			cursorEnvs: map[string]string{
				"CURSOR_CLI_HOST": "localhost:1234",
			},
			expectedPlatform: PlatformCursor,
		},
		{
			name: "cursor env var CURSOR_HOOK_ID",
			cursorEnvs: map[string]string{
				"CURSOR_HOOK_ID": "hook-789",
			},
			expectedPlatform: PlatformCursor,
		},
		{
			name: "cursor env with whitespace is valid",
			cursorEnvs: map[string]string{
				"CURSOR_AGENT_ROOT": "  /some/path  ",
			},
			expectedPlatform: PlatformCursor,
		},
		{
			name: "cursor env with only whitespace is ignored",
			cursorEnvs: map[string]string{
				"CURSOR_AGENT_ROOT": "   ",
			},
			expectedPlatform: PlatformClaude,
		},
		{
			name:             "no env vars defaults to claude",
			expectedPlatform: PlatformClaude,
		},
		{
			name:   "override takes precedence over cursor env",
			envVar: "claude",
			cursorEnvs: map[string]string{
				"CURSOR_AGENT_ROOT": "/some/path",
			},
			expectedPlatform: PlatformClaude,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean environment
			os.Unsetenv("BT_HOOK_PLATFORM")
			cursorVars := []string{
				"CURSOR_AGENT_ROOT",
				"CURSOR_WORKSPACE_ID",
				"CURSOR_SESSION_ID",
				"CURSOR_CLI_HOST",
				"CURSOR_HOOK_ID",
			}
			for _, key := range cursorVars {
				os.Unsetenv(key)
			}

			// Set up test environment
			if tt.envVar != "" {
				os.Setenv("BT_HOOK_PLATFORM", tt.envVar)
				defer os.Unsetenv("BT_HOOK_PLATFORM")
			}

			for key, value := range tt.cursorEnvs {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			// Test
			result := DetectPlatform()
			if result != tt.expectedPlatform {
				t.Errorf("DetectPlatform() = %q, want %q", result, tt.expectedPlatform)
			}
		})
	}
}

func TestPlatformSupportsCursorAsk(t *testing.T) {
	tests := []struct {
		platform Platform
		expected bool
	}{
		{PlatformCursor, true},
		{PlatformClaude, false},
		{PlatformUnknown, false},
		{Platform("custom"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.platform), func(t *testing.T) {
			result := tt.platform.SupportsCursorAsk()
			if result != tt.expected {
				t.Errorf("Platform(%q).SupportsCursorAsk() = %v, want %v", tt.platform, result, tt.expected)
			}
		})
	}
}

func TestPlatformOrDefault(t *testing.T) {
	// Set up a known environment for testing
	os.Setenv("BT_HOOK_PLATFORM", "cursor")
	defer os.Unsetenv("BT_HOOK_PLATFORM")

	tests := []struct {
		name     string
		platform Platform
		expected Platform
	}{
		{
			name:     "unknown platform uses detection",
			platform: PlatformUnknown,
			expected: PlatformCursor, // based on env var
		},
		{
			name:     "cursor platform stays cursor",
			platform: PlatformCursor,
			expected: PlatformCursor,
		},
		{
			name:     "claude platform stays claude",
			platform: PlatformClaude,
			expected: PlatformClaude,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.platform.OrDefault()
			if result != tt.expected {
				t.Errorf("Platform(%q).OrDefault() = %q, want %q", tt.platform, result, tt.expected)
			}
		})
	}
}
