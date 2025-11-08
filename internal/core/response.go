package core

import (
	"encoding/json"

	"github.com/brads3290/cchooks"
)

// DualMessagePreToolResponse wraps PreToolUse responses with dual messages.
// It embeds the actual cchooks.PreToolUseResponse to satisfy the interface,
// and adds separate messages for end-users and AI agents.
type DualMessagePreToolResponse struct {
	*cchooks.PreToolUseResponse
	userMessage  string
	agentMessage string
}

// DualMessagePostToolResponse wraps PostToolUse responses with dual messages.
// It embeds the actual cchooks.PostToolUseResponse to satisfy the interface,
// and adds separate messages for end-users and AI agents.
type DualMessagePostToolResponse struct {
	*cchooks.PostToolUseResponse
	userMessage  string
	agentMessage string
}

// CursorPermissionResponse represents responses that can map to Cursor's permission schema.
type CursorPermissionResponse interface {
	cchooks.PreToolUseResponseInterface
	CursorPermission() string
	ClaudeFallback() cchooks.PreToolUseResponseInterface
}

// AskPreToolResponse is a specialized response for permission prompts.
//
// Cursor expects "permission: ask" semantics, while Claude Code continues to
// interpret the fallback approve response.
type AskPreToolResponse struct {
	*DualMessagePreToolResponse
}

// CursorPermission returns the Cursor permission token.
func (r *AskPreToolResponse) CursorPermission() string {
	return "ask"
}

// ClaudeFallback returns the Claude-compatible response when Cursor permissions aren't supported.
func (r *AskPreToolResponse) ClaudeFallback() cchooks.PreToolUseResponseInterface {
	return r.DualMessagePreToolResponse
}

// MarshalJSON emits Cursor-compatible JSON for ask responses.
func (r *AskPreToolResponse) MarshalJSON() ([]byte, error) {
	payload := struct {
		Permission   string `json:"permission"`
		UserMessage  string `json:"userMessage,omitempty"`
		AgentMessage string `json:"agentMessage,omitempty"`
	}{
		Permission:   r.CursorPermission(),
		UserMessage:  r.GetUserMessage(),
		AgentMessage: r.GetAgentMessage(),
	}
	return json.Marshal(payload)
}

// BlockWithMessages creates a blocking response for PreToolUse events with
// separate messages for users and agents.
//
// If agentMsg is omitted, userMsg is sent to both audiences.
// If agentMsg is provided, userMsg goes to the user and agentMsg goes to the agent.
// If multiple agentMsg values are provided, only the first is used.
//
// Usage:
//
//	// Single message (same for both audiences)
//	return core.BlockWithMessages("Operation blocked")
//
//	// Dual messages (different for each audience)
//	return core.BlockWithMessages(
//	    "This command was blocked for security reasons.",
//	    "Blocked dangerous pattern: sudo. Type: privilege_escalation",
//	)
func BlockWithMessages(userMsg string, agentMsg ...string) cchooks.PreToolUseResponseInterface {
	agent := userMsg
	if len(agentMsg) > 0 {
		agent = agentMsg[0]
	}

	return &DualMessagePreToolResponse{
		PreToolUseResponse: cchooks.Block(userMsg),
		userMessage:        userMsg,
		agentMessage:       agent,
	}
}

// ApproveWithMessages creates an approval response for PreToolUse events with
// optional context messages for users and agents.
//
// If agentMsg is omitted, userMsg is sent to both audiences.
// If agentMsg is provided, userMsg goes to the user and agentMsg goes to the agent.
//
// Usage:
//
//	// Simple approval (no messages)
//	return cchooks.Approve()  // Still works!
//
//	// Approval with context
//	return core.ApproveWithMessages(
//	    "Security check passed",
//	    "All patterns validated successfully",
//	)
func ApproveWithMessages(userMsg string, agentMsg ...string) cchooks.PreToolUseResponseInterface {
	agent := userMsg
	if len(agentMsg) > 0 {
		agent = agentMsg[0]
	}

	return &DualMessagePreToolResponse{
		PreToolUseResponse: cchooks.Approve(),
		userMessage:        userMsg,
		agentMessage:       agent,
	}
}

