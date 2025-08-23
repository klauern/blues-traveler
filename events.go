package main

// EventType represents a Claude Code hook event
type EventType string

// All supported Claude Code hook events
const (
	PreToolUseEvent       EventType = "PreToolUse"
	PostToolUseEvent      EventType = "PostToolUse"
	UserPromptSubmitEvent EventType = "UserPromptSubmit"
	NotificationEvent     EventType = "Notification"
	StopEvent             EventType = "Stop"
	SubagentStopEvent     EventType = "SubagentStop"
	PreCompactEvent       EventType = "PreCompact"
	SessionStartEvent     EventType = "SessionStart"
	SessionEndEvent       EventType = "SessionEnd"
)

// ClaudeCodeEvent represents a Claude Code hook event type with metadata
type ClaudeCodeEvent struct {
	Type               EventType
	Name               string
	Description        string
	SupportedByCCHooks bool
}

// AllClaudeCodeEvents returns all available Claude Code hook events
func AllClaudeCodeEvents() []ClaudeCodeEvent {
	return []ClaudeCodeEvent{
		{
			Type:               PreToolUseEvent,
			Name:               string(PreToolUseEvent),
			Description:        "Runs after Claude creates tool parameters and before processing the tool call",
			SupportedByCCHooks: true,
		},
		{
			Type:               PostToolUseEvent,
			Name:               string(PostToolUseEvent),
			Description:        "Runs immediately after a tool completes successfully",
			SupportedByCCHooks: true,
		},
		{
			Type:               NotificationEvent,
			Name:               string(NotificationEvent),
			Description:        "Runs when Claude needs permission to use a tool or when input has been idle for 60 seconds",
			SupportedByCCHooks: true,
		},
		{
			Type:               StopEvent,
			Name:               string(StopEvent),
			Description:        "Runs when the main Claude Code agent has finished responding",
			SupportedByCCHooks: true,
		},
		{
			Type:               UserPromptSubmitEvent,
			Name:               string(UserPromptSubmitEvent),
			Description:        "Runs when the user submits a prompt, before Claude processes it",
			SupportedByCCHooks: false,
		},
		{
			Type:               SubagentStopEvent,
			Name:               string(SubagentStopEvent),
			Description:        "Runs when a Claude Code subagent (Task tool call) has finished responding",
			SupportedByCCHooks: false,
		},
		{
			Type:               PreCompactEvent,
			Name:               string(PreCompactEvent),
			Description:        "Runs before Claude Code is about to run a compact operation",
			SupportedByCCHooks: false,
		},
		{
			Type:               SessionStartEvent,
			Name:               string(SessionStartEvent),
			Description:        "Runs when Claude Code starts a new session or resumes an existing session",
			SupportedByCCHooks: false,
		},
		{
			Type:               SessionEndEvent,
			Name:               string(SessionEndEvent),
			Description:        "Runs when a Claude Code session ends",
			SupportedByCCHooks: false,
		},
	}
}

// ValidEventTypes returns a slice of all valid event type names
func ValidEventTypes() []string {
	events := AllClaudeCodeEvents()
	names := make([]string, len(events))
	for i, event := range events {
		names[i] = event.Name
	}
	return names
}

// IsValidEventType checks if an event type string is valid
func IsValidEventType(eventType string) bool {
	for _, event := range AllClaudeCodeEvents() {
		if event.Name == eventType {
			return true
		}
	}
	return false
}
