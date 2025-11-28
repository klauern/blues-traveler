package cursor

import (
	"testing"

	"github.com/klauern/blues-traveler/internal/core"
)

func TestEventForAction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		action        Action
		wantEvent     core.EventType
		wantSupported bool
		wantOK        bool
	}{
		{
			name:          "before shell execution maps to PreToolUse",
			action:        ActionBeforeShellExecution,
			wantEvent:     core.PreToolUseEvent,
			wantSupported: true,
			wantOK:        true,
		},
		{
			name:          "after shell execution maps to PostToolUse",
			action:        ActionAfterShellExecution,
			wantEvent:     core.PostToolUseEvent,
			wantSupported: true,
			wantOK:        true,
		},
		{
			name:          "after file edit maps to PostToolUse",
			action:        ActionAfterFileEdit,
			wantEvent:     core.PostToolUseEvent,
			wantSupported: true,
			wantOK:        true,
		},
		{
			name:          "before read file maps to PreToolUse",
			action:        ActionBeforeReadFile,
			wantEvent:     core.PreToolUseEvent,
			wantSupported: true,
			wantOK:        true,
		},
		{
			name:          "before MCP execution maps to PreToolUse",
			action:        ActionBeforeMCPExecution,
			wantEvent:     core.PreToolUseEvent,
			wantSupported: true,
			wantOK:        true,
		},
		{
			name:          "after MCP execution maps to PostToolUse",
			action:        ActionAfterMCPExecution,
			wantEvent:     core.PostToolUseEvent,
			wantSupported: true,
			wantOK:        true,
		},
		{
			name:          "before submit prompt maps to UserPromptSubmit but unsupported",
			action:        ActionBeforeSubmitPrompt,
			wantEvent:     core.UserPromptSubmitEvent,
			wantSupported: false,
			wantOK:        true,
		},
		{
			name:          "stop maps to Stop event",
			action:        ActionStop,
			wantEvent:     core.StopEvent,
			wantSupported: true,
			wantOK:        true,
		},
		{
			name:          "after agent response maps to Stop event",
			action:        ActionAfterAgentResponse,
			wantEvent:     core.StopEvent,
			wantSupported: true,
			wantOK:        true,
		},
		{
			name:          "unknown action returns false",
			action:        Action("Unknown"),
			wantEvent:     "",
			wantSupported: false,
			wantOK:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mapping, ok := EventForAction(tt.action)
			if ok != tt.wantOK {
				t.Fatalf("EventForAction(%q) ok = %v, want %v", tt.action, ok, tt.wantOK)
			}

			if !ok {
				return
			}

			if mapping.Event != tt.wantEvent {
				t.Fatalf("EventForAction(%q) event = %v, want %v", tt.action, mapping.Event, tt.wantEvent)
			}

			if mapping.Supported != tt.wantSupported {
				t.Fatalf("EventForAction(%q) supported = %v, want %v", tt.action, mapping.Supported, tt.wantSupported)
			}
		})
	}
}

