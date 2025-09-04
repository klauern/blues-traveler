package compat

import (
	"sort"

	"github.com/klauern/blues-traveler/internal/config"
	"github.com/klauern/blues-traveler/internal/core"
)

// DEPRECATED REGISTRY SHIM
// This file now provides thin compatibility helpers mapping the historical
// plugin registry API onto the internal/hooks global registry directly.
// The unified pipeline uses hooks.BuildUnifiedRunner / BuildRunnerForKeys;
// new code should reference hooks.* functions instead of these shims.

// Plugin alias retained for backward compatibility.
type Plugin = core.Hook

// GetPlugin constructs a fresh hook instance by key using internal registry.
// Returns (nil,false) if key not found.
func GetPlugin(key string) (Plugin, bool) {
	h, err := core.CreateHook(key)
	if err != nil || h == nil {
		return nil, false
	}
	return h, true
}

// PluginKeys returns sorted hook keys from internal registry.
func PluginKeys() []string {
	keys := core.GetHookKeys()
	sort.Strings(keys)
	return keys
}

// IsPluginEnabled is a wrapper around config.IsPluginEnabled for compatibility
func IsPluginEnabled(pluginKey string) bool {
	return config.IsPluginEnabled(pluginKey)
}

// init: ensure HookContext has proper settings checker (previous behavior)
func init() {
	core.SetGlobalContext(&core.HookContext{
		FileSystem:      &core.RealFileSystem{},
		CommandExecutor: &core.RealCommandExecutor{},
		RunnerFactory:   core.DefaultRunnerFactory,
		SettingsChecker: config.IsPluginEnabled,
	})
}
