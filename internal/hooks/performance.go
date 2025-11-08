package hooks

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/brads3290/cchooks"
	"github.com/klauern/blues-traveler/internal/core"
)

// PerformanceHook implements performance monitoring logic
type PerformanceHook struct {
	*core.BaseHook
	logger      *log.Logger
	logFile     *os.File
	startTimes  map[string]time.Time
	totalTime   time.Duration
	toolCount   int
}

// NewPerformanceHook creates a new performance hook instance
func NewPerformanceHook(ctx *core.HookContext) core.Hook {
	base := core.NewBaseHook(
		"performance",
		"Performance Hook",
		"Monitors hook execution performance and resource usage",
		ctx,
	)
	return &PerformanceHook{
		BaseHook:   base,
		startTimes: make(map[string]time.Time),
	}
}

// Run executes the performance hook
func (h *PerformanceHook) Run() error {
	if !h.IsEnabled() {
		fmt.Println("Performance plugin disabled - skipping")
		return nil
	}

	h.ensureLogger()
	if h.logger == nil {
		return fmt.Errorf("failed to initialize logger")
	}

	defer func() {
		if h.logFile != nil {
			// Log summary statistics
			h.logSummary()
			if err := h.logFile.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "performance log close error: %v\n", err)
			}
		}
	}()

	runner := h.Context().RunnerFactory(
		h.preToolUseHandler,
		h.postToolUseHandler,
		h.CreateRawHandler(),
	)
	fmt.Println("Performance hook started - monitoring to .claude/hooks/performance.log")
	runner.Run()
	return nil
}

func (h *PerformanceHook) ensureLogger() {
	if h.logger != nil {
		return
	}

	// Ensure directory exists
	logPath := ".claude/hooks/performance.log"
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0o750); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create performance log dir %s: %v\n", logDir, err)
		return
	}

	var err error
	h.logFile, err = h.Context().FileSystem.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open performance log file %s: %v\n", logPath, err)
		return
	}
	h.logger = log.New(h.logFile, "", log.LstdFlags)
	h.logger.Println("=== Performance Monitoring Session Started ===")
}

func (h *PerformanceHook) preToolUseHandler(
	_ context.Context,
	event *cchooks.PreToolUseEvent,
) cchooks.PreToolUseResponseInterface {
	h.ensureLogger()

	// Record start time
	eventID := fmt.Sprintf("%s-%d", event.ToolName, time.Now().UnixNano())
	h.startTimes[eventID] = time.Now()
	h.toolCount++

	if h.logger != nil {
		h.logger.Printf("START: %s [ID: %s]", event.ToolName, eventID)
	}

	// Log detailed event data if logging is enabled
	if h.Context().LoggingEnabled {
		details := make(map[string]interface{})
		rawData := make(map[string]interface{})
		rawData["tool_name"] = event.ToolName
		rawData["event_id"] = eventID
		rawData["timestamp"] = time.Now().Format(time.RFC3339Nano)

		h.LogHookEvent("performance_start", event.ToolName, rawData, details)
	}

	return cchooks.Approve()
}

func (h *PerformanceHook) postToolUseHandler(
	_ context.Context,
	event *cchooks.PostToolUseEvent,
) cchooks.PostToolUseResponseInterface {
	h.ensureLogger()

	// Find the most recent start time for this tool
	var eventID string
	var startTime time.Time
	for id, st := range h.startTimes {
		if len(id) > len(event.ToolName) && id[:len(event.ToolName)] == event.ToolName {
			if startTime.IsZero() || st.After(startTime) {
				eventID = id
				startTime = st
			}
		}
	}

	if !startTime.IsZero() {
		elapsed := time.Since(startTime)
		h.totalTime += elapsed
		delete(h.startTimes, eventID)

		if h.logger != nil {
			h.logger.Printf("END: %s [ID: %s] Duration: %v", event.ToolName, eventID, elapsed)
		}

		// Log detailed event data if logging is enabled
		if h.Context().LoggingEnabled {
			details := make(map[string]interface{})
			rawData := make(map[string]interface{})
			rawData["tool_name"] = event.ToolName
			rawData["event_id"] = eventID
			rawData["duration_ms"] = elapsed.Milliseconds()
			rawData["duration_ns"] = elapsed.Nanoseconds()
			rawData["timestamp"] = time.Now().Format(time.RFC3339Nano)

			details["elapsed_time"] = elapsed.String()

			h.LogHookEvent("performance_end", event.ToolName, rawData, details)
		}
	} else if h.logger != nil {
		h.logger.Printf("END: %s [no matching start time]", event.ToolName)
	}

	return cchooks.Allow()
}

func (h *PerformanceHook) logSummary() {
	if h.logger == nil {
		return
	}

	h.logger.Println("=== Performance Monitoring Summary ===")
	h.logger.Printf("Total tools executed: %d", h.toolCount)
	h.logger.Printf("Total execution time: %v", h.totalTime)
	if h.toolCount > 0 {
		avgTime := h.totalTime / time.Duration(h.toolCount)
		h.logger.Printf("Average execution time: %v", avgTime)
	}

	// Log any incomplete operations
	if len(h.startTimes) > 0 {
		h.logger.Printf("Incomplete operations: %d", len(h.startTimes))
		for id := range h.startTimes {
			h.logger.Printf("  - %s (started but not completed)", id)
		}
	}

	h.logger.Println("=== Performance Monitoring Session Ended ===")
}
