package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// HookCommand represents a single hook command configuration with type, command, and optional timeout
type HookCommand struct {
	Type    string `json:"type"`
	Command string `json:"command"`
	Timeout *int   `json:"timeout,omitempty"`
}

// HookMatcher represents a matcher pattern with associated hook commands
type HookMatcher struct {
	Matcher string        `json:"matcher,omitempty"`
	Hooks   []HookCommand `json:"hooks"`
}

// HooksConfig represents the hooks configuration organized by event type
type HooksConfig struct {
	PreToolUse       []HookMatcher `json:"PreToolUse,omitempty"`
	PostToolUse      []HookMatcher `json:"PostToolUse,omitempty"`
	UserPromptSubmit []HookMatcher `json:"UserPromptSubmit,omitempty"`
	Notification     []HookMatcher `json:"Notification,omitempty"`
	Stop             []HookMatcher `json:"Stop,omitempty"`
	SubagentStop     []HookMatcher `json:"SubagentStop,omitempty"`
	PreCompact       []HookMatcher `json:"PreCompact,omitempty"`
	SessionStart     []HookMatcher `json:"SessionStart,omitempty"`
	SessionEnd       []HookMatcher `json:"SessionEnd,omitempty"`
}

// PluginConfig stores per-plugin settings (extendable later with plugin-specific fields).
// A nil Enabled means default (enabled). If Enabled=false, the plugin is disabled.
type PluginConfig struct {
	Enabled *bool `json:"enabled,omitempty"`
}

// Settings represents the complete settings structure including hooks, plugins, and other configuration
type Settings struct {
	Hooks        HooksConfig             `json:"hooks,omitempty"`
	Plugins      map[string]PluginConfig `json:"plugins,omitempty"`
	DefaultModel string                  `json:"defaultModel,omitempty"`
	Other        map[string]interface{}  `json:"-"`
}

// GetSettingsPath returns the path to the settings file (global or project-specific)
func GetSettingsPath(global bool) (string, error) {
	if global {
		// Global settings: ~/.claude/settings.json
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		return filepath.Join(homeDir, ".claude", "settings.json"), nil
	}
	// Project settings: ./.claude/settings.json
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}
	return filepath.Join(cwd, ".claude", "settings.json"), nil
}

// LoadSettings loads settings from the specified path, preserving unknown JSON fields
func LoadSettings(settingsPath string) (*Settings, error) {
	settings := &Settings{
		Other: make(map[string]interface{}),
	}

	// Check if file exists
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		// File doesn't exist, return empty settings
		return settings, nil
	}

	data, err := os.ReadFile(settingsPath) // #nosec G304 - controlled settings paths
	if err != nil {
		return nil, fmt.Errorf("failed to read settings file: %w", err)
	}

	// First unmarshal into a generic map to preserve unknown fields
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse settings JSON: %w", err)
	}

	// Extract known fields
	if err := json.Unmarshal(data, settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings: %w", err)
	}

	// Store unknown fields (remove known keys first)
	delete(raw, "hooks")
	delete(raw, "plugins")
	delete(raw, "defaultModel")
	settings.Other = raw

	// Ensure maps initialized
	if settings.Plugins == nil {
		settings.Plugins = make(map[string]PluginConfig)
	}

	return settings, nil
}

