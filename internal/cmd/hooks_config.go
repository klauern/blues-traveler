package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	btconfig "github.com/klauern/blues-traveler/internal/config"
	"github.com/urfave/cli/v3"
	yaml "gopkg.in/yaml.v3"
)

// generateHooksREADME creates comprehensive documentation for the hooks directory
func generateHooksREADME() string {
	return `# Custom Hooks Configuration

This directory contains your Blues Traveler custom hooks configuration. Custom hooks allow you to run scripts and commands in response to Claude Code events.

## Quick Start

1. **Create a hooks configuration file** (hooks.yml or blues-traveler-config.json)
2. **Define hook groups and jobs** for specific Claude Code events
3. **Test your hooks** with ` + "`blues-traveler run config:group:job`" + `
4. **Install into Claude Code** with ` + "`blues-traveler config sync`" + `

## Configuration Format

### YAML Format (hooks.yml)
` + "```yaml" + `
my-group:
  PreToolUse:
    jobs:
      - name: security-check
        run: ./scripts/security-check.sh
        glob: ["*.sh", "*.py"]
        only: ${TOOL_NAME} == "Bash"
  PostToolUse:
    jobs:
      - name: format-code
        run: gofmt -w ${TOOL_OUTPUT_FILE}
        only: ${TOOL_NAME} == "Edit" || ${TOOL_NAME} == "Write"
        glob: ["*.go"]
` + "```" + `

### JSON Format (blues-traveler-config.json)
` + "```json" + `
{
  "customHooks": {
    "my-group": {
      "PreToolUse": {
        "jobs": [
          {
            "name": "security-check",
            "run": "./scripts/security-check.sh",
            "glob": ["*.sh", "*.py"],
            "only": "${TOOL_NAME} == \"Bash\""
          }
        ]
      }
    }
  }
}
` + "```" + `

## Available Events

| Event | Description | Best Use Cases |
|-------|-------------|----------------|
| **PreToolUse** | Before Claude Code runs a tool | Security checks, validation |
| **PostToolUse** | After Claude Code runs a tool | Formatting, testing, cleanup |
| **UserPromptSubmit** | When user submits a prompt | Logging, preprocessing |
| **Notification** | System notifications | Alerts, monitoring |
| **Stop** | When Claude Code stops | Cleanup, reporting |
| **SubagentStop** | When a subagent stops | Subagent-specific cleanup |
| **PreCompact** | Before context compaction | Data preservation |
| **SessionStart** | Session begins | Initialization, setup |
| **SessionEnd** | Session ends | Teardown, reporting |

## Environment Variables

These variables are available in your hook scripts:

| Variable | Description | Example |
|----------|-------------|---------|
| ` + "`TOOL_NAME`" + ` | Name of the tool being used | Bash, Edit, Write |
| ` + "`TOOL_OUTPUT_FILE`" + ` | File path for Edit/Write tools | /path/to/file.go |
| ` + "`FILES_CHANGED`" + ` | Files modified (comma-separated) | file1.py,file2.js |
| ` + "`USER_PROMPT`" + ` | User's prompt text | "Fix this bug" |
| ` + "`EVENT_NAME`" + ` | Current event name | PreToolUse |

## Job Properties

### Required
- **name**: Unique identifier for the job
- **run**: Command to execute

### Optional
- **glob**: File patterns to match (` + "`[\"*.py\", \"*.js\"]`" + `)
- **only**: Condition for when to run (` + "`${TOOL_NAME} == \"Edit\"`" + `)
- **skip**: Condition for when to skip (` + "`${FILES_CHANGED} regex \"test\"`" + `)
- **timeout**: Timeout in seconds (default: 30)
- **env**: Environment variables (` + "`{\"VAR\": \"value\"}`" + `)
- **workdir**: Working directory for the command

## Filtering Examples

### Tool-based Filtering
` + "```yaml" + `
# Only run on Bash commands
only: ${TOOL_NAME} == "Bash"

# Run on Edit or Write operations
only: ${TOOL_NAME} == "Edit" || ${TOOL_NAME} == "Write"
` + "```" + `

### File-based Filtering  
` + "```yaml" + `
# Only Python files
glob: ["*.py"]

# Multiple file types
glob: ["*.go", "*.mod", "*.sum"]

# Regex on changed files
only: ${FILES_CHANGED} regex ".*\\.py$"
` + "```" + `

## Real-World Examples

### Code Formatting
` + "```yaml" + `
format-group:
  PostToolUse:
    jobs:
      - name: format-go
        run: gofmt -w ${TOOL_OUTPUT_FILE}
        only: ${TOOL_NAME} == "Edit" || ${TOOL_NAME} == "Write"
        glob: ["*.go"]
      - name: format-python
        run: black ${TOOL_OUTPUT_FILE}
        only: ${TOOL_NAME} == "Edit" || ${TOOL_NAME} == "Write"
        glob: ["*.py"]
` + "```" + `

### Security Checks
` + "```yaml" + `
security-group:
  PreToolUse:
    jobs:
      - name: dangerous-commands
        run: |
          if echo "$TOOL_ARGS" | grep -E "(rm -rf|sudo|curl.*\\|.*sh)"; then
            echo "❌ Dangerous command detected!"
            exit 1
          fi
        only: ${TOOL_NAME} == "Bash"
` + "```" + `

### Testing
` + "```yaml" + `
test-group:
  PostToolUse:
    jobs:
      - name: run-tests
        run: go test ./...
        only: ${TOOL_NAME} == "Edit" || ${TOOL_NAME} == "Write"
        glob: ["*.go"]
        skip: ${FILES_CHANGED} regex ".*_test\\.go$"
` + "```" + `

## Built-in Hooks vs Custom Hooks

**Built-in Hooks** (via ` + "`blues-traveler install`" + `):
- Pre-built, tested functionality
- Security, formatting, debugging, audit
- Installed directly into Claude Code settings

**Custom Hooks** (this configuration):
- Your own scripts and commands
- Project or personal automation
- Requires ` + "`blues-traveler config sync`" + ` to install

## Commands

` + "```bash" + `
# Initialize configuration
blues-traveler config init [--global]

# Validate configuration
blues-traveler config validate

# Show current configuration  
blues-traveler config show

# Test a specific hook
blues-traveler run config:group:job

# Install hooks into Claude Code
blues-traveler config sync [--global]

# List hook groups
blues-traveler config groups
` + "```" + `

## Troubleshooting

### Hook Not Running
1. Check ` + "`blues-traveler config validate`" + ` for syntax errors
2. Verify the event type matches your use case
3. Test conditions with ` + "`only`" + ` and ` + "`skip`" + ` filters
4. Check file glob patterns match your files

### Permission Errors
1. Ensure script files are executable (` + "`chmod +x script.sh`" + `)
2. Use absolute paths for commands
3. Check working directory with ` + "`workdir`" + ` property

### Environment Variables
1. Use ` + "`echo $TOOL_NAME`" + ` in your script to debug
2. Variables are only available during hook execution
3. Empty variables mean the event doesn't provide that data

## Best Practices

### Security
- ✅ Validate input from environment variables
- ✅ Use absolute paths for scripts
- ✅ Limit file permissions on hook scripts
- ❌ Don't trust user input blindly

### Performance
- ✅ Use specific glob patterns to reduce overhead
- ✅ Set reasonable timeouts
- ✅ Use ` + "`skip`" + ` conditions to avoid unnecessary work
- ❌ Don't run expensive operations on every event

### Maintainability
- ✅ Use descriptive job names
- ✅ Comment your hook configurations
- ✅ Test hooks before deploying
- ✅ Version control your hook configurations

## Getting Help

- **Documentation**: [Blues Traveler GitHub](https://github.com/klauern/blues-traveler)
- **Built-in Help**: ` + "`blues-traveler --help`" + `
- **Validation**: ` + "`blues-traveler config validate`" + `
- **Testing**: ` + "`blues-traveler run config:group:job`" + `
`
}

