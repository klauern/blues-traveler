package main

import (
	"fmt"
	"sort"

	"github.com/klauern/klauer-hooks/internal/hooks"
)

// DEPRECATED REGISTRY SHIM
// This file now provides thin compatibility helpers mapping the historical
// plugin registry API onto the internal/hooks global registry directly.
// The unified pipeline uses hooks.BuildUnifiedRunner / BuildRunnerForKeys;
// new code should reference hooks.* functions instead of these shims.

// Plugin alias retained for backward compatibility.
type Plugin = hooks.Hook

// GetPlugin constructs a fresh hook instance by key using internal registry.
// Returns (nil,false) if key not found.
func GetPlugin(key string) (Plugin, bool) {
	h, err := hooks.CreateHook(key)
	if err != nil || h == nil {
		return nil, false
	}
	return h, true
}

// PluginKeys returns sorted hook keys from internal registry.
func PluginKeys() []string {
	keys := hooks.GetHookKeys()
	sort.Strings(keys)
	return keys
}

// ListPlugins returns fresh instances for all keys (legacy shape).
func ListPlugins() map[string]Plugin {
	out := map[string]Plugin{}
	for _, k := range hooks.GetHookKeys() {
		if h, err := hooks.CreateHook(k); err == nil {
			out[k] = h
		}
	}
	return out
}

// RegisterPlugin / MustRegisterPlugin kept as no-ops to avoid compile breakage
// for external code that might still attempt dynamic registration.
// In future major version, remove these.
func RegisterPlugin(key string, p Plugin) error {
	// No dynamic registration path; advise migration.
	return fmt.Errorf("dynamic plugin registration deprecated; add hook in internal/hooks/registry.go")
}

func MustRegisterPlugin(key string, p Plugin) {
	if err := RegisterPlugin(key, p); err != nil {
		panic(err)
	}
}

// init: ensure HookContext has proper settings checker (previous behavior)
func init() {
	hooks.SetGlobalContext(&hooks.HookContext{
		FileSystem:      &hooks.RealFileSystem{},
		CommandExecutor: &hooks.RealCommandExecutor{},
		RunnerFactory:   hooks.DefaultRunnerFactory,
		SettingsChecker: isPluginEnabled,
	})
}
