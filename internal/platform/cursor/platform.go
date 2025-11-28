package cursor

import (
	"strings"

	"github.com/klauern/blues-traveler/internal/core"
)

// Action represents a Cursor hook action name.
type Action string

// Supported Cursor hook actions.
const (
	ActionBeforeToolUse        Action = "BeforeToolUse"
	ActionBeforeShellExecution Action = "BeforeShellExecution"
	ActionBeforeFileEdit       Action = "BeforeFileEdit"
	ActionBeforeFileWrite      Action = "BeforeFileWrite"
	ActionBeforeReadFile       Action = "BeforeReadFile"
	ActionBeforeMCPExecution   Action = "BeforeMCPExecution"
	ActionAfterToolUse         Action = "AfterToolUse"
	ActionAfterShellExecution  Action = "AfterShellExecution"
	ActionAfterFileEdit        Action = "AfterFileEdit"
	ActionAfterFileWrite       Action = "AfterFileWrite"
	ActionAfterMCPExecution    Action = "AfterMCPExecution"
	ActionOnNotification       Action = "OnNotification"
	ActionOnPermissionRequest  Action = "OnPermissionRequest"
	ActionOnStop               Action = "OnStop"
	ActionOnAgentStop          Action = "OnAgentStop"
	ActionAfterResponse        Action = "AfterResponse"
	ActionAfterAgentResponse   Action = "AfterAgentResponse"
	ActionOnPromptSubmit       Action = "OnPromptSubmit"
	ActionBeforePrompt         Action = "BeforePrompt"
	ActionOnUserInput          Action = "OnUserInput"
	ActionBeforeSubmitPrompt   Action = "BeforeSubmitPrompt"
	ActionOnSubagentStop       Action = "OnSubagentStop"
	ActionAfterSubagent        Action = "AfterSubagent"
	ActionOnTaskComplete       Action = "OnTaskComplete"
	ActionBeforeCompact        Action = "BeforeCompact"
	ActionOnCompact            Action = "OnCompact"
	ActionOnSessionStart       Action = "OnSessionStart"
	ActionOnStart              Action = "OnStart"
	ActionOnSessionBegin       Action = "OnSessionBegin"
	ActionOnSessionEnd         Action = "OnSessionEnd"
	ActionOnEnd                Action = "OnEnd"
	ActionOnSessionClose       Action = "OnSessionClose"
	ActionStop                 Action = "Stop"
)

// EventMapping describes the Claude Code event metadata for a Cursor action.
type EventMapping struct {
	Event     core.EventType
	Supported bool
}

