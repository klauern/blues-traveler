// Package hooks provides built-in hook implementations for security, formatting, and auditing
package hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/brads3290/cchooks"
	"github.com/klauern/blues-traveler/internal/constants"
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
	return h.StandardRun(h.preToolUseHandler, h.postToolUseHandler)
}

// addToolSpecificDetails adds tool-specific details to the audit entry
func (h *AuditHook) addToolSpecificDetails(entry *AuditEntry, event *cchooks.PreToolUseEvent) {
	switch event.ToolName {
	case constants.ToolBash:
		h.addBashDetails(entry, event)
	case constants.ToolEdit:
		h.addEditDetails(entry, event)
	case constants.ToolWrite:
		h.addWriteDetails(entry, event)
	case constants.ToolRead:
		h.addReadDetails(entry, event)
	case constants.ToolGlob:
		h.addGlobDetails(entry, event)
	case constants.ToolGrep:
		h.addGrepDetails(entry, event)
	}
}

func (h *AuditHook) addBashDetails(entry *AuditEntry, event *cchooks.PreToolUseEvent) {
	if bash, err := event.AsBash(); err == nil {
		entry.Details["command"] = bash.Command
		entry.Details["description"] = bash.Description
	}
}

func (h *AuditHook) addEditDetails(entry *AuditEntry, event *cchooks.PreToolUseEvent) {
	if edit, err := event.AsEdit(); err == nil {
		entry.Details["file_path"] = edit.FilePath
		entry.Details["old_string_length"] = len(edit.OldString)
		entry.Details["new_string_length"] = len(edit.NewString)
	}
}

func (h *AuditHook) addWriteDetails(entry *AuditEntry, event *cchooks.PreToolUseEvent) {
	if write, err := event.AsWrite(); err == nil {
		entry.Details["file_path"] = write.FilePath
		entry.Details["content_length"] = len(write.Content)
	}
}

func (h *AuditHook) addReadDetails(entry *AuditEntry, event *cchooks.PreToolUseEvent) {
	if read, err := event.AsRead(); err == nil {
		entry.Details["file_path"] = read.FilePath
	}
}

func (h *AuditHook) addGlobDetails(entry *AuditEntry, event *cchooks.PreToolUseEvent) {
	if glob, err := event.AsGlob(); err == nil {
		entry.Details["pattern"] = glob.Pattern
	}
}

func (h *AuditHook) addGrepDetails(entry *AuditEntry, event *cchooks.PreToolUseEvent) {
	if grep, err := event.AsGrep(); err == nil {
		entry.Details["pattern"] = grep.Pattern
	}
}

func (h *AuditHook) preToolUseHandler(_ context.Context, event *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface {
	entry := AuditEntry{
		Event:    "pre_tool_use",
		ToolName: event.ToolName,
		Details:  make(map[string]interface{}),
	}

	h.addToolSpecificDetails(&entry, event)
	h.logAuditEntry(entry)

	// Also use the new detailed logging if enabled
	if h.Context().LoggingEnabled {
		rawData := make(map[string]interface{})
		rawData["tool_name"] = event.ToolName
		h.LogHookEvent("pre_tool_use", event.ToolName, rawData, entry.Details)
	}

	return cchooks.Approve()
}

func (h *AuditHook) postToolUseHandler(_ context.Context, event *cchooks.PostToolUseEvent) cchooks.PostToolUseResponseInterface {
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
