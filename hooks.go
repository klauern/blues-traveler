package main

// Built-in hook plugin implementations are registered in plugin.go via init().
// This file now only contains the concrete hook runner functions invoked by the registered plugins.

// isPluginEnabled checks (project first, then global) settings to see if a plugin is enabled.
// Defaults to enabled if settings cannot be loaded or plugin key absent.
func isPluginEnabled(pluginKey string) bool {
	// Project settings
	if projectPath, err := getSettingsPath(false); err == nil {
		if s, err := loadSettings(projectPath); err == nil {
			if !s.IsPluginEnabled(pluginKey) {
				return false
			}
		}
	}
	// Global settings fallback
	if globalPath, err := getSettingsPath(true); err == nil {
		if s, err := loadSettings(globalPath); err == nil {
			if !s.IsPluginEnabled(pluginKey) {
				return false
			}
		}
	}
	return true
}

// runSecurityHook implements security blocking logic (enhanced for macOS)
// Strategy:
//  1. Parse Bash command
//  2. Run a series of detectors (token based + regex patterns)
//  3. Block immediately on high‑risk destructive / persistence / system reconfiguration ops
//  4. Provide specific rationale to aid user correction
