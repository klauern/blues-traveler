package core

import (
	"reflect"
	"testing"
)

// testHook is a simple hook implementation for testing
type testHook struct {
	*BaseHook
}

func (h *testHook) Run() error {
	return nil
}

func newTestHook(key, name, description string, ctx *HookContext) Hook {
	base := NewBaseHook(key, name, description, ctx)
	return &testHook{BaseHook: base}
}

func TestRegistry(t *testing.T) {
	ctx := TestHookContext(nil)
	registry := NewRegistry(ctx)

	// Test registering a hook
	factory := func(ctx *HookContext) Hook {
		return newTestHook("test", "Test Hook", "Test description", ctx)
	}

	err := registry.Register("test", factory)
	if err != nil {
		t.Errorf("Failed to register hook: %v", err)
	}

	// Test creating hook
	hook, err := registry.Create("test")
	if err != nil {
		t.Errorf("Failed to create hook: %v", err)
	}

	if hook.Key() != "test" {
		t.Errorf("Expected hook key 'test', got '%s'", hook.Key())
	}
}

func TestRegistryDuplicateKey(t *testing.T) {
	ctx := TestHookContext(nil)
	registry := NewRegistry(ctx)

	factory := func(ctx *HookContext) Hook {
		return newTestHook("test", "Test Hook", "Test description", ctx)
	}

	// Register first time - should succeed
	err := registry.Register("test", factory)
	if err != nil {
		t.Errorf("First registration failed: %v", err)
	}

	// Register second time - should fail
	err = registry.Register("test", factory)
	if err == nil {
		t.Error("Expected error when registering duplicate key")
	}
}

func TestRegistryMustRegister(t *testing.T) {
	ctx := TestHookContext(nil)
	registry := NewRegistry(ctx)

	factory := func(ctx *HookContext) Hook {
		return newTestHook("test", "Test Hook", "Test description", ctx)
	}

	// Should not panic for first registration
	registry.MustRegister("test", factory)

	// Should panic for duplicate registration
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when registering duplicate key with MustRegister")
		}
	}()

	registry.MustRegister("test", factory)
}

func TestRegistryKeys(t *testing.T) {
	ctx := TestHookContext(nil)
	registry := NewRegistry(ctx)

	// Register some hooks
	factories := map[string]HookFactory{
		"z_hook": func(ctx *HookContext) Hook { return newTestHook("z_hook", "Z Hook", "Z description", ctx) },
		"a_hook": func(ctx *HookContext) Hook { return newTestHook("a_hook", "A Hook", "A description", ctx) },
		"m_hook": func(ctx *HookContext) Hook { return newTestHook("m_hook", "M Hook", "M description", ctx) },
	}

	for key, factory := range factories {
		registry.MustRegister(key, factory)
	}

	keys := registry.Keys()
	expected := []string{"a_hook", "m_hook", "z_hook"}

	if !reflect.DeepEqual(keys, expected) {
		t.Errorf("Expected keys %v, got %v", expected, keys)
	}
}

func TestRegistryCreateNotFound(t *testing.T) {
	ctx := TestHookContext(nil)
	registry := NewRegistry(ctx)

	_, err := registry.Create("nonexistent")
	if err == nil {
		t.Error("Expected error when creating non-existent hook")
	}
}

func TestRegistrySetContext(t *testing.T) {
	ctx1 := TestHookContext(func(string) bool { return true })
	ctx2 := TestHookContext(func(string) bool { return false })

	registry := NewRegistry(ctx1)

	factory := func(ctx *HookContext) Hook {
		return newTestHook("test", "Test Hook", "Test description", ctx)
	}
	registry.MustRegister("test", factory)

	// Create hook with first context
	hook1, _ := registry.Create("test")
	if !hook1.IsEnabled() {
		t.Error("Expected hook to be enabled with first context")
	}

	// Change context
	registry.SetContext(ctx2)

	// Create hook with second context
	hook2, _ := registry.Create("test")
	if hook2.IsEnabled() {
		t.Error("Expected hook to be disabled with second context")
	}
}

func TestGlobalRegistry(t *testing.T) {
	// Test that global registry functions work
	_ = GetHookKeys() // Just test that it doesn't panic

	// Test that we can register and retrieve hooks
	testFactory := func(ctx *HookContext) Hook {
		return newTestHook("test-global", "Test Global Hook", "Test hook", ctx)
	}

	// Register a test hook
	err := globalRegistry.Register("test-global", testFactory)
	if err != nil {
		t.Errorf("Failed to register test hook: %v", err)
	}

	// Test creating the registered hook
	hook, err := CreateHook("test-global")
	if err != nil {
		t.Errorf("Failed to create hook 'test-global': %v", err)
	}
	if hook.Key() != "test-global" {
		t.Errorf("Hook key mismatch: expected 'test-global', got '%s'", hook.Key())
	}

	// Clean up
	globalRegistry = NewRegistry(DefaultHookContext())
}

func TestRegistryList(t *testing.T) {
	ctx := TestHookContext(nil)
	registry := NewRegistry(ctx)

	// Register test hooks
	factories := map[string]HookFactory{
		"test1": func(ctx *HookContext) Hook { return newTestHook("test1", "Test 1", "Test 1 description", ctx) },
		"test2": func(ctx *HookContext) Hook { return newTestHook("test2", "Test 2", "Test 2 description", ctx) },
	}

	for key, factory := range factories {
		registry.MustRegister(key, factory)
	}

	// Test listing hooks
	hooks := registry.List()

	if len(hooks) != 2 {
		t.Errorf("Expected 2 hooks, got %d", len(hooks))
	}

	for key := range factories {
		if hook, exists := hooks[key]; !exists {
			t.Errorf("Expected hook '%s' to be in list", key)
		} else if hook.Key() != key {
			t.Errorf("Hook key mismatch: expected '%s', got '%s'", key, hook.Key())
		}
	}
}