var actionEventMap = map[Action]EventMapping{
	// PreToolUse mappings
	ActionBeforeToolUse:        {Event: core.PreToolUseEvent, Supported: true},
	ActionBeforeShellExecution: {Event: core.PreToolUseEvent, Supported: true},
	ActionBeforeFileEdit:       {Event: core.PreToolUseEvent, Supported: true},
	ActionBeforeFileWrite:      {Event: core.PreToolUseEvent, Supported: true},
	ActionBeforeReadFile:       {Event: core.PreToolUseEvent, Supported: true},
	ActionBeforeMCPExecution:   {Event: core.PreToolUseEvent, Supported: true},

	// PostToolUse mappings
	ActionAfterToolUse:        {Event: core.PostToolUseEvent, Supported: true},
	ActionAfterShellExecution: {Event: core.PostToolUseEvent, Supported: true},
	ActionAfterFileEdit:       {Event: core.PostToolUseEvent, Supported: true},
	ActionAfterFileWrite:      {Event: core.PostToolUseEvent, Supported: true},
	ActionAfterMCPExecution:   {Event: core.PostToolUseEvent, Supported: true},

	// Notification mappings
	ActionOnNotification:      {Event: core.NotificationEvent, Supported: true},
	ActionOnPermissionRequest: {Event: core.NotificationEvent, Supported: true},

	// Stop event mappings
	ActionOnStop:             {Event: core.StopEvent, Supported: true},
	ActionOnAgentStop:        {Event: core.StopEvent, Supported: true},
	ActionAfterResponse:      {Event: core.StopEvent, Supported: true},
	ActionAfterAgentResponse: {Event: core.StopEvent, Supported: true},
	ActionStop:               {Event: core.StopEvent, Supported: true},

	// UserPromptSubmit mappings
	ActionOnPromptSubmit:     {Event: core.UserPromptSubmitEvent, Supported: false},
	ActionBeforePrompt:       {Event: core.UserPromptSubmitEvent, Supported: false},
	ActionOnUserInput:        {Event: core.UserPromptSubmitEvent, Supported: false},
	ActionBeforeSubmitPrompt: {Event: core.UserPromptSubmitEvent, Supported: false},

	// SubagentStop mappings
	ActionOnSubagentStop: {Event: core.SubagentStopEvent, Supported: true},
	ActionAfterSubagent:  {Event: core.SubagentStopEvent, Supported: true},
	ActionOnTaskComplete: {Event: core.SubagentStopEvent, Supported: true},

	// PreCompact mappings
	ActionBeforeCompact: {Event: core.PreCompactEvent, Supported: true},
	ActionOnCompact:     {Event: core.PreCompactEvent, Supported: true},

	// SessionStart mappings
	ActionOnSessionStart: {Event: core.SessionStartEvent, Supported: true},
	ActionOnStart:        {Event: core.SessionStartEvent, Supported: true},
	ActionOnSessionBegin: {Event: core.SessionStartEvent, Supported: true},

	// SessionEnd mappings
	ActionOnSessionEnd:   {Event: core.SessionEndEvent, Supported: true},
	ActionOnEnd:          {Event: core.SessionEndEvent, Supported: true},
	ActionOnSessionClose: {Event: core.SessionEndEvent, Supported: true},
}

// eventToCursorActions builds reverse map from core events to Cursor actions
var eventToCursorActions = func() map[core.EventType][]string {
	m := make(map[core.EventType][]string)
	for action, mapping := range actionEventMap {
		m[mapping.Event] = append(m[mapping.Event], string(action))
	}
	return m
}()

// EventForAction returns the Claude Code event metadata for the given Cursor action.
// The boolean return value reports whether the action is recognized.
func EventForAction(action Action) (EventMapping, bool) {
	mapping, ok := actionEventMap[action]
	return mapping, ok
}

// NormalizeAction normalizes a Cursor action string to the canonical Action form.
func NormalizeAction(action string) (Action, bool) {
	trimmed := strings.TrimSpace(action)
	if trimmed == "" {
		return "", false
	}

	candidate := Action(trimmed)
	if _, ok := actionEventMap[candidate]; ok {
		return candidate, true
	}

	// Support lowerCamelCase variants from Cursor documentation/configuration.
	upperFirst := toUpperFirst(trimmed)
	candidate = Action(upperFirst)
	if _, ok := actionEventMap[candidate]; ok {
		return candidate, true
	}

	return "", false
}

// EventAliases returns the Cursor IDE event aliases associated with the provided core event type.
func EventAliases(event core.EventType) []string {
	aliases, ok := eventToCursorActions[event]
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
	// Try direct canonical event name match first
	canonicalEvents := []core.EventType{
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

	for _, event := range canonicalEvents {
		if string(event) == name {
			return event, true
		}
	}

	// Try as a Cursor action
	action, ok := NormalizeAction(name)
	if !ok {
		return "", false
	}

	mapping, ok := EventForAction(action)
	if !ok {
		return "", false
	}

	return mapping.Event, true
}

func toUpperFirst(s string) string {
	if s == "" {
		return s
	}

	if len(s) == 1 {
		return strings.ToUpper(s)
	}

	return strings.ToUpper(s[:1]) + s[1:]
}
