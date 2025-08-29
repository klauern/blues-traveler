package hooks

import (
	"testing"

	"github.com/klauern/blues-traveler/internal/core"
)

func TestFormatHook(t *testing.T) {
	ctx := core.TestHookContext(nil)
	hook := NewFormatHook(ctx)

	// Test basic properties
	if hook.Key() != "format" {
		t.Errorf("Expected key 'format', got '%s'", hook.Key())
	}

	if hook.Name() != "Format Hook" {
		t.Errorf("Expected name 'Format Hook', got '%s'", hook.Name())
	}

	// Test that hook is enabled by default
	if !hook.IsEnabled() {
		t.Error("Expected hook to be enabled by default")
	}

	// Test running the hook (should not error)
	err := hook.Run()
	if err != nil {
		t.Errorf("Hook run failed: %v", err)
	}
}

func TestFormatHookGoFile(t *testing.T) {
	mockCmd := core.NewMockCommandExecutor()
	ctx := &core.HookContext{
		FileSystem:      core.NewMockFileSystem(),
		CommandExecutor: mockCmd,
		RunnerFactory:   core.MockRunnerFactory,
		SettingsChecker: func(string) bool { return true },
	}

	hook := NewFormatHook(ctx).(*FormatHook)

	// Test formatting Go file
	_ = hook.formatFile("test.go")

	// Check that either gofumpt or gofmt was called (prefers gofumpt when available)
	gofumptCalled := mockCmd.WasCommandExecuted("gofumpt", "-w", "test.go")
	gofmtCalled := mockCmd.WasCommandExecuted("gofmt", "-w", "test.go")
	
	if !gofumptCalled && !gofmtCalled {
		t.Error("Expected either gofumpt or gofmt to be executed for Go file")
	}
}

func TestFormatHookJavaScriptFile(t *testing.T) {
	mockCmd := core.NewMockCommandExecutor()
	ctx := &core.HookContext{
		FileSystem:      core.NewMockFileSystem(),
		CommandExecutor: mockCmd,
		RunnerFactory:   core.MockRunnerFactory,
		SettingsChecker: func(string) bool { return true },
	}

	hook := NewFormatHook(ctx).(*FormatHook)

	// Test formatting different JS/TS files
	files := []string{"test.js", "test.ts", "test.jsx", "test.tsx"}

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			_ = hook.formatFile(file)

			// Check that prettier was called
			if !mockCmd.WasCommandExecuted("prettier", "--write", file) {
				t.Errorf("Expected prettier to be executed for file %s", file)
			}
		})
	}
}

func TestFormatHookPythonFile(t *testing.T) {
	mockCmd := core.NewMockCommandExecutor()
	ctx := &core.HookContext{
		FileSystem:      core.NewMockFileSystem(),
		CommandExecutor: mockCmd,
		RunnerFactory:   core.MockRunnerFactory,
		SettingsChecker: func(string) bool { return true },
	}

	hook := NewFormatHook(ctx).(*FormatHook)

	// Test formatting Python file
	_ = hook.formatFile("test.py")

	// Check that ruff format was called
	if !mockCmd.WasCommandExecuted("uvx", "ruff", "format", "test.py") {
		t.Error("Expected uvx ruff format to be executed for Python file")
	}
}

func TestFormatHookYAMLFile(t *testing.T) {
	mockCmd := core.NewMockCommandExecutor()
	ctx := &core.HookContext{
		FileSystem:      core.NewMockFileSystem(),
		CommandExecutor: mockCmd,
		RunnerFactory:   core.MockRunnerFactory,
		SettingsChecker: func(string) bool { return true },
	}

	hook := NewFormatHook(ctx).(*FormatHook)

	// Test formatting different YAML files
	files := []string{"test.yml", "test.yaml"}

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			_ = hook.formatFile(file)

			// Check that prettier was called
			if !mockCmd.WasCommandExecuted("prettier", "--write", file) {
				t.Errorf("Expected prettier to be executed for file %s", file)
			}
		})
	}
}

func TestFormatHookUnsupportedFile(t *testing.T) {
	mockCmd := core.NewMockCommandExecutor()
	ctx := &core.HookContext{
		FileSystem:      core.NewMockFileSystem(),
		CommandExecutor: mockCmd,
		RunnerFactory:   core.MockRunnerFactory,
		SettingsChecker: func(string) bool { return true },
	}

	hook := NewFormatHook(ctx).(*FormatHook)

	// Test unsupported file extension
	_ = hook.formatFile("test.txt")

	// Check that no commands were executed
	commands := mockCmd.GetExecutedCommands()
	if len(commands) > 0 {
		t.Errorf("Expected no commands to be executed for unsupported file, got %d commands", len(commands))
	}
}
