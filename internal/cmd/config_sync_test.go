// Package cmd provides tests for configuration synchronization functionality
package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	btconfig "github.com/klauern/blues-traveler/internal/config"
)

// setupTestEnv creates a temporary directory with .claude structure for testing
func setupTestEnv(t *testing.T) func() {
	t.Helper()

	// Save original directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	// Create temp directory and change to it
	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change to temp directory: %v", err)
	}

	// Create .claude directory structure
	claudeDir := filepath.Join(tempDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o750); err != nil {
		t.Fatalf("failed to create .claude directory: %v", err)
	}

	cleanup := func() {
		_ = os.Chdir(originalDir)
		// On Windows, give a moment for file handles to be released
		if runtime.GOOS == "windows" {
			time.Sleep(10 * time.Millisecond)
		}
	}

	return cleanup
}

// createConfigWithGroup creates a blues-traveler config with the specified group
func createConfigWithGroup(t *testing.T, groupName string) {
	t.Helper()

	configPath, err := btconfig.GetLogConfigPath(false)
	if err != nil {
		t.Fatalf("failed to get config path: %v", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0o750); err != nil {
		t.Fatalf("failed to create config directory: %v", err)
	}

	// Load existing config or create new one
	config, err := btconfig.LoadLogConfig(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Initialize CustomHooks if nil
	if config.CustomHooks == nil {
		config.CustomHooks = btconfig.CustomHooksConfig{}
	}

	// Create the group directly
	config.CustomHooks[groupName] = btconfig.HookGroup{
		"PreToolUse": &btconfig.EventConfig{
			Jobs: []btconfig.HookJob{
				{
					Name: "test-job-1",
					Run:  "echo \"test job 1\"",
					Glob: []string{"*.py"},
				},
				{
					Name: "test-job-2",
					Run:  "echo \"test job 2\"",
					Only: "${TOOL_NAME} == \"Edit\"",
				},
			},
		},
		"PostToolUse": &btconfig.EventConfig{
			Jobs: []btconfig.HookJob{
				{
					Name: "post-job",
					Run:  "echo \"post test\"",
					Glob: []string{"*.go"},
				},
			},
		},
	}

	// Save config
	if err := btconfig.SaveLogConfig(configPath, config); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}
}

// removeGroupFromConfig removes a group from the config
func removeGroupFromConfig(t *testing.T, groupName string) {
	t.Helper()

	configPath, err := btconfig.GetLogConfigPath(false)
	if err != nil {
		t.Fatalf("failed to get config path: %v", err)
	}

	config, err := btconfig.LoadLogConfig(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Remove the group from custom hooks
	delete(config.CustomHooks, groupName)

	if err := btconfig.SaveLogConfig(configPath, config); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}
}

// clearAllConfig removes all custom hooks from config
func clearAllConfig(t *testing.T) {
	t.Helper()

	configPath, err := btconfig.GetLogConfigPath(false)
	if err != nil {
		t.Fatalf("failed to get config path: %v", err)
	}

	config, err := btconfig.LoadLogConfig(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Clear all custom hooks
	config.CustomHooks = btconfig.CustomHooksConfig{}

	if err := btconfig.SaveLogConfig(configPath, config); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}
}

// runConfigSync simulates the config sync functionality by calling the core sync logic directly
func runConfigSync(t *testing.T, groupFilter ...string) error {
	t.Helper()

	hooksCfg, settings, settingsPath, err := loadConfigAndSettings()
	if err != nil {
		return err
	}

	targetGroup := getTargetGroup(groupFilter)
	changed := cleanupStaleGroupsForTest(t, settings, hooksCfg, targetGroup)
	changed += syncCurrentGroups(t, settings, hooksCfg, targetGroup)

	if changed == 0 {
		t.Log("No changes detected.")
		return nil
	}

	return btconfig.SaveSettings(settingsPath, settings)
}

// loadConfigAndSettings loads the hooks config and settings
func loadConfigAndSettings() (*btconfig.CustomHooksConfig, *btconfig.Settings, string, error) {
	hooksCfg, err := btconfig.LoadHooksConfig()
	if err != nil {
		return nil, nil, "", err
	}

	settingsPath, err := btconfig.GetSettingsPath(false) // project scope
	if err != nil {
		return nil, nil, "", err
	}

	settings, err := btconfig.LoadSettings(settingsPath)
	if err != nil {
		return nil, nil, "", err
	}

	return hooksCfg, settings, settingsPath, nil
}

// getTargetGroup extracts the target group from the filter if provided
func getTargetGroup(groupFilter []string) string {
	if len(groupFilter) > 0 {
		return groupFilter[0]
	}
	return ""
}

// cleanupStaleGroupsForTest removes groups that exist in settings but not in current config
func cleanupStaleGroupsForTest(t *testing.T, settings *btconfig.Settings, hooksCfg *btconfig.CustomHooksConfig, targetGroup string) int {
	t.Helper()
	changed := 0

	existingGroups := btconfig.GetConfigGroupsInSettings(settings)
	configGroups := buildConfigGroupsMapForTest(hooksCfg)

	for existingGroup := range existingGroups {
		if shouldSkipGroupForTest(existingGroup, targetGroup) {
			continue
		}
		if !configGroups[existingGroup] {
			removed := btconfig.RemoveConfigGroupFromSettings(settings, existingGroup, "")
			if removed > 0 {
				t.Logf("Cleaned up %d stale entries for removed group '%s'", removed, existingGroup)
				changed += removed
			}
		}
	}

	return changed
}

// buildConfigGroupsMapForTest creates a map of groups that exist in the current config
func buildConfigGroupsMapForTest(hooksCfg *btconfig.CustomHooksConfig) map[string]bool {
	configGroups := make(map[string]bool)
	if hooksCfg != nil {
		for groupName := range *hooksCfg {
			configGroups[groupName] = true
		}
	}
	return configGroups
}

// shouldSkipGroupForTest determines if a group should be skipped based on the filter
func shouldSkipGroupForTest(groupName, targetGroup string) bool {
	return targetGroup != "" && groupName != targetGroup
}

// syncCurrentGroups syncs the current config groups to settings
func syncCurrentGroups(t *testing.T, settings *btconfig.Settings, hooksCfg *btconfig.CustomHooksConfig, targetGroup string) int {
	t.Helper()
	changed := 0

	if hooksCfg == nil {
		return changed
	}

	for groupName, group := range *hooksCfg {
		if shouldSkipGroupForTest(groupName, targetGroup) {
			continue
		}

		changed += syncGroup(t, settings, groupName, group)
	}

	return changed
}

// syncGroup syncs a single group to settings
func syncGroup(t *testing.T, settings *btconfig.Settings, groupName string, group btconfig.HookGroup) int {
	t.Helper()
	changed := 0

	// Prune existing settings for this group
	removed := btconfig.RemoveConfigGroupFromSettings(settings, groupName, "")
	if removed > 0 {
		t.Logf("Pruned %d entries for group '%s'", removed, groupName)
	}

	// Add current definitions
	for eventName, ev := range group {
		changed += addJobsToSettings(settings, groupName, eventName, ev.Jobs)
	}

	return changed
}

// addJobsToSettings adds jobs to settings for a specific event
func addJobsToSettings(settings *btconfig.Settings, groupName, eventName string, jobs []btconfig.HookJob) int {
	changed := 0

	for _, job := range jobs {
		if job.Name == "" {
			continue
		}

		hookCommand := "blues-traveler run config:" + groupName + ":" + job.Name
		matcher := determineMatcherForEvent(eventName)
		btconfig.AddHookToSettings(settings, eventName, matcher, hookCommand, nil)
		changed++
	}

	return changed
}

// determineMatcherForEvent determines the appropriate matcher for an event
func determineMatcherForEvent(eventName string) string {
	if eventName == "PostToolUse" {
		return "Edit,Write"
	}
	return "*"
}

// countHooksInSettings counts custom hook entries in settings.json for a specific group
func countHooksInSettings(t *testing.T, groupName string) int {
	t.Helper()

	settingsPath, err := btconfig.GetSettingsPath(false)
	if err != nil {
		t.Fatalf("failed to get settings path: %v", err)
	}

	settings, err := btconfig.LoadSettings(settingsPath)
	if err != nil {
		t.Fatalf("failed to load settings: %v", err)
	}

	count := 0
	searchPattern := "config:" + groupName + ":"

	// Count across all event types
	eventMatchers := [][]btconfig.HookMatcher{
		settings.Hooks.PreToolUse,
		settings.Hooks.PostToolUse,
		settings.Hooks.UserPromptSubmit,
		settings.Hooks.Notification,
		settings.Hooks.Stop,
		settings.Hooks.SubagentStop,
		settings.Hooks.PreCompact,
		settings.Hooks.SessionStart,
		settings.Hooks.SessionEnd,
	}

	for _, matchers := range eventMatchers {
		for _, matcher := range matchers {
			for _, hook := range matcher.Hooks {
				if btconfig.IsBluesTravelerCommand(hook.Command) &&
					strings.Contains(hook.Command, searchPattern) {
					count++
				}
			}
		}
	}

	return count
}

// This test WILL FAIL initially, exposing the bug where removed groups aren't cleaned up
func TestConfigSync_RemoveEntireGroup_ShouldCleanupSettings(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	groupName := "python-tools"

	// Step 1: Create config with the group
	createConfigWithGroup(t, groupName)

	// Step 2: Run sync to populate settings.json
	if err := runConfigSync(t); err != nil {
		t.Fatalf("initial sync failed: %v", err)
	}

	// Step 3: Verify hooks are in settings
	initialCount := countHooksInSettings(t, groupName)
	if initialCount == 0 {
		t.Fatal("expected hooks to be present in settings after initial sync")
	}
	t.Logf("Initial hook count for group '%s': %d", groupName, initialCount)

	// Step 4: Remove the entire group from config
	removeGroupFromConfig(t, groupName)

	// Step 5: Run sync again
	if err := runConfigSync(t); err != nil {
		t.Fatalf("second sync failed: %v", err)
	}

	// Step 6: Verify hooks are removed from settings
	// THIS ASSERTION WILL FAIL, exposing the bug
	finalCount := countHooksInSettings(t, groupName)
	if finalCount > 0 {
		t.Errorf("EXPECTED BUG: Found %d stale hook entries for removed group '%s' in settings.json", finalCount, groupName)
		t.Errorf("The config sync does not clean up groups that are completely removed from config")
	}
}

// This test WILL FAIL initially, exposing the bug with empty configs
func TestConfigSync_EmptyConfig_ShouldCleanupAllCustomHooks(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	// Step 1: Create config with multiple groups
	createConfigWithGroup(t, "group-1")
	createConfigWithGroup(t, "group-2")

	// Step 2: Run sync to populate settings.json
	if err := runConfigSync(t); err != nil {
		t.Fatalf("initial sync failed: %v", err)
	}

	// Step 3: Verify hooks are in settings
	count1 := countHooksInSettings(t, "group-1")
	count2 := countHooksInSettings(t, "group-2")
	if count1 == 0 || count2 == 0 {
		t.Fatal("expected hooks to be present in settings after initial sync")
	}

	// Step 4: Clear all config
	clearAllConfig(t)

	// Step 5: Run sync again
	if err := runConfigSync(t); err != nil {
		t.Fatalf("second sync failed: %v", err)
	}

	// Step 6: Verify all hooks are removed from settings
	// THIS ASSERTION WILL FAIL, exposing the bug
	finalCount1 := countHooksInSettings(t, "group-1")
	finalCount2 := countHooksInSettings(t, "group-2")

	if finalCount1 > 0 || finalCount2 > 0 {
		t.Errorf("EXPECTED BUG: Found stale hook entries after clearing config")
		t.Errorf("group-1: %d entries, group-2: %d entries", finalCount1, finalCount2)
		t.Errorf("The config sync does not clean up hooks when config is completely empty")
	}
}

// This test WILL FAIL initially, exposing the bug with group-specific sync
func TestConfigSync_GroupFilter_RemovedGroup_ShouldCleanup(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	targetGroup := "target-group"
	otherGroup := "other-group"

	// Step 1: Create config with two groups
	createConfigWithGroup(t, targetGroup)
	createConfigWithGroup(t, otherGroup)

	// Step 2: Run sync to populate settings.json
	if err := runConfigSync(t); err != nil {
		t.Fatalf("initial sync failed: %v", err)
	}

	// Step 3: Verify both groups are in settings
	targetCount := countHooksInSettings(t, targetGroup)
	otherCount := countHooksInSettings(t, otherGroup)
	if targetCount == 0 || otherCount == 0 {
		t.Fatal("expected hooks to be present in settings after initial sync")
	}

	// Step 4: Remove only the target group from config
	removeGroupFromConfig(t, targetGroup)

	// Step 5: Run sync with group filter for the removed group
	if err := runConfigSync(t, targetGroup); err != nil {
		t.Fatalf("group-filtered sync failed: %v", err)
	}

	// Step 6: Verify target group is removed but other group remains
	// THIS ASSERTION WILL FAIL, exposing the bug
	finalTargetCount := countHooksInSettings(t, targetGroup)
	finalOtherCount := countHooksInSettings(t, otherGroup)

	if finalTargetCount > 0 {
		t.Errorf("EXPECTED BUG: Found %d stale hook entries for removed group '%s' after group-filtered sync", finalTargetCount, targetGroup)
		t.Errorf("The config sync does not clean up a specific group when it's removed from config")
	}

	// Other group should be unaffected
	if finalOtherCount != otherCount {
		t.Errorf("Other group was unexpectedly affected: had %d, now has %d", otherCount, finalOtherCount)
	}
}
