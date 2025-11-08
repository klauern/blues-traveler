package core

import (
	"os"
	"strings"
)

// Platform identifies the runtime environment Blues Traveler is running under.
type Platform string

const (
	// PlatformUnknown represents an unspecified runtime. Defaults map to Claude Code semantics.
	PlatformUnknown Platform = ""
	// PlatformClaude represents the Claude Code hook runtime (cchooks).
	PlatformClaude Platform = "claude"
	// PlatformCursor represents the Cursor IDE hook runtime.
	PlatformCursor Platform = "cursor"
)

// DetectPlatform attempts to infer the current runtime platform.
//
// Detection order:
//  1. Explicit override via BT_HOOK_PLATFORM environment variable ("claude" or "cursor")
//  2. Presence of well-known Cursor-specific environment variables
//  3. Default to Claude semantics when no signal is present
func DetectPlatform() Platform {
	if override := strings.TrimSpace(os.Getenv("BT_HOOK_PLATFORM")); override != "" {
		switch strings.ToLower(override) {
		case string(PlatformCursor):
			return PlatformCursor
		case string(PlatformClaude):
			return PlatformClaude
		default:
			// Unknown override - fall back to default detection
		}
	}

	cursorEnvVars := []string{
		"CURSOR_AGENT_ROOT",
		"CURSOR_WORKSPACE_ID",
		"CURSOR_SESSION_ID",
		"CURSOR_CLI_HOST",
		"CURSOR_HOOK_ID",
	}
	for _, key := range cursorEnvVars {
		if value := os.Getenv(key); strings.TrimSpace(value) != "" {
			return PlatformCursor
		}
	}

	return PlatformClaude
}

// SupportsCursorAsk indicates whether the platform natively handles permission requests.
func (p Platform) SupportsCursorAsk() bool {
	return p == PlatformCursor
}

// OrDefault returns the platform if specified or falls back to DetectPlatform.
func (p Platform) OrDefault() Platform {
	if p == PlatformUnknown {
		return DetectPlatform()
	}
	return p
}
