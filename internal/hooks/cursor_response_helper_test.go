package hooks

import (
	"testing"

	"github.com/brads3290/cchooks"
	"github.com/klauern/blues-traveler/internal/core"
)

type stubLogger struct {
	events []loggedEvent
}

type loggedEvent struct {
	event    string
	toolName string
	rawData  map[string]interface{}
	details  map[string]interface{}
}

func (s *stubLogger) LogHookEvent(event string, toolName string, rawData map[string]interface{}, details map[string]interface{}) {
	s.events = append(s.events, loggedEvent{event: event, toolName: toolName, rawData: rawData, details: details})
}

func TestTranslateCursorResponsePre(t *testing.T) {
	jobName := "unit-test"
	logger := &stubLogger{}

	t.Run("continue false blocks execution", func(t *testing.T) {
		cont := false
		resp, _ := translateCursorResponse(jobName, core.PreToolUseEvent, logger, &CursorHookResponse{
			Continue:    &cont,
			UserMessage: "explicit user",
		})

		dual, ok := resp.(*core.DualMessagePreToolResponse)
		if !ok {
			t.Fatalf("expected DualMessagePreToolResponse, got %T", resp)
		}

		if got := dual.GetUserMessage(); got != "explicit user" {
			t.Errorf("user message = %q, want %q", got, "explicit user")
		}

		if got := dual.GetAgentMessage(); got != "explicit user" {
			t.Errorf("agent message = %q, want %q", got, "explicit user")
		}
	})

	t.Run("deny falls back to defaults", func(t *testing.T) {
		resp, _ := translateCursorResponse(jobName, core.PreToolUseEvent, logger, &CursorHookResponse{
			Permission: "deny",
		})

		dual, ok := resp.(*core.DualMessagePreToolResponse)
		if !ok {
			t.Fatalf("expected DualMessagePreToolResponse, got %T", resp)
		}

		want := "Hook 'unit-test' denied permission"
		if got := dual.GetUserMessage(); got != want {
			t.Errorf("user message = %q, want %q", got, want)
		}
	})

	t.Run("ask logs and requests confirmation", func(t *testing.T) {
		logger.events = nil
		resp, _ := translateCursorResponse(jobName, core.PreToolUseEvent, logger, &CursorHookResponse{
			Permission:   "ask",
			UserMessage:  "Need approval",
			AgentMessage: "Explain details",
		})

		// Check if it's a DualMessagePreToolResponse (new ask implementation)
		dual, ok := resp.(*core.DualMessagePreToolResponse)
		if !ok {
			t.Fatalf("expected DualMessagePreToolResponse, got %T", resp)
		}

		if got := dual.GetUserMessage(); got != "Need approval" {
			t.Errorf("user message = %q, want %q", got, "Need approval")
		}

		if len(logger.events) != 1 {
			t.Fatalf("expected 1 log event, got %d", len(logger.events))
		}

		evt := logger.events[0]
		if evt.event != "hook_ask_permission" {
			t.Errorf("event = %q, want %q", evt.event, "hook_ask_permission")
		}

		if evt.toolName != jobName {
			t.Errorf("toolName = %q, want %q", evt.toolName, jobName)
		}
	})

	t.Run("allow without messages returns approve", func(t *testing.T) {
		resp, _ := translateCursorResponse(jobName, core.PreToolUseEvent, logger, &CursorHookResponse{
			Permission: "allow",
		})

		if _, ok := resp.(*cchooks.PreToolUseResponse); !ok {
			t.Fatalf("expected *cchooks.PreToolUseResponse, got %T", resp)
		}
	})

	t.Run("unknown permission blocks", func(t *testing.T) {
		resp, _ := translateCursorResponse(jobName, core.PreToolUseEvent, logger, &CursorHookResponse{
			Permission: "unexpected",
		})

		dual, ok := resp.(*core.DualMessagePreToolResponse)
		if !ok {
			t.Fatalf("expected DualMessagePreToolResponse, got %T", resp)
		}

		want := "Hook 'unit-test' returned unknown permission: unexpected"
		if got := dual.GetUserMessage(); got != want {
			t.Errorf("user message = %q, want %q", got, want)
		}
	})
}

func TestTranslateCursorResponsePost(t *testing.T) {
	jobName := "unit-test"
	logger := &stubLogger{}

	t.Run("continue false blocks execution", func(t *testing.T) {
		cont := false
		_, resp := translateCursorResponse(jobName, core.PostToolUseEvent, logger, &CursorHookResponse{
			Continue: &cont,
		})

		dual, ok := resp.(*core.DualMessagePostToolResponse)
		if !ok {
			t.Fatalf("expected DualMessagePostToolResponse, got %T", resp)
		}

		want := "Hook 'unit-test' blocked execution"
		if got := dual.GetUserMessage(); got != want {
			t.Errorf("user message = %q, want %q", got, want)
		}
	})

	t.Run("deny with explicit messages", func(t *testing.T) {
		_, resp := translateCursorResponse(jobName, core.PostToolUseEvent, logger, &CursorHookResponse{
			Permission:   "deny",
			UserMessage:  "User",
			AgentMessage: "Agent",
		})

		dual, ok := resp.(*core.DualMessagePostToolResponse)
		if !ok {
			t.Fatalf("expected DualMessagePostToolResponse, got %T", resp)
		}

		if got := dual.GetAgentMessage(); got != "Agent" {
			t.Errorf("agent message = %q, want %q", got, "Agent")
		}
	})

	t.Run("ask logs and allows with messages", func(t *testing.T) {
		logger.events = nil
		_, resp := translateCursorResponse(jobName, core.PostToolUseEvent, logger, &CursorHookResponse{
			Permission:   "ask",
			UserMessage:  "Allow?",
			AgentMessage: "Please",
		})

		dual, ok := resp.(*core.DualMessagePostToolResponse)
		if !ok {
			t.Fatalf("expected DualMessagePostToolResponse, got %T", resp)
		}

		if got := dual.GetUserMessage(); got != "Allow?" {
			t.Errorf("user message = %q, want %q", got, "Allow?")
		}

		if len(logger.events) != 1 {
			t.Fatalf("expected 1 log event, got %d", len(logger.events))
		}

		if logger.events[0].event != "hook_ask_permission_post" {
			t.Errorf("event = %q, want %q", logger.events[0].event, "hook_ask_permission_post")
		}
	})

	t.Run("allow without messages returns allow", func(t *testing.T) {
		_, resp := translateCursorResponse(jobName, core.PostToolUseEvent, logger, &CursorHookResponse{
			Permission: "allow",
		})

		if _, ok := resp.(*cchooks.PostToolUseResponse); !ok {
			t.Fatalf("expected *cchooks.PostToolUseResponse, got %T", resp)
		}
	})
}
