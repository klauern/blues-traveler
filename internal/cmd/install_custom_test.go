package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	btconfig "github.com/klauern/blues-traveler/internal/config"
)

func TestInstallCustom_WritesSettings(t *testing.T) {
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

	// Write hooks.yml
	hooks := `ruby:
  PreToolUse:
    jobs:
      - name: rubocop
        run: rubocop ${FILES_CHANGED}
        glob: ["*.rb"]
`
	if _, err := btconfig.WriteSampleHooksConfig(false, hooks, true); err != nil {
		t.Fatalf("write hooks.yml: %v", err)
	}

	// Load and validate
	cfg, err := btconfig.LoadHooksConfig()
	if err != nil || cfg == nil {
		t.Fatalf("load config: %v", err)
	}
	if err := btconfig.ValidateHooksConfig(cfg); err != nil {
		t.Fatalf("validate: %v", err)
	}

	// Ensure we're back in original directory before cleanup
	// This helps avoid Windows file locking issues
	_ = os.Chdir(cwd)
}
