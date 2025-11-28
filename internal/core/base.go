// Package core provides the fundamental hook system interfaces, base implementations, and execution context
package core

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"

	"github.com/brads3290/cchooks"
	"github.com/klauern/blues-traveler/internal/config"
)

// Hook defines the interface that all hook implementations must satisfy
type Hook interface {
	// Key returns the unique identifier for this hook
	Key() string
	// Name returns the human-readable name for this hook
	Name() string
	// Description returns a description of what this hook does
	Description() string
	// Run executes the hook and returns any error
	Run() error
	// IsEnabled checks if this hook is enabled in the current context
	IsEnabled() bool
}

// BaseHook provides common functionality for all hooks
type BaseHook struct {
	key         string
	name        string
	description string
	context     *HookContext
}

// Key returns the hook key
func (h *BaseHook) Key() string {
	return h.key
}

// Name returns the hook name
func (h *BaseHook) Name() string {
	return h.name
}

// Description returns the hook description
func (h *BaseHook) Description() string {
	return h.description
}

// IsEnabled checks if the hook is enabled by consulting settings
func (h *BaseHook) IsEnabled() bool {
	return h.context.SettingsChecker(h.key)
}

// Context returns the hook context
func (h *BaseHook) Context() *HookContext {
	return h.context
}

// NewBaseHook creates a new BaseHook with the given metadata
func NewBaseHook(key, name, description string, ctx *HookContext) *BaseHook {
	if ctx == nil {
		ctx = DefaultHookContext()
	}
	return &BaseHook{
		key:         key,
		name:        name,
		description: description,
		context:     ctx,
	}
}

// FileSystem interface for dependency injection in testing
type FileSystem interface {
	WriteFile(filename string, data []byte, perm os.FileMode) error
	OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
	Stat(name string) (os.FileInfo, error)
}

// RealFileSystem implements FileSystem using the real filesystem
type RealFileSystem struct{}

// WriteFile writes data to a file with the specified permissions
func (fs *RealFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}

// OpenFile opens a file with the specified flags and permissions
func (fs *RealFileSystem) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, perm) // #nosec G304 - filesystem interface, paths controlled by caller
}

