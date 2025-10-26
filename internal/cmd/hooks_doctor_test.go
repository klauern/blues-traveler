package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	btconfig "github.com/klauern/blues-traveler/internal/config"
	"github.com/klauern/blues-traveler/internal/constants"
)

func TestRunDoctorCheck_NoConfiguration(t *testing.T) {
	// Setup temp project with .claude
	cwd, _ := os.Getwd()
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
		// On Windows, give a moment for file handles to be released
		if runtime.GOOS == constants.GOOSWindows {
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

// setupDoctorTest sets up a temporary test environment with a hooks config
func setupDoctorTest(t *testing.T, hooksConfig string) func() {
	t.Helper()
	cwd, _ := os.Getwd()
	cleanup := func() {
		_ = os.Chdir(cwd)
		// On Windows, give a moment for file handles to be released
		if runtime.GOOS == constants.GOOSWindows {
			time.Sleep(10 * time.Millisecond)
		}
	}
	t.Cleanup(cleanup)

	dir := t.TempDir()
	_ = os.Chdir(dir)
	_ = os.MkdirAll(filepath.Join(dir, ".claude"), 0o755)

	if _, err := btconfig.WriteSampleHooksConfig(false, hooksConfig, true); err != nil {
		t.Fatalf("write hooks.yml: %v", err)
	}

	return func() {
		_ = os.Chdir(cwd)
	}
}

func TestRunDoctorCheck_WithHooksConfig(t *testing.T) {
	hooks := `test-group:
  PreToolUse:
    jobs:
      - name: test-job
        run: echo "test"
        glob: ["*"]
`
	cleanup := setupDoctorTest(t, hooks)
	defer cleanup()

	err := runDoctorCheck(false)
	if err != nil {
		t.Errorf("runDoctorCheck failed: %v", err)
	}
}

func TestRunDoctorCheck_VerboseMode(t *testing.T) {
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
	cleanup := setupDoctorTest(t, hooks)
	defer cleanup()

	err := runDoctorCheck(true)
	if err != nil {
		t.Errorf("runDoctorCheck with verbose failed: %v", err)
	}
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
	cwd, _ := os.Getwd()

	for _, p := range paths {
		// Global paths are directly under ~/.claude
		// Project paths are under <current-dir>/.claude
		if home != "" && strings.HasPrefix(p, filepath.Join(home, ".claude")) &&
			!strings.HasPrefix(p, filepath.Join(cwd, ".claude")) {
			hasGlobal = true
		} else if strings.HasPrefix(p, filepath.Join(cwd, ".claude")) {
			hasProject = true
		}
	}

	if !hasProject {
		t.Error("expected project-level paths in candidates")
	}

	if home != "" && !hasGlobal {
		t.Error("expected global-level paths in candidates")
	}
}
