package hooks

import "github.com/klauern/klauer-hooks/internal/core"

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
}
