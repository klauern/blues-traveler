package hooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/brads3290/cchooks"
	"github.com/klauern/blues-traveler/internal/config"
	"github.com/klauern/blues-traveler/internal/core"
)

// ConfigHook implements running jobs from hooks.yml groups
type ConfigHook struct {
	*core.BaseHook
	job         config.HookJob
	event       string
	groupName   string
	envProvider core.EnvironmentProvider
	lastRaw     string
}

// NewConfigHook constructs a hook from config data
func NewConfigHook(groupName, jobName string, job config.HookJob, event string, ctx *core.HookContext) core.Hook {
	key := fmt.Sprintf("config:%s:%s", groupName, jobName)
	base := core.NewBaseHook(key, jobName, fmt.Sprintf("Config job '%s' for %s", jobName, event), ctx)
	return &ConfigHook{
		BaseHook:    base,
		job:         job,
		event:       event,
		groupName:   groupName,
		envProvider: core.NewClaudeCodeEnvironmentProvider(),
	}
}

// CursorHookResponse represents the JSON response format from Cursor-compatible hooks
// Spec: https://cursor.com/docs/agent/hooks
type CursorHookResponse struct {
	Permission   string `json:"permission"`   // "allow", "deny", or "ask"
	UserMessage  string `json:"userMessage"`  // Message displayed to the user
	AgentMessage string `json:"agentMessage"` // Message sent to the AI agent
	Continue     *bool  `json:"continue"`     // Whether to continue execution (nil if not specified)
}

// hookExecutionResult captures the result of running a hook command
type hookExecutionResult struct {
	exitCode int
	stdout   string
	stderr   string
	err      error
}

// parseCursorResponse attempts to parse JSON output from a hook script
// Returns nil if output is not valid JSON or doesn't match Cursor format
func parseCursorResponse(output string) (*CursorHookResponse, error) {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" || !strings.HasPrefix(trimmed, "{") {
		return nil, nil // Not JSON, use fallback behavior
	}

	var response CursorHookResponse
	if err := json.Unmarshal([]byte(trimmed), &response); err != nil {
		return nil, fmt.Errorf("invalid JSON in hook output: %w", err)
	}

	return &response, nil
}

