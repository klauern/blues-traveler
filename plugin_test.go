package main

import (
	"testing"
)

// Test that registering a new plugin works and duplicate keys error.
func TestRegisterPluginDuplicate(t *testing.T) {
	t.Attr("category", "registration")
	t.Attr("component", "plugin-registry")
	key := "test-temp-plugin"
	// Ensure key not already present
	if _, exists := GetPlugin(key); exists {
		t.Fatalf("test precondition failed: plugin %s already exists", key)
	}
	// First registration should succeed
	err := RegisterPlugin(key, NewFuncPlugin(key, "Temp", "Temporary test plugin", func() error { return nil }))
	if err != nil {
		t.Fatalf("expected first registration success, got error: %v", err)
	}
	// Second registration should fail
	err = RegisterPlugin(key, NewFuncPlugin(key+"2", "Temp2", "Another", func() error { return nil }))
	if err == nil {
		t.Fatalf("expected duplicate registration error, got nil")
	}
}

// Test that PluginKeys returns sorted keys (lexicographically ascending).
func TestPluginKeysSorted(t *testing.T) {
	t.Attr("category", "ordering")
	t.Attr("component", "plugin-registry")
	keys := PluginKeys()
	for i := 1; i < len(keys); i++ {
		if keys[i-1] > keys[i] {
			t.Fatalf("PluginKeys not sorted: %s > %s at positions %d,%d", keys[i-1], keys[i], i-1, i)
		}
	}
	// Sanity: built-in keys should be present
	required := []string{"security", "format", "debug", "audit"}
	for _, r := range required {
		found := false
		for _, k := range keys {
			if k == r {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected built-in plugin key %q not found in PluginKeys", r)
		}
	}
}

// Test Settings.IsPluginEnabled logic (default enabled, explicit disable).
func TestIsPluginEnabled(t *testing.T) {
	t.Attr("category", "settings")
	t.Attr("component", "plugin-config")
	s := &Settings{
		Plugins: map[string]PluginConfig{},
	}
	// Default: absent plugin treated as enabled
	if !s.IsPluginEnabled("nonexistent") {
		t.Fatalf("expected absent plugin to be enabled by default")
	}

	// Explicitly enable
	trueVal := true
	s.Plugins["alpha"] = PluginConfig{Enabled: &trueVal}
	if !s.IsPluginEnabled("alpha") {
		t.Fatalf("expected explicitly enabled plugin to be enabled")
	}

	// Explicitly disable
	falseVal := false
	s.Plugins["beta"] = PluginConfig{Enabled: &falseVal}
	if s.IsPluginEnabled("beta") {
		t.Fatalf("expected explicitly disabled plugin to be disabled")
	}
}
