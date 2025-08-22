package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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

func getSettingsPath(global bool) (string, error) {
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

func loadSettings(settingsPath string) (*Settings, error) {
	settings := &Settings{
		Other: make(map[string]interface{}),
	}

	// Check if file exists
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		// File doesn't exist, return empty settings
		return settings, nil
	}

	data, err := os.ReadFile(settingsPath)
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

func saveSettings(settingsPath string, settings *Settings) error {
	// Ensure directory exists
	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
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
	if !isHooksConfigEmpty(settings.Hooks) {
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

	if err := os.WriteFile(settingsPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write settings file: %v", err)
	}

	return nil
}

func isHooksConfigEmpty(hooks HooksConfig) bool {
	return len(hooks.PreToolUse) == 0 &&
		len(hooks.PostToolUse) == 0 &&
		len(hooks.UserPromptSubmit) == 0 &&
		len(hooks.Notification) == 0 &&
		len(hooks.Stop) == 0 &&
		len(hooks.SubagentStop) == 0 &&
		len(hooks.PreCompact) == 0 &&
		len(hooks.SessionStart) == 0
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

func addHookToSettings(settings *Settings, event, matcher, command string, timeout *int) {
	hookCmd := HookCommand{
		Type:    "command",
		Command: command,
		Timeout: timeout,
	}

	hookMatcher := HookMatcher{
		Matcher: matcher,
		Hooks:   []HookCommand{hookCmd},
	}

	switch event {
	case "PreToolUse":
		settings.Hooks.PreToolUse = mergeHookMatcher(settings.Hooks.PreToolUse, hookMatcher)
	case "PostToolUse":
		settings.Hooks.PostToolUse = mergeHookMatcher(settings.Hooks.PostToolUse, hookMatcher)
	case "UserPromptSubmit":
		settings.Hooks.UserPromptSubmit = mergeHookMatcher(settings.Hooks.UserPromptSubmit, hookMatcher)
	case "Notification":
		settings.Hooks.Notification = mergeHookMatcher(settings.Hooks.Notification, hookMatcher)
	case "Stop":
		settings.Hooks.Stop = mergeHookMatcher(settings.Hooks.Stop, hookMatcher)
	case "SubagentStop":
		settings.Hooks.SubagentStop = mergeHookMatcher(settings.Hooks.SubagentStop, hookMatcher)
	case "PreCompact":
		settings.Hooks.PreCompact = mergeHookMatcher(settings.Hooks.PreCompact, hookMatcher)
	case "SessionStart":
		settings.Hooks.SessionStart = mergeHookMatcher(settings.Hooks.SessionStart, hookMatcher)
	}
}

func mergeHookMatcher(existing []HookMatcher, new HookMatcher) []HookMatcher {
	// Look for existing matcher
	for i, matcher := range existing {
		if matcher.Matcher == new.Matcher {
			// Append to existing matcher
			existing[i].Hooks = append(existing[i].Hooks, new.Hooks...)
			return existing
		}
	}
	// No existing matcher found, add new one
	return append(existing, new)
}

func getDefaultClaudeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	if runtime.GOOS == "windows" {
		return filepath.Join(homeDir, ".claude")
	}
	return filepath.Join(homeDir, ".claude")
}

func removeHookFromSettings(settings *Settings, command string) bool {
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