// Stat returns file information for the specified path
func (fs *RealFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// CommandExecutor interface for dependency injection in testing
type CommandExecutor interface {
	ExecuteCommand(name string, args ...string) ([]byte, error)
}

// RealCommandExecutor implements CommandExecutor using real system commands
type RealCommandExecutor struct{}

// ExecuteCommand executes a system command with the specified arguments and returns the combined output
// #nosec G204 - Command name is controlled by hooks, not user input; args are hook-defined
func (ce *RealCommandExecutor) ExecuteCommand(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.CombinedOutput()
}

// Runner interface allows for mocking in tests
type Runner interface {
	Run()
}

// RunnerFactory creates a Runner with the provided handlers
type RunnerFactory func(preHook func(context.Context, *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface,
	postHook func(context.Context, *cchooks.PostToolUseEvent) cchooks.PostToolUseResponseInterface,
	rawHook func(context.Context, string) *cchooks.RawResponse) Runner

// DefaultRunnerFactory creates a standard cchooks.Runner
func DefaultRunnerFactory(preHook func(context.Context, *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface,
	postHook func(context.Context, *cchooks.PostToolUseEvent) cchooks.PostToolUseResponseInterface,
	rawHook func(context.Context, string) *cchooks.RawResponse,
) Runner {
	runner := &cchooks.Runner{}
	if preHook != nil {
		runner.PreToolUse = preHook
	}
	if postHook != nil {
		runner.PostToolUse = postHook
	}
	if rawHook != nil {
		runner.Raw = rawHook
	}
	return runner
}

// HookContext provides dependencies that hooks may need
type HookContext struct {
	FileSystem      FileSystem
	CommandExecutor CommandExecutor
	RunnerFactory   RunnerFactory
	SettingsChecker func(string) bool
	LoggingEnabled  bool
	LoggingDir      string
	LoggingFormat   string
	// Platform identifies the runtime environment (e.g., Claude, Cursor)
	Platform Platform
}

// DefaultHookContext returns a context with real implementations
func DefaultHookContext() *HookContext {
	return &HookContext{
		FileSystem:      &RealFileSystem{},
		CommandExecutor: &RealCommandExecutor{},
		RunnerFactory:   DefaultRunnerFactory,
		SettingsChecker: defaultIsPluginEnabled,
		LoggingEnabled:  false,
		LoggingDir:      ".claude/hooks",
		LoggingFormat:   config.LoggingFormatJSONL,
		Platform:        DetectPlatform(),
	}
}

// defaultIsPluginEnabled is the default implementation - always returns true
// This will be replaced by the main package when registering hooks
func defaultIsPluginEnabled(_ string) bool {
	return true
}

// SetDefaultSettingsChecker allows the main package to set the real settings checker
// (removed) SetDefaultSettingsChecker unused; settings checker is injected via global context.

// LogHookEvent delegates to shared logging utility (see logging.go)
func (h *BaseHook) LogHookEvent(event string, toolName string, rawData map[string]interface{}, details map[string]interface{}) {
	if !h.context.LoggingEnabled {
		return
	}
	logHookEvent(h.context, h.key, event, toolName, rawData, details)
}

// CreateRawHandler creates a raw handler that logs all incoming JSON data when logging is enabled
func (h *BaseHook) CreateRawHandler() func(context.Context, string) *cchooks.RawResponse {
	if !h.context.LoggingEnabled {
		return nil
	}

	return func(_ context.Context, rawJSON string) *cchooks.RawResponse {
		// Parse the raw JSON to extract basic event information
		var rawEvent map[string]interface{}
		if err := json.Unmarshal([]byte(rawJSON), &rawEvent); err != nil {
			// Log parsing error but continue
			h.LogHookEvent("raw_event_parse_error", "unknown", map[string]interface{}{
				"raw_json_string": rawJSON,
				"error":           err.Error(),
			}, nil)
			return nil
		}

		// Extract basic information
		eventName, _ := rawEvent["hook_event_name"].(string)
		toolName, _ := rawEvent["tool_name"].(string)

		// Log the complete raw event data with the parsed JSON as a nested object
		h.LogHookEvent("raw_event", toolName, map[string]interface{}{
			"hook_event_name": eventName,
		}, rawEvent) // Pass the parsed JSON directly as details for readable formatting

		// Return nil to continue with normal processing
		return nil
	}
}

// StandardRun executes the hook with the provided handlers.
// Concrete hooks should call this in their Run() method.
func (h *BaseHook) StandardRun(
	preHandler func(context.Context, *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface,
	postHandler func(context.Context, *cchooks.PostToolUseEvent) cchooks.PostToolUseResponseInterface,
) error {
	if !h.IsEnabled() {
		// fmt.Printf("%s plugin disabled - skipping\n", h.Name()) // Optional: print to stdout
		return nil
	}

	runner := h.Context().RunnerFactory(preHandler, postHandler, h.CreateRawHandler())
	runner.Run()
	return nil
}

// LogError logs a standard error event
func (h *BaseHook) LogError(eventType, toolName string, err error) {
	if h.Context().LoggingEnabled {
		h.LogHookEvent(eventType, toolName, map[string]interface{}{"error": err.Error()}, nil)
	}
}

// LogApproval logs a standard approval event
func (h *BaseHook) LogApproval(eventType, toolName string, details map[string]interface{}) {
	if h.Context().LoggingEnabled {
		h.LogHookEvent(eventType, toolName, details, nil)
	}
}

// LogBlock logs a standard block event
func (h *BaseHook) LogBlock(eventType, toolName string, details map[string]interface{}) {
	if h.Context().LoggingEnabled {
		h.LogHookEvent(eventType, toolName, details, nil)
	}
}

// PreToolUseHandler interface for hooks that handle pre-tool-use events
type PreToolUseHandler interface {
	PreToolUse(context.Context, *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface
}

// PostToolUseHandler interface for hooks that handle post-tool-use events
type PostToolUseHandler interface {
	PostToolUse(context.Context, *cchooks.PostToolUseEvent) cchooks.PostToolUseResponseInterface
}
