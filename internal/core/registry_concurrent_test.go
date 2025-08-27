package core

import (
	"sync"
	"testing"
	// Note: testing/synctest may not be available in all Go 1.25 builds yet
	// Using standard concurrent testing patterns for now
)

// Test hook implementations for testing
func NewTestSecurityHook(ctx *HookContext) Hook {
	return newTestHook("test-security", "Test Security Hook", "Test security hook", ctx)
}

func NewTestFormatHook(ctx *HookContext) Hook {
	return newTestHook("test-format", "Test Format Hook", "Test format hook", ctx)
}

func NewTestDebugHook(ctx *HookContext) Hook {
	return newTestHook("test-debug", "Test Debug Hook", "Test debug hook", ctx)
}

func NewTestAuditHook(ctx *HookContext) Hook {
	return newTestHook("test-audit", "Test Audit Hook", "Test audit hook", ctx)
}

// TestRegistryConcurrentOperations tests concurrent access to the registry
func TestRegistryConcurrentOperations(t *testing.T) {
	registry := NewRegistry(DefaultHookContext())

	// Test concurrent registration
	var wg sync.WaitGroup
	errors := make(chan error, 10)

	// Register multiple hooks concurrently
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			key := "concurrent-test-hook"
			factory := NewTestSecurityHook
			if err := registry.Register(key, factory); err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Count errors - should have 4 errors (duplicates) and 1 success
	errorCount := 0
	for err := range errors {
		if err != nil {
			errorCount++
		}
	}

	if errorCount != 4 {
		t.Fatalf("expected 4 duplicate registration errors, got %d", errorCount)
	}

	// Verify the hook was registered once
	keys := registry.Keys()
	found := false
	for _, key := range keys {
		if key == "concurrent-test-hook" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected hook to be registered successfully once")
	}
}

// TestRegistryBatchOperations tests batch registration
func TestRegistryBatchOperations(t *testing.T) {
	registry := NewRegistry(DefaultHookContext())

	// Test concurrent batch registrations
	var wg sync.WaitGroup
	errors := make(chan error, 5)

	// Multiple goroutines trying to register the same batch
	batch := map[string]HookFactory{
		"batch-test-1": NewTestSecurityHook,
		"batch-test-2": NewTestFormatHook,
	}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := registry.RegisterBatch(batch); err != nil {
				errors <- err
			} else {
				errors <- nil
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Count results - should have 4 errors and 1 success
	successCount := 0
	errorCount := 0
	for err := range errors {
		if err == nil {
			successCount++
		} else {
			errorCount++
		}
	}

	if successCount != 1 {
		t.Fatalf("expected 1 successful batch registration, got %d", successCount)
	}
	if errorCount != 4 {
		t.Fatalf("expected 4 duplicate batch registration errors, got %d", errorCount)
	}

	// Verify both hooks were registered
	keys := registry.Keys()
	expectedKeys := []string{"batch-test-1", "batch-test-2"}
	for _, expected := range expectedKeys {
		found := false
		for _, key := range keys {
			if key == expected {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected hook %s to be registered", expected)
		}
	}
}

// TestRegistryConcurrentListAndCreate tests concurrent read operations
func TestRegistryConcurrentListAndCreate(t *testing.T) {
	registry := NewRegistry(DefaultHookContext())

	// Pre-populate registry
	batch := map[string]HookFactory{
		"read-test-1": NewTestSecurityHook,
		"read-test-2": NewTestFormatHook,
		"read-test-3": NewTestDebugHook,
		"read-test-4": NewTestAuditHook,
	}
	registry.MustRegisterBatch(batch)

	// Test concurrent reading operations
	var wg sync.WaitGroup
	results := make(chan map[string]Hook, 10)
	createResults := make(chan Hook, 10)

	// Multiple goroutines listing hooks
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			hooks := registry.List()
			results <- hooks
		}()
	}

	// Multiple goroutines creating individual hooks
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			hook, _ := registry.Create("read-test-1")
			createResults <- hook
		}()
	}

	wg.Wait()
	close(results)
	close(createResults)

	// Verify all List operations returned the same number of hooks
	expectedCount := len(batch)
	for hooks := range results {
		if len(hooks) != expectedCount {
			t.Fatalf("expected %d hooks in List(), got %d", expectedCount, len(hooks))
		}
	}

	// Verify all Create operations succeeded
	createCount := 0
	for hook := range createResults {
		if hook != nil {
			createCount++
		}
	}
	if createCount != 5 {
		t.Fatalf("expected 5 successful Create operations, got %d", createCount)
	}
}

// TestRegistryContextUpdate tests concurrent context updates
func TestRegistryContextUpdate(t *testing.T) {
	registry := NewRegistry(DefaultHookContext())

	// Register a test hook
	registry.MustRegister("context-test", NewTestSecurityHook)

	var wg sync.WaitGroup
	contextUpdates := make(chan bool, 5)
	hookCreations := make(chan bool, 5)

	// Concurrent context updates
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			newCtx := DefaultHookContext()
			registry.SetContext(newCtx)
			contextUpdates <- true
		}()
	}

	// Concurrent hook creations during context updates
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := registry.Create("context-test")
			hookCreations <- err == nil
		}()
	}

	wg.Wait()
	close(contextUpdates)
	close(hookCreations)

	// Verify context updates completed
	updateCount := 0
	for updated := range contextUpdates {
		if updated {
			updateCount++
		}
	}
	if updateCount != 3 {
		t.Fatalf("expected 3 context updates, got %d", updateCount)
	}

	// Verify hook creations succeeded despite concurrent context updates
	successCount := 0
	for success := range hookCreations {
		if success {
			successCount++
		}
	}
	if successCount != 5 {
		t.Fatalf("expected 5 successful hook creations, got %d", successCount)
	}
}

// TestRegistryWaitGroupGo tests the new sync.WaitGroup.Go() method usage
func TestRegistryWaitGroupGo(t *testing.T) {
	registry := NewRegistry(DefaultHookContext())

	// Pre-populate with test hooks
	batch := map[string]HookFactory{
		"wg-test-1": NewTestSecurityHook,
		"wg-test-2": NewTestFormatHook,
		"wg-test-3": NewTestDebugHook,
	}
	registry.MustRegisterBatch(batch)

	// Test the List() method which uses WaitGroup.Go()
	hooks := registry.List()

	if len(hooks) != len(batch) {
		t.Fatalf("expected %d hooks from List(), got %d", len(batch), len(hooks))
	}

	// Verify all expected hooks are present
	for key := range batch {
		if _, ok := hooks[key]; !ok {
			t.Fatalf("expected hook %s not found in List() result", key)
		}
	}
}
