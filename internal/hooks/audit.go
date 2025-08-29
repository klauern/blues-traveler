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

// AuditHook implements comprehensive audit logging
type AuditHook struct {
	*core.BaseHook
}

// AuditEntry represents an audit log entry
type AuditEntry struct {
	Timestamp string                 `json:"timestamp"`
	Event     string                 `json:"event"`
	ToolName  string                 `json:"tool_name"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// NewAuditHook creates a new audit hook instance
func NewAuditHook(ctx *core.HookContext) core.Hook {
	base := core.NewBaseHook("audit", "Audit Hook", "Comprehensive audit logging with JSON output", ctx)
	return &AuditHook{BaseHook: base}
}

// Run executes the audit hook.
func (h *AuditHook) Run() error {
	if !h.IsEnabled() {
		fmt.Println("Audit plugin disabled - skipping")
		return nil
	}

	runner := h.Context().RunnerFactory(h.preToolUseHandler, h.postToolUseHandler, h.CreateRawHandler())
	runner.Run()
	return nil
}

func (h *AuditHook) preToolUseHandler(ctx context.Context, event *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface {
	entry := AuditEntry{
		Event:    "pre_tool_use",
		ToolName: event.ToolName,
		Details:  make(map[string]interface{}),
	}

	// Add tool-specific details
	switch event.ToolName {
	case "Bash":
		if bash, err := event.AsBash(); err == nil {
			entry.Details["command"] = bash.Command
			entry.Details["description"] = bash.Description
		}
	case "Edit":
		if edit, err := event.AsEdit(); err == nil {
			entry.Details["file_path"] = edit.FilePath
			entry.Details["old_string_length"] = len(edit.OldString)
			entry.Details["new_string_length"] = len(edit.NewString)
		}
	case "Write":
		if write, err := event.AsWrite(); err == nil {
			entry.Details["file_path"] = write.FilePath
			entry.Details["content_length"] = len(write.Content)
		}
	case "Read":
		if read, err := event.AsRead(); err == nil {
			entry.Details["file_path"] = read.FilePath
		}
	case "Glob":
		if glob, err := event.AsGlob(); err == nil {
			entry.Details["pattern"] = glob.Pattern
		}
	case "Grep":
		if grep, err := event.AsGrep(); err == nil {
			entry.Details["pattern"] = grep.Pattern
		}
	}

	h.logAuditEntry(entry)

	// Also use the new detailed logging if enabled
	if h.Context().LoggingEnabled {
		rawData := make(map[string]interface{})
		rawData["tool_name"] = event.ToolName
		h.LogHookEvent("pre_tool_use", event.ToolName, rawData, entry.Details)
	}

	return cchooks.Approve()
}

func (h *AuditHook) postToolUseHandler(ctx context.Context, event *cchooks.PostToolUseEvent) cchooks.PostToolUseResponseInterface {
	entry := AuditEntry{
		Event:    "post_tool_use",
		ToolName: event.ToolName,
		Details:  make(map[string]interface{}),
	}

	h.logAuditEntry(entry)

	// Also use the new detailed logging if enabled
	if h.Context().LoggingEnabled {
		rawData := make(map[string]interface{})
		rawData["tool_name"] = event.ToolName
		h.LogHookEvent("post_tool_use", event.ToolName, rawData, entry.Details)
	}

	return cchooks.Allow()
}

func (h *AuditHook) logAuditEntry(entry AuditEntry) {
	entry.Timestamp = time.Now().Format(time.RFC3339)

	jsonData, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal audit entry: %v\n", err)
		return
	}

	fmt.Println(string(jsonData))
}
