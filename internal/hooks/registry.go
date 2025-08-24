package hooks

import (
	"fmt"
	"sort"
	"sync"
)

// HookFactory is a function that creates a Hook instance
type HookFactory func(ctx *HookContext) Hook

// Registry manages hook registration and creation
type Registry struct {
	mu        sync.RWMutex
	factories map[string]HookFactory
	context   *HookContext
}

// NewRegistry creates a new hook registry
func NewRegistry(ctx *HookContext) *Registry {
	if ctx == nil {
		ctx = DefaultHookContext()
	}
	return &Registry{
		factories: make(map[string]HookFactory),
		context:   ctx,
	}
}

// Register registers a hook factory with the given key
func (r *Registry) Register(key string, factory HookFactory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[key]; exists {
		return fmt.Errorf("hook with key '%s' already registered", key)
	}

	r.factories[key] = factory
	return nil
}

// MustRegister is like Register but panics on error
func (r *Registry) MustRegister(key string, factory HookFactory) {
	if err := r.Register(key, factory); err != nil {
		panic(err)
	}
}

// RegisterBatch registers multiple hooks concurrently for better initialization performance
func (r *Registry) RegisterBatch(hooks map[string]HookFactory) error {
	// First, validate all keys don't exist to avoid partial registration
	r.mu.RLock()
	for key := range hooks {
		if _, exists := r.factories[key]; exists {
			r.mu.RUnlock()
			return fmt.Errorf("hook with key '%s' already registered", key)
		}
	}
	r.mu.RUnlock()

	// Now register all at once
	r.mu.Lock()
	defer r.mu.Unlock()
	for key, factory := range hooks {
		r.factories[key] = factory
	}
	return nil
}

// MustRegisterBatch is like RegisterBatch but panics on error
func (r *Registry) MustRegisterBatch(hooks map[string]HookFactory) {
	if err := r.RegisterBatch(hooks); err != nil {
		panic(err)
	}
}

// Create creates a hook instance by key
func (r *Registry) Create(key string) (Hook, error) {
	r.mu.RLock()
	factory, exists := r.factories[key]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("hook with key '%s' not found", key)
	}

	return factory(r.context), nil
}

// Keys returns all registered hook keys in sorted order
func (r *Registry) Keys() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	keys := make([]string, 0, len(r.factories))
	for k := range r.factories {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// List returns a map of all hooks (key -> hook instance)
func (r *Registry) List() map[string]Hook {
	r.mu.RLock()
	factories := make(map[string]HookFactory, len(r.factories))
	for k, v := range r.factories {
		factories[k] = v
	}
	ctx := r.context
	r.mu.RUnlock()

	result := make(map[string]Hook, len(factories))
	resultMu := sync.Mutex{}
	var wg sync.WaitGroup

	for key, factory := range factories {
		wg.Go(func() {
			// Capture variables for the closure
			k, f := key, factory
			hook := f(ctx)
			resultMu.Lock()
			result[k] = hook
			resultMu.Unlock()
		})
	}

	wg.Wait()
	return result
}

// SetContext updates the context used for creating hook instances
func (r *Registry) SetContext(ctx *HookContext) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.context = ctx
}

// Global registry instance
var globalRegistry = NewRegistry(nil)

// RegisterHook registers a hook factory globally
func RegisterHook(key string, factory HookFactory) error {
	return globalRegistry.Register(key, factory)
}

// MustRegisterHook registers a hook factory globally, panics on error
func MustRegisterHook(key string, factory HookFactory) {
	globalRegistry.MustRegister(key, factory)
}

// CreateHook creates a hook instance by key from the global registry
func CreateHook(key string) (Hook, error) {
	return globalRegistry.Create(key)
}

// GetHookKeys returns all registered hook keys from the global registry
func GetHookKeys() []string {
	return globalRegistry.Keys()
}

// ListHooks returns all hooks from the global registry
func ListHooks() map[string]Hook {
	return globalRegistry.List()
}

// SetGlobalContext updates the global registry's context
func SetGlobalContext(ctx *HookContext) {
	globalRegistry.SetContext(ctx)
}

// SetGlobalLoggingConfig updates the global registry's context with logging configuration
func SetGlobalLoggingConfig(enabled bool, logDir string) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	if globalRegistry.context != nil {
		globalRegistry.context.LoggingEnabled = enabled
		globalRegistry.context.LoggingDir = logDir
	}
}

// GetGlobalRegistry returns the global registry instance
func GetGlobalRegistry() *Registry {
	return globalRegistry
}

// init registers all built-in hooks using batch registration for better performance
func init() {
	builtinHooks := map[string]HookFactory{
		"security": NewSecurityHook,
		"format":   NewFormatHook,
		"debug":    NewDebugHook,
		"audit":    NewAuditHook,
		"vet":      NewVetHook,
		// "performance": NewPerformanceHook, // TODO: Enable when performance.go is properly integrated
	}
	globalRegistry.MustRegisterBatch(builtinHooks)
}
