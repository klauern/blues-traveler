// Package constants provides application-wide constants - single source of truth for naming throughout the codebase
package constants

// Application constants - single source of truth for naming throughout the codebase
const (
	// Core application identity
	AppName        = "Blues Traveler"
	BinaryName     = "blues-traveler"
	ProjectTagline = "The hook brings you back"

	// Module and repository
	ModulePath    = "github.com/klauern/blues-traveler"
	RepositoryURL = "https://github.com/klauern/blues-traveler"

	// Configuration files
	ConfigFileName   = "blues-traveler-config.json"
	SettingsFileName = "settings.json"
	BlockedUrlsFile  = "blocked-urls.txt"

	// Log files
	DefaultLogFile = "blues-traveler.log"
	DebugLogFile   = "debug.log"
	FormatLogFile  = "format.log"

	// Directory paths
	ClaudeDir        = ".claude"
	HooksSubDir      = "hooks"
	InternalHooksDir = "internal/hooks"

	// Command patterns for settings
	CommandPattern = BinaryName + " run"

	// Platform and OS
	GOOSWindows = "windows"

	// Configuration sources
	XDGSource = "xdg"

	// Test constants
	TestProjectPath = "/Users/user/dev/project"

	// Scope constants
	ScopeProject = "project"
	ScopeGlobal  = "global"

	// Tool names
	ToolBash  = "Bash"
	ToolEdit  = "Edit"
	ToolWrite = "Write"
	ToolRead  = "Read"
	ToolGlob  = "Glob"
	ToolGrep  = "Grep"
)

// GetConfigPath returns the full config file path
func GetConfigPath(baseDir string) string {
	return baseDir + "/" + ClaudeDir + "/" + HooksSubDir + "/" + ConfigFileName
}
