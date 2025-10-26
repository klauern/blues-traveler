package hooks

import (
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

func (h *ConfigHook) runCommandWithEnv(env map[string]string) error {
	// Prepare environment
	mergedEnv := os.Environ()
	for k, v := range env {
		mergedEnv = append(mergedEnv, fmt.Sprintf("%s=%s", k, v))
	}
	for k, v := range h.job.Env {
		mergedEnv = append(mergedEnv, fmt.Sprintf("%s=%s", k, v))
	}

	// Build command
	cmd := exec.Command("bash", "-lc", h.job.Run) // #nosec G204 -- user-configured command execution is intentional and safe
	// If we have the original raw JSON for this event, pass it to child stdin so
	// nested blues-traveler invocations can consume it.
	if h.lastRaw != "" {
		cmd.Stdin = strings.NewReader(h.lastRaw)
	}
	if h.job.WorkDir != "" {
		cmd.Dir = h.job.WorkDir
	}
	cmd.Env = mergedEnv

	// Handle timeout
	var timer *time.Timer
	done := make(chan error, 1)
	go func() { done <- cmd.Run() }()

	if h.job.Timeout > 0 {
		timer = time.NewTimer(time.Duration(h.job.Timeout) * time.Second)
		defer timer.Stop()
		select {
		case err := <-done:
			return err
		case <-timer.C:
			_ = cmd.Process.Kill()
			return fmt.Errorf("command timed out after %ds", h.job.Timeout)
		}
	}
	return <-done
}

func (h *ConfigHook) preHandler(ctx context.Context, ev *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface {
	c := core.BuildPreToolUseContext(ctx, ev)
	env := h.envProvider.GetEnvironment(string(core.PreToolUseEvent), c)

	if err := h.executeIfShouldRun(env); err != nil {
		return cchooks.Block(err.Error())
	}
	return cchooks.Approve()
}

func (h *ConfigHook) postHandler(ctx context.Context, ev *cchooks.PostToolUseEvent) cchooks.PostToolUseResponseInterface {
	c := core.BuildPostToolUseContext(ctx, ev)
	env := h.envProvider.GetEnvironment(string(core.PostToolUseEvent), c)

	if err := h.executeIfShouldRun(env); err != nil {
		return cchooks.PostBlock(err.Error())
	}
	return cchooks.Allow()
}

// executeIfShouldRun checks if the hook should run and executes it
func (h *ConfigHook) executeIfShouldRun(env map[string]string) error {
	ok, err := h.shouldRun(env)
	if err != nil {
		return fmt.Errorf("config hook error: %v", err)
	}
	if !ok {
		return nil
	}
	if err := h.runCommandWithEnv(env); err != nil {
		return fmt.Errorf("job '%s' failed: %v", h.job.Name, err)
	}
	return nil
}

// rawHandler handles unsupported events (e.g., UserPromptSubmit) by parsing the raw JSON
// and executing the configured job when the event name matches this hook's event.
func (h *ConfigHook) rawHandler() func(context.Context, string) *cchooks.RawResponse {
	return func(_ context.Context, rawJSON string) *cchooks.RawResponse {
		var rawEvent map[string]interface{}
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
		ctxData := map[string]interface{}{}
		if v, ok := rawEvent["tool_name"].(string); ok {
			ctxData["tool_name"] = v
		}
		if v, ok := rawEvent["user_prompt"].(string); ok {
			ctxData["user_prompt"] = v
		}
		env := h.envProvider.GetEnvironment(evName, ctxData)
		if ok, err := h.shouldRun(env); err == nil && ok {
			_ = h.runCommandWithEnv(env)
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
