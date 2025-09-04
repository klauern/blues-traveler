package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/klauern/blues-traveler/internal/constants"
	yaml "gopkg.in/yaml.v3"
)

// HookJob represents a single job within an event in a named group
type HookJob struct {
	Name    string            `yaml:"name" json:"name"`
	Run     string            `yaml:"run" json:"run"`
	Glob    []string          `yaml:"glob,omitempty" json:"glob,omitempty"`
	Skip    string            `yaml:"skip,omitempty" json:"skip,omitempty"`
	Only    string            `yaml:"only,omitempty" json:"only,omitempty"`
	Timeout int               `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Env     map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	WorkDir string            `yaml:"workdir,omitempty" json:"workdir,omitempty"`
}

// EventConfig contains jobs for a given Claude Code event, and execution hints
type EventConfig struct {
	Parallel bool      `yaml:"parallel,omitempty" json:"parallel,omitempty"`
	Jobs     []HookJob `yaml:"jobs" json:"jobs"`
}

// HookGroup is a set of EventName -> EventConfig
type HookGroup map[string]*EventConfig

// HooksConfig is the root structure: GroupName -> HookGroup
type CustomHooksConfig map[string]HookGroup

// candidateConfigPaths returns the list of possible config file locations in
// priority order (earlier paths have higher precedence).
// The loader will merge from lowest to highest priority so earlier entries win.
func candidateConfigPaths() ([]string, error) {
	var paths []string

	// Project scope
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %v", err)
	}
	proj := filepath.Join(cwd, ".claude")
	// Prefer new canonical file under hooks/
	paths = append(paths,
		filepath.Join(proj, "hooks", "hooks.yml"),
		filepath.Join(proj, "hooks", "hooks.yaml"),
	)
	// Legacy locations for backward compatibility
	paths = append(paths,
		filepath.Join(proj, "hooks.yml"),
		filepath.Join(proj, "hooks.yaml"),
		filepath.Join(proj, "hooks.json"),
	)
	// Per-group files in .claude/hooks/
	projGroups := filepath.Join(proj, "hooks")
	if entries, err := os.ReadDir(projGroups); err == nil {
		// deterministic order
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			low := strings.ToLower(name)
			if name == constants.ConfigFileName {
				continue // skip app config files
			}
			if strings.HasSuffix(low, ".yml") || strings.HasSuffix(low, ".yaml") {
				names = append(names, name)
			}
		}
		sort.Strings(names)
		for _, n := range names {
			paths = append(paths, filepath.Join(projGroups, n))
		}
	}
	// Local override last in project scope
	paths = append(paths, filepath.Join(proj, "hooks-local.yml"))

	// Global scope
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}
	glob := filepath.Join(home, ".claude")
	paths = append(paths,
		filepath.Join(glob, "hooks", "hooks.yml"),
		filepath.Join(glob, "hooks", "hooks.yaml"),
	)
	// Legacy top-level
	paths = append(paths,
		filepath.Join(glob, "hooks.yml"),
		filepath.Join(glob, "hooks.yaml"),
		filepath.Join(glob, "hooks.json"),
	)
	// Per-group files in ~/.claude/hooks/
	globGroups := filepath.Join(glob, "hooks")
	if entries, err := os.ReadDir(globGroups); err == nil {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			low := strings.ToLower(name)
			if name == constants.ConfigFileName {
				continue
			}
			if strings.HasSuffix(low, ".yml") || strings.HasSuffix(low, ".yaml") {
				names = append(names, name)
			}
		}
		sort.Strings(names)
		for _, n := range names {
			paths = append(paths, filepath.Join(globGroups, n))
		}
	}
	// Global local override last
	paths = append(paths, filepath.Join(glob, "hooks-local.yml"))

	return paths, nil
}

// LoadHooksConfig discovers, parses, and merges all available config files.
// Higher-priority files (earlier in candidate list) override lower-priority ones.
func LoadHooksConfig() (*CustomHooksConfig, error) {
	// 1) Try embedded in main config first
	if embedded := loadEmbeddedHooksConfig(); embedded != nil {
		return embedded, nil
	}
	// 2) Fallback to file discovery (legacy)
	candidates, err := candidateConfigPaths()
	if err != nil {
		return nil, err
	}

	type parsed struct {
		path string
		cfg  CustomHooksConfig
	}
	var found []parsed
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			cfg, err := parseHooksConfigFile(p)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %s: %w", p, err)
			}
			found = append(found, parsed{path: p, cfg: cfg})
		}
	}

	if len(found) == 0 {
		empty := CustomHooksConfig{}
		return &empty, nil
	}

	// Merge from lowest to highest priority so first candidate wins on conflicts
	// i.e., apply in reverse order, then overlay forward.
	eff := CustomHooksConfig{}
	for i := len(found) - 1; i >= 0; i-- {
		eff = mergeHooksConfigs(eff, found[i].cfg)
	}
	return &eff, nil
}

// loadEmbeddedHooksConfig attempts to read custom hooks embedded in the
// Blues Traveler main config file (blues-traveler-config.json) under
// the key "customHooks". Project scope is checked before global.
func loadEmbeddedHooksConfig() *CustomHooksConfig {
	// project then global
	for _, global := range []bool{false, true} {
		path, err := GetLogConfigPath(global)
		if err != nil {
			continue
		}
		cfg, err := LoadLogConfig(path)
		if err != nil || cfg == nil {
			continue
		}
		if len(cfg.CustomHooks) > 0 {
			cp := cloneHooksConfig(cfg.CustomHooks)
			return &cp
		}
	}
	return nil
}

// MergeHooksConfigs merges two HooksConfig structures.
// base provides existing values; override entries replace or extend base.
func MergeHooksConfigs(base, override *CustomHooksConfig) *CustomHooksConfig {
	if base == nil && override == nil {
		out := CustomHooksConfig{}
		return &out
	}
	if base == nil {
		cp := cloneHooksConfig(*override)
		return &cp
	}
	if override == nil {
		cp := cloneHooksConfig(*base)
		return &cp
	}
	merged := mergeHooksConfigs(*base, *override)
	return &merged
}

func mergeHooksConfigs(base CustomHooksConfig, override CustomHooksConfig) CustomHooksConfig {
	out := cloneHooksConfig(base)
	for groupName, oGroup := range override {
		if oGroup == nil {
			continue
		}
		bGroup, ok := out[groupName]
		if !ok || bGroup == nil {
			out[groupName] = cloneHookGroup(oGroup)
			continue
		}
		// Merge events under the group
		for eventName, oEvent := range oGroup {
			if oEvent == nil {
				continue
			}
			bEvent, exists := bGroup[eventName]
			if !exists || bEvent == nil {
				bGroup[eventName] = cloneEventConfig(oEvent)
				continue
			}
			// Merge EventConfig: override Parallel flag, merge Jobs by name
			merged := &EventConfig{
				Parallel: oEvent.Parallel || bEvent.Parallel, // prefer true if any requests it
				Jobs:     mergeJobsByName(bEvent.Jobs, oEvent.Jobs),
			}
			bGroup[eventName] = merged
		}
	}
	return out
}

func cloneHooksConfig(in CustomHooksConfig) CustomHooksConfig {
	out := CustomHooksConfig{}
	for g, grp := range in {
		out[g] = cloneHookGroup(grp)
	}
	return out
}

func cloneHookGroup(in HookGroup) HookGroup {
	if in == nil {
		return nil
	}
	out := HookGroup{}
	for e, ec := range in {
		out[e] = cloneEventConfig(ec)
	}
	return out
}

func cloneEventConfig(in *EventConfig) *EventConfig {
	if in == nil {
		return nil
	}
	out := &EventConfig{Parallel: in.Parallel}
	if len(in.Jobs) > 0 {
		out.Jobs = make([]HookJob, len(in.Jobs))
		copy(out.Jobs, in.Jobs)
	}
	return out
}

func mergeJobsByName(base, override []HookJob) []HookJob {
	result := make([]HookJob, 0, len(base)+len(override))
	index := map[string]int{}
	for i, j := range base {
		result = append(result, j)
		if j.Name != "" {
			index[j.Name] = i
		}
	}
	for _, j := range override {
		if j.Name == "" {
			// Append anonymous job as-is
			result = append(result, j)
			continue
		}
		if idx, ok := index[j.Name]; ok {
			// Replace existing job with same name
			result[idx] = j
		} else {
			result = append(result, j)
		}
	}
	return result
}

// parseHooksConfigFile decodes YAML or JSON based on extension
func parseHooksConfigFile(path string) (CustomHooksConfig, error) {
	data, err := os.ReadFile(path) // #nosec G304 - paths are restricted to known .claude dirs
	if err != nil {
		return nil, err
	}

	var cfg CustomHooksConfig
	switch strings.ToLower(filepath.Ext(path)) {
	case ".yml", ".yaml":
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}
	case ".json":
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported config file extension: %s", path)
	}
	if cfg == nil {
		cfg = CustomHooksConfig{}
	}
	return cfg, nil
}

// ValidateHooksConfig performs basic checks for structure and required fields.
func ValidateHooksConfig(cfg *CustomHooksConfig) error {
	if cfg == nil {
		return errors.New("nil config")
	}
	for groupName, grp := range *cfg {
		if grp == nil {
			continue
		}
		for eventName, ec := range grp {
			if ec == nil {
				return fmt.Errorf("group '%s' event '%s' has nil config", groupName, eventName)
			}
			for i, j := range ec.Jobs {
				if strings.TrimSpace(j.Name) == "" {
					return fmt.Errorf("group '%s' event '%s' job[%d] missing name", groupName, eventName, i)
				}
				if strings.TrimSpace(j.Run) == "" {
					return fmt.Errorf("group '%s' event '%s' job '%s' missing run command", groupName, eventName, j.Name)
				}
			}
		}
	}
	return nil
}

// ListHookGroups returns sorted group names from the config
func ListHookGroups(cfg *CustomHooksConfig) []string {
	if cfg == nil {
		return nil
	}
	names := make([]string, 0, len(*cfg))
	for name := range *cfg {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// EnsureClaudeDir ensures the .claude directory exists in the chosen scope
func EnsureClaudeDir(global bool) (string, error) {
	var dir string
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, constants.ClaudeDir)
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(cwd, constants.ClaudeDir)
	}
	if err := os.MkdirAll(filepath.Join(dir, constants.HooksSubDir), 0o750); err != nil {
		return "", err
	}
	return dir, nil
}

// WriteSampleHooksConfig writes a minimal sample hooks.yml to the chosen scope
func WriteSampleHooksConfig(global bool, content string, overwrite bool) (string, error) {
	// Parse YAML content into CustomHooksConfig
	var fromYAML CustomHooksConfig
	if err := yaml.Unmarshal([]byte(content), &fromYAML); err != nil {
		return "", fmt.Errorf("invalid sample yaml: %w", err)
	}

	// Load existing main config JSON
	cfgPath, err := GetLogConfigPath(global)
	if err != nil {
		return "", err
	}
	logCfg, err := LoadLogConfig(cfgPath)
	if err != nil {
		return "", err
	}

	// If not overwrite and there is already CustomHooks, refuse to overwrite
	if !overwrite && logCfg.CustomHooks != nil && len(logCfg.CustomHooks) > 0 {
		return cfgPath, fs.ErrExist
	}

	// Merge (overwrite existing groups with sample)
	merged := MergeHooksConfigs(&logCfg.CustomHooks, &fromYAML)
	logCfg.CustomHooks = *merged

	if err := SaveLogConfig(cfgPath, logCfg); err != nil {
		return "", err
	}
	return cfgPath, nil
}
