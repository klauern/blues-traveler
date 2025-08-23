package main

import (
	"fmt"
	"sort"
	"sync"

	"github.com/klauern/klauer-hooks/internal/hooks"
)

// Plugin defines the behavior for all dynamically registered hook plugins.
// This interface remains for backward compatibility but now wraps the new Hook interface
type Plugin interface {
	Name() string
	Description() string
	Run() error
}

// hookAdapter wraps a Hook to implement the Plugin interface
type hookAdapter struct {
	hook hooks.Hook
}

func (h *hookAdapter) Name() string        { return h.hook.Name() }
func (h *hookAdapter) Description() string { return h.hook.Description() }
func (h *hookAdapter) Run() error          { return h.hook.Run() }

// funcPlugin is an adapter to allow simple function-based plugins to be registered quickly.
type funcPlugin struct {
	name        string
	description string
	fn          func() error
}

func (f *funcPlugin) Name() string        { return f.name }
func (f *funcPlugin) Description() string { return f.description }
func (f *funcPlugin) Run() error          { return f.fn() }

// NewFuncPlugin creates a simple Plugin from a function and metadata.
func NewFuncPlugin(name, description string, fn func() error) Plugin {
	return &funcPlugin{name: name, description: description, fn: fn}
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

// RegisterHookAsPlugin registers a hook from the modular system as a plugin
func RegisterHookAsPlugin(key string) error {
	hook, err := hooks.CreateHook(key)
	if err != nil {
		return fmt.Errorf("failed to create hook '%s': %v", key, err)
	}

	adapter := &hookAdapter{hook: hook}
	return RegisterPlugin(key, adapter)
}

// MustRegisterHookAsPlugin registers a hook as a plugin, panics on error
func MustRegisterHookAsPlugin(key string) {
	if err := RegisterHookAsPlugin(key); err != nil {
		panic(err)
	}
}

// init registers all built-in hooks from the modular system
func init() {
	// Setup the settings checker for the hooks system
	hooks.SetGlobalContext(&hooks.HookContext{
		FileSystem:      &hooks.RealFileSystem{},
		CommandExecutor: &hooks.RealCommandExecutor{},
		RunnerFactory:   hooks.DefaultRunnerFactory,
		SettingsChecker: isPluginEnabled, // Use the existing function from main
	})

	// Register all hooks from the modular system
	for _, key := range hooks.GetHookKeys() {
		MustRegisterHookAsPlugin(key)
	}
}
