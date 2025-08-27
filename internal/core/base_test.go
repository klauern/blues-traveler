package core

import (
	"testing"
)

func TestBaseHook(t *testing.T) {
	ctx := TestHookContext(nil)

	hook := NewBaseHook("test", "Test Hook", "Test description", ctx)

	// Test basic properties
	if hook.Key() != "test" {
		t.Errorf("Expected key 'test', got '%s'", hook.Key())
	}

	if hook.Name() != "Test Hook" {
		t.Errorf("Expected name 'Test Hook', got '%s'", hook.Name())
	}

	if hook.Description() != "Test description" {
		t.Errorf("Expected description 'Test description', got '%s'", hook.Description())
	}

	// Test enabled by default
	if !hook.IsEnabled() {
		t.Error("Expected hook to be enabled by default")
	}

	if hook.Context() != ctx {
		t.Error("Expected context to match provided context")
	}
}

func TestBaseHookDisabled(t *testing.T) {
	ctx := TestHookContext(func(string) bool { return false })

	hook := NewBaseHook("test", "Test Hook", "Test description", ctx)

	if hook.IsEnabled() {
		t.Error("Expected hook to be disabled")
	}
}

func TestBaseHookNilContext(t *testing.T) {
	hook := NewBaseHook("test", "Test Hook", "Test description", nil)

	// Should get default context
	if hook.Context() == nil {
		t.Error("Expected default context when nil provided")
	}
}
