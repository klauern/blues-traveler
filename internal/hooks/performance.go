package hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/brads3290/cchooks"
	"github.com/klauern/blues-traveler/internal/core"
)

// PerformanceHook implements performance monitoring and timing
type PerformanceHook struct {
	*core.BaseHook
	startTimes map[string]time.Time
}

// PerformanceEntry represents a performance log entry
type PerformanceEntry struct {
	Timestamp   string  `json:"timestamp"`
	Event       string  `json:"event"`
	ToolName    string  `json:"tool_name"`
	DurationMS  float64 `json:"duration_ms,omitempty"`
	Description string  `json:"description,omitempty"`
}

// NewPerformanceHook creates a new performance hook instance
func NewPerformanceHook(ctx *core.HookContext) core.Hook {
	base := core.NewBaseHook("performance", "Performance Hook", "Monitors tool execution timing and performance metrics", ctx)
	return &PerformanceHook{
		BaseHook:   base,
		startTimes: make(map[string]time.Time),
	}
}

// Run executes the performance hook.
func (h *PerformanceHook) Run() error {
	if !h.IsEnabled() {
		fmt.Println("Performance plugin disabled - skipping")
		return nil
	}

	runner := h.Context().RunnerFactory(h.preToolUseHandler, h.postToolUseHandler, h.CreateRawHandler())
	fmt.Println("Performance monitoring started")
	runner.Run()
	return nil
}

func (h *PerformanceHook) preToolUseHandler(_ context.Context, event *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface {
	// Record start time for this tool invocation
	h.startTimes[event.ToolName] = time.Now()

	entry := PerformanceEntry{
		Timestamp:   time.Now().Format(time.RFC3339),
		Event:       "tool_start",
		ToolName:    event.ToolName,
		Description: "Tool execution started",
	}

	h.logPerformanceEntry(entry)

	// Also use the detailed logging if enabled
	if h.Context().LoggingEnabled {
		rawData := map[string]interface{}{
			"tool_name": event.ToolName,
			"event":     "start",
		}
		details := map[string]interface{}{
			"timestamp": entry.Timestamp,
		}
		h.LogHookEvent("performance_pre", event.ToolName, rawData, details)
	}

	return cchooks.Approve()
}

func (h *PerformanceHook) postToolUseHandler(_ context.Context, event *cchooks.PostToolUseEvent) cchooks.PostToolUseResponseInterface {
	// Calculate duration if we have a start time
	var durationMS float64
	if startTime, ok := h.startTimes[event.ToolName]; ok {
		duration := time.Since(startTime)
		durationMS = float64(duration.Milliseconds())
		delete(h.startTimes, event.ToolName) // Clean up
	}

	entry := PerformanceEntry{
		Timestamp:   time.Now().Format(time.RFC3339),
		Event:       "tool_complete",
		ToolName:    event.ToolName,
		DurationMS:  durationMS,
		Description: fmt.Sprintf("Tool execution completed in %.2fms", durationMS),
	}

	h.logPerformanceEntry(entry)

	// Also use the detailed logging if enabled
	if h.Context().LoggingEnabled {
		rawData := map[string]interface{}{
			"tool_name":   event.ToolName,
			"event":       "complete",
			"duration_ms": durationMS,
		}
		details := map[string]interface{}{
			"timestamp":   entry.Timestamp,
			"duration_ms": durationMS,
		}
		h.LogHookEvent("performance_post", event.ToolName, rawData, details)
	}

	return cchooks.Allow()
}

func (h *PerformanceHook) logPerformanceEntry(entry PerformanceEntry) {
	jsonData, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal performance entry: %v\n", err)
		return
	}

	fmt.Println(string(jsonData))
}
