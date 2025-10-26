package hooks

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/brads3290/cchooks"
	"github.com/klauern/blues-traveler/internal/constants"
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
	defer func() {
		if h.logFile != nil {
			if err := h.logFile.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "debug log close error: %v\n", err)
			}
		}
	}()
	runner := h.Context().RunnerFactory(h.preToolUseHandler, h.postToolUseHandler, h.CreateRawHandler())
	fmt.Println("Debug hook started - logging to .claude/hooks/debug.log")
	runner.Run()
	return nil
}

func (h *DebugHook) ensureLogger() {
	if h.logger != nil {
		return
	}

	// Ensure directory exists
	logPath := ".claude/hooks/debug.log"
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0o750); err != nil {
		// Fallback: leave logger nil, but surface the error
		fmt.Fprintf(os.Stderr, "failed to create debug log dir %s: %v\n", logDir, err)
		return
	}

	var err error
	h.logFile, err = h.Context().FileSystem.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		// Fallback: leave logger nil
		fmt.Fprintf(os.Stderr, "failed to open debug log file %s: %v\n", logPath, err)
		return
	}
	h.logger = log.New(h.logFile, "", log.LstdFlags)
}

func (h *DebugHook) preToolUseHandler(_ context.Context, event *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface {
	h.ensureLogger()
	if h.logger != nil {
		h.logger.Printf("PRE-TOOL: %s", event.ToolName)
	}

	// Prepare logging data
	details := make(map[string]interface{})
	rawData := make(map[string]interface{})
	if h.Context().LoggingEnabled {
		rawData["tool_name"] = event.ToolName
	}

	// Log tool-specific details
	h.logPreToolDetails(event, details)

	// Log detailed event data if logging is enabled
	if h.Context().LoggingEnabled {
		h.LogHookEvent("pre_tool_use", event.ToolName, rawData, details)
	}

	return cchooks.Approve()
}

// logPreToolDetails logs tool-specific details for pre-tool events
func (h *DebugHook) logPreToolDetails(event *cchooks.PreToolUseEvent, details map[string]interface{}) {
	switch event.ToolName {
	case constants.ToolBash:
		h.logBashDetails(event, details)
	case constants.ToolEdit:
		h.logEditDetails(event, details)
	case constants.ToolWrite:
		h.logWriteDetails(event, details)
	case constants.ToolRead:
		h.logReadDetails(event, details)
	case constants.ToolGlob:
		h.logGlobDetails(event, details)
	case constants.ToolGrep:
		h.logGrepDetails(event, details)
	}
}

// logBashDetails logs details for Bash tool events
func (h *DebugHook) logBashDetails(event *cchooks.PreToolUseEvent, details map[string]interface{}) {
	bash, err := event.AsBash()
	if err != nil {
		return
	}
	if h.logger != nil {
		h.logger.Printf("  Command: %s", bash.Command)
	}
	details["command"] = bash.Command
	details["description"] = bash.Description
}

// logEditDetails logs details for Edit tool events
func (h *DebugHook) logEditDetails(event *cchooks.PreToolUseEvent, details map[string]interface{}) {
	edit, err := event.AsEdit()
	if err != nil {
		return
	}
	if h.logger != nil {
		h.logger.Printf("  File: %s", edit.FilePath)
	}
	details["file_path"] = edit.FilePath
	details["old_string_length"] = len(edit.OldString)
	details["new_string_length"] = len(edit.NewString)
}

// logWriteDetails logs details for Write tool events
func (h *DebugHook) logWriteDetails(event *cchooks.PreToolUseEvent, details map[string]interface{}) {
	write, err := event.AsWrite()
	if err != nil {
		return
	}
	if h.logger != nil {
		h.logger.Printf("  File: %s", write.FilePath)
	}
	details["file_path"] = write.FilePath
	details["content_length"] = len(write.Content)
}

// logReadDetails logs details for Read tool events
func (h *DebugHook) logReadDetails(event *cchooks.PreToolUseEvent, details map[string]interface{}) {
	read, err := event.AsRead()
	if err != nil {
		return
	}
	if h.logger != nil {
		h.logger.Printf("  File: %s", read.FilePath)
	}
	details["file_path"] = read.FilePath
}

// logGlobDetails logs details for Glob tool events
func (h *DebugHook) logGlobDetails(event *cchooks.PreToolUseEvent, details map[string]interface{}) {
	glob, err := event.AsGlob()
	if err != nil {
		return
	}
	details["pattern"] = glob.Pattern
}

// logGrepDetails logs details for Grep tool events
func (h *DebugHook) logGrepDetails(event *cchooks.PreToolUseEvent, details map[string]interface{}) {
	grep, err := event.AsGrep()
	if err != nil {
		return
	}
	details["pattern"] = grep.Pattern
}

func (h *DebugHook) postToolUseHandler(_ context.Context, event *cchooks.PostToolUseEvent) cchooks.PostToolUseResponseInterface {
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
