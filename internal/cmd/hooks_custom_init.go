package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/klauern/blues-traveler/internal/config"
	"github.com/urfave/cli/v3"
)

// generateSampleConfig creates the sample configuration content
func generateSampleConfig(global bool, group string) string {
	if global {
		return fmt.Sprintf(`# Global hooks configuration for group '%s'
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
	}

	return fmt.Sprintf(`# Sample hooks configuration for group '%s'
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

// sanitizeFileName validates and sanitizes a filename to prevent path traversal
func sanitizeFileName(fileName string) (string, error) {
	// Reject empty, ".", or ".." names
	if fileName == "" || fileName == "." || fileName == ".." {
		return "", fmt.Errorf("invalid filename: %q", fileName)
	}

	// Reject absolute paths
	if filepath.IsAbs(fileName) {
		return "", fmt.Errorf("absolute paths not allowed: %q", fileName)
	}

	// Reject paths containing separators
	if strings.Contains(fileName, string(filepath.Separator)) || strings.Contains(fileName, "/") || strings.Contains(fileName, "\\") {
		return "", fmt.Errorf("path separators not allowed in filename: %q", fileName)
	}

	// Use filepath.Base as additional safety (should be a no-op after above checks)
	base := filepath.Base(fileName)

	// Ensure .yml or .yaml extension
	if !strings.HasSuffix(strings.ToLower(base), ".yml") && !strings.HasSuffix(strings.ToLower(base), ".yaml") {
		base += ".yml"
	}

	return base, nil
}

// writePerGroupConfig writes a per-group config file to .claude/hooks/<name>.yml
func writePerGroupConfig(global bool, fileName string, sample string, overwrite bool) (string, error) {
	dir, err := config.EnsureClaudeDir(global)
	if err != nil {
		return "", err
	}

	hooksDir := filepath.Join(dir, "hooks")
	if err := os.MkdirAll(hooksDir, 0o750); err != nil {
		return "", err
	}

	// Sanitize filename to prevent path traversal
	base, err := sanitizeFileName(fileName)
	if err != nil {
		return "", err
	}

	target := filepath.Join(hooksDir, base)
	if !overwrite {
		if _, err := os.Stat(target); err == nil {
			fmt.Printf("File already exists: %s (use --overwrite to replace)\n", target)
			return target, nil
		}
	}

	if err := os.WriteFile(target, []byte(sample), 0o600); err != nil {
		return "", err
	}

	return target, nil
}

// writeGlobalDefaultConfig creates a minimal global configuration
func writeGlobalDefaultConfig(overwrite bool) (string, error) {
	configPath, err := config.GetLogConfigPath(true)
	if err != nil {
		return "", err
	}

	// Load existing config or create default
	logCfg, err := config.LoadLogConfig(configPath)
	if err != nil {
		return "", err
	}

	// Check for existing config without overwrite
	if !overwrite && logCfg.CustomHooks != nil && len(logCfg.CustomHooks) > 0 {
		fmt.Printf("File already exists: %s (use --overwrite to replace)\n", configPath)
		return configPath, nil
	}

	// Create minimal hooks structure (empty)
	logCfg.CustomHooks = config.CustomHooksConfig{}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0o750); err != nil {
		return "", err
	}

	// Save the minimal config
	if err := config.SaveLogConfig(configPath, logCfg); err != nil {
		return "", err
	}

	return configPath, nil
}

// newHooksCustomInitCommand creates the init command for custom hooks
func newHooksCustomInitCommand() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Create a sample hooks configuration file",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "global", Aliases: []string{"g"}, Usage: "Create in ~/.claude"},
			&cli.BoolFlag{Name: "overwrite", Usage: "Overwrite existing file if present"},
			&cli.StringFlag{Name: "group", Aliases: []string{"G"}, Value: "example", Usage: "Group name for this config"},
			&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Usage: "Filename for per-group config (writes .claude/hooks/<name>.yml)"},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			global := cmd.Bool("global")
			overwrite := cmd.Bool("overwrite")
			group := cmd.String("group")
			fileName := cmd.String("name")

			sample := generateSampleConfig(global, group)

			var path string
			var err error

			// If --name provided, create .claude/hooks/<name>.yml
			switch {
			case fileName != "":
				path, err = writePerGroupConfig(global, fileName, sample, overwrite)
				if err != nil {
					return err
				}
			case global:
				path, err = writeGlobalDefaultConfig(overwrite)
				if err != nil {
					return err
				}
			default:
				// For project configs, use existing sample logic
				path, err = config.WriteSampleHooksConfig(global, sample, overwrite)
				if err != nil {
					return err
				}
			}

			fmt.Printf("Created sample hooks config at %s\n", path)
			return nil
		},
	}
}
