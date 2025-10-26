package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMergeHooksConfigs_GroupEventJobMerge(t *testing.T) {
	base := CustomHooksConfig{
		"ruby": HookGroup{
			"PreToolUse": &EventConfig{Jobs: []HookJob{{Name: "rubocop", Run: "rubocop"}}},
		},
	}
	override := CustomHooksConfig{
		"ruby": HookGroup{
			"PreToolUse": &EventConfig{Jobs: []HookJob{{Name: "rubocop", Run: "bundle exec rubocop"}, {Name: "brakeman", Run: "brakeman"}}},
		},
	}

	merged := MergeHooksConfigs(&base, &override)
	ev := (*merged)["ruby"]["PreToolUse"]
	if len(ev.Jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(ev.Jobs))
	}
	if ev.Jobs[0].Run != "bundle exec rubocop" {
		t.Fatalf("expected rubocop to be replaced, got %q", ev.Jobs[0].Run)
	}
}

func TestParseHooksConfigFile_YAMLAndJSON(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "hooks.yml")
	jsonp := filepath.Join(dir, "hooks.json")
	yamlContent := []byte("ruby:\n  PreToolUse:\n    jobs:\n      - name: rubocop\n        run: rubocop\n")
	jsonContent := []byte(`{"ruby":{"PostToolUse":{"jobs":[{"name":"rspec","run":"rspec"}]}}}`)

	if err := os.WriteFile(yml, yamlContent, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(jsonp, jsonContent, 0o600); err != nil {
		t.Fatal(err)
	}

	cfgY, err := parseHooksConfigFile(yml)
	if err != nil || cfgY["ruby"] == nil {
		t.Fatalf("yaml parse failed: %v", err)
	}
	cfgJ, err := parseHooksConfigFile(jsonp)
	if err != nil || cfgJ["ruby"] == nil {
		t.Fatalf("json parse failed: %v", err)
	}
}

func TestIsValidHookConfigFile(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		want     bool
	}{
		{"yml file", "hooks.yml", true},
		{"yaml file", "hooks.yaml", true},
		{"uppercase YML", "HOOKS.YML", true},
		{"config file to skip", "blues-traveler-config.json", false},
		{"json file", "hooks.json", false},
		{"txt file", "hooks.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidHookConfigFile(tt.fileName)
			if got != tt.want {
				t.Errorf("isValidHookConfigFile(%q) = %v, want %v", tt.fileName, got, tt.want)
			}
		})
	}
}

func TestCollectPerGroupFiles(t *testing.T) {
	dir := t.TempDir()
	hooksDir := filepath.Join(dir, "hooks")
	if err := os.MkdirAll(hooksDir, 0o750); err != nil {
		t.Fatal(err)
	}

	// Create test files
	testFiles := []string{
		"ruby.yml",
		"python.yaml",
		"go.yml",
		"blues-traveler-config.json", // should be skipped
		"readme.txt",                 // should be skipped
	}

	for _, f := range testFiles {
		if err := os.WriteFile(filepath.Join(hooksDir, f), []byte{}, 0o600); err != nil {
			t.Fatal(err)
		}
	}

	// Create a subdirectory (should be ignored)
	subdir := filepath.Join(hooksDir, "subdir")
	if err := os.MkdirAll(subdir, 0o750); err != nil {
		t.Fatal(err)
	}

	paths := collectPerGroupFiles(hooksDir)

	// Should only include yml/yaml files, sorted alphabetically
	expectedCount := 3 // ruby.yml, python.yaml, go.yml
	if len(paths) != expectedCount {
		t.Errorf("collectPerGroupFiles() returned %d files, want %d", len(paths), expectedCount)
	}

	// Verify files are sorted
	expectedFiles := []string{"go.yml", "python.yaml", "ruby.yml"}
	for i, expected := range expectedFiles {
		if filepath.Base(paths[i]) != expected {
			t.Errorf("collectPerGroupFiles()[%d] basename = %s, want %s", i, filepath.Base(paths[i]), expected)
		}
	}
}

func TestAddProjectPaths(t *testing.T) {
	baseDir := "/test/.claude"
	paths := addProjectPaths(baseDir)

	// Verify it includes canonical paths
	expectedPaths := []string{
		filepath.Join(baseDir, "hooks", "hooks.yml"),
		filepath.Join(baseDir, "hooks", "hooks.yaml"),
		filepath.Join(baseDir, "hooks.yml"),
		filepath.Join(baseDir, "hooks.yaml"),
		filepath.Join(baseDir, "hooks.json"),
	}

	for _, expected := range expectedPaths {
		found := false
		for _, path := range paths {
			if path == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("addProjectPaths() missing expected path: %s", expected)
		}
	}

	// Verify it includes local override (should be first for highest precedence)
	localOverride := filepath.Join(baseDir, "hooks-local.yml")
	if paths[0] != localOverride {
		t.Errorf("addProjectPaths() first path should be local override, got %s", paths[0])
	}
}

func TestAddGlobalPaths(t *testing.T) {
	baseDir := "/home/user/.claude"
	paths := addGlobalPaths(baseDir)

	// Verify it includes canonical paths
	expectedPaths := []string{
		filepath.Join(baseDir, "hooks", "hooks.yml"),
		filepath.Join(baseDir, "hooks", "hooks.yaml"),
		filepath.Join(baseDir, "hooks.yml"),
	}

	for _, expected := range expectedPaths {
		found := false
		for _, path := range paths {
			if path == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("addGlobalPaths() missing expected path: %s", expected)
		}
	}
}
