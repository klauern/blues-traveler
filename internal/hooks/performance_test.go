package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/klauern/blues-traveler/internal/core"
)

func TestPerformanceHook(t *testing.T) {
	ctx := core.TestHookContext(nil)
	hook := NewPerformanceHook(ctx)

	// Test basic properties
	if hook.Key() != "performance" {
		t.Errorf("Expected key 'performance', got '%s'", hook.Key())
	}

	if hook.Name() != "Performance Hook" {
		t.Errorf("Expected name 'Performance Hook', got '%s'", hook.Name())
	}

	// Test that hook is enabled by default
	if !hook.IsEnabled() {
		t.Error("Expected hook to be enabled by default")
	}
}

func TestPerformanceHookDescription(t *testing.T) {
	ctx := core.TestHookContext(nil)
	hook := NewPerformanceHook(ctx)

	description := hook.Description()
	expectedDescription := "Monitors hook execution performance and resource usage"

	if description != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, description)
	}
}

func TestPerformanceHookDisabled(t *testing.T) {
	ctx := &core.HookContext{
		FileSystem:      core.NewMockFileSystem(),
		CommandExecutor: core.NewMockCommandExecutor(),
		RunnerFactory:   core.MockRunnerFactory,
		SettingsChecker: func(string) bool { return false }, // Disabled
	}

	hook := NewPerformanceHook(ctx)

	// Test that hook respects disabled state
	if hook.IsEnabled() {
		t.Error("Expected hook to be disabled")
	}

	// Running a disabled hook should not error
	err := hook.Run()
	if err != nil {
		t.Errorf("Disabled hook run failed: %v", err)
	}
}

func TestPerformanceHookRun(t *testing.T) {
	// Create a temporary directory for the test
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Create .claude/hooks directory
	hooksDir := filepath.Join(tempDir, ".claude", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatalf("failed to create hooks directory: %v", err)
	}

	ctx := core.TestHookContext(nil)
	hook := NewPerformanceHook(ctx)

	// Test running the hook (should not error)
	err := hook.Run()
	if err != nil {
		t.Errorf("Hook run failed: %v", err)
	}

	// Verify that performance.log was created
	logPath := filepath.Join(tempDir, ".claude", "hooks", "performance.log")
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("Expected performance.log to be created")
	}
}

func TestPerformanceHookLogging(t *testing.T) {
	// Create a temporary directory for the test
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Create .claude/hooks directory
	hooksDir := filepath.Join(tempDir, ".claude", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatalf("failed to create hooks directory: %v", err)
	}

	ctx := &core.HookContext{
		FileSystem:      core.NewMockFileSystem(),
		CommandExecutor: core.NewMockCommandExecutor(),
		RunnerFactory:   core.MockRunnerFactory,
		SettingsChecker: func(string) bool { return true },
		LoggingEnabled:  true, // Enable detailed logging
	}

	hook := NewPerformanceHook(ctx)

	// Run the hook
	err := hook.Run()
	if err != nil {
		t.Errorf("Hook run failed: %v", err)
	}

	// Verify that performance.log was created
	logPath := filepath.Join(tempDir, ".claude", "hooks", "performance.log")
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("Expected performance.log to be created")
	}

	// Read log file and verify it has expected content
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)

	// Check for session markers
	if !strings.Contains(logContent, "Performance Monitoring Session Started") {
		t.Error("Expected log to contain session start marker")
	}

	if !strings.Contains(logContent, "Performance Monitoring Session Ended") {
		t.Error("Expected log to contain session end marker")
	}

	if !strings.Contains(logContent, "Performance Monitoring Summary") {
		t.Error("Expected log to contain summary section")
	}
}

func TestPerformanceHookStructure(t *testing.T) {
	ctx := core.TestHookContext(nil)
	hook := NewPerformanceHook(ctx).(*PerformanceHook)

	// Verify that the hook has the expected structure
	if hook.BaseHook == nil {
		t.Error("Expected BaseHook to be initialized")
	}

	if hook.startTimes == nil {
		t.Error("Expected startTimes map to be initialized")
	}

	// Verify initial state
	if hook.totalTime != 0 {
		t.Error("Expected totalTime to be zero initially")
	}

	if hook.toolCount != 0 {
		t.Error("Expected toolCount to be zero initially")
	}
}

func TestPerformanceHookLoggerInitialization(t *testing.T) {
	// Create a temporary directory for the test
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Create .claude/hooks directory
	hooksDir := filepath.Join(tempDir, ".claude", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatalf("failed to create hooks directory: %v", err)
	}

	ctx := core.TestHookContext(nil)
	hook := NewPerformanceHook(ctx).(*PerformanceHook)

	// Initially, logger should be nil
	if hook.logger != nil {
		t.Error("Expected logger to be nil before ensureLogger is called")
	}

	// Call ensureLogger
	hook.ensureLogger()

	// After ensureLogger, logger should be initialized
	if hook.logger == nil {
		t.Error("Expected logger to be initialized after ensureLogger is called")
	}

	// Calling ensureLogger again should not change the logger
	originalLogger := hook.logger
	hook.ensureLogger()
	if hook.logger != originalLogger {
		t.Error("Expected logger to remain the same after multiple ensureLogger calls")
	}
}

func TestPerformanceHookLogDirectory(t *testing.T) {
	// Create a temporary directory for the test
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	ctx := core.TestHookContext(nil)
	hook := NewPerformanceHook(ctx).(*PerformanceHook)

	// Call ensureLogger which should create the directory
	hook.ensureLogger()

	// Verify that .claude/hooks directory was created
	hooksDir := filepath.Join(tempDir, ".claude", "hooks")
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		t.Error("Expected .claude/hooks directory to be created")
	}
}
