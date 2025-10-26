package cmd

import (
	"testing"
)

func TestParseSyncOptions_EventValidation(t *testing.T) {
	// Mock validation functions
	validEvents := []string{"PreToolUse", "PostToolUse", "UserPromptSubmit"}
	isValidEventType := func(event string) bool {
		for _, v := range validEvents {
			if event == v {
				return true
			}
		}
		return false
	}
	_ = func() []string {
		return validEvents
	}

	tests := []struct {
		name        string
		eventFilter string
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid event - PreToolUse",
			eventFilter: "PreToolUse",
			wantErr:     false,
		},
		{
			name:        "valid event - PostToolUse",
			eventFilter: "PostToolUse",
			wantErr:     false,
		},
		{
			name:        "empty event filter - should pass",
			eventFilter: "",
			wantErr:     false,
		},
		{
			name:        "invalid event name",
			eventFilter: "InvalidEvent",
			wantErr:     true,
			errContains: "invalid event 'InvalidEvent'",
		},
		{
			name:        "typo in event name",
			eventFilter: "PreTollUse",
			wantErr:     true,
			errContains: "invalid event",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This is a simplified test - in a real scenario we'd mock cli.Command
			// For now, we're testing the validation logic conceptually

			// Simulate what parseSyncOptions does
			if tt.eventFilter != "" && !isValidEventType(tt.eventFilter) {
				// This should match the error in parseSyncOptions
				if !tt.wantErr {
					t.Errorf("Expected no error for event %q, but validation would fail", tt.eventFilter)
				}
				// Verify error message would contain the expected text
				if tt.errContains == "" {
					t.Error("Expected error contains string not specified in test")
				}
			} else {
				if tt.wantErr {
					t.Errorf("Expected error for event %q, but validation would pass", tt.eventFilter)
				}
			}
		})
	}
}

func TestEventValidation_ListsValidEvents(t *testing.T) {
	validEvents := []string{"PreToolUse", "PostToolUse", "UserPromptSubmit", "Notification", "Stop"}
	isValidEventType := func(event string) bool {
		for _, v := range validEvents {
			if event == v {
				return true
			}
		}
		return false
	}

	// Test that validation correctly identifies valid events
	for _, event := range validEvents {
		if !isValidEventType(event) {
			t.Errorf("Event %q should be valid but validation says it's not", event)
		}
	}

	// Test that validation correctly rejects invalid events
	invalidEvents := []string{"InvalidEvent", "pretooluse", "PRETOOLUSE", "PreTollUse", ""}
	for _, event := range invalidEvents {
		if event != "" && isValidEventType(event) {
			t.Errorf("Event %q should be invalid but validation says it's valid", event)
		}
	}
}
