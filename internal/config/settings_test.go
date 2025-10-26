package config

import (
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