// writeFileAtomic writes data to a file atomically by writing to a temp file
// and then renaming it. This prevents corruption from power loss or crashes.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	// Create temp file in same directory to ensure same filesystem
	dir := filepath.Dir(path)
	tempFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()

	// Clean up temp file on error
	defer func() {
		if tempFile != nil {
			tempFile.Close()
			os.Remove(tempPath)
		}
	}()

	// Write data to temp file
	if _, err := tempFile.Write(data); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Sync to disk to ensure data is persisted
	if err := tempFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	// Close temp file before rename
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	tempFile = nil // Prevent defer from closing again

	// Set permissions before rename
	if err := os.Chmod(tempPath, perm); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Atomic rename (on POSIX systems)
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// SaveSettings saves settings to the specified path with proper formatting
func SaveSettings(settingsPath string, settings *Settings) error {
	// Ensure directory exists
	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create directory %s: %v\n", dir, err)
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Merge known and unknown fields
	output := make(map[string]interface{})

	// Add other fields first
	for k, v := range settings.Other {
		output[k] = v
	}

	// Add known fields
	if settings.DefaultModel != "" {
		output["defaultModel"] = settings.DefaultModel
	}

	// Only add hooks if they're not empty
	if !IsHooksConfigEmpty(settings.Hooks) {
		output["hooks"] = settings.Hooks
	}

	// Only add plugins if non-empty
	if len(settings.Plugins) > 0 {
		output["plugins"] = settings.Plugins
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := writeFileAtomic(settingsPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
}

// IsHooksConfigEmpty returns true if the hooks configuration has no hooks defined for any event
func IsHooksConfigEmpty(hooks HooksConfig) bool {
	return len(hooks.PreToolUse) == 0 &&
		len(hooks.PostToolUse) == 0 &&
		len(hooks.UserPromptSubmit) == 0 &&
		len(hooks.Notification) == 0 &&
		len(hooks.Stop) == 0 &&
		len(hooks.SubagentStop) == 0 &&
		len(hooks.PreCompact) == 0 &&
		len(hooks.SessionStart) == 0 &&
		len(hooks.SessionEnd) == 0
}

// IsPluginEnabled returns true if the plugin is enabled (default) or explicitly enabled.
// Returns false if explicitly disabled in settings.
func (s *Settings) IsPluginEnabled(key string) bool {
	if s == nil {
		return true
	}
	if s.Plugins == nil {
		return true
	}
	cfg, ok := s.Plugins[key]
	if !ok || cfg.Enabled == nil {
		return true
	}
	return *cfg.Enabled
}

// AddHookToSettings adds or merges a hook into settings for the specified event
func AddHookToSettings(settings *Settings, event, matcher, command string, timeout *int) MergeResult {
	hookCmd := HookCommand{
		Type:    "command",
		Command: command,
		Timeout: timeout,
	}

	hookMatcher := HookMatcher{
		Matcher: matcher,
		Hooks:   []HookCommand{hookCmd},
	}

	// Resolve Cursor event aliases to canonical names
	// This allows hooks to be installed using Cursor event names
	// Note: We don't import core package to avoid circular dependency,
	// so alias resolution happens at the command level, not here.
	// This function receives already-resolved canonical event names.

	var result MergeResult
	switch event {
	case "PreToolUse":
		result = mergeHookMatcher(settings.Hooks.PreToolUse, hookMatcher)
		settings.Hooks.PreToolUse = result.Matchers
	case "PostToolUse":
		result = mergeHookMatcher(settings.Hooks.PostToolUse, hookMatcher)
		settings.Hooks.PostToolUse = result.Matchers
	case "UserPromptSubmit":
		result = mergeHookMatcher(settings.Hooks.UserPromptSubmit, hookMatcher)
		settings.Hooks.UserPromptSubmit = result.Matchers
	case "Notification":
		result = mergeHookMatcher(settings.Hooks.Notification, hookMatcher)
		settings.Hooks.Notification = result.Matchers
	case "Stop":
		result = mergeHookMatcher(settings.Hooks.Stop, hookMatcher)
		settings.Hooks.Stop = result.Matchers
	case "SubagentStop":
		result = mergeHookMatcher(settings.Hooks.SubagentStop, hookMatcher)
		settings.Hooks.SubagentStop = result.Matchers
	case "PreCompact":
		result = mergeHookMatcher(settings.Hooks.PreCompact, hookMatcher)
		settings.Hooks.PreCompact = result.Matchers
	case "SessionStart":
		result = mergeHookMatcher(settings.Hooks.SessionStart, hookMatcher)
		settings.Hooks.SessionStart = result.Matchers
	case "SessionEnd":
		result = mergeHookMatcher(settings.Hooks.SessionEnd, hookMatcher)
		settings.Hooks.SessionEnd = result.Matchers
	}
	return result
}

// MergeResult represents the result of merging hook matchers
type MergeResult struct {
	Matchers      []HookMatcher
	WasDuplicate  bool
	DuplicateInfo string
}

// extractHookType extracts the hook type from a blues-traveler command
// Example: "/path/to/blues-traveler run debug --log" -> "debug"
// Also handles: "/path/to/blues-traveler hooks run debug --log" -> "debug"
func extractHookType(command string) string {
	// Match both "blues-traveler run" and "blues-traveler hooks run" patterns
	// This correctly captures config hooks like 'config:python:post-sample'.
	re := regexp.MustCompile(`blues-traveler\s+(?:hooks\s+)?run\s+([^\s]+)`) // capture until whitespace
	matches := re.FindStringSubmatch(command)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// isBluesTravelerCommand checks if a command is a blues-traveler command
func isBluesTravelerCommand(command string) bool {
	return strings.Contains(command, "blues-traveler run") || strings.Contains(command, "blues-traveler hooks run")
}

// checkExactDuplicate checks if a hook command is an exact duplicate
func checkExactDuplicate(existingHook HookCommand, newHook HookCommand, matcherName string) *MergeResult {
	if existingHook.Command == newHook.Command {
		return &MergeResult{
			Matchers:      nil,
			WasDuplicate:  true,
			DuplicateInfo: fmt.Sprintf("Hook command '%s' already exists for matcher '%s'", newHook.Command, matcherName),
		}
	}
	return nil
}

// checkBluesTravelerConflict checks if two blues-traveler hooks conflict
// Creates a copy of the input slice to avoid side effects on the original data
func checkBluesTravelerConflict(existingHook HookCommand, newHook HookCommand, matcherName string, matcherIndex, hookIndex int, existing []HookMatcher) *MergeResult {
	if !isBluesTravelerCommand(existingHook.Command) || !isBluesTravelerCommand(newHook.Command) {
		return nil
	}

	existingType := extractHookType(existingHook.Command)
	newType := extractHookType(newHook.Command)

	if existingType != "" && existingType == newType {
		// Create a copy of the existing matchers to avoid mutating the input
		result := make([]HookMatcher, len(existing))
		copy(result, existing)

		// Copy the hooks slice for the affected matcher
		result[matcherIndex].Hooks = make([]HookCommand, len(existing[matcherIndex].Hooks))
		copy(result[matcherIndex].Hooks, existing[matcherIndex].Hooks)

		// Replace the existing hook with the new one in the copy
		result[matcherIndex].Hooks[hookIndex] = newHook

		return &MergeResult{
			Matchers:      result,
			WasDuplicate:  true,
			DuplicateInfo: fmt.Sprintf("Replaced existing %s hook with updated command for matcher '%s'", newType, matcherName),
		}
	}

	return nil
}

// checkHookConflicts checks for conflicts between existing and new hooks
func checkHookConflicts(existing []HookMatcher, newMatcher HookMatcher, matcherIndex int) *MergeResult {
	for j, existingHook := range existing[matcherIndex].Hooks {
		for _, newHook := range newMatcher.Hooks {
			// Exact duplicate check
			if result := checkExactDuplicate(existingHook, newHook, existing[matcherIndex].Matcher); result != nil {
				result.Matchers = existing
				return result
			}

			// Check if both are blues-traveler commands with the same hook type
			if result := checkBluesTravelerConflict(existingHook, newHook, existing[matcherIndex].Matcher, matcherIndex, j, existing); result != nil {
				return result
			}
		}
	}
	return nil
}

func mergeHookMatcher(existing []HookMatcher, newMatcher HookMatcher) MergeResult {
	// Look for existing matcher
	for i, matcher := range existing {
		if matcher.Matcher == newMatcher.Matcher {
			// Check for conflicts
			if result := checkHookConflicts(existing, newMatcher, i); result != nil {
				return *result
			}

			// No conflicts found, append to existing matcher
			existing[i].Hooks = append(existing[i].Hooks, newMatcher.Hooks...)
			return MergeResult{
				Matchers:     existing,
				WasDuplicate: false,
			}
		}
	}

	// No existing matcher found, add new one
	return MergeResult{
		Matchers:     append(existing, newMatcher),
		WasDuplicate: false,
	}
}

// RemoveHookFromSettings removes all occurrences of a specific hook command from settings
func RemoveHookFromSettings(settings *Settings, command string) bool {
	return removeFromAllEvents(settings, func(matchers []HookMatcher, removed *bool) []HookMatcher {
		return removeHookFromMatchers(matchers, command, removed)
	})
}

// RemoveHookTypeFromSettings removes all hooks matching a hook type pattern.
// This handles cases where hooks were installed with flags (--log, --format) or
// when the executable path has changed.
func RemoveHookTypeFromSettings(settings *Settings, hookType string) bool {
	return removeFromAllEvents(settings, func(matchers []HookMatcher, removed *bool) []HookMatcher {
		return removeHookTypeFromMatchers(matchers, hookType, removed)
	})
}

// removeFromAllEvents applies a removal function to all event types in settings
func removeFromAllEvents(settings *Settings, removalFn func([]HookMatcher, *bool) []HookMatcher) bool {
	removed := false

	settings.Hooks.PreToolUse = removalFn(settings.Hooks.PreToolUse, &removed)
	settings.Hooks.PostToolUse = removalFn(settings.Hooks.PostToolUse, &removed)
	settings.Hooks.UserPromptSubmit = removalFn(settings.Hooks.UserPromptSubmit, &removed)
	settings.Hooks.Notification = removalFn(settings.Hooks.Notification, &removed)
	settings.Hooks.Stop = removalFn(settings.Hooks.Stop, &removed)
	settings.Hooks.SubagentStop = removalFn(settings.Hooks.SubagentStop, &removed)
	settings.Hooks.PreCompact = removalFn(settings.Hooks.PreCompact, &removed)
	settings.Hooks.SessionStart = removalFn(settings.Hooks.SessionStart, &removed)
	settings.Hooks.SessionEnd = removalFn(settings.Hooks.SessionEnd, &removed)

	return removed
}

func removeHookFromMatchers(matchers []HookMatcher, command string, removed *bool) []HookMatcher {
	var result []HookMatcher

	for _, matcher := range matchers {
		var filteredHooks []HookCommand
		for _, hook := range matcher.Hooks {
			if hook.Command != command {
				filteredHooks = append(filteredHooks, hook)
			} else {
				*removed = true
			}
		}

		// Only keep matcher if it still has hooks
		if len(filteredHooks) > 0 {
			matcher.Hooks = filteredHooks
			result = append(result, matcher)
		}
	}

	return result
}

func removeHookTypeFromMatchers(matchers []HookMatcher, hookType string, removed *bool) []HookMatcher {
	var result []HookMatcher

	for _, matcher := range matchers {
		var filteredHooks []HookCommand
		for _, hook := range matcher.Hooks {
			// Check if this is a blues-traveler command matching the hook type
			// Matches: "hooks run <hookType>" or "blues-traveler run <hookType>"
			// Ignores: executable path, flags like --log, --log-format
			if !matchesHookType(hook.Command, hookType) {
				filteredHooks = append(filteredHooks, hook)
			} else {
				*removed = true
			}
		}

		// Only keep matcher if it still has hooks
		if len(filteredHooks) > 0 {
			matcher.Hooks = filteredHooks
			result = append(result, matcher)
		}
	}

	return result
}

// matchesHookType checks if a command matches a hook type pattern
// Example: matchesHookType("/path/blues-traveler hooks run security --log", "security") -> true
// Example: matchesHookType("/path/blues-traveler run security --log", "security") -> true
// Example: matchesHookType("/path/blues-traveler hooks run config:group:job", "config:group:job") -> true
func matchesHookType(command, hookType string) bool {
	// Must be a blues-traveler command
	if !IsBluesTravelerCommand(command) {
		return false
	}

	// Look for "hooks run <hookType>" or "blues-traveler run <hookType>" pattern
	// The command may have additional flags after the hook type
	patterns := []string{"hooks run " + hookType, "blues-traveler run " + hookType}
	idx := -1
	patternLen := 0

	for _, pattern := range patterns {
		if foundIdx := strings.Index(command, pattern); foundIdx != -1 {
			idx = foundIdx
			patternLen = len(pattern)
			break
		}
	}

	if idx == -1 {
		return false
	}

	// Verify it's either at the end or followed by a space (for flags)
	endIdx := idx + patternLen
	if endIdx == len(command) {
		return true
	}
	if endIdx < len(command) && (command[endIdx] == ' ' || command[endIdx] == '\t') {
		return true
	}

	return false
}

// CountBluesTravelerInSettings counts all blues-traveler commands in the settings
func CountBluesTravelerInSettings(settings *Settings) int {
	count := 0

	// Define a helper function to count hooks in a slice of matchers
	countInMatchers := func(matchers []HookMatcher) int {
		c := 0
		for _, matcher := range matchers {
			for _, hook := range matcher.Hooks {
				if IsBluesTravelerCommand(hook.Command) {
					c++
				}
			}
		}
		return c
	}

	count += countInMatchers(settings.Hooks.PreToolUse)
	count += countInMatchers(settings.Hooks.PostToolUse)
	count += countInMatchers(settings.Hooks.UserPromptSubmit)
	count += countInMatchers(settings.Hooks.Notification)
	count += countInMatchers(settings.Hooks.Stop)
	count += countInMatchers(settings.Hooks.SubagentStop)
	count += countInMatchers(settings.Hooks.PreCompact)
	count += countInMatchers(settings.Hooks.SessionStart)

	return count
}

// IsBluesTravelerCommand checks if a command is from blues-traveler
func IsBluesTravelerCommand(command string) bool {
	return strings.Contains(command, "blues-traveler run") || strings.Contains(command, "hooks run")
}

// PrintBluesTravelerToRemove shows which blues-traveler hooks will be removed
func PrintBluesTravelerToRemove(settings *Settings) {
	// Define a helper function to print hooks from a slice of matchers
	printFromMatchers := func(eventName string, matchers []HookMatcher) {
		found := false
		for _, matcher := range matchers {
			for _, hook := range matcher.Hooks {
				if IsBluesTravelerCommand(hook.Command) {
					if !found {
						fmt.Printf("%s:\n", eventName)
						found = true
					}
					fmt.Printf("  Matcher: %s\n", matcher.Matcher)
					fmt.Printf("    - %s\n", hook.Command)
				}
			}
		}
		if found {
			fmt.Println()
		}
	}

	printFromMatchers("PreToolUse", settings.Hooks.PreToolUse)
	printFromMatchers("PostToolUse", settings.Hooks.PostToolUse)
	printFromMatchers("UserPromptSubmit", settings.Hooks.UserPromptSubmit)
	printFromMatchers("Notification", settings.Hooks.Notification)
	printFromMatchers("Stop", settings.Hooks.Stop)
	printFromMatchers("SubagentStop", settings.Hooks.SubagentStop)
	printFromMatchers("PreCompact", settings.Hooks.PreCompact)
	printFromMatchers("SessionStart", settings.Hooks.SessionStart)
}

// RemoveAllBluesTravelerFromSettings removes all blues-traveler hooks from settings and returns count removed
func RemoveAllBluesTravelerFromSettings(settings *Settings) int {
	removed := 0

	// Define a helper function to remove blues-traveler hooks from a slice of matchers
	removeFromMatchers := func(matchers []HookMatcher) []HookMatcher {
		var result []HookMatcher

		for _, matcher := range matchers {
			var filteredHooks []HookCommand

			// Keep only non-blues-traveler hooks
			for _, hook := range matcher.Hooks {
				if !IsBluesTravelerCommand(hook.Command) {
					filteredHooks = append(filteredHooks, hook)
				} else {
					removed++
				}
			}

			// Only keep the matcher if it has remaining hooks
			if len(filteredHooks) > 0 {
				matcher.Hooks = filteredHooks
				result = append(result, matcher)
			}
		}

		return result
	}

	settings.Hooks.PreToolUse = removeFromMatchers(settings.Hooks.PreToolUse)
	settings.Hooks.PostToolUse = removeFromMatchers(settings.Hooks.PostToolUse)
	settings.Hooks.UserPromptSubmit = removeFromMatchers(settings.Hooks.UserPromptSubmit)
	settings.Hooks.Notification = removeFromMatchers(settings.Hooks.Notification)
	settings.Hooks.Stop = removeFromMatchers(settings.Hooks.Stop)
	settings.Hooks.SubagentStop = removeFromMatchers(settings.Hooks.SubagentStop)
	settings.Hooks.PreCompact = removeFromMatchers(settings.Hooks.PreCompact)
	settings.Hooks.SessionStart = removeFromMatchers(settings.Hooks.SessionStart)

	return removed
}

// RemoveConfigGroupFromSettings removes all blues-traveler config:<group>:* commands
// from the specified event (or from all events if event == ""). Returns count removed.
func RemoveConfigGroupFromSettings(settings *Settings, group string, event string) int {
	if settings == nil || group == "" {
		return 0
	}

	removed := 0
	matchPattern := "config:" + group + ":"

	// Create filter function that removes matching hooks
	filter := makeConfigGroupFilter(matchPattern, &removed)

	// Apply filter to specified event or all events
	if event == "" {
		filterAllEvents(settings, filter)
	} else {
		filterSingleEvent(settings, event, filter)
	}

	return removed
}

// makeConfigGroupFilter creates a filter function that removes hooks matching a config group
func makeConfigGroupFilter(matchPattern string, removed *int) func([]HookMatcher) []HookMatcher {
	return func(matchers []HookMatcher) []HookMatcher {
		var result []HookMatcher
		for _, m := range matchers {
			var hooks []HookCommand
			for _, h := range m.Hooks {
				if IsBluesTravelerCommand(h.Command) && strings.Contains(h.Command, matchPattern) {
					*removed++
					continue
				}
				hooks = append(hooks, h)
			}
			if len(hooks) > 0 {
				m.Hooks = hooks
				result = append(result, m)
			}
		}
		return result
	}
}

// filterAllEvents applies the filter to all event types
func filterAllEvents(settings *Settings, filter func([]HookMatcher) []HookMatcher) {
	settings.Hooks.PreToolUse = filter(settings.Hooks.PreToolUse)
	settings.Hooks.PostToolUse = filter(settings.Hooks.PostToolUse)
	settings.Hooks.UserPromptSubmit = filter(settings.Hooks.UserPromptSubmit)
	settings.Hooks.Notification = filter(settings.Hooks.Notification)
	settings.Hooks.Stop = filter(settings.Hooks.Stop)
	settings.Hooks.SubagentStop = filter(settings.Hooks.SubagentStop)
	settings.Hooks.PreCompact = filter(settings.Hooks.PreCompact)
	settings.Hooks.SessionStart = filter(settings.Hooks.SessionStart)
	settings.Hooks.SessionEnd = filter(settings.Hooks.SessionEnd)
}

// filterSingleEvent applies the filter to a specific event type
func filterSingleEvent(settings *Settings, event string, filter func([]HookMatcher) []HookMatcher) {
	switch event {
	case "PreToolUse":
		settings.Hooks.PreToolUse = filter(settings.Hooks.PreToolUse)
	case "PostToolUse":
		settings.Hooks.PostToolUse = filter(settings.Hooks.PostToolUse)
	case "UserPromptSubmit":
		settings.Hooks.UserPromptSubmit = filter(settings.Hooks.UserPromptSubmit)
	case "Notification":
		settings.Hooks.Notification = filter(settings.Hooks.Notification)
	case "Stop":
		settings.Hooks.Stop = filter(settings.Hooks.Stop)
	case "SubagentStop":
		settings.Hooks.SubagentStop = filter(settings.Hooks.SubagentStop)
	case "PreCompact":
		settings.Hooks.PreCompact = filter(settings.Hooks.PreCompact)
	case "SessionStart":
		settings.Hooks.SessionStart = filter(settings.Hooks.SessionStart)
	case "SessionEnd":
		settings.Hooks.SessionEnd = filter(settings.Hooks.SessionEnd)
	}
}

// GetConfigGroupsInSettings returns a set of all config group names found in settings
func GetConfigGroupsInSettings(settings *Settings) map[string]bool {
	groups := make(map[string]bool)
	if settings == nil {
		return groups
	}

	// Get all matcher slices from hooks
	allMatchers := getAllHookMatchers(&settings.Hooks)

	// Extract group names from all matchers
	for _, matchers := range allMatchers {
		extractGroupsFromMatchers(matchers, groups)
	}

	return groups
}

// getAllHookMatchers returns all hook matcher slices from a HooksConfig
func getAllHookMatchers(hooks *HooksConfig) [][]HookMatcher {
	return [][]HookMatcher{
		hooks.PreToolUse,
		hooks.PostToolUse,
		hooks.UserPromptSubmit,
		hooks.Notification,
		hooks.Stop,
		hooks.SubagentStop,
		hooks.PreCompact,
		hooks.SessionStart,
		hooks.SessionEnd,
	}
}

// extractGroupsFromMatchers extracts config group names from hook matchers
func extractGroupsFromMatchers(matchers []HookMatcher, groups map[string]bool) {
	for _, matcher := range matchers {
		for _, hook := range matcher.Hooks {
			if groupName := extractConfigGroupName(hook.Command); groupName != "" {
				groups[groupName] = true
			}
		}
	}
}

// extractConfigGroupName extracts the group name from a config hook command
// Returns empty string if not a config hook command
func extractConfigGroupName(command string) string {
	if !IsBluesTravelerCommand(command) || !strings.Contains(command, "config:") {
		return ""
	}

	// Extract group name from "blues-traveler run config:groupname:jobname"
	// Find the "config:" substring to avoid splitting Windows paths like "C:\..."
	configIdx := strings.Index(command, "config:")
	if configIdx == -1 {
		return ""
	}

	// Parse only from the "config:" part onwards
	configPart := command[configIdx+len("config:"):]
	parts := strings.Split(configPart, ":")
	if len(parts) >= 2 && parts[0] != "" {
		return parts[0]
	}
	return ""
}

// IsPluginEnabled checks (project first, then global) settings to see if a plugin is enabled.
// Defaults to enabled if settings cannot be loaded or plugin key absent.
func IsPluginEnabled(pluginKey string) bool {
	// Project settings - if explicit value is set, use it
	if projectPath, err := GetSettingsPath(false); err == nil {
		if s, err := LoadSettings(projectPath); err == nil {
			if cfg, ok := s.Plugins[pluginKey]; ok && cfg.Enabled != nil {
				return *cfg.Enabled
			}
		}
	}
	// Global settings fallback - if explicit value is set, use it
	if globalPath, err := GetSettingsPath(true); err == nil {
		if s, err := LoadSettings(globalPath); err == nil {
			if cfg, ok := s.Plugins[pluginKey]; ok && cfg.Enabled != nil {
				return *cfg.Enabled
			}
		}
	}
	return true
}
