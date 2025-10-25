package claude

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/klauern/blues-traveler/internal/core"
	"github.com/klauern/blues-traveler/internal/platform"
)

// ClaudeCodePlatform implements Platform for Claude Code
type ClaudeCodePlatform struct{}

// New creates a new Claude Code platform instance
func New() platform.Platform {
	return &ClaudeCodePlatform{}
}

// Type returns the platform type
func (p *ClaudeCodePlatform) Type() platform.Type {
	return platform.ClaudeCode
}

// Name returns the human-readable platform name
func (p *ClaudeCodePlatform) Name() string {
	return "Claude Code"
}

// ConfigPath returns the path to the Claude Code settings.json file
func (p *ClaudeCodePlatform) ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(home, ".claude", "settings.json"), nil
}

// SupportsEvent returns true if Claude Code supports the given event
func (p *ClaudeCodePlatform) SupportsEvent(event core.EventType) bool {
	// Claude Code supports all core events
	switch event {
	case core.PreToolUseEvent,
		core.PostToolUseEvent,
		core.UserPromptSubmitEvent,
		core.NotificationEvent,
		core.StopEvent,
		core.SubagentStopEvent,
		core.PreCompactEvent,
		core.SessionStartEvent,
		core.SessionEndEvent:
		return true
	default:
		return false
	}
}

// MapEventFromGeneric maps a generic event type to Claude Code event name
func (p *ClaudeCodePlatform) MapEventFromGeneric(event core.EventType) []string {
	// For Claude Code, the mapping is 1:1
	eventName := string(event)
	if p.SupportsEvent(event) {
		return []string{eventName}
	}
	return nil
}

// MapEventToGeneric maps a Claude Code event name to generic event type
func (p *ClaudeCodePlatform) MapEventToGeneric(platformEvent string) (core.EventType, bool) {
	event := core.EventType(platformEvent)
	return event, p.SupportsEvent(event)
}

// ValidateEventName returns true if the event name is valid for Claude Code
func (p *ClaudeCodePlatform) ValidateEventName(eventName string) bool {
	return p.SupportsEvent(core.EventType(eventName))
}

// AllEvents returns all events supported by Claude Code
func (p *ClaudeCodePlatform) AllEvents() []platform.PlatformEvent {
	events := core.AllClaudeCodeEvents()
	platformEvents := make([]platform.PlatformEvent, len(events))

	for i, e := range events {
		platformEvents[i] = platform.PlatformEvent{
			Name:           e.Name,
			Description:    e.Description,
			GenericEvent:   e.Type,
			RequiresStdio:  false, // Claude Code uses env vars, not stdio
			SupportsFilter: true,  // Claude Code supports regex matchers in config
		}
	}

	return platformEvents
}
