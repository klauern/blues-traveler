package hooks

import (
	"fmt"

	"github.com/klauern/blues-traveler/internal/config"
	"github.com/klauern/blues-traveler/internal/core"
)

// init registers all built-in hooks using batch registration for better performance
func init() {
	builtinHooks := map[string]core.HookFactory{
		"security":      NewSecurityHook,
		"format":        NewFormatHook,
		"debug":         NewDebugHook,
		"audit":         NewAuditHook,
		"vet":           NewVetHook,
		"fetch-blocker": NewFetchBlockerHook,
		"find-blocker":  NewFindBlockerHook,
		// "performance": NewPerformanceHook, // TODO: Enable when performance.go is properly integrated
	}
	core.RegisterBuiltinHooks(builtinHooks)

	// Attempt to load and register config-based hooks.
	// Errors are non-fatal and will be surfaced at runtime via logs.
	registerConfigBasedHooks()
}

// registerConfigBasedHooks loads and registers hooks from configuration files
func registerConfigBasedHooks() {
	cfg, err := config.LoadHooksConfig()
	if err != nil || cfg == nil {
		return
	}

	factories := buildConfigHookFactories(cfg)
	if len(factories) > 0 {
		core.RegisterBuiltinHooks(factories)
	}
}

// buildConfigHookFactories creates hook factories from configuration
func buildConfigHookFactories(cfg *config.CustomHooksConfig) map[string]core.HookFactory {
	factories := make(map[string]core.HookFactory)

	for groupName, group := range *cfg {
		for eventName, eventCfg := range group {
			if eventCfg == nil {
				continue
			}
			addJobFactories(factories, groupName, eventName, eventCfg.Jobs)
		}
	}

	return factories
}

// addJobFactories adds hook factories for each job in the configuration
func addJobFactories(factories map[string]core.HookFactory, groupName, eventName string, jobs []config.HookJob) {
	for _, job := range jobs {
		if job.Name == "" {
			continue
		}
		key := fmt.Sprintf("config:%s:%s", groupName, job.Name)
		// Capture variables for closure
		g, j, e := groupName, job, eventName
		factories[key] = func(ctx *core.HookContext) core.Hook {
			return NewConfigHook(g, j.Name, j, e, ctx)
		}
	}
}
