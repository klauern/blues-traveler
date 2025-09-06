package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

// XDGConfig handles XDG Base Directory Specification compliant configuration
type XDGConfig struct {
	BaseDir string
}

// ProjectRegistry maps project paths to their configuration files
type ProjectRegistry struct {
	Version  string                   `json:"version"`
	Projects map[string]ProjectConfig `json:"projects"`
}

// ProjectConfig contains metadata about a project's configuration
type ProjectConfig struct {
	ConfigFile   string `json:"configFile"`
	LastModified string `json:"lastModified"`
	ConfigFormat string `json:"configFormat"`
}

// NewXDGConfig creates a new XDG configuration manager
func NewXDGConfig() *XDGConfig {
	baseDir := os.Getenv("XDG_CONFIG_HOME")
	if baseDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// Fallback to current directory if home directory cannot be determined
			baseDir = ".config"
		} else {
			baseDir = filepath.Join(homeDir, ".config")
		}
	}

	return &XDGConfig{
		BaseDir: filepath.Join(baseDir, "blues-traveler"),
	}
}

// GetConfigDir returns the XDG configuration directory for blues-traveler
func (x *XDGConfig) GetConfigDir() string {
	return x.BaseDir
}

// GetGlobalConfigPath returns the path to the global configuration file
func (x *XDGConfig) GetGlobalConfigPath(format string) string {
	if format == "" {
		format = "json"
	}
	return filepath.Join(x.BaseDir, fmt.Sprintf("global.%s", format))
}

// GetProjectsDir returns the directory where project-specific configs are stored
func (x *XDGConfig) GetProjectsDir() string {
	return filepath.Join(x.BaseDir, "projects")
}

// GetRegistryPath returns the path to the project registry file
func (x *XDGConfig) GetRegistryPath() string {
	return filepath.Join(x.BaseDir, "registry.json")
}

// SanitizeProjectPath converts a project path to a safe filename
func (x *XDGConfig) SanitizeProjectPath(projectPath string) string {
	// Convert absolute path to a safe filename
	// Replace path separators and special characters with hyphens
	sanitized := strings.ReplaceAll(projectPath, "/", "-")
	sanitized = strings.ReplaceAll(sanitized, "\\", "-")
	sanitized = strings.ReplaceAll(sanitized, ":", "-")
	sanitized = strings.ReplaceAll(sanitized, " ", "-")
	sanitized = strings.ReplaceAll(sanitized, "~", "home")

	// Replace multiple consecutive hyphens with single hyphen
	re := regexp.MustCompile(`-+`)
	sanitized = re.ReplaceAllString(sanitized, "-")

	// Remove leading/trailing hyphens
	sanitized = strings.Trim(sanitized, "-")

	// Limit length to avoid filesystem issues
	if len(sanitized) > 200 {
		sanitized = sanitized[:200]
	}

	return sanitized
}

// GetProjectConfigPath returns the path to a project's configuration file
func (x *XDGConfig) GetProjectConfigPath(projectPath, format string) string {
	if format == "" {
		format = "json"
	}
	sanitized := x.SanitizeProjectPath(projectPath)
	filename := fmt.Sprintf("%s.%s", sanitized, format)
	return filepath.Join(x.GetProjectsDir(), filename)
}

