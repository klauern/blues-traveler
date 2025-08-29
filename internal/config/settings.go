package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type HookCommand struct {
	Type    string `json:"type"`
	Command string `json:"command"`
	Timeout *int   `json:"timeout,omitempty"`
}

type HookMatcher struct {
	Matcher string        `json:"matcher,omitempty"`
	Hooks   []HookCommand `json:"hooks"`
}

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

type Settings struct {
	Hooks        HooksConfig             `json:"hooks,omitempty"`
	Plugins      map[string]PluginConfig `json:"plugins,omitempty"`
	DefaultModel string                  `json:"defaultModel,omitempty"`
	Other        map[string]interface{}  `json:"-"`
}

func GetSettingsPath(global bool) (string, error) {
	if global {
		// Global settings: ~/.claude/settings.json
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %v", err)
		}
		return filepath.Join(homeDir, ".claude", "settings.json"), nil
	} else {
		// Project settings: ./.claude/settings.json
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %v", err)
		}
		return filepath.Join(cwd, ".claude", "settings.json"), nil
	}
}

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
		return nil, fmt.Errorf("failed to read settings file: %v", err)
	}

	// First unmarshal into a generic map to preserve unknown fields
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse settings JSON: %v", err)
	}

	// Extract known fields
	if err := json.Unmarshal(data, settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings: %v", err)
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

func SaveSettings(settingsPath string, settings *Settings) error {
	// Ensure directory exists
	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
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
		return fmt.Errorf("failed to marshal settings: %v", err)
	}

	if err := os.WriteFile(settingsPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write settings file: %v", err)
	}

	return nil
}

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
func extractHookType(command string) string {
	// Pattern to match "blues-traveler run <hooktype>" with optional flags
	re := regexp.MustCompile(`blues-traveler\s+run\s+(\w+)`)
	matches := re.FindStringSubmatch(command)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// isBluesTravelerCommand checks if a command is a blues-traveler command
func isBluesTravelerCommand(command string) bool {
	return strings.Contains(command, "blues-traveler run")
}

func mergeHookMatcher(existing []HookMatcher, new HookMatcher) MergeResult {
	// Look for existing matcher
	for i, matcher := range existing {
		if matcher.Matcher == new.Matcher {
			// Check for blues-traveler command conflicts within this matcher
			for j, existingHook := range existing[i].Hooks {
				for _, newHook := range new.Hooks {
					// Exact duplicate check
					if existingHook.Command == newHook.Command {
						return MergeResult{
							Matchers:      existing,
							WasDuplicate:  true,
							DuplicateInfo: fmt.Sprintf("Hook command '%s' already exists for matcher '%s'", newHook.Command, matcher.Matcher),
						}
					}

					// Check if both are blues-traveler commands with the same hook type
					if isBluesTravelerCommand(existingHook.Command) && isBluesTravelerCommand(newHook.Command) {
						existingType := extractHookType(existingHook.Command)
						newType := extractHookType(newHook.Command)
						if existingType != "" && existingType == newType {
							// Replace the existing hook with the new one
							existing[i].Hooks[j] = newHook
							return MergeResult{
								Matchers:      existing,
								WasDuplicate:  true,
								DuplicateInfo: fmt.Sprintf("Replaced existing %s hook with updated command for matcher '%s'", newType, matcher.Matcher),
							}
						}
					}
				}
			}
			// No conflicts found, append to existing matcher
			existing[i].Hooks = append(existing[i].Hooks, new.Hooks...)
			return MergeResult{
				Matchers:     existing,
				WasDuplicate: false,
			}
		}
	}
	// No existing matcher found, add new one
	return MergeResult{
		Matchers:     append(existing, new),
		WasDuplicate: false,
	}
}

func RemoveHookFromSettings(settings *Settings, command string) bool {
	removed := false

	settings.Hooks.PreToolUse = removeHookFromMatchers(settings.Hooks.PreToolUse, command, &removed)
	settings.Hooks.PostToolUse = removeHookFromMatchers(settings.Hooks.PostToolUse, command, &removed)
	settings.Hooks.UserPromptSubmit = removeHookFromMatchers(settings.Hooks.UserPromptSubmit, command, &removed)
	settings.Hooks.Notification = removeHookFromMatchers(settings.Hooks.Notification, command, &removed)
	settings.Hooks.Stop = removeHookFromMatchers(settings.Hooks.Stop, command, &removed)
	settings.Hooks.SubagentStop = removeHookFromMatchers(settings.Hooks.SubagentStop, command, &removed)
	settings.Hooks.PreCompact = removeHookFromMatchers(settings.Hooks.PreCompact, command, &removed)
	settings.Hooks.SessionStart = removeHookFromMatchers(settings.Hooks.SessionStart, command, &removed)

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

// IsPluginEnabled checks (project first, then global) settings to see if a plugin is enabled.
// Defaults to enabled if settings cannot be loaded or plugin key absent.
func IsPluginEnabled(pluginKey string) bool {
	// Project settings
	if projectPath, err := GetSettingsPath(false); err == nil {
		if s, err := LoadSettings(projectPath); err == nil {
			if !s.IsPluginEnabled(pluginKey) {
				return false
			}
		}
	}
	// Global settings fallback
	if globalPath, err := GetSettingsPath(true); err == nil {
		if s, err := LoadSettings(globalPath); err == nil {
			if !s.IsPluginEnabled(pluginKey) {
				return false
			}
		}
	}
	return true
}