// Run executes the custom hook based on its configured event type and matcher
func (h *ConfigHook) Run() error {
	if !h.IsEnabled() {
		return nil
	}
	// For events not natively supported by cchooks (anything other than Pre/Post),
	// handle via raw JSON read from stdin to avoid "unknown event type" errors.
	if h.event != string(core.PreToolUseEvent) && h.event != string(core.PostToolUseEvent) {
		return h.processRawFromStdin()
	}
	var pre func(context.Context, *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface
	var post func(context.Context, *cchooks.PostToolUseEvent) cchooks.PostToolUseResponseInterface
	raw := h.rawHandler()

	switch h.event {
	case string(core.PreToolUseEvent):
		pre = h.preHandler
	case string(core.PostToolUseEvent):
		post = h.postHandler
	default:
		// For events not supported by cchooks, just no-op runner
	}

	runner := h.Context().RunnerFactory(pre, post, raw)
	runner.Run()
	return nil
}

func (h *ConfigHook) shouldRun(env map[string]string) (bool, error) {
	if strings.TrimSpace(h.job.Skip) != "" {
		ok, err := core.EvalExpression(h.job.Skip, env)
		if err != nil {
			return false, err
		}
		if ok {
			return false, nil
		}
	}
	if strings.TrimSpace(h.job.Only) != "" {
		ok, err := core.EvalExpression(h.job.Only, env)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

func (h *ConfigHook) runCommandWithEnv(env map[string]string) (*hookExecutionResult, error) {
	// Prepare environment
	mergedEnv := os.Environ()
	for k, v := range env {
		mergedEnv = append(mergedEnv, fmt.Sprintf("%s=%s", k, v))
	}
	for k, v := range h.job.Env {
		mergedEnv = append(mergedEnv, fmt.Sprintf("%s=%s", k, v))
	}

	// Build command (with timeout-aware context)
	cmdCtx := context.Background()
	if h.job.Timeout > 0 {
		var cancel context.CancelFunc
		cmdCtx, cancel = context.WithTimeout(cmdCtx, time.Duration(h.job.Timeout)*time.Second)
		defer cancel()
	}
	cmd := exec.CommandContext(cmdCtx, "bash", "-lc", h.job.Run) // #nosec G204 -- user-configured command execution is intentional and safe

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// If we have the original raw JSON for this event, pass it to child stdin so
	// nested blues-traveler invocations can consume it.
	if h.lastRaw != "" {
		cmd.Stdin = strings.NewReader(h.lastRaw)
	}
	if h.job.WorkDir != "" {
		cmd.Dir = h.job.WorkDir
	}
	cmd.Env = mergedEnv

	// Run and capture result
	result := &hookExecutionResult{
		stdout: stdout.String(),
		stderr: stderr.String(),
	}

	err := cmd.Run()
	result.err = err

	if err != nil {
		// Translate deadline exceeded into a friendly timeout error
		if cmdCtx.Err() == context.DeadlineExceeded && h.job.Timeout > 0 {
			return result, fmt.Errorf("command timed out after %ds", h.job.Timeout)
		}
		// Try to extract exit code
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.exitCode = exitErr.ExitCode()
		} else {
			result.exitCode = 1
		}
		return result, err
	}

	result.exitCode = 0
	return result, nil
}

func (h *ConfigHook) preHandler(ctx context.Context, ev *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface {
	c := core.BuildPreToolUseContext(ctx, ev)
	env := h.envProvider.GetEnvironment(string(core.PreToolUseEvent), c)

	result, err := h.executeIfShouldRunWithResult(env)
	if err != nil {
		// User-friendly message + technical details for agent
		userMsg := fmt.Sprintf("Hook '%s' execution failed", h.job.Name)
		agentMsg := err.Error()
		return core.BlockWithMessages(userMsg, agentMsg)
	}

	// Try to parse Cursor JSON response
	if result != nil && result.stdout != "" {
		cursorResp, parseErr := parseCursorResponse(result.stdout)

		// Rule 3: Invalid JSON = block with "hook broken" message
		if parseErr != nil {
			userMsg := fmt.Sprintf("Hook '%s' returned invalid JSON", h.job.Name)
			agentMsg := fmt.Sprintf("Hook output parsing failed: %v. Output: %s", parseErr, result.stdout)
			return core.BlockWithMessages(userMsg, agentMsg)
		}

		// Rule 2: Partial JSON = proceed with available fields
		if cursorResp != nil {
			if resp, _ := translateCursorResponse(h.job.Name, core.PreToolUseEvent, h, cursorResp); resp != nil {
				return resp
			}
		}
	}

	// Rule 1: Non-zero exit + no JSON = block with alert + error message
	if result != nil && result.exitCode != 0 {
		userMsg := fmt.Sprintf("Hook '%s' failed with exit code %d", h.job.Name, result.exitCode)
		agentMsg := fmt.Sprintf("Exit code: %d, stderr: %s", result.exitCode, result.stderr)
		return core.BlockWithMessages(userMsg, agentMsg)
	}

	return cchooks.Approve()
}

func (h *ConfigHook) postHandler(ctx context.Context, ev *cchooks.PostToolUseEvent) cchooks.PostToolUseResponseInterface {
	c := core.BuildPostToolUseContext(ctx, ev)
	env := h.envProvider.GetEnvironment(string(core.PostToolUseEvent), c)

	result, err := h.executeIfShouldRunWithResult(env)
	if err != nil {
		// User-friendly message + technical details for agent
		userMsg := fmt.Sprintf("Hook '%s' execution failed", h.job.Name)
		agentMsg := err.Error()
		return core.PostBlockWithMessages(userMsg, agentMsg)
	}

	// Try to parse Cursor JSON response
	if result != nil && result.stdout != "" {
		cursorResp, parseErr := parseCursorResponse(result.stdout)

		// Rule 3: Invalid JSON = block with "hook broken" message
		if parseErr != nil {
			userMsg := fmt.Sprintf("Hook '%s' returned invalid JSON", h.job.Name)
			agentMsg := fmt.Sprintf("Hook output parsing failed: %v. Output: %s", parseErr, result.stdout)
			return core.PostBlockWithMessages(userMsg, agentMsg)
		}

		// Rule 2: Partial JSON = proceed with available fields
		if cursorResp != nil {
			if _, resp := translateCursorResponse(h.job.Name, core.PostToolUseEvent, h, cursorResp); resp != nil {
				return resp
			}
		}
	}

	// Rule 1: Non-zero exit + no JSON = block with alert + error message
	if result != nil && result.exitCode != 0 {
		userMsg := fmt.Sprintf("Hook '%s' failed with exit code %d", h.job.Name, result.exitCode)
		agentMsg := fmt.Sprintf("Exit code: %d, stderr: %s", result.exitCode, result.stderr)
		return core.PostBlockWithMessages(userMsg, agentMsg)
	}

	return cchooks.Allow()
}

// executeIfShouldRun checks if the hook should run and executes it (legacy interface)
func (h *ConfigHook) executeIfShouldRun(env map[string]string) error {
	_, err := h.executeIfShouldRunWithResult(env)
	return err
}

// executeIfShouldRunWithResult checks if the hook should run and executes it, returning the result
func (h *ConfigHook) executeIfShouldRunWithResult(env map[string]string) (*hookExecutionResult, error) {
	ok, err := h.shouldRun(env)
	if err != nil {
		return nil, fmt.Errorf("config hook error: %w", err)
	}
	if !ok {
		return nil, nil
	}
	result, err := h.runCommandWithEnv(env)
	if err != nil {
		return result, fmt.Errorf("job '%s' failed: %w", h.job.Name, err)
	}
	return result, nil
}

// handleCursorResponsePre processes a Cursor JSON response for PreToolUse events
// rawHandler handles unsupported events (e.g., UserPromptSubmit) by parsing the raw JSON
// and executing the configured job when the event name matches this hook's event.
func (h *ConfigHook) rawHandler() func(context.Context, string) *cchooks.RawResponse {
	return func(_ context.Context, rawJSON string) *cchooks.RawResponse {
		var rawEvent map[string]any
		if err := json.Unmarshal([]byte(rawJSON), &rawEvent); err != nil {
			return nil
		}
		evName, _ := rawEvent["hook_event_name"].(string)
		if evName == "" || evName != h.event {
			return nil
		}
		// Store raw JSON to feed to any nested commands launched by this hook
		h.lastRaw = rawJSON
		// Build minimal context for env provider
		ctxData := map[string]any{}
		if v, ok := rawEvent["tool_name"].(string); ok {
			ctxData["tool_name"] = v
		}
		if v, ok := rawEvent["user_prompt"].(string); ok {
			ctxData["user_prompt"] = v
		}
		env := h.envProvider.GetEnvironment(evName, ctxData)
		if ok, err := h.shouldRun(env); err == nil && ok {
			_, _ = h.runCommandWithEnv(env)
		}
		return nil
	}
}

func (h *ConfigHook) processRawFromStdin() error {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil // fail open
	}
	handler := h.rawHandler()
	if handler != nil {
		handler(context.Background(), string(data))
	}
	return nil
}
