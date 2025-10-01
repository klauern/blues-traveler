package cursor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/klauern/blues-traveler/internal/core"
	"github.com/klauern/blues-traveler/internal/platform"
)

// validCursorEvents is a map for efficient event name validation
var validCursorEvents = map[string]bool{
	BeforeShellExecution: true,
	BeforeMCPExecution:   true,
	AfterFileEdit:        true,
	BeforeReadFile:       true,
	BeforeSubmitPrompt:   true,
	Stop:                 true,
}

// CursorPlatform implements Platform for Cursor IDE
type CursorPlatform struct{}

// New creates a new Cursor platform instance
func New() platform.Platform {
	return &CursorPlatform{}
}

// Type returns the platform type
func (p *CursorPlatform) Type() platform.Type {
	return platform.Cursor
}

// Name returns the human-readable platform name
func (p *CursorPlatform) Name() string {
	return "Cursor"
}

// ConfigPath returns the path to the Cursor hooks.json file
func (p *CursorPlatform) ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(home, ".cursor", "hooks.json"), nil
}

// SupportsEvent returns true if Cursor supports the given generic event
func (p *CursorPlatform) SupportsEvent(event core.EventType) bool {
	events := p.MapEventFromGeneric(event)
	return len(events) > 0
}

// MapEventFromGeneric maps a generic event type to Cursor-specific event name(s)
func (p *CursorPlatform) MapEventFromGeneric(event core.EventType) []string {
	switch event {
	case core.PreToolUseEvent:
		// PreToolUse maps to shell, MCP, and file read events
		return []string{BeforeShellExecution, BeforeMCPExecution, BeforeReadFile}
	case core.PostToolUseEvent:
		// PostToolUse only maps to file edits in Cursor
		return []string{AfterFileEdit}
	case core.UserPromptSubmitEvent:
		return []string{BeforeSubmitPrompt}
	case core.StopEvent:
		return []string{Stop}
	default:
		// Cursor doesn't support: Notification, SubagentStop, PreCompact, SessionStart, SessionEnd
		return nil
	}
}

// MapEventToGeneric maps a Cursor-specific event name to generic event type
func (p *CursorPlatform) MapEventToGeneric(platformEvent string) (core.EventType, bool) {
	switch platformEvent {
	case BeforeShellExecution, BeforeMCPExecution, BeforeReadFile:
		return core.PreToolUseEvent, true
	case AfterFileEdit:
		return core.PostToolUseEvent, true
	case BeforeSubmitPrompt:
		return core.UserPromptSubmitEvent, true
	case Stop:
		return core.StopEvent, true
	default:
		return "", false
	}
}

// ValidateEventName returns true if the event name is valid for Cursor
func (p *CursorPlatform) ValidateEventName(eventName string) bool {
	return validCursorEvents[eventName]
}

// AllEvents returns all events supported by Cursor
func (p *CursorPlatform) AllEvents() []platform.PlatformEvent {
	return []platform.PlatformEvent{
		{
			Name:           BeforeShellExecution,
			Description:    "Before shell command execution",
			GenericEvent:   core.PreToolUseEvent,
			RequiresStdio:  true,
			SupportsFilter: false,
		},
		{
			Name:           BeforeMCPExecution,
			Description:    "Before MCP tool execution",
			GenericEvent:   core.PreToolUseEvent,
			RequiresStdio:  true,
			SupportsFilter: false,
		},
		{
			Name:           AfterFileEdit,
			Description:    "After file edit operation",
			GenericEvent:   core.PostToolUseEvent,
			RequiresStdio:  true,
			SupportsFilter: false,
		},
		{
			Name:           BeforeReadFile,
			Description:    "Before agent reads a file (access control)",
			GenericEvent:   core.PreToolUseEvent,
			RequiresStdio:  true,
			SupportsFilter: false,
		},
		{
			Name:           BeforeSubmitPrompt,
			Description:    "Before user prompt is submitted",
			GenericEvent:   core.UserPromptSubmitEvent,
			RequiresStdio:  true,
			SupportsFilter: false,
		},
		{
			Name:           Stop,
			Description:    "When agent loop ends",
			GenericEvent:   core.StopEvent,
			RequiresStdio:  true,
			SupportsFilter: false,
		},
	}
}

// LoadConfig loads the Cursor hooks configuration from disk
func (p *CursorPlatform) LoadConfig() (*Config, error) {
	configPath, err := p.ConfigPath()
	if err != nil {
		return nil, err
	}

	// If config doesn't exist, return new empty config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return NewConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// SaveConfig saves the Cursor hooks configuration to disk
func (p *CursorPlatform) SaveConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	configPath, err := p.ConfigPath()
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
