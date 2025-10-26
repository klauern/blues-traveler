package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	btconfig "github.com/klauern/blues-traveler/internal/config"
)

func TestRunDoctorCheck_NoConfiguration(t *testing.T) {
	// Setup temp project with .claude
	cwd, _ := os.Getwd()
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
		// On Windows, give a moment for file handles to be released
		if runtime.GOOS == "windows" {
			time.Sleep(10 * time.Millisecond)
		}
	})
	dir := t.TempDir()
	_ = os.Chdir(dir)
	_ = os.MkdirAll(filepath.Join(dir, ".claude"), 0o755)

	// Run doctor check - should not error even with no configuration
	err := runDoctorCheck(false)
	if err != nil {
		t.Errorf("runDoctorCheck failed: %v", err)
	}

	// Ensure we're back in original directory before cleanup
	_ = os.Chdir(cwd)
}

func TestRunDoctorCheck_WithHooksConfig(t *testing.T) {
	// Setup temp project with .claude
	cwd, _ := os.Getwd()
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
		// On Windows, give a moment for file handles to be released
		if runtime.GOOS == "windows" {
			time.Sleep(10 * time.Millisecond)
		}
	})
	dir := t.TempDir()
	_ = os.Chdir(dir)
	_ = os.MkdirAll(filepath.Join(dir, ".claude"), 0o755)

	// Write a sample hooks.yml
	hooks := `test-group:
  PreToolUse:
    jobs:
      - name: test-job
        run: echo "test"
        glob: ["*"]
`
	if _, err := btconfig.WriteSampleHooksConfig(false, hooks, true); err != nil {
		t.Fatalf("write hooks.yml: %v", err)
	}

	// Run doctor check
	err := runDoctorCheck(false)
	if err != nil {
		t.Errorf("runDoctorCheck failed: %v", err)
	}

	// Ensure we're back in original directory before cleanup
	_ = os.Chdir(cwd)
}

func TestRunDoctorCheck_VerboseMode(t *testing.T) {
	// Setup temp project with .claude
	cwd, _ := os.Getwd()
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
		// On Windows, give a moment for file handles to be released
		if runtime.GOOS == "windows" {
			time.Sleep(10 * time.Millisecond)
		}
	})
	dir := t.TempDir()
	_ = os.Chdir(dir)
	_ = os.MkdirAll(filepath.Join(dir, ".claude"), 0o755)

	// Write a sample hooks.yml with multiple groups
	hooks := `test-group-1:
  PreToolUse:
    jobs:
      - name: job1
        run: echo "test1"
        glob: ["*.go"]
test-group-2:
  PostToolUse:
    jobs:
      - name: job2
        run: echo "test2"
        glob: ["*.md"]
`
	if _, err := btconfig.WriteSampleHooksConfig(false, hooks, true); err != nil {
		t.Fatalf("write hooks.yml: %v", err)
	}

	// Run doctor check in verbose mode
	err := runDoctorCheck(true)
	if err != nil {
		t.Errorf("runDoctorCheck with verbose failed: %v", err)
	}

	// Ensure we're back in original directory before cleanup
	_ = os.Chdir(cwd)
}

func TestCountHooksByEvent(t *testing.T) {
	tests := []struct {
		name     string
		hooks    btconfig.HooksConfig
		expected map[string]int
	}{
		{
			name:     "empty hooks",
			hooks:    btconfig.HooksConfig{},
			expected: map[string]int{},
		},
		{
			name: "single event with one hook",
			hooks: btconfig.HooksConfig{
				PreToolUse: []btconfig.HookMatcher{
					{
						Matcher: "*",
						Hooks: []btconfig.HookCommand{
							{Command: "echo test"},
						},
					},
				},
			},
			expected: map[string]int{"PreToolUse": 1},
		},
		{
			name: "multiple events with multiple hooks",
			hooks: btconfig.HooksConfig{
				PreToolUse: []btconfig.HookMatcher{
					{
						Matcher: "*",
						Hooks: []btconfig.HookCommand{
							{Command: "echo test1"},
							{Command: "echo test2"},
						},
					},
				},
				PostToolUse: []btconfig.HookMatcher{
					{
						Matcher: "Read",
						Hooks: []btconfig.HookCommand{
							{Command: "echo test3"},
						},
					},
				},
			},
			expected: map[string]int{"PreToolUse": 2, "PostToolUse": 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countHooksByEvent(tt.hooks)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d events, got %d", len(tt.expected), len(result))
			}
			for event, count := range tt.expected {
				if result[event] != count {
					t.Errorf("event %s: expected count %d, got %d", event, count, result[event])
				}
			}
		})
	}
}

func TestGetCandidateConfigPaths(t *testing.T) {
	paths, err := getCandidateConfigPaths()
	if err != nil {
		t.Fatalf("getCandidateConfigPaths failed: %v", err)
	}

	if len(paths) == 0 {
		t.Error("expected at least some candidate paths, got none")
	}

	// Check that we have both project and global paths
	hasProject := false
	hasGlobal := false
	home, _ := os.UserHomeDir()

	for _, p := range paths {
		if home != "" {
			if filepath.HasPrefix(p, home) {
				hasGlobal = true
			} else {
				hasProject = true
			}
		}
	}

	if !hasProject {
		t.Error("expected project-level paths in candidates")
	}

	if home != "" && !hasGlobal {
		t.Error("expected global-level paths in candidates")
	}
}
