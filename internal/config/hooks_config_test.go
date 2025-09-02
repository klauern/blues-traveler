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