// EnsureDirectories creates the necessary XDG directories
func (x *XDGConfig) EnsureDirectories() error {
	dirs := []string{
		x.GetConfigDir(),
		x.GetProjectsDir(),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o750); err != nil { // #nosec G301 - XDG directories should be user-only accessible
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// LoadRegistry loads the project registry
func (x *XDGConfig) LoadRegistry() (*ProjectRegistry, error) {
	registryPath := x.GetRegistryPath()

	// If registry doesn't exist, return empty registry
	if _, err := os.Stat(registryPath); os.IsNotExist(err) {
		return &ProjectRegistry{
			Version:  "1.0",
			Projects: make(map[string]ProjectConfig),
		}, nil
	}

	data, err := os.ReadFile(registryPath) // #nosec G304 - registryPath is internally controlled via GetRegistryPath()
	if err != nil {
		return nil, fmt.Errorf("failed to read registry file: %w", err)
	}

	var registry ProjectRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("failed to parse registry JSON: %w", err)
	}

	// Ensure projects map is initialized
	if registry.Projects == nil {
		registry.Projects = make(map[string]ProjectConfig)
	}

	return &registry, nil
}

// SaveRegistry saves the project registry
func (x *XDGConfig) SaveRegistry(registry *ProjectRegistry) error {
	if err := x.EnsureDirectories(); err != nil {
		return err
	}

	registryPath := x.GetRegistryPath()

	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	if err := os.WriteFile(registryPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write registry file: %w", err)
	}

	return nil
}

// RegisterProject adds or updates a project in the registry
func (x *XDGConfig) RegisterProject(projectPath, configFormat string) error {
	registry, err := x.LoadRegistry()
	if err != nil {
		return err
	}

	configFile := x.GetProjectConfigPath(projectPath, configFormat)
	relativeConfigFile := filepath.Join("projects", filepath.Base(configFile))

	registry.Projects[projectPath] = ProjectConfig{
		ConfigFile:   relativeConfigFile,
		LastModified: time.Now().UTC().Format(time.RFC3339),
		ConfigFormat: configFormat,
	}

	return x.SaveRegistry(registry)
}

// GetProjectConfig returns the configuration metadata for a project
func (x *XDGConfig) GetProjectConfig(projectPath string) (*ProjectConfig, error) {
	registry, err := x.LoadRegistry()
	if err != nil {
		return nil, err
	}

	config, exists := registry.Projects[projectPath]
	if !exists {
		return nil, fmt.Errorf("project not found in registry: %s", projectPath)
	}

	return &config, nil
}

// ListProjects returns all registered projects
func (x *XDGConfig) ListProjects() ([]string, error) {
	registry, err := x.LoadRegistry()
	if err != nil {
		return nil, err
	}

	projects := make([]string, 0, len(registry.Projects))
	for projectPath := range registry.Projects {
		projects = append(projects, projectPath)
	}

	return projects, nil
}

// LoadProjectConfig loads configuration data for a specific project
func (x *XDGConfig) LoadProjectConfig(projectPath string) (map[string]interface{}, error) {
	config, err := x.GetProjectConfig(projectPath)
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(x.GetConfigDir(), config.ConfigFile)

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return make(map[string]interface{}), nil
	}

	data, err := os.ReadFile(configPath) // #nosec G304 - configPath is internally controlled via project registry
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var configData map[string]interface{}

	switch config.ConfigFormat {
	case "json":
		if err := json.Unmarshal(data, &configData); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	case "toml":
		if err := toml.Unmarshal(data, &configData); err != nil {
			return nil, fmt.Errorf("failed to parse TOML config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config format: %s", config.ConfigFormat)
	}

	return configData, nil
}

// SaveProjectConfig saves configuration data for a specific project
func (x *XDGConfig) SaveProjectConfig(projectPath string, configData map[string]interface{}, format string) error {
	if format == "" {
		format = "json"
	}

	if err := x.EnsureDirectories(); err != nil {
		return err
	}

	configPath := x.GetProjectConfigPath(projectPath, format)

	var data []byte
	var err error

	switch format {
	case "json":
		data, err = json.MarshalIndent(configData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON config: %w", err)
		}
	case "toml":
		var buf strings.Builder
		encoder := toml.NewEncoder(&buf)
		if err := encoder.Encode(configData); err != nil {
			return fmt.Errorf("failed to marshal TOML config: %w", err)
		}
		data = []byte(buf.String())
	default:
		return fmt.Errorf("unsupported config format: %s", format)
	}

	// Ensure the projects directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0o750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Register the project in the registry
	return x.RegisterProject(projectPath, format)
}

// LoadGlobalConfig loads the global configuration
func (x *XDGConfig) LoadGlobalConfig(format string) (map[string]interface{}, error) {
	if format == "" {
		format = "json"
	}

	configPath := x.GetGlobalConfigPath(format)

	// If global config doesn't exist, return empty config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return make(map[string]interface{}), nil
	}

	data, err := os.ReadFile(configPath) // #nosec G304 - configPath is internally controlled via project registry
	if err != nil {
		return nil, fmt.Errorf("failed to read global config file: %w", err)
	}

	var configData map[string]interface{}

	switch format {
	case "json":
		if err := json.Unmarshal(data, &configData); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	case "toml":
		if err := toml.Unmarshal(data, &configData); err != nil {
			return nil, fmt.Errorf("failed to parse TOML config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config format: %s", format)
	}

	return configData, nil
}

// SaveGlobalConfig saves the global configuration
func (x *XDGConfig) SaveGlobalConfig(configData map[string]interface{}, format string) error {
	if format == "" {
		format = "json"
	}

	if err := x.EnsureDirectories(); err != nil {
		return err
	}

	configPath := x.GetGlobalConfigPath(format)

	var data []byte
	var err error

	switch format {
	case "json":
		data, err = json.MarshalIndent(configData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON config: %w", err)
		}
	case "toml":
		var buf strings.Builder
		encoder := toml.NewEncoder(&buf)
		if err := encoder.Encode(configData); err != nil {
			return fmt.Errorf("failed to marshal TOML config: %w", err)
		}
		data = []byte(buf.String())
	default:
		return fmt.Errorf("unsupported config format: %s", format)
	}

	// Ensure the config directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0o750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write global config file: %w", err)
	}

	return nil
}

// CleanupOrphanedConfigs removes configuration files for projects that no longer exist
func (x *XDGConfig) CleanupOrphanedConfigs() ([]string, error) {
	registry, err := x.LoadRegistry()
	if err != nil {
		return nil, err
	}

	var orphaned []string

	for projectPath, config := range registry.Projects {
		// Check if project directory still exists
		if _, err := os.Stat(projectPath); os.IsNotExist(err) {
			configPath := filepath.Join(x.GetConfigDir(), config.ConfigFile)

			// Remove the config file
			if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
				return orphaned, fmt.Errorf("failed to remove orphaned config %s: %w", configPath, err)
			}

			// Remove from registry
			delete(registry.Projects, projectPath)
			orphaned = append(orphaned, projectPath)
		}
	}

	if len(orphaned) > 0 {
		if err := x.SaveRegistry(registry); err != nil {
			return orphaned, err
		}
	}

	return orphaned, nil
}
