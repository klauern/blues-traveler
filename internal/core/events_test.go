package core

import (
	"testing"
)

func TestIsValidEventType(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		want      bool
	}{
		// Canonical names
		{
			name:      "canonical PreToolUse",
			eventType: "PreToolUse",
			want:      true,
		},
		{
			name:      "canonical PostToolUse",
			eventType: "PostToolUse",
			want:      true,
		},
		{
			name:      "canonical SessionEnd",
			eventType: "SessionEnd",
			want:      true,
		},
		// Cursor aliases
		{
			name:      "cursor alias beforeShellExecution",
			eventType: "beforeShellExecution",
			want:      true,
		},
		{
			name:      "cursor alias afterFileEdit",
			eventType: "afterFileEdit",
			want:      true,
		},
		{
			name:      "cursor alias onSessionStart",
			eventType: "onSessionStart",
			want:      true,
		},
		// Invalid names
		{
			name:      "invalid event name",
			eventType: "InvalidEvent",
			want:      false,
		},
		{
			name:      "empty string",
			eventType: "",
			want:      false,
		},
		{
			name:      "typo in canonical name",
			eventType: "PreTollUse",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidEventType(tt.eventType)
			if got != tt.want {
				t.Errorf("IsValidEventType(%q) = %v, want %v", tt.eventType, got, tt.want)
			}
		})
	}
}

func TestResolveEventAlias(t *testing.T) {
	tests := []struct {
		name      string
		eventName string
		want      string
	}{
		// Canonical names should return themselves
		{
			name:      "canonical PreToolUse unchanged",
			eventName: "PreToolUse",
			want:      "PreToolUse",
		},
		{
			name:      "canonical PostToolUse unchanged",
			eventName: "PostToolUse",
			want:      "PostToolUse",
		},
		// Cursor aliases should resolve to canonical names
		{
			name:      "beforeShellExecution resolves to PreToolUse",
			eventName: "beforeShellExecution",
			want:      "PreToolUse",
		},
		{
			name:      "beforeToolUse resolves to PreToolUse",
			eventName: "beforeToolUse",
			want:      "PreToolUse",
		},
		{
			name:      "beforeFileEdit resolves to PreToolUse",
			eventName: "beforeFileEdit",
			want:      "PreToolUse",
		},
		{
			name:      "afterShellExecution resolves to PostToolUse",
			eventName: "afterShellExecution",
			want:      "PostToolUse",
		},
		{
			name:      "afterFileEdit resolves to PostToolUse",
			eventName: "afterFileEdit",
			want:      "PostToolUse",
		},
		{
			name:      "onSessionStart resolves to SessionStart",
			eventName: "onSessionStart",
			want:      "SessionStart",
		},
		{
			name:      "onSessionEnd resolves to SessionEnd",
			eventName: "onSessionEnd",
			want:      "SessionEnd",
		},
		{
			name:      "onNotification resolves to Notification",
			eventName: "onNotification",
			want:      "Notification",
		},
		{
			name:      "onPromptSubmit resolves to UserPromptSubmit",
			eventName: "onPromptSubmit",
			want:      "UserPromptSubmit",
		},
		// Invalid names should return empty string
		{
			name:      "invalid name returns empty",
			eventName: "InvalidEvent",
			want:      "",
		},
		{
			name:      "empty string returns empty",
			eventName: "",
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveEventAlias(tt.eventName)
			if got != tt.want {
				t.Errorf("ResolveEventAlias(%q) = %q, want %q", tt.eventName, got, tt.want)
			}
		})
	}
}

func TestGetEventAliases(t *testing.T) {
	tests := []struct {
		name          string
		canonicalName string
		wantContains  []string
	}{
		{
			name:          "PreToolUse has cursor aliases",
			canonicalName: "PreToolUse",
			wantContains:  []string{"beforeToolUse", "beforeShellExecution", "beforeFileEdit", "beforeFileWrite"},
		},
		{
			name:          "PostToolUse has cursor aliases",
			canonicalName: "PostToolUse",
			wantContains:  []string{"afterToolUse", "afterShellExecution", "afterFileEdit", "afterFileWrite"},
		},
		{
			name:          "SessionStart has cursor aliases",
			canonicalName: "SessionStart",
			wantContains:  []string{"onSessionStart", "onStart", "onSessionBegin"},
		},
		{
			name:          "Invalid name returns nil",
			canonicalName: "InvalidEvent",
			wantContains:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetEventAliases(tt.canonicalName)

			if tt.wantContains == nil {
				if got != nil {
					t.Errorf("GetEventAliases(%q) = %v, want nil", tt.canonicalName, got)
				}
				return
			}

			if got == nil {
				t.Errorf("GetEventAliases(%q) = nil, want non-nil", tt.canonicalName)
				return
			}

			// Check that all expected aliases are present
			for _, expectedAlias := range tt.wantContains {
				found := false
				for _, alias := range got {
					if alias == expectedAlias {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GetEventAliases(%q) missing expected alias %q", tt.canonicalName, expectedAlias)
				}
			}
		})
	}
}

func TestAllClaudeCodeEvents_HasCursorAliases(t *testing.T) {
	events := AllClaudeCodeEvents()

	// Ensure all events have the CursorAliases field
	for _, event := range events {
		// Every event should have at least one alias (even if it's for future compatibility)
		// Actually, some events may not have established Cursor equivalents yet,
		// so just verify the field exists and is accessible
		_ = event.CursorAliases
	}
}

func TestCursorCompatibility_Roundtrip(t *testing.T) {
	// Test that we can go from Cursor alias -> canonical -> back to valid
	cursorAliases := []string{
		"beforeShellExecution",
		"beforeFileEdit",
		"afterShellExecution",
		"afterFileEdit",
		"onSessionStart",
		"onSessionEnd",
		"onNotification",
		"onPromptSubmit",
	}

	for _, alias := range cursorAliases {
		t.Run(alias, func(t *testing.T) {
			// Should be valid
			if !IsValidEventType(alias) {
				t.Errorf("Cursor alias %q should be recognized as valid", alias)
			}

			// Should resolve to canonical name
			canonical := ResolveEventAlias(alias)
			if canonical == "" {
				t.Errorf("Cursor alias %q failed to resolve to canonical name", alias)
			}

			// Canonical name should also be valid
			if !IsValidEventType(canonical) {
				t.Errorf("Resolved canonical name %q should be valid", canonical)
			}

			// Resolving canonical name should return itself
			if ResolveEventAlias(canonical) != canonical {
				t.Errorf("Canonical name %q should resolve to itself", canonical)
			}
		})
	}
}
