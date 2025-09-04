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
)

// GetConfigPath returns the full config file path
func GetConfigPath(baseDir string) string {
	return baseDir + "/" + ClaudeDir + "/" + HooksSubDir + "/" + ConfigFileName
}
