package hooks

import (
	"context"
	"testing"
	"time"

	"github.com/brads3290/cchooks"
	"github.com/klauern/blues-traveler/internal/core"
)

func TestPerformanceHook_NewPerformanceHook(t *testing.T) {
	ctx := core.TestHookContext(nil)
	hook := NewPerformanceHook(ctx)

	if hook == nil {
		t.Fatal("Expected hook to be created")
	}

	if hook.Key() != "performance" {
		t.Errorf("Expected key 'performance', got %s", hook.Key())
	}

	if hook.Name() != "Performance Hook" {
		t.Errorf("Expected name 'Performance Hook', got %s", hook.Name())
	}
}

func TestPerformanceHook_Disabled(t *testing.T) {
	// Create a hook context with performance disabled
	ctx := core.TestHookContext(func(string) bool { return false })
	hook := NewPerformanceHook(ctx)

	if hook.IsEnabled() {
		t.Error("Expected hook to be disabled")
	}

	// Run should not error when disabled
	err := hook.Run()
	if err != nil {
		t.Errorf("Expected no error when disabled, got %v", err)
	}
}

func TestPerformanceHook_PreToolUseHandler(t *testing.T) {
	ctx := core.TestHookContext(nil)
	perfHook := NewPerformanceHook(ctx).(*PerformanceHook)

	// Create a pre-tool event
	event := &cchooks.PreToolUseEvent{
		ToolName: "Bash",
	}

	// Call the handler
	response := perfHook.preToolUseHandler(context.Background(), event)

	if response == nil {
		t.Fatal("Expected response to be non-nil")
	}

	// Verify that the start time was recorded
	if _, ok := perfHook.startTimes["Bash"]; !ok {
		t.Error("Expected start time to be recorded for Bash tool")
	}
}

func TestPerformanceHook_PostToolUseHandler(t *testing.T) {
	ctx := core.TestHookContext(nil)
	perfHook := NewPerformanceHook(ctx).(*PerformanceHook)

	// Set up a start time
	toolName := "Bash"
	perfHook.startTimes[toolName] = time.Now().Add(-100 * time.Millisecond)

	// Create a post-tool event
	event := &cchooks.PostToolUseEvent{
		ToolName: toolName,
	}

	// Call the handler
	response := perfHook.postToolUseHandler(context.Background(), event)

	if response == nil {
		t.Fatal("Expected response to be non-nil")
	}

	// Verify that the start time was cleaned up
	if _, ok := perfHook.startTimes[toolName]; ok {
		t.Error("Expected start time to be cleaned up after post-tool handler")
	}
}

func TestPerformanceHook_TimingAccuracy(t *testing.T) {
	ctx := core.TestHookContext(nil)
	perfHook := NewPerformanceHook(ctx).(*PerformanceHook)

	toolName := "TestTool"

	// Record start time
	preEvent := &cchooks.PreToolUseEvent{
		ToolName: toolName,
	}
	perfHook.preToolUseHandler(context.Background(), preEvent)

	// Simulate some work
	time.Sleep(50 * time.Millisecond)

	// Record end time
	postEvent := &cchooks.PostToolUseEvent{
		ToolName: toolName,
	}
	perfHook.postToolUseHandler(context.Background(), postEvent)

	// Verify timing was tracked (start time should be cleaned up)
	if _, ok := perfHook.startTimes[toolName]; ok {
		t.Error("Expected start time to be cleaned up")
	}
}

func TestPerformanceHook_MultipleTools(t *testing.T) {
	ctx := core.TestHookContext(nil)
	perfHook := NewPerformanceHook(ctx).(*PerformanceHook)

	// Start multiple tools
	tools := []string{"Bash", "Edit", "Read"}
	for _, tool := range tools {
		event := &cchooks.PreToolUseEvent{ToolName: tool}
		perfHook.preToolUseHandler(context.Background(), event)
	}

	// Verify all tools have start times
	if len(perfHook.startTimes) != len(tools) {
		t.Errorf("Expected %d start times, got %d", len(tools), len(perfHook.startTimes))
	}

	// Complete one tool
	postEvent := &cchooks.PostToolUseEvent{ToolName: "Bash"}
	perfHook.postToolUseHandler(context.Background(), postEvent)

	// Verify only the completed tool was cleaned up
	if len(perfHook.startTimes) != len(tools)-1 {
		t.Errorf("Expected %d start times after completion, got %d", len(tools)-1, len(perfHook.startTimes))
	}

	if _, ok := perfHook.startTimes["Bash"]; ok {
		t.Error("Expected Bash start time to be cleaned up")
	}
}
