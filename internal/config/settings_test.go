package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMatchesHookType(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		hookType string
		want     bool
	}{
		{
			name:     "exact match without flags",
			command:  "/usr/local/bin/blues-traveler hooks run security",
			hookType: "security",
			want:     true,
		},
		{
			name:     "match with --log flag",
			command:  "/usr/local/bin/blues-traveler hooks run security --log",
			hookType: "security",
			want:     true,
		},
		{
			name:     "match with --log --log-format flags",
			command:  "/usr/local/bin/blues-traveler hooks run format --log --log-format pretty",
			hookType: "format",
			want:     true,
		},
		{
			name:     "match with different executable path",
			command:  "/different/path/blues-traveler hooks run audit",
			hookType: "audit",
			want:     true,
		},
		{
			name:     "match with config:group:job pattern",
			command:  "/path/blues-traveler hooks run config:mygroup:myjob",
			hookType: "config:mygroup:myjob",
			want:     true,
		},
		{
			name:     "match with config pattern and flags",
			command:  "/path/blues-traveler hooks run config:test:check --log",
			hookType: "config:test:check",
			want:     true,
		},
		{
			name:     "no match - different hook type",
			command:  "/usr/local/bin/blues-traveler hooks run security",
			hookType: "audit",
			want:     false,
		},
		{
			name:     "no match - partial hook type",
			command:  "/usr/local/bin/blues-traveler hooks run security-test",
			hookType: "security",
			want:     false,
		},
		{
			name:     "no match - not a blues-traveler command",
			command:  "/usr/bin/some-other-tool run security",
			hookType: "security",
			want:     false,
		},
		{
			name:     "no match - empty hook type",
			command:  "/usr/local/bin/blues-traveler hooks run security",
			hookType: "",
			want:     false,
		},
		{
			name:     "match with blues-traveler run (direct)",
			command:  "/path/blues-traveler run security",
			hookType: "security",
			want:     true, // Should match "blues-traveler run" pattern
		},
		{
			name:     "match at end of command",
			command:  "blues-traveler hooks run test",
			hookType: "test",
			want:     true,
		},
		{
			name:     "match with tab after hook type",
			command:  "/path/blues-traveler hooks run debug\t--log",
			hookType: "debug",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesHookType(tt.command, tt.hookType)
			if got != tt.want {
				t.Errorf("matchesHookType(%q, %q) = %v, want %v", tt.command, tt.hookType, got, tt.want)
			}
		})
	}
}