// AskWithMessages creates a permission request for PreToolUse events with
// context messages for users and agents. This function provides a consistent API
// for "ask" mode across different platforms.
//
// Platform Behavior:
//   - Cursor IDE: Native "ask" support - prompts user for manual approval
//   - Claude Code: Falls back to Approve() with messages (Claude Code doesn't support ask mode)
//
// The fallback behavior is intentional since blues-traveler works with both Cursor hooks
// and Claude Code hooks. When running under Cursor, the native ask behavior is used.
// When running under Claude Code (via cchooks), it gracefully degrades to approve with messages.
//
// If agentMsg is omitted, userMsg is sent to both audiences.
// If agentMsg is provided, userMsg goes to the user and agentMsg goes to the agent.
//
// Usage:
//
//	// Request user approval with context (works in Cursor, falls back in Claude Code)
//	return core.AskWithMessages(
//	    "This operation requires your confirmation",
//	    "Attempting to execute: sudo rm -rf /tmp/cache",
//	)
func AskWithMessages(userMsg string, agentMsg ...string) cchooks.PreToolUseResponseInterface {
	agent := userMsg
	if len(agentMsg) > 0 {
		agent = agentMsg[0]
	}

	dual := &DualMessagePreToolResponse{
		PreToolUseResponse: cchooks.Approve(),
		userMessage:        userMsg,
		agentMessage:       agent,
	}

	return &AskPreToolResponse{DualMessagePreToolResponse: dual}
}

// PostBlockWithMessages creates a blocking response for PostToolUse events with
// separate messages for users and agents.
//
// If agentMsg is omitted, userMsg is sent to both audiences.
// If agentMsg is provided, userMsg goes to the user and agentMsg goes to the agent.
//
// Usage:
//
//	// Single message
//	return core.PostBlockWithMessages("Operation failed")
//
//	// Dual messages with technical details for agent
//	return core.PostBlockWithMessages(
//	    "Code formatting failed",
//	    fmt.Sprintf("Black formatter failed: %s\nStderr: %v", filePath, err),
//	)
func PostBlockWithMessages(userMsg string, agentMsg ...string) cchooks.PostToolUseResponseInterface {
	agent := userMsg
	if len(agentMsg) > 0 {
		agent = agentMsg[0]
	}

	return &DualMessagePostToolResponse{
		PostToolUseResponse: cchooks.PostBlock(userMsg),
		userMessage:         userMsg,
		agentMessage:        agent,
	}
}

// AllowWithMessages creates an allow response for PostToolUse events with
// optional status messages for users and agents.
//
// If agentMsg is omitted, userMsg is sent to both audiences.
// If agentMsg is provided, userMsg goes to the user and agentMsg goes to the agent.
//
// Usage:
//
//	// Simple allow (no messages)
//	return cchooks.Allow()  // Still works!
//
//	// Allow with status
//	return core.AllowWithMessages(
//	    "Operation completed successfully",
//	    "Job completed in 245ms, 0 issues found",
//	)
func AllowWithMessages(userMsg string, agentMsg ...string) cchooks.PostToolUseResponseInterface {
	agent := userMsg
	if len(agentMsg) > 0 {
		agent = agentMsg[0]
	}

	return &DualMessagePostToolResponse{
		PostToolUseResponse: cchooks.Allow(),
		userMessage:         userMsg,
		agentMessage:        agent,
	}
}

// GetUserMessage returns the message intended for the end-user.
func (r *DualMessagePreToolResponse) GetUserMessage() string {
	return r.userMessage
}

// GetAgentMessage returns the message intended for the AI agent.
func (r *DualMessagePreToolResponse) GetAgentMessage() string {
	return r.agentMessage
}

// GetUserMessage returns the message intended for the end-user.
func (r *DualMessagePostToolResponse) GetUserMessage() string {
	return r.userMessage
}

// GetAgentMessage returns the message intended for the AI agent.
func (r *DualMessagePostToolResponse) GetAgentMessage() string {
	return r.agentMessage
}
