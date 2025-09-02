package cmd

import (
	"os"
	"path/filepath"
	"testing"

	btconfig "github.com/klauern/blues-traveler/internal/config"
)

func TestInstallCustom_WritesSettings(t *testing.T) {
	// Setup temp project with .claude
	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
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
}
