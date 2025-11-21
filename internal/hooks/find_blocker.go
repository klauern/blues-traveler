package hooks

import (
	"context"
	"fmt"
	"strings"

	"github.com/brads3290/cchooks"
	"github.com/klauern/blues-traveler/internal/constants"
	"github.com/klauern/blues-traveler/internal/core"
)

// FindBlockerHook implements logic to block find commands and suggest fd instead
type FindBlockerHook struct {
	*core.BaseHook
}

// NewFindBlockerHook creates a new find blocker hook instance
func NewFindBlockerHook(ctx *core.HookContext) core.Hook {
	base := core.NewBaseHook("find-blocker", "Find Command Blocker", "Blocks find commands and suggests using fd instead for better performance", ctx)
	return &FindBlockerHook{BaseHook: base}
}

// Run executes the find blocker hook.
func (h *FindBlockerHook) Run() error {
	return h.StandardRun(h.preToolUseHandler, nil)
}

func (h *FindBlockerHook) preToolUseHandler(_ context.Context, event *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface {
	// Log detailed event data
	if h.Context().LoggingEnabled {
		details := make(map[string]interface{})
		rawData := make(map[string]interface{})
		rawData["tool_name"] = event.ToolName

		if event.ToolName == constants.ToolBash {
			if bash, err := event.AsBash(); err == nil {
				details["command"] = bash.Command
				details["description"] = bash.Description
			}
		}

		h.LogHookEvent("pre_tool_use_find_check", event.ToolName, rawData, details)
	}

	// Only check Bash commands
	if event.ToolName != constants.ToolBash {
		return cchooks.Approve()
	}

	bash, err := event.AsBash()
	if err != nil {
		h.LogError("find_blocker_error", event.ToolName, err)
		return cchooks.Block("failed to parse bash command")
	}

	// Check if this is a find command
	if blocked, suggestion := h.isFindCommand(bash.Command); blocked {
		h.LogBlock("find_blocker_block", constants.ToolBash, map[string]interface{}{
			"command":    bash.Command,
			"suggestion": suggestion,
		})
		// User-friendly message + detailed fd suggestions for agent
		return core.BlockWithMessages(
			"The 'find' command was blocked. Please use 'fd' instead for better performance.",
			suggestion,
		)
	}

	// Log approved commands
	h.LogApproval("find_blocker_approved", constants.ToolBash, map[string]interface{}{
		"command": bash.Command,
	})

	return cchooks.Approve()
}

// isFindCommand checks if a command uses find and provides fd alternatives
func (h *FindBlockerHook) isFindCommand(command string) (bool, string) {
	// Normalize the command
	cmd := strings.TrimSpace(command)
	tokens := strings.Fields(cmd)

	if len(tokens) == 0 {
		return false, ""
	}

	// Check if the command starts with find or contains find in pipes
	if tokens[0] == "find" || h.containsFindInPipeline(cmd) {
		return true, h.generateFdSuggestion(cmd)
	}

	return false, ""
}

// containsFindInPipeline checks if find is used in a pipeline (e.g., "find . -name '*.go' | xargs grep")
func (h *FindBlockerHook) containsFindInPipeline(command string) bool {
	// Look for patterns like "find ..." in pipes or command substitutions
	// But avoid matching "find" within quoted strings

	// Simple heuristic: check for find with typical command-like contexts
	// This is not perfect but catches most practical cases
	patterns := []string{
		" find ",   // find with spaces around it
		"|find ",   // piped to find
		"$(find ",  // command substitution
		"`find ",   // backtick command substitution
		"; find ",  // after semicolon
		"&& find ", // after &&
		"|| find ", // after ||
	}

	for _, pattern := range patterns {
		if strings.Contains(command, pattern) {
			// Additional check: make sure it's not in a quoted string
			if !core.IsInQuotedString(command, pattern) {
				return true
			}
		}
	}

	return false
}

// generateFdSuggestion creates helpful fd command suggestions based on common find patterns
func (h *FindBlockerHook) generateFdSuggestion(findCommand string) string {
	cmd := strings.TrimSpace(findCommand)

	baseMessage := fmt.Sprintf("Command blocked: 'find' usage detected. Use 'fd' instead for better performance and usability.\n\nOriginal: %s", cmd)

	// Try to provide specific suggestions based on the command
	// Check for more specific patterns first (order matters)
	if strings.Contains(cmd, "-type f") {
		return fmt.Sprintf("%s\n\nSuggestion: Use 'fd --type f pattern' for files only", baseMessage)
	}
	if strings.Contains(cmd, "-type d") {
		return fmt.Sprintf("%s\n\nSuggestion: Use 'fd --type d pattern' for directories only", baseMessage)
	}
	if strings.Contains(cmd, "-maxdepth") {
		return fmt.Sprintf("%s\n\nSuggestion: Use 'fd --max-depth N pattern' to limit search depth", baseMessage)
	}
	if strings.Contains(cmd, "-iname") {
		return fmt.Sprintf("%s\n\nSuggestion: Use 'fd --ignore-case pattern' for case-insensitive search", baseMessage)
	}
	if strings.Contains(cmd, "-name") {
		return fmt.Sprintf("%s\n\nSuggestion: Use 'fd pattern' instead of 'find . -name pattern'", baseMessage)
	}

	// Generic suggestions for common patterns
	if strings.Contains(cmd, "find .") {
		return fmt.Sprintf("%s\n\nGeneric examples:\n- find . -name '*.go' → fd '\\.go$'\n- find . -type f -name '*.txt' → fd --type f '\\.txt$'\n- find /path -maxdepth 2 -name pattern → fd --max-depth 2 pattern /path", baseMessage)
	}

	// Fallback generic message
	return fmt.Sprintf("%s\n\nTip: 'fd' is faster, has better defaults, and simpler syntax. See 'fd --help' for usage.", baseMessage)
}