//nolint:gocognit,funlen // Comprehensive table-driven test with extensive test cases
func TestRemoveHookTypeFromSettings(t *testing.T) {
	tests := []struct {
		name     string
		settings *Settings
		hookType string
		want     bool
		check    func(t *testing.T, s *Settings)
	}{
		{
			name: "remove hook with no flags",
			settings: &Settings{
				Hooks: HooksConfig{
					PreToolUse: []HookMatcher{
						{
							Matcher: "*",
							Hooks: []HookCommand{
								{Command: "/path/blues-traveler hooks run security"},
							},
						},
					},
				},
			},
			hookType: "security",
			want:     true,
			check: func(t *testing.T, s *Settings) {
				if len(s.Hooks.PreToolUse) != 0 {
					t.Error("Expected PreToolUse to be empty")
				}
			},
		},
		{
			name: "remove hook with flags",
			settings: &Settings{
				Hooks: HooksConfig{
					PreToolUse: []HookMatcher{
						{
							Matcher: "*",
							Hooks: []HookCommand{
								{Command: "/path/blues-traveler hooks run security --log --log-format pretty"},
							},
						},
					},
				},
			},
			hookType: "security",
			want:     true,
			check: func(t *testing.T, s *Settings) {
				if len(s.Hooks.PreToolUse) != 0 {
					t.Error("Expected PreToolUse to be empty")
				}
			},
		},
		{
			name: "remove hook with different executable path",
			settings: &Settings{
				Hooks: HooksConfig{
					PostToolUse: []HookMatcher{
						{
							Matcher: "Edit,Write",
							Hooks: []HookCommand{
								{Command: "/different/path/blues-traveler hooks run audit"},
							},
						},
					},
				},
			},
			hookType: "audit",
			want:     true,
			check: func(t *testing.T, s *Settings) {
				if len(s.Hooks.PostToolUse) != 0 {
					t.Error("Expected PostToolUse to be empty")
				}
			},
		},
		{
			name: "remove across multiple event types",
			settings: &Settings{
				Hooks: HooksConfig{
					PreToolUse: []HookMatcher{
						{
							Matcher: "*",
							Hooks: []HookCommand{
								{Command: "/path/blues-traveler hooks run security"},
							},
						},
					},
					PostToolUse: []HookMatcher{
						{
							Matcher: "Edit",
							Hooks: []HookCommand{
								{Command: "/path/blues-traveler hooks run security --log"},
							},
						},
					},
				},
			},
			hookType: "security",
			want:     true,
			check: func(t *testing.T, s *Settings) {
				if len(s.Hooks.PreToolUse) != 0 || len(s.Hooks.PostToolUse) != 0 {
					t.Error("Expected both PreToolUse and PostToolUse to be empty")
				}
			},
		},
		{
			name: "preserve other hooks in same matcher",
			settings: &Settings{
				Hooks: HooksConfig{
					PreToolUse: []HookMatcher{
						{
							Matcher: "*",
							Hooks: []HookCommand{
								{Command: "/path/blues-traveler hooks run security"},
								{Command: "/path/blues-traveler hooks run audit"},
							},
						},
					},
				},
			},
			hookType: "security",
			want:     true,
			check: func(t *testing.T, s *Settings) {
				if len(s.Hooks.PreToolUse) != 1 {
					t.Fatal("Expected one matcher")
				}
				if len(s.Hooks.PreToolUse[0].Hooks) != 1 {
					t.Fatal("Expected one hook remaining")
				}
				if !matchesHookType(s.Hooks.PreToolUse[0].Hooks[0].Command, "audit") {
					t.Error("Expected audit hook to remain")
				}
			},
		},
		{
			name: "no match - hook type not found",
			settings: &Settings{
				Hooks: HooksConfig{
					PreToolUse: []HookMatcher{
						{
							Matcher: "*",
							Hooks: []HookCommand{
								{Command: "/path/blues-traveler hooks run security"},
							},
						},
					},
				},
			},
			hookType: "nonexistent",
			want:     false,
			check: func(t *testing.T, s *Settings) {
				if len(s.Hooks.PreToolUse) != 1 {
					t.Error("Expected PreToolUse to remain unchanged")
				}
			},
		},
		{
			name: "remove config:group:job pattern",
			settings: &Settings{
				Hooks: HooksConfig{
					PreToolUse: []HookMatcher{
						{
							Matcher: "*",
							Hooks: []HookCommand{
								{Command: "/path/blues-traveler hooks run config:mygroup:myjob"},
							},
						},
					},
				},
			},
			hookType: "config:mygroup:myjob",
			want:     true,
			check: func(t *testing.T, s *Settings) {
				if len(s.Hooks.PreToolUse) != 0 {
					t.Error("Expected PreToolUse to be empty")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemoveHookTypeFromSettings(tt.settings, tt.hookType)
			if got != tt.want {
				t.Errorf("RemoveHookTypeFromSettings() = %v, want %v", got, tt.want)
			}
			if tt.check != nil {
				tt.check(t, tt.settings)
			}
		})
	}
}

//nolint:gocognit,funlen // Comprehensive table-driven test with extensive test cases
func TestSettingsPrecedence(t *testing.T) {
	// Save original working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("Failed to restore working directory: %v", err)
		}
	}()

	// Create temp directories for global and project settings
	tempDir := t.TempDir()
	globalDir := filepath.Join(tempDir, "global")
	projectDir := filepath.Join(tempDir, "project")

	if err := os.MkdirAll(filepath.Join(globalDir, ".claude"), 0o755); err != nil {
		t.Fatalf("Failed to create global .claude dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(projectDir, ".claude"), 0o755); err != nil {
		t.Fatalf("Failed to create project .claude dir: %v", err)
	}

	// Helper to write settings file
	writeSettings := func(dir string, settings *Settings) error {
		path := filepath.Join(dir, ".claude", "settings.json")
		data, err := json.MarshalIndent(settings, "", "  ")
		if err != nil {
			return err
		}
		return os.WriteFile(path, data, 0o644)
	}

	// Helper to create bool pointer
	boolPtr := func(b bool) *bool {
		return &b
	}

	tests := []struct {
		name            string
		globalSettings  *Settings
		projectSettings *Settings
		pluginKey       string
		want            bool
	}{
		{
			name: "project enabled + global disabled = enabled",
			globalSettings: &Settings{
				Plugins: map[string]PluginConfig{
					"security": {Enabled: boolPtr(false)},
				},
			},
			projectSettings: &Settings{
				Plugins: map[string]PluginConfig{
					"security": {Enabled: boolPtr(true)},
				},
			},
			pluginKey: "security",
			want:      true,
		},
		{
			name: "project disabled + global enabled = disabled",
			globalSettings: &Settings{
				Plugins: map[string]PluginConfig{
					"security": {Enabled: boolPtr(true)},
				},
			},
			projectSettings: &Settings{
				Plugins: map[string]PluginConfig{
					"security": {Enabled: boolPtr(false)},
				},
			},
			pluginKey: "security",
			want:      false,
		},
		{
			name: "project nil + global enabled = enabled",
			globalSettings: &Settings{
				Plugins: map[string]PluginConfig{
					"security": {Enabled: boolPtr(true)},
				},
			},
			projectSettings: &Settings{
				Plugins: map[string]PluginConfig{
					// No security plugin configured
				},
			},
			pluginKey: "security",
			want:      true,
		},
		{
			name: "project nil + global disabled = disabled",
			globalSettings: &Settings{
				Plugins: map[string]PluginConfig{
					"security": {Enabled: boolPtr(false)},
				},
			},
			projectSettings: &Settings{
				Plugins: map[string]PluginConfig{
					// No security plugin configured
				},
			},
			pluginKey: "security",
			want:      false,
		},
		{
			name: "project nil + global nil = default enabled",
			globalSettings: &Settings{
				Plugins: map[string]PluginConfig{
					// No security plugin configured
				},
			},
			projectSettings: &Settings{
				Plugins: map[string]PluginConfig{
					// No security plugin configured
				},
			},
			pluginKey: "security",
			want:      true, // Default is enabled
		},
		{
			name: "project enabled + global nil = enabled",
			globalSettings: &Settings{
				Plugins: map[string]PluginConfig{
					// No security plugin configured
				},
			},
			projectSettings: &Settings{
				Plugins: map[string]PluginConfig{
					"security": {Enabled: boolPtr(true)},
				},
			},
			pluginKey: "security",
			want:      true,
		},
		{
			name: "project disabled + global nil = disabled",
			globalSettings: &Settings{
				Plugins: map[string]PluginConfig{
					// No security plugin configured
				},
			},
			projectSettings: &Settings{
				Plugins: map[string]PluginConfig{
					"security": {Enabled: boolPtr(false)},
				},
			},
			pluginKey: "security",
			want:      false,
		},
		{
			name: "project nil + global enabled for different plugin = default enabled",
			globalSettings: &Settings{
				Plugins: map[string]PluginConfig{
					"format": {Enabled: boolPtr(true)},
				},
			},
			projectSettings: &Settings{
				Plugins: map[string]PluginConfig{
					"audit": {Enabled: boolPtr(true)},
				},
			},
			pluginKey: "security",
			want:      true, // Default is enabled when plugin not found
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set HOME to global directory (Unix-like systems)
			t.Setenv("HOME", globalDir)
			// Set USERPROFILE for Windows compatibility
			t.Setenv("USERPROFILE", globalDir)

			// Write global settings
			if err := writeSettings(globalDir, tt.globalSettings); err != nil {
				t.Fatalf("Failed to write global settings: %v", err)
			}

			// Write project settings
			if err := writeSettings(projectDir, tt.projectSettings); err != nil {
				t.Fatalf("Failed to write project settings: %v", err)
			}

			// Change to project directory
			if err := os.Chdir(projectDir); err != nil {
				t.Fatalf("Failed to change to project directory: %v", err)
			}

			// Test IsPluginEnabled
			got := IsPluginEnabled(tt.pluginKey)
			if got != tt.want {
				t.Errorf("IsPluginEnabled(%q) = %v, want %v", tt.pluginKey, got, tt.want)
			}
		})
	}
}
