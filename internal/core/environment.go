package core

import (
    "context"
    "os"
    "strings"

    "github.com/brads3290/cchooks"
)

// EnvironmentProvider defines how to produce environment variables for hooks
type EnvironmentProvider interface {
    // GetEnvironment returns a map of environment variables for the given event
    GetEnvironment(event string, ctxData map[string]interface{}) map[string]string
}

// claudeCodeEnvironmentProvider implements EnvironmentProvider using cchooks event data
type claudeCodeEnvironmentProvider struct{}

// NewClaudeCodeEnvironmentProvider creates a provider that extracts context from Claude Code events
func NewClaudeCodeEnvironmentProvider() EnvironmentProvider {
    return &claudeCodeEnvironmentProvider{}
}

// GetEnvironment builds a set of common environment variables from loosely typed context
// ctxData may contain: "tool_name" string, "files_changed" []string, "project_root" string, "user_prompt" string
func (p *claudeCodeEnvironmentProvider) GetEnvironment(event string, ctxData map[string]interface{}) map[string]string {
    env := map[string]string{
        "EVENT_NAME": event,
    }
    if v, ok := ctxData["tool_name"].(string); ok && v != "" {
        env["TOOL_NAME"] = v
    }
    if v, ok := ctxData["files_changed"].([]string); ok && len(v) > 0 {
        env["FILES_CHANGED"] = strings.Join(v, " ")
    }
    if v, ok := ctxData["project_root"].(string); ok && v != "" {
        env["PROJECT_ROOT"] = v
    }
    if v, ok := ctxData["user_prompt"].(string); ok && v != "" {
        env["USER_PROMPT"] = v
    }
    return env
}

// Helpers to extract context from cchooks events

// BuildPreToolUseContext extracts a minimal context map from a PreToolUseEvent
func BuildPreToolUseContext(_ context.Context, ev *cchooks.PreToolUseEvent) map[string]interface{} {
    ctx := map[string]interface{}{
        "tool_name": ev.ToolName,
    }
    if wd, err := os.Getwd(); err == nil {
        ctx["project_root"] = wd
    }
    // For PreToolUse we conservatively avoid parsing tool inputs except for Bash in other hooks.
    // Rely on PostToolUse for file-specific context.
    return ctx
}

// BuildPostToolUseContext extracts a minimal context map from a PostToolUseEvent
func BuildPostToolUseContext(_ context.Context, ev *cchooks.PostToolUseEvent) map[string]interface{} {
    ctx := map[string]interface{}{
        "tool_name": ev.ToolName,
    }
    if wd, err := os.Getwd(); err == nil {
        ctx["project_root"] = wd
    }
    // Attempt best-effort extraction using known helpers in hooks
    // Avoid referencing cchooks helpers that may not exist in this version.
    return ctx
}
