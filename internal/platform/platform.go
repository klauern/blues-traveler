package platform

import "github.com/klauern/blues-traveler/internal/core"

// Type represents the platform/IDE type
type Type string

const (
	ClaudeCode Type = "claudecode"
	Cursor     Type = "cursor"
)

// Platform represents an AI IDE platform that supports hooks
type Platform interface {
	// Type returns the platform type
	Type() Type

	// Name returns the human-readable platform name
	Name() string

	// ConfigPath returns the path to the hooks configuration file
	ConfigPath() (string, error)

	// SupportsEvent returns true if the platform supports the given event
	SupportsEvent(event core.EventType) bool

	// MapEventFromGeneric maps a generic event type to platform-specific event name(s)
	// Returns multiple event names if the generic event maps to multiple platform events
	MapEventFromGeneric(event core.EventType) []string

	// MapEventToGeneric maps a platform-specific event name to generic event type
	MapEventToGeneric(platformEvent string) (core.EventType, bool)

	// ValidateEventName returns true if the event name is valid for this platform
	ValidateEventName(eventName string) bool

	// AllEvents returns all events supported by this platform
	AllEvents() []PlatformEvent
}

// PlatformEvent represents a platform-specific hook event
type PlatformEvent struct {
	Name           string
	Description    string
	GenericEvent   core.EventType
	RequiresStdio  bool // true if event uses stdin/stdout protocol
	SupportsFilter bool // true if platform supports config-level filtering
}

// Detector provides platform auto-detection
type Detector interface {
	Detect() (Platform, error)
	DetectType() (Type, error)
}

// Factory creates Platform instances
type Factory interface {
	Create(t Type) Platform
}
