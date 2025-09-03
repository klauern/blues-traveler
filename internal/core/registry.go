package core

import (
	"fmt"
	"sort"
	"sync"

	"github.com/klauern/blues-traveler/internal/config"
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

// Register registers a hook factory with the given key (used by tests)
func (r *Registry) Register(key string, factory HookFactory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[key]; exists {
		return fmt.Errorf("hook with key '%s' already registered", key)
	}

	r.factories[key] = factory
	return nil
}

// MustRegister is like Register but panics on error (used by tests)
func (r *Registry) MustRegister(key string, factory HookFactory) {
	if err := r.Register(key, factory); err != nil {
		panic(err)
	}
}


// RegisterBatch registers multiple hooks concurrently for better initialization performance
func (r *Registry) RegisterBatch(hooks map[string]HookFactory) error {
	// Register all at once under write lock to avoid race conditions
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicates under the write lock
	for key := range hooks {
		if _, exists := r.factories[key]; exists {
			return fmt.Errorf("hook with key '%s' already registered", key)
		}
	}

	// Register all hooks
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
	context := r.context
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("hook with key '%s' not found", key)
	}

	return factory(context), nil
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

// List returns a map of all hooks (key -> hook instance) (used by tests)
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


// CreateHook creates a hook instance by key from the global registry
func CreateHook(key string) (Hook, error) {
	return globalRegistry.Create(key)
}

// GetHookKeys returns all registered hook keys from the global registry
func GetHookKeys() []string {
	return globalRegistry.Keys()
}

// (removed) ListHooks unused externally; avoid exporting broader surface.

// SetGlobalContext updates the global registry's context
func SetGlobalContext(ctx *HookContext) {
	globalRegistry.SetContext(ctx)
}

// SetGlobalLoggingConfig updates the global registry's context with logging configuration
func SetGlobalLoggingConfig(enabled bool, logDir string, format string) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	if globalRegistry.context != nil {
		globalRegistry.context.LoggingEnabled = enabled
		globalRegistry.context.LoggingDir = logDir
		if config.IsValidLoggingFormat(format) {
			globalRegistry.context.LoggingFormat = format
		}
		// else: leave default format when empty or invalid
	}
}

// (removed) GetGlobalRegistry unused; keep internal-only access.

// RegisterBuiltinHooks can be called by the hooks package to register all built-in hooks
func RegisterBuiltinHooks(hooks map[string]HookFactory) {
	globalRegistry.MustRegisterBatch(hooks)
}
