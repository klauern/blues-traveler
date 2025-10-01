package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/brads3290/cchooks"
)

// MockFileSystem implements FileSystem interface for testing
type MockFileSystem struct {
	Files    map[string][]byte
	Dirs     map[string]bool
	WriteErr error
	OpenErr  error
	StatErr  error
	mu       sync.RWMutex
}

// NewMockFileSystem creates a new mock filesystem for testing
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		Files: make(map[string][]byte),
		Dirs:  make(map[string]bool),
	}
}

// WriteFile implements FileSystem.WriteFile for testing
func (m *MockFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	if m.WriteErr != nil {
		return m.WriteErr
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	m.Dirs[dir] = true

	// Write file
	m.Files[filename] = make([]byte, len(data))
	copy(m.Files[filename], data)
	return nil
}

// OpenFile implements FileSystem.OpenFile for testing
func (m *MockFileSystem) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	if m.OpenErr != nil {
		return nil, m.OpenErr
	}

	// For testing, we can return a temporary file or a mock
	// This is a simplified implementation for testing hooks
	return os.CreateTemp("", "mock_*")
}

// Stat implements FileSystem.Stat for testing
func (m *MockFileSystem) Stat(name string) (os.FileInfo, error) {
	if m.StatErr != nil {
		return nil, m.StatErr
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, exists := m.Files[name]; exists {
		return &mockFileInfo{name: name, size: int64(len(m.Files[name]))}, nil
	}

	return nil, os.ErrNotExist
}

// (helpers removed) GetWrittenFile and HasDirectory were unused by tests.

type mockFileInfo struct {
	name string
	size int64
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return m.size }
func (m *mockFileInfo) Mode() os.FileMode  { return 0o644 }
func (m *mockFileInfo) ModTime() time.Time { return time.Now() }
func (m *mockFileInfo) IsDir() bool        { return false }
func (m *mockFileInfo) Sys() interface{}   { return nil }

// MockCommandExecutor implements CommandExecutor interface for testing
type MockCommandExecutor struct {
	Commands  []MockCommand
	Responses map[string]MockCommandResponse
	mu        sync.RWMutex
}

// MockCommand represents a command that was executed
type MockCommand struct {
	Name string
	Args []string
}

// MockCommandResponse represents the response for a mocked command
type MockCommandResponse struct {
	Output []byte
	Error  error
}

// NewMockCommandExecutor creates a new mock command executor for testing
func NewMockCommandExecutor() *MockCommandExecutor {
	return &MockCommandExecutor{
		Commands:  []MockCommand{},
		Responses: make(map[string]MockCommandResponse),
	}
}

// ExecuteCommand implements CommandExecutor.ExecuteCommand for testing
func (m *MockCommandExecutor) ExecuteCommand(name string, args ...string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Record the command
	m.Commands = append(m.Commands, MockCommand{
		Name: name,
		Args: append([]string{}, args...),
	})

	// Create command key for lookup
	key := name
	if len(args) > 0 {
		key = fmt.Sprintf("%s %s", name, args[0])
	}

	// Return response if configured
	if response, exists := m.Responses[key]; exists {
		return response.Output, response.Error
	}

	// Default success response
	return []byte("mock command output"), nil
}

// SetResponse configures a response for a specific command
func (m *MockCommandExecutor) SetResponse(command string, output []byte, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Responses[command] = MockCommandResponse{
		Output: output,
		Error:  err,
	}
}

// GetExecutedCommands returns all executed commands (used in tests)
func (m *MockCommandExecutor) GetExecutedCommands() []MockCommand {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]MockCommand, len(m.Commands))
	copy(result, m.Commands)
	return result
}

// WasCommandExecuted checks if a command was executed
func (m *MockCommandExecutor) WasCommandExecuted(name string, args ...string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, cmd := range m.Commands {
		if cmd.Name == name {
			if len(args) == 0 {
				return true
			}
			if len(cmd.Args) >= len(args) {
				match := true
				for i, arg := range args {
					if cmd.Args[i] != arg {
						match = false
						break
					}
				}
				if match {
					return true
				}
			}
		}
	}
	return false
}

// MockRunner implements a test runner for cchooks that mimics cchooks.Runner structure
type MockRunner struct {
	PreToolUse  func(context.Context, *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface
	PostToolUse func(context.Context, *cchooks.PostToolUseEvent) cchooks.PostToolUseResponseInterface
	RawHook     func(context.Context, string) *cchooks.RawResponse
	RunCalled   bool
}

// Run implements Runner.Run for testing
func (m *MockRunner) Run() {
	m.RunCalled = true
	// Don't actually read from stdin in tests
}

// MockRunnerFactory creates MockRunner instances
func MockRunnerFactory(preHook func(context.Context, *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface,
	postHook func(context.Context, *cchooks.PostToolUseEvent) cchooks.PostToolUseResponseInterface,
	rawHook func(context.Context, string) *cchooks.RawResponse,
) Runner {
	// Create a mock runner that doesn't actually read from stdin
	// This prevents the "failed to decode stdin" error in tests
	return &MockRunner{
		PreToolUse:  preHook,
		PostToolUse: postHook,
		RawHook:     rawHook,
		RunCalled:   false,
	}
}

// TestHookContext creates a context suitable for testing
func TestHookContext(settingsChecker func(string) bool) *HookContext {
	if settingsChecker == nil {
		settingsChecker = func(string) bool { return true }
	}

	return &HookContext{
		FileSystem:      NewMockFileSystem(),
		CommandExecutor: NewMockCommandExecutor(),
		RunnerFactory:   MockRunnerFactory,
		SettingsChecker: settingsChecker,
	}
}

// TestEvent helpers for creating test events

// NewMockPreToolUseEvent creates a mock PreToolUseEvent for testing
func NewMockPreToolUseEvent(toolName string) *cchooks.PreToolUseEvent {
	// This would need to be implemented based on the cchooks library structure
	// For now, returning nil as we'd need to examine the cchooks library more closely
	return nil
}

// NewMockPostToolUseEvent creates a mock PostToolUseEvent for testing
func NewMockPostToolUseEvent(toolName string) *cchooks.PostToolUseEvent {
	// This would need to be implemented based on the cchooks library structure
	// For now, returning nil as we'd need to examine the cchooks library more closely
	return nil
}
