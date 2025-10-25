package cursor

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/klauern/blues-traveler/internal/constants"
)

// DEPRECATED: Wrapper scripts are no longer used as of v0.3.0-alpha.
// The installation now registers blues-traveler commands directly in ~/.cursor/hooks.json
// with the --cursor-mode flag, eliminating the need for intermediate shell scripts.
// This file is retained for reference and backward compatibility but is not used
// in the current implementation.

// WrapperConfig holds the configuration for generating a wrapper script
type WrapperConfig struct {
	HookKey     string
	CursorEvent string
	Matcher     string
	BinaryPath  string
	Description string
}

// GenerateWrapper creates a bash wrapper script that translates Cursor JSON to blues-traveler format
func GenerateWrapper(config WrapperConfig) (string, error) {
	tmpl, err := template.New("wrapper").Parse(wrapperTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse wrapper template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, config); err != nil {
		return "", fmt.Errorf("failed to execute wrapper template: %w", err)
	}

	return buf.String(), nil
}

// WrapperScriptPath returns the recommended path for a wrapper script
func WrapperScriptPath(hookKey, event string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	filename := fmt.Sprintf("%s-%s-%s.sh", constants.BinaryName, hookKey, event)
	return filepath.Join(home, ".cursor", "hooks", filename), nil
}

const wrapperTemplate = `#!/bin/bash
# Auto-generated Cursor hook wrapper for {{.BinaryPath}}
# Hook: {{.HookKey}}
# Event: {{.CursorEvent}}
# Description: {{.Description}}
{{- if .Matcher }}
# Matcher: {{.Matcher}}
{{- end }}
#
# This script translates Cursor's JSON protocol to {{.BinaryPath}} environment variables

set -euo pipefail

# Read JSON input from stdin
input=$(cat)

# Parse common fields
export CONVERSATION_ID=$(echo "$input" | jq -r '.conversation_id // ""')
export GENERATION_ID=$(echo "$input" | jq -r '.generation_id // ""')
export EVENT_NAME=$(echo "$input" | jq -r '.hook_event_name // ""')
export WORKSPACE_ROOTS=$(echo "$input" | jq -r '.workspace_roots // [] | join(":")')

# Parse event-specific fields based on event type
case "$EVENT_NAME" in
  beforeShellExecution)
    export TOOL_NAME="Bash"
    export TOOL_ARGS=$(echo "$input" | jq -r '.command // ""')
    export CWD=$(echo "$input" | jq -r '.cwd // ""')
    ;;
  beforeMCPExecution)
    export TOOL_NAME=$(echo "$input" | jq -r '.tool_name // ""')
    export TOOL_ARGS=$(echo "$input" | jq -r '.tool_input // ""')
    export MCP_URL=$(echo "$input" | jq -r '.url // .command // ""')
    ;;
  afterFileEdit)
    export FILE_PATH=$(echo "$input" | jq -r '.file_path // ""')
    export FILE_EDITS=$(echo "$input" | jq -c '.edits // []')
    ;;
  beforeReadFile)
    export FILE_PATH=$(echo "$input" | jq -r '.file_path // ""')
    export FILE_CONTENT=$(echo "$input" | jq -r '.content // ""')
    ;;
  beforeSubmitPrompt)
    export USER_PROMPT=$(echo "$input" | jq -r '.prompt // ""')
    export PROMPT_ATTACHMENTS=$(echo "$input" | jq -c '.attachments // []')
    ;;
  stop)
    export STOP_STATUS=$(echo "$input" | jq -r '.status // ""')
    ;;
esac
{{- if .Matcher }}

# Apply matcher filter
# Note: Cursor doesn't support config-level matchers, so we implement it here
matcher="{{.Matcher}}"
# Convert wildcard matcher to valid regex (default * becomes .*)
if [[ "$matcher" == "*" ]]; then
  matcher=".*"
fi
check_value=""

case "$EVENT_NAME" in
  beforeShellExecution)
    check_value="$TOOL_ARGS"
    ;;
  afterFileEdit|beforeReadFile)
    check_value="$FILE_PATH"
    ;;
  beforeMCPExecution)
    check_value="$TOOL_NAME"
    ;;
esac

if [[ -n "$check_value" && -n "$matcher" ]]; then
  if ! echo "$check_value" | grep -E "$matcher" > /dev/null 2>&1; then
    # Matcher didn't match, allow operation
    echo '{"permission": "allow"}'
    exit 0
  fi
fi
{{- end }}

# Run blues-traveler in Cursor mode
if "{{.BinaryPath}}" hooks run "{{.HookKey}}" --cursor-mode <<< "$input"; then
  # Hook succeeded, allow operation
  exit 0
else
  # Hook failed, deny operation
  exit 3
fi
`