// NewHooksConfigCmd provides `config`-like helpers for hooks.yml management
func NewHooksConfigCmd() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "Manage custom hooks configuration (hooks.yml)",
		Commands: []*cli.Command{
			{
				Name:  "init",
				Usage: "Create a sample hooks configuration file",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "global", Aliases: []string{"g"}, Usage: "Create in ~/.claude"},
					&cli.BoolFlag{Name: "overwrite", Usage: "Overwrite existing file if present"},
					&cli.StringFlag{Name: "group", Aliases: []string{"G"}, Value: "example", Usage: "Group name for this config"},
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Usage: "Filename for per-group config (writes .claude/hooks/<name>.yml)"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					global := cmd.Bool("global")
					overwrite := cmd.Bool("overwrite")
					group := cmd.String("group")
					fileName := cmd.String("name")
					
					var sample string
					if global {
						// Minimal global config - no example hooks to avoid accidental installation
						sample = fmt.Sprintf(`# Global hooks configuration for group '%s'
# This is your personal global configuration. Add real hooks here.
# See README.md in this directory for documentation and examples.
%s:
  # Add your custom hooks here
  # Example structure:
  # PreToolUse:
  #   jobs:
  #     - name: my-security-check
  #       run: ./my-script.sh
  #       glob: ["*.go"]
`, group, group)
					} else {
						// Project config with comprehensive examples for learning
						sample = fmt.Sprintf(`# Sample hooks configuration for group '%s'
%s:
  PreToolUse:
    jobs:
      - name: pre-sample
        run: echo "PreToolUse TOOL=${TOOL_NAME}"
        glob: ["*"]
  PostToolUse:
    jobs:
      - name: post-format-sample
        # Demonstrates file-based action with TOOL_OUTPUT_FILE for Edit/Write
        run: ruff format --fix ${TOOL_OUTPUT_FILE}
        only: ${TOOL_NAME} == "Edit" || ${TOOL_NAME} == "Write"
        glob: ["*.py"]
      - name: post-regex-sample
        # Demonstrates regex filtering on FILES_CHANGED
        run: echo "Matched regex on ${FILES_CHANGED}"
        only: ${FILES_CHANGED} regex ".*\\.py$"
  UserPromptSubmit:
    jobs:
      - name: user-prompt-sample
        run: echo "UserPrompt ${USER_PROMPT}"
  Notification:
    jobs:
      - name: notification-sample
        run: echo "Notification EVENT=${EVENT_NAME}"
  Stop:
    jobs:
      - name: stop-sample
        run: echo "Stop EVENT=${EVENT_NAME}"
  SubagentStop:
    jobs:
      - name: subagent-stop-sample
        run: echo "SubagentStop EVENT=${EVENT_NAME}"
  PreCompact:
    jobs:
      - name: precompact-sample
        run: echo "PreCompact EVENT=${EVENT_NAME}"
  SessionStart:
    jobs:
      - name: session-start-sample
        run: echo "SessionStart EVENT=${EVENT_NAME}"
  SessionEnd:
    jobs:
      - name: session-end-sample
        run: echo "SessionEnd EVENT=${EVENT_NAME}"
`, group, group)
					}
					// If --name provided, create .claude/hooks/<name>.yml; else .claude/hooks.yml
					var path string
					if fileName != "" {
						dir, err := btconfig.EnsureClaudeDir(global)
						if err != nil {
							return err
						}
						hooksDir := filepath.Join(dir, "hooks")
						if err := os.MkdirAll(hooksDir, 0o750); err != nil {
							return err
						}
						// sanitize minimal: ensure .yml extension
						base := fileName
						if !strings.HasSuffix(strings.ToLower(base), ".yml") && !strings.HasSuffix(strings.ToLower(base), ".yaml") {
							base = base + ".yml"
						}
						target := filepath.Join(hooksDir, base)
						if !overwrite {
							if _, err := os.Stat(target); err == nil {
								fmt.Printf("File already exists: %s (use --overwrite to replace)\n", target)
								return nil
							}
						}
						if err := os.WriteFile(target, []byte(sample), 0o600); err != nil {
							return err
						}
						path = target
					} else {
						if global {
							// For global configs, create minimal config directly
							configPath, err := btconfig.GetLogConfigPath(global)
							if err != nil {
								return err
							}
							
							// Load existing config or create default
							logCfg, err := btconfig.LoadLogConfig(configPath)
							if err != nil {
								return err
							}
							
							// Check for existing config without overwrite
							if !overwrite && logCfg.CustomHooks != nil && len(logCfg.CustomHooks) > 0 {
								fmt.Printf("File already exists: %s (use --overwrite to replace)\n", configPath)
								return nil
							}
							
							// Create minimal hooks structure (empty)
							logCfg.CustomHooks = btconfig.CustomHooksConfig{}
							
							// Ensure directory exists
							if err := os.MkdirAll(filepath.Dir(configPath), 0o750); err != nil {
								return err
							}
							
							// Save the minimal config
							if err := btconfig.SaveLogConfig(configPath, logCfg); err != nil {
								return err
							}
							path = configPath
						} else {
							// For project configs, use existing sample logic
							var werr error
							path, werr = btconfig.WriteSampleHooksConfig(global, sample, overwrite)
							if errors.Is(werr, fs.ErrExist) {
								fmt.Printf("File already exists: %s (use --overwrite to replace)\n", path)
								return nil
							}
							if werr != nil {
								return werr
							}
						}
					}
					
					// Create README.md in the hooks directory
					hooksDir := filepath.Dir(path)
					readmePath := filepath.Join(hooksDir, "README.md")
					
					// Check if README.md already exists
					if !overwrite {
						if _, err := os.Stat(readmePath); err == nil {
							fmt.Printf("Created sample hooks config at %s\n", path)
							fmt.Printf("README.md already exists at %s\n", readmePath)
							return nil
						}
					}
					
					// Write README.md
					readmeContent := generateHooksREADME()
					if err := os.WriteFile(readmePath, []byte(readmeContent), 0o644); err != nil {
						// Don't fail the whole operation if README creation fails
						fmt.Printf("Warning: Could not create README.md: %v\n", err)
					} else {
						fmt.Printf("Created hooks documentation at %s\n", readmePath)
					}
					
					fmt.Printf("Created sample hooks config at %s\n", path)
					return nil
				},
			},
			{
				Name:  "validate",
				Usage: "Validate hooks.yml syntax",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cfg, err := btconfig.LoadHooksConfig()
					if err != nil {
						return fmt.Errorf("load error: %v", err)
					}
					if err := btconfig.ValidateHooksConfig(cfg); err != nil {
						return fmt.Errorf("invalid hooks config: %v", err)
					}
					fmt.Println("hooks config is valid")
					return nil
				},
			},
			{
				Name:  "groups",
				Usage: "List available custom hook groups",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cfg, err := btconfig.LoadHooksConfig()
					if err != nil {
						return fmt.Errorf("load error: %v", err)
					}
					groups := btconfig.ListHookGroups(cfg)
					if len(groups) == 0 {
						fmt.Println("No custom hook groups found")
						return nil
					}
					for _, g := range groups {
						fmt.Println(g)
					}
					return nil
				},
			},
			{
				Name:  "show",
				Usage: "Display the effective custom hooks configuration",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "format", Aliases: []string{"f"}, Value: "yaml", Usage: "Output format: yaml or json"},
					&cli.BoolFlag{Name: "global", Aliases: []string{"g"}, Usage: "Prefer global config when showing embedded sections"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					// Load merged hooks config (project over global, including embedded and legacy)
					hooksCfg, err := btconfig.LoadHooksConfig()
					if err != nil {
						return fmt.Errorf("load hooks config: %v", err)
					}

					// Load embedded blocked URLs for display (prefer project unless --global)
					useGlobal := cmd.Bool("global")
					cfgPath, err := btconfig.GetLogConfigPath(useGlobal)
					if err != nil {
						return fmt.Errorf("get config path: %v", err)
					}
					logCfg, err := btconfig.LoadLogConfig(cfgPath)
					if err != nil {
						return fmt.Errorf("load main config: %v", err)
					}

					// Build output view
					out := map[string]interface{}{
						"customHooks": hooksCfg,
					}
					if len(logCfg.BlockedURLs) > 0 {
						out["blockedUrls"] = logCfg.BlockedURLs
					}

					switch strings.ToLower(cmd.String("format")) {
					case "json":
						b, err := json.MarshalIndent(out, "", "  ")
						if err != nil {
							return err
						}
						fmt.Println(string(b))
					default:
						b, err := yaml.Marshal(out)
						if err != nil {
							return err
						}
						fmt.Print(string(b))
					}
					return nil
				},
			},
			{
				Name:  "blocked",
				Usage: "Manage blocked URL prefixes used by fetch-blocker",
				Commands: []*cli.Command{
					{
						Name:  "list",
						Usage: "List blocked URL prefixes",
						Flags: []cli.Flag{&cli.BoolFlag{Name: "global", Aliases: []string{"g"}}},
						Action: func(ctx context.Context, cmd *cli.Command) error {
							global := cmd.Bool("global")
							path, err := btconfig.GetLogConfigPath(global)
							if err != nil {
								return err
							}
							lc, err := btconfig.LoadLogConfig(path)
							if err != nil {
								return err
							}
							scope := "project"
							if global {
								scope = "global"
							}
							fmt.Printf("Blocked URLs (%s config: %s):\n", scope, path)
							if len(lc.BlockedURLs) == 0 {
								fmt.Println("(none)")
								return nil
							}
							for _, b := range lc.BlockedURLs {
								if b.Suggestion != "" {
									fmt.Printf("- %s | %s\n", b.Prefix, b.Suggestion)
								} else {
									fmt.Printf("- %s\n", b.Prefix)
								}
							}
							return nil
						},
					},
					{
						Name:  "add",
						Usage: "Add a blocked URL prefix",
						Flags: []cli.Flag{
							&cli.BoolFlag{Name: "global", Aliases: []string{"g"}},
							&cli.StringFlag{Name: "suggestion", Aliases: []string{"s"}},
						},
						ArgsUsage: "<prefix>",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							args := cmd.Args().Slice()
							if len(args) != 1 {
								return fmt.Errorf("exactly one argument required: <prefix>")
							}
							prefix := strings.TrimSpace(args[0])
							if prefix == "" {
								return fmt.Errorf("prefix cannot be empty")
							}
							global := cmd.Bool("global")
							suggestion := cmd.String("suggestion")
							path, err := btconfig.GetLogConfigPath(global)
							if err != nil {
								return err
							}
							lc, err := btconfig.LoadLogConfig(path)
							if err != nil {
								return err
							}
							// Check duplicate
							for _, b := range lc.BlockedURLs {
								if b.Prefix == prefix {
									fmt.Println("Prefix already present; no change.")
									return nil
								}
							}
							lc.BlockedURLs = append(lc.BlockedURLs, btconfig.BlockedURL{Prefix: prefix, Suggestion: suggestion})
							if err := btconfig.SaveLogConfig(path, lc); err != nil {
								return err
							}
							fmt.Printf("Added blocked prefix to %s: %s\n", path, prefix)
							return nil
						},
					},
					{
						Name:      "remove",
						Usage:     "Remove a blocked URL prefix",
						Flags:     []cli.Flag{&cli.BoolFlag{Name: "global", Aliases: []string{"g"}}},
						ArgsUsage: "<prefix>",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							args := cmd.Args().Slice()
							if len(args) != 1 {
								return fmt.Errorf("exactly one argument required: <prefix>")
							}
							prefix := strings.TrimSpace(args[0])
							global := cmd.Bool("global")
							path, err := btconfig.GetLogConfigPath(global)
							if err != nil {
								return err
							}
							lc, err := btconfig.LoadLogConfig(path)
							if err != nil {
								return err
							}
							filtered := lc.BlockedURLs[:0]
							removed := false
							for _, b := range lc.BlockedURLs {
								if b.Prefix == prefix {
									removed = true
									continue
								}
								filtered = append(filtered, b)
							}
							if !removed {
								fmt.Println("Prefix not found; no change.")
								return nil
							}
							lc.BlockedURLs = filtered
							if err := btconfig.SaveLogConfig(path, lc); err != nil {
								return err
							}
							fmt.Printf("Removed blocked prefix from %s: %s\n", path, prefix)
							return nil
						},
					},
					{
						Name:  "clear",
						Usage: "Clear all blocked URL prefixes",
						Flags: []cli.Flag{&cli.BoolFlag{Name: "global", Aliases: []string{"g"}}},
						Action: func(ctx context.Context, cmd *cli.Command) error {
							global := cmd.Bool("global")
							path, err := btconfig.GetLogConfigPath(global)
							if err != nil {
								return err
							}
							lc, err := btconfig.LoadLogConfig(path)
							if err != nil {
								return err
							}
							if len(lc.BlockedURLs) == 0 {
								fmt.Println("Blocked URLs already empty; no change.")
								return nil
							}
							lc.BlockedURLs = nil
							if err := btconfig.SaveLogConfig(path, lc); err != nil {
								return err
							}
							fmt.Printf("Cleared blocked URLs in %s\n", path)
							return nil
						},
					},
				},
			},
			{
				Name:      "sync",
				Usage:     "Sync custom hooks from blues-traveler-config.json into Claude settings",
				ArgsUsage: "[group]",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "global", Aliases: []string{"g"}, Usage: "Sync to global settings (~/.claude/settings.json)"},
					&cli.BoolFlag{Name: "dry-run", Aliases: []string{"n"}, Usage: "Show intended changes without writing"},
					&cli.StringFlag{Name: "event", Aliases: []string{"e"}, Usage: "Restrict sync to a single event (e.g., PreToolUse, PostToolUse)"},
					&cli.StringFlag{Name: "matcher", Aliases: []string{"m"}, Value: "*", Usage: "Default tool matcher for events (e.g., '*')"},
					&cli.StringFlag{Name: "post-matcher", Value: "Edit,Write", Usage: "Matcher for PostToolUse when not overridden"},
					&cli.IntFlag{Name: "timeout", Aliases: []string{"t"}, Usage: "Override timeout in seconds for installed commands"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					args := cmd.Args().Slice()
					var groupFilter string
					if len(args) > 0 {
						if len(args) > 1 {
							return fmt.Errorf("at most one [group] argument is allowed")
						}
						groupFilter = args[0]
					}

					useGlobal := cmd.Bool("global")
					dry := cmd.Bool("dry-run")
					eventFilter := strings.TrimSpace(cmd.String("event"))
					defaultMatcher := cmd.String("matcher")
					postMatcher := cmd.String("post-matcher")
					timeoutOverride := cmd.Int("timeout")

					// Load config (embedded + legacy merge)
					hooksCfg, err := btconfig.LoadHooksConfig()
					if err != nil {
						return fmt.Errorf("load hooks config: %v", err)
					}
					if hooksCfg == nil || len(*hooksCfg) == 0 {
						fmt.Println("No custom hooks found in config.")
						return nil
					}

					// Load settings
					settingsPath, err := btconfig.GetSettingsPath(useGlobal)
					if err != nil {
						return err
					}
					settings, err := btconfig.LoadSettings(settingsPath)
					if err != nil {
						return err
					}

					// Resolve a stable blues-traveler path for settings entries:
					// prefer PATH lookup, then local ./blues-traveler, then current executable.
					execPath := func() string {
						if p, err := exec.LookPath("blues-traveler"); err == nil && p != "" {
							return p
						}
						if _, err := os.Stat("./blues-traveler"); err == nil {
							if abs, err2 := filepath.Abs("./blues-traveler"); err2 == nil {
								return abs
							}
							return "./blues-traveler"
						}
						if p, err := os.Executable(); err == nil {
							return p
						}
						return "blues-traveler"
					}()

					// Helper to choose matcher
					pickMatcher := func(event string) string {
						if event == "PostToolUse" {
							return postMatcher
						}
						return defaultMatcher
					}

					changed := 0

					// Iterate groups
					for groupName, group := range *hooksCfg {
						if groupFilter != "" && groupName != groupFilter {
							continue
						}

						// Prune existing settings for this group (optionally event-filtered)
						removed := btconfig.RemoveConfigGroupFromSettings(settings, groupName, eventFilter)
						if removed > 0 {
							fmt.Printf("Pruned %d entries for group '%s'%s\n", removed, groupName, func() string {
								if eventFilter != "" {
									return " (event: " + eventFilter + ")"
								}
								return ""
							}())
						}

						// Add current definitions
						for eventName, ev := range group {
							if eventFilter != "" && eventFilter != eventName {
								continue
							}
							for _, job := range ev.Jobs {
								if job.Name == "" {
									continue
								}
								// Build command to run this job
								hookCommand := fmt.Sprintf("%s run config:%s:%s", execPath, groupName, job.Name)
								// Timeout preference: CLI override > job.Timeout
								var timeout *int
								if timeoutOverride > 0 {
									timeout = &timeoutOverride
								} else if job.Timeout > 0 {
									t := job.Timeout
									timeout = &t
								}
								// Matcher
								matcher := pickMatcher(eventName)
								// Add to settings
								res := btconfig.AddHookToSettings(settings, eventName, matcher, hookCommand, timeout)
								_ = res
								changed++
								if dry {
									fmt.Printf("Would add: [%s] matcher=%q command=%q\n", eventName, matcher, hookCommand)
								}
							}
						}
					}

					if changed == 0 {
						fmt.Println("No changes detected.")
						return nil
					}

					if dry {
						fmt.Println("Dry run; not writing settings.")
						return nil
					}

					if err := btconfig.SaveSettings(settingsPath, settings); err != nil {
						return err
					}
					scope := "project"
					if useGlobal {
						scope = "global"
					}
					fmt.Printf("Synced %d entries into %s settings: %s\n", changed, scope, settingsPath)
					return nil
				},
			},
		},
	}
}
