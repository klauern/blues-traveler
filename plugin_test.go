package main

import "testing"

// Test that PluginKeys returns sorted keys and includes required built-ins.
func TestPluginKeysSortedAndBuiltinPresence(t *testing.T) {
	t.Attr("category", "registry")
	t.Attr("component", "plugin-registry")

	keys := PluginKeys()
	if len(keys) == 0 {
		t.Fatalf("expected at least one plugin key")
	}
	for i := 1; i < len(keys); i++ {
		if keys[i-1] > keys[i] {
			t.Fatalf("PluginKeys not sorted: %s > %s at positions %d,%d", keys[i-1], keys[i], i-1, i)
		}
	}

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
			t.Fatalf("missing required plugin key %q", r)
		}
	}
}

// Test Settings.IsPluginEnabled logic (default enabled, explicit enable/disable).
func TestSettingsIsPluginEnabled(t *testing.T) {
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
