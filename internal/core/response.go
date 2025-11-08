package core

import (
	"github.com/brads3290/cchooks"
)

// Decision constants for hook responses
// These extend the cchooks constants to support the "ask" permission mode
const (
	PreToolUseAsk  = "ask"
	PostToolUseAsk = "ask"
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

// Ask creates a PreToolUseResponse that requests user confirmation before proceeding.
// This is part of the 3-way permission model (allow/deny/ask) for Cursor compatibility.
func Ask(reason string) *cchooks.PreToolUseResponse {
	return &cchooks.PreToolUseResponse{Decision: PreToolUseAsk, Reason: reason}
}

// AskPost creates a PostToolUseResponse that requests user confirmation.
// This is part of the 3-way permission model (allow/deny/ask) for Cursor compatibility.
func AskPost(reason string) *cchooks.PostToolUseResponse {
	return &cchooks.PostToolUseResponse{Decision: PostToolUseAsk, Reason: reason}
}

// AskWithMessages creates an ask response for PreToolUse events with
// separate messages for users and agents. This prompts the user for manual
// approval before proceeding with the tool use.
//
// If agentMsg is omitted, userMsg is sent to both audiences.
// If agentMsg is provided, userMsg goes to the user and agentMsg goes to the agent.
//
// Usage:
//
//	// Ask with single message
//	return core.AskWithMessages("Do you want to proceed?")
//
//	// Ask with different messages for user and agent
//	return core.AskWithMessages(
//	    "Hook 'security-check' requests confirmation",
//	    "Potentially sensitive operation detected: modifying .env file",
//	)
func AskWithMessages(userMsg string, agentMsg ...string) cchooks.PreToolUseResponseInterface {
	agent := userMsg
	if len(agentMsg) > 0 {
		agent = agentMsg[0]
	}

	return &DualMessagePreToolResponse{
		PreToolUseResponse: Ask(userMsg),
		userMessage:        userMsg,
		agentMessage:       agent,
	}
}

// AskPostWithMessages creates an ask response for PostToolUse events with
// separate messages for users and agents. This prompts the user for manual
// approval after a tool has completed.
//
// If agentMsg is omitted, userMsg is sent to both audiences.
// If agentMsg is provided, userMsg goes to the user and agentMsg goes to the agent.
//
// Usage:
//
//	// Ask with single message
//	return core.AskPostWithMessages("Confirm the result?")
//
//	// Ask with different messages for user and agent
//	return core.AskPostWithMessages(
//	    "Hook 'audit' requests review",
//	    "Detected 5 changes to sensitive files - please review",
//	)
func AskPostWithMessages(userMsg string, agentMsg ...string) cchooks.PostToolUseResponseInterface {
	agent := userMsg
	if len(agentMsg) > 0 {
		agent = agentMsg[0]
	}

	return &DualMessagePostToolResponse{
		PostToolUseResponse: AskPost(userMsg),
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
