package main

import (
	"fmt"
	"sort"
	"sync"

	"github.com/klauern/klauer-hooks/internal/hooks"
)

// Plugin is now an alias for Hook interface for backward compatibility
// The Hook interface from internal/hooks is more complete and serves the same purpose
type Plugin = hooks.Hook

// funcPlugin is an adapter to allow simple function-based plugins to be registered quickly.
type funcPlugin struct {
	key         string
	name        string
	description string
	fn          func() error
}

func (f *funcPlugin) Key() string         { return f.key }
func (f *funcPlugin) Name() string        { return f.name }
func (f *funcPlugin) Description() string { return f.description }
func (f *funcPlugin) Run() error          { return f.fn() }
func (f *funcPlugin) IsEnabled() bool     { return true } // Simple plugins are always enabled

// NewFuncPlugin creates a simple Plugin from a function and metadata.
func NewFuncPlugin(key, name, description string, fn func() error) Plugin {
	return &funcPlugin{key: key, name: name, description: description, fn: fn}
}

var (
	registryMu     sync.RWMutex
	pluginRegistry = map[string]Plugin{}
)

// RegisterPlugin registers a plugin under a unique key (e.g. "security", "format").
// It returns an error if the key is already registered.
func RegisterPlugin(key string, p Plugin) error {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, exists := pluginRegistry[key]; exists {
		return fmt.Errorf("plugin with key '%s' already registered", key)
	}
	pluginRegistry[key] = p
	return nil
}

// MustRegisterPlugin is like RegisterPlugin but panics on error (suitable for init()).
func MustRegisterPlugin(key string, p Plugin) {
	if err := RegisterPlugin(key, p); err != nil {
		panic(err)
	}
}

// GetPlugin retrieves a plugin by key.
func GetPlugin(key string) (Plugin, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	p, ok := pluginRegistry[key]
	return p, ok
}

// ListPlugins returns a snapshot copy of all registered plugins keyed by their registration key.
func ListPlugins() map[string]Plugin {
	registryMu.RLock()
	defer registryMu.RUnlock()
	out := make(map[string]Plugin, len(pluginRegistry))
	for k, v := range pluginRegistry {
		out[k] = v
	}
	return out
}

// PluginKeys returns sorted registration keys for stable UI output.
func PluginKeys() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	keys := make([]string, 0, len(pluginRegistry))
	for k := range pluginRegistry {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// init registers all built-in hooks from the modular system
func init() {
	// Setup the settings checker for the hooks system
	hooks.SetGlobalContext(&hooks.HookContext{
		FileSystem:      &hooks.RealFileSystem{},
		CommandExecutor: &hooks.RealCommandExecutor{},
		RunnerFactory:   hooks.DefaultRunnerFactory,
		SettingsChecker: isPluginEnabled, // Use the existing function from settings.go
	})

	// Register all hooks from the modular system directly
	for _, key := range hooks.GetHookKeys() {
		hook, err := hooks.CreateHook(key)
		if err != nil {
			panic(fmt.Errorf("failed to create hook '%s': %v", key, err))
		}
		MustRegisterPlugin(key, hook)
	}
}
