package main

import (
	"fmt"
	"sort"
	"sync"
)

// Plugin defines the behavior for all dynamically registered hook plugins.
// Implementations may hold internal state/config if needed.
type Plugin interface {
	Name() string
	Description() string
	Run() error
}

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

// ----- Registration of built-in plugins -----
// Each built-in plugin registers itself in init() for automatic discovery.

func init() {
	// Security hook
	MustRegisterPlugin("security", NewFuncPlugin(
		"Security Hook",
		"Blocks dangerous commands and provides security controls",
		runSecurityHook,
	))
	// Format hook
	MustRegisterPlugin("format", NewFuncPlugin(
		"Format Hook",
		"Enforces code formatting standards",
		runFormatHook,
	))
	// Debug hook
	MustRegisterPlugin("debug", NewFuncPlugin(
		"Debug Hook",
		"Logs all tool usage for debugging purposes",
		runDebugHook,
	))
	// Audit hook
	MustRegisterPlugin("audit", NewFuncPlugin(
		"Audit Hook",
		"Comprehensive audit logging with JSON output",
		runAuditHook,
	))
}
