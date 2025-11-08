package cursor

import "github.com/klauern/blues-traveler/internal/core"

var canonicalEvents = []core.EventType{
	core.PreToolUseEvent,
	core.PostToolUseEvent,
	core.NotificationEvent,
	core.StopEvent,
	core.UserPromptSubmitEvent,
	core.SubagentStopEvent,
	core.PreCompactEvent,
	core.SessionStartEvent,
	core.SessionEndEvent,
}

var eventToCursor = map[core.EventType][]string{
	core.PreToolUseEvent: {
		"BeforeToolUse",
		"BeforeShellExecution",
		"BeforeFileEdit",
		"BeforeFileWrite",
		"BeforeReadFile",
	},
	core.PostToolUseEvent: {
		"AfterToolUse",
		"AfterShellExecution",
		"AfterFileEdit",
		"AfterFileWrite",
	},
	core.NotificationEvent: {
		"OnNotification",
		"OnPermissionRequest",
	},
	core.StopEvent: {
		"OnStop",
		"OnAgentStop",
		"AfterResponse",
	},
	core.UserPromptSubmitEvent: {
		"OnPromptSubmit",
		"BeforePrompt",
		"OnUserInput",
	},
	core.SubagentStopEvent: {
		"OnSubagentStop",
		"AfterSubagent",
		"OnTaskComplete",
	},
	core.PreCompactEvent: {
		"BeforeCompact",
		"OnCompact",
	},
	core.SessionStartEvent: {
		"OnSessionStart",
		"OnStart",
		"OnSessionBegin",
	},
	core.SessionEndEvent: {
		"OnSessionEnd",
		"OnEnd",
		"OnSessionClose",
	},
}

var cursorToEvent = func() map[string]core.EventType {
	m := make(map[string]core.EventType)
	for event, aliases := range eventToCursor {
		for _, alias := range aliases {
			m[alias] = event
		}
	}
	return m
}()

// EventAliases returns the Cursor IDE event aliases associated with the provided core event type.
func EventAliases(event core.EventType) []string {
	aliases, ok := eventToCursor[event]
	if !ok {
		return nil
	}
	result := make([]string, len(aliases))
	copy(result, aliases)
	return result
}

// ResolveCursorEvent resolves a Cursor IDE event name to its canonical core event type.
// It returns the resolved event and true when a mapping exists, otherwise returns false.
func ResolveCursorEvent(name string) (core.EventType, bool) {
	for _, event := range canonicalEvents {
		if string(event) == name {
			return event, true
		}
	}
	resolved, ok := cursorToEvent[name]
	return resolved, ok
}
