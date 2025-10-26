package hooks

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/brads3290/cchooks"
	"github.com/klauern/blues-traveler/internal/core"
)

// DebugHook implements debug logging logic
type DebugHook struct {
	*core.BaseHook
	logger  *log.Logger
	logFile *os.File
}

// NewDebugHook creates a new debug hook instance
func NewDebugHook(ctx *core.HookContext) core.Hook {
	base := core.NewBaseHook("debug", "Debug Hook", "Logs all tool usage for debugging purposes", ctx)
	return &DebugHook{BaseHook: base}
}

// Run executes the debug hook.
func (h *DebugHook) Run() error {
	if !h.IsEnabled() {
		fmt.Println("Debug plugin disabled - skipping")
		return nil
	}
	h.ensureLogger()
	if h.logger == nil {
		return fmt.Errorf("failed to initialize logger")
	}
	runner := h.Context().RunnerFactory(h.preToolUseHandler, h.postToolUseHandler, h.CreateRawHandler())
	fmt.Println("Debug hook started - logging to blues-traveler.log")
	runner.Run()
	return nil
}

func (h *DebugHook) ensureLogger() {
	if h.logger != nil {
		return
	}
	var err error
	h.logFile, err = h.Context().FileSystem.OpenFile(".claude/hooks/debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		// Fallback: leave logger nil
		return
	}
	h.logger = log.New(h.logFile, "", log.LstdFlags)
}

func (h *DebugHook) preToolUseHandler(ctx context.Context, event *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface {
	h.ensureLogger()
	if h.logger != nil {
		h.logger.Printf("PRE-TOOL: %s", event.ToolName)
	}

	// Prepare detailed logging if enabled
	details := make(map[string]interface{})
	rawData := make(map[string]interface{})

	// Capture raw data for detailed logging
	if h.Context().LoggingEnabled {
		rawData["tool_name"] = event.ToolName
		// Add more raw data as available from the event
	}

	// Log specific tool details
	switch event.ToolName {
	case ToolBash:
		if bash, err := event.AsBash(); err == nil {
			h.logger.Printf("  Command: %s", bash.Command)
			details["command"] = bash.Command
			details["description"] = bash.Description
		}
	case ToolEdit:
		if edit, err := event.AsEdit(); err == nil {
			h.logger.Printf("  File: %s", edit.FilePath)
			details["file_path"] = edit.FilePath
			details["old_string_length"] = len(edit.OldString)
			details["new_string_length"] = len(edit.NewString)
		}
	case ToolWrite:
		if write, err := event.AsWrite(); err == nil {
			h.logger.Printf("  File: %s", write.FilePath)
			details["file_path"] = write.FilePath
			details["content_length"] = len(write.Content)
		}
	case "Read":
		if read, err := event.AsRead(); err == nil {
			h.logger.Printf("  File: %s", read.FilePath)
			details["file_path"] = read.FilePath
		}
	case "Glob":
		if glob, err := event.AsGlob(); err == nil {
			details["pattern"] = glob.Pattern
		}
	case "Grep":
		if grep, err := event.AsGrep(); err == nil {
			details["pattern"] = grep.Pattern
		}
	}

	// Log detailed event data if logging is enabled
	h.LogHookEvent("pre_tool_use", event.ToolName, rawData, details)

	return cchooks.Approve()
}

func (h *DebugHook) postToolUseHandler(ctx context.Context, event *cchooks.PostToolUseEvent) cchooks.PostToolUseResponseInterface {
	h.ensureLogger()
	if h.logger != nil {
		h.logger.Printf("POST-TOOL: %s", event.ToolName)
	}

	// Log detailed event data if logging is enabled
	if h.Context().LoggingEnabled {
		details := make(map[string]interface{})
		rawData := make(map[string]interface{})

		rawData["tool_name"] = event.ToolName
		// Add any available post-tool event data

		h.LogHookEvent("post_tool_use", event.ToolName, rawData, details)
	}

	return cchooks.Allow()
}
