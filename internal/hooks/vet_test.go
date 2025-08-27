package hooks

import (
	"testing"

	"github.com/klauern/klauer-hooks/internal/core"
)

func TestVetHook(t *testing.T) {
	ctx := &core.HookContext{
		CommandExecutor: &core.MockCommandExecutor{},
		RunnerFactory:   core.MockRunnerFactory,
		LoggingEnabled:  false,
	}

	hook := NewVetHook(ctx)
	if hook == nil {
		t.Fatal("Expected hook to not be nil")
	}

	if hook.Key() != "vet" {
		t.Errorf("Expected hook key to be 'vet', got %s", hook.Key())
	}

	if hook.Description() != "Performs Python type checking using ty" {
		t.Errorf("Expected hook description to be 'Performs Python type checking using ty', got %s", hook.Description())
	}
}

func TestVetHookIsPythonFile(t *testing.T) {
	ctx := core.DefaultHookContext()
	hook := NewVetHook(ctx).(*VetHook)

	tests := []struct {
		filePath string
		expected bool
	}{
		{"main.py", true},
		{"script.py", true},
		{"test.PY", true},
		{"main.go", false},
		{"script.js", false},
		{"README.md", false},
		{"config.yml", false},
	}

	for _, test := range tests {
		result := hook.isPythonFile(test.filePath)
		if result != test.expected {
			t.Errorf("isPythonFile(%s) = %t, expected %t", test.filePath, result, test.expected)
		}
	}
}