func TestNormalizeAction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input  string
		want   Action
		wantOK bool
	}{
		{input: "BeforeShellExecution", want: ActionBeforeShellExecution, wantOK: true},
		{input: "beforeShellExecution", want: ActionBeforeShellExecution, wantOK: true},
		{input: " beforeShellExecution \t", want: ActionBeforeShellExecution, wantOK: true},
		{input: "AfterFileEdit", want: ActionAfterFileEdit, wantOK: true},
		{input: "afterFileEdit", want: ActionAfterFileEdit, wantOK: true},
		{input: "BeforeReadFile", want: ActionBeforeReadFile, wantOK: true},
		{input: "beforeReadFile", want: ActionBeforeReadFile, wantOK: true},
		{input: "", want: "", wantOK: false},
		{input: "unknown", want: "", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			got, ok := NormalizeAction(tt.input)
			if ok != tt.wantOK {
				t.Fatalf("NormalizeAction(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
			}

			if ok && got != tt.want {
				t.Fatalf("NormalizeAction(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEventAliasesIncludesBeforeReadFile(t *testing.T) {
	t.Parallel()

	aliases := EventAliases(core.PreToolUseEvent)
	if len(aliases) == 0 {
		t.Fatalf("expected aliases for PreToolUseEvent, got none")
	}

	found := false
	for _, alias := range aliases {
		if alias == "BeforeReadFile" {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("expected BeforeReadFile alias for PreToolUseEvent, got %v", aliases)
	}
}

func TestEventAliases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		event        core.EventType
		wantContains []string
		wantNil      bool
	}{
		{
			name:         "PreToolUse has multiple aliases",
			event:        core.PreToolUseEvent,
			wantContains: []string{"BeforeReadFile", "BeforeShellExecution", "BeforeMCPExecution"},
		},
		{
			name:         "PostToolUse has multiple aliases",
			event:        core.PostToolUseEvent,
			wantContains: []string{"AfterFileEdit", "AfterShellExecution", "AfterMCPExecution"},
		},
		{
			name:         "Stop has multiple aliases",
			event:        core.StopEvent,
			wantContains: []string{"OnStop", "Stop"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := EventAliases(tt.event)

			if tt.wantNil {
				if got != nil {
					t.Fatalf("EventAliases(%q) = %v, want nil", tt.event, got)
				}
				return
			}

			if got == nil {
				t.Fatalf("EventAliases(%q) = nil, want non-nil", tt.event)
			}

			for _, wantAlias := range tt.wantContains {
				found := false
				for _, alias := range got {
					if alias == wantAlias {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("EventAliases(%q) missing expected alias %q, got %v", tt.event, wantAlias, got)
				}
			}
		})
	}
}

func TestResolveCursorEventRecognizesBeforeReadFile(t *testing.T) {
	t.Parallel()

	resolved, ok := ResolveCursorEvent("BeforeReadFile")
	if !ok {
		t.Fatalf("expected BeforeReadFile to resolve to a core event")
	}

	if resolved != core.PreToolUseEvent {
		t.Fatalf("expected BeforeReadFile to resolve to PreToolUseEvent, got %q", resolved)
	}
}

func TestResolveCursorEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantEvent core.EventType
		wantOK    bool
	}{
		{
			name:      "canonical event name",
			input:     string(core.PreToolUseEvent),
			wantEvent: core.PreToolUseEvent,
			wantOK:    true,
		},
		{
			name:      "cursor alias BeforeReadFile",
			input:     "BeforeReadFile",
			wantEvent: core.PreToolUseEvent,
			wantOK:    true,
		},
		{
			name:      "cursor alias lowerCamelCase",
			input:     "beforeReadFile",
			wantEvent: core.PreToolUseEvent,
			wantOK:    true,
		},
		{
			name:      "AfterFileEdit",
			input:     "AfterFileEdit",
			wantEvent: core.PostToolUseEvent,
			wantOK:    true,
		},
		{
			name:      "Stop",
			input:     "Stop",
			wantEvent: core.StopEvent,
			wantOK:    true,
		},
		{
			name:      "unknown event",
			input:     "UnknownEvent",
			wantEvent: "",
			wantOK:    false,
		},
		{
			name:      "empty string",
			input:     "",
			wantEvent: "",
			wantOK:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := ResolveCursorEvent(tt.input)
			if ok != tt.wantOK {
				t.Fatalf("ResolveCursorEvent(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
			}

			if ok && got != tt.wantEvent {
				t.Fatalf("ResolveCursorEvent(%q) = %q, want %q", tt.input, got, tt.wantEvent)
			}
		})
	}
}

func TestEventAliasesReturnsDefensiveCopy(t *testing.T) {
	t.Parallel()

	aliases1 := EventAliases(core.PreToolUseEvent)
	aliases2 := EventAliases(core.PreToolUseEvent)

	if len(aliases1) == 0 {
		t.Fatal("expected non-empty aliases")
	}

	// Modify first slice
	aliases1[0] = "Modified"

	// Second slice should be unchanged
	if aliases2[0] == "Modified" {
		t.Fatal("EventAliases did not return a defensive copy")
	}
}
