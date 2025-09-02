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
					sample := fmt.Sprintf(`# Sample hooks configuration for group '%s'
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
