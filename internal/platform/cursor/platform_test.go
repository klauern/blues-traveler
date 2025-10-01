package cursor

import (
	"testing"

	"github.com/klauern/blues-traveler/internal/core"
)

func TestCursorPlatform_Type(t *testing.T) {
	p := New()
	if string(p.Type()) != "cursor" {
		t.Errorf("Expected type 'cursor', got %s", p.Type())
	}
}

func TestCursorPlatform_Name(t *testing.T) {
	p := New()
	if p.Name() != "Cursor" {
		t.Errorf("Expected name 'Cursor', got %s", p.Name())
	}
}

func TestCursorPlatform_MapEventFromGeneric(t *testing.T) {
	p := New()

	tests := []struct {
		name         string
		genericEvent core.EventType
		wantEvents   []string
	}{
		{
			name:         "PreToolUse maps to shell, MCP, and file read",
			genericEvent: core.PreToolUseEvent,
			wantEvents:   []string{BeforeShellExecution, BeforeMCPExecution, BeforeReadFile},
		},
		{
			name:         "PostToolUse maps to file edit",
			genericEvent: core.PostToolUseEvent,
			wantEvents:   []string{AfterFileEdit},
		},
		{
			name:         "UserPromptSubmit maps to beforeSubmitPrompt",
			genericEvent: core.UserPromptSubmitEvent,
			wantEvents:   []string{BeforeSubmitPrompt},
		},
		{
			name:         "Stop maps to stop",
			genericEvent: core.StopEvent,
			wantEvents:   []string{Stop},
		},
		{
			name:         "Notification not supported",
			genericEvent: core.NotificationEvent,
			wantEvents:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events := p.MapEventFromGeneric(tt.genericEvent)
			if len(events) != len(tt.wantEvents) {
				t.Errorf("Expected %d events, got %d", len(tt.wantEvents), len(events))
				return
			}
			for i, want := range tt.wantEvents {
				if events[i] != want {
					t.Errorf("Event %d: expected %s, got %s", i, want, events[i])
				}
			}
		})
	}
}

func TestCursorPlatform_MapEventToGeneric(t *testing.T) {
	p := New()

	tests := []struct {
		name          string
		cursorEvent   string
		wantEvent     core.EventType
		wantSupported bool
	}{
		{
			name:          "beforeShellExecution maps to PreToolUse",
			cursorEvent:   BeforeShellExecution,
			wantEvent:     core.PreToolUseEvent,
			wantSupported: true,
		},
		{
			name:          "beforeMCPExecution maps to PreToolUse",
			cursorEvent:   BeforeMCPExecution,
			wantEvent:     core.PreToolUseEvent,
			wantSupported: true,
		},
		{
			name:          "afterFileEdit maps to PostToolUse",
			cursorEvent:   AfterFileEdit,
			wantEvent:     core.PostToolUseEvent,
			wantSupported: true,
		},
		{
			name:          "beforeReadFile maps to PreToolUse",
			cursorEvent:   BeforeReadFile,
			wantEvent:     core.PreToolUseEvent,
			wantSupported: true,
		},
		{
			name:          "invalid event not supported",
			cursorEvent:   "invalidEvent",
			wantEvent:     "",
			wantSupported: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, supported := p.MapEventToGeneric(tt.cursorEvent)
			if supported != tt.wantSupported {
				t.Errorf("Expected supported=%v, got %v", tt.wantSupported, supported)
			}
			if event != tt.wantEvent {
				t.Errorf("Expected event %s, got %s", tt.wantEvent, event)
			}
		})
	}
}

func TestCursorPlatform_ValidateEventName(t *testing.T) {
	p := New()

	validEvents := []string{
		BeforeShellExecution,
		BeforeMCPExecution,
		AfterFileEdit,
		BeforeReadFile,
		BeforeSubmitPrompt,
		Stop,
	}

	for _, event := range validEvents {
		if !p.ValidateEventName(event) {
			t.Errorf("Event %s should be valid", event)
		}
	}

	invalidEvents := []string{
		"PreToolUse", // Claude Code event, not Cursor
		"invalidEvent",
		"",
	}

	for _, event := range invalidEvents {
		if p.ValidateEventName(event) {
			t.Errorf("Event %s should be invalid", event)
		}
	}
}

func TestCursorPlatform_AllEvents(t *testing.T) {
	p := New()
	events := p.AllEvents()

	if len(events) != 6 {
		t.Errorf("Expected 6 events, got %d", len(events))
	}

	// All Cursor events should require stdio
	for _, event := range events {
		if !event.RequiresStdio {
			t.Errorf("Event %s should require stdio", event.Name)
		}
		if event.SupportsFilter {
			t.Errorf("Event %s should not support filters (Cursor doesn't have config-level filters)", event.Name)
		}
	}
}
