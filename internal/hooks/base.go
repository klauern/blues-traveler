package hooks

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"

	"github.com/brads3290/cchooks"
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

func (fs *RealFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}

func (fs *RealFileSystem) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, perm)
}

func (fs *RealFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// CommandExecutor interface for dependency injection in testing
type CommandExecutor interface {
	ExecuteCommand(name string, args ...string) ([]byte, error)
}

// RealCommandExecutor implements CommandExecutor using real system commands
type RealCommandExecutor struct{}

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
	}
}

// defaultIsPluginEnabled is the default implementation - always returns true
// This will be replaced by the main package when registering hooks
func defaultIsPluginEnabled(pluginKey string) bool {
	return true
}

// SetDefaultSettingsChecker allows the main package to set the real settings checker
func SetDefaultSettingsChecker(checker func(string) bool) {
	if checker != nil {
		// Update the default context's checker
		defaultContext := DefaultHookContext()
		defaultContext.SettingsChecker = checker
	}
}

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

	return func(ctx context.Context, rawJSON string) *cchooks.RawResponse {
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
