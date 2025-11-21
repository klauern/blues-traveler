package hooks

import (
	"fmt"
	"strings"

	"github.com/brads3290/cchooks"
	"github.com/klauern/blues-traveler/internal/core"
)

type hookEventLogger interface {
	LogHookEvent(event string, toolName string, rawData map[string]interface{}, details map[string]interface{})
}

type cursorResponseAdapter[T any] struct {
	approve             func() T
	approveWithMessages func(string, string) T
	block               func(string, string) T
	ask                 func(string, string) T
	askEvent            string
	askNote             string
}

func translateCursorResponse(jobName string, event core.EventType, logger hookEventLogger, resp *CursorHookResponse) (cchooks.PreToolUseResponseInterface, cchooks.PostToolUseResponseInterface) {
	if resp == nil {
		return nil, nil
	}

	switch event {
	case core.PreToolUseEvent:
		adapter := cursorResponseAdapter[cchooks.PreToolUseResponseInterface]{
			approve: func() cchooks.PreToolUseResponseInterface {
				return cchooks.Approve()
			},
			approveWithMessages: func(user, agent string) cchooks.PreToolUseResponseInterface {
				return core.ApproveWithMessages(user, agent)
			},
			block: func(user, agent string) cchooks.PreToolUseResponseInterface {
				return core.BlockWithMessages(user, agent)
			},
			ask: func(user, agent string) cchooks.PreToolUseResponseInterface {
				return core.AskWithMessages(user, agent)
			},
			askEvent: "hook_ask_permission",
			askNote:  "Ask mode: works in Cursor IDE, falls back to approve in Claude Code",
		}
		return translateCursorResponseWithAdapter(adapter, jobName, logger, resp), nil
	case core.PostToolUseEvent:
		adapter := cursorResponseAdapter[cchooks.PostToolUseResponseInterface]{
			approve: func() cchooks.PostToolUseResponseInterface {
				return cchooks.Allow()
			},
			approveWithMessages: func(user, agent string) cchooks.PostToolUseResponseInterface {
				return core.AllowWithMessages(user, agent)
			},
			block: func(user, agent string) cchooks.PostToolUseResponseInterface {
				return core.PostBlockWithMessages(user, agent)
			},
			ask: func(user, agent string) cchooks.PostToolUseResponseInterface {
				return core.AllowWithMessages(user, agent)
			},
			askEvent: "hook_ask_permission_post",
			askNote:  "Ask mode: works in Cursor IDE, falls back to allow in Claude Code",
		}
		return nil, translateCursorResponseWithAdapter(adapter, jobName, logger, resp)
	default:
		return nil, nil
	}
}

func translateCursorResponseWithAdapter[T any](adapter cursorResponseAdapter[T], jobName string, logger hookEventLogger, resp *CursorHookResponse) T {
	if resp.Continue != nil && !*resp.Continue {
		userMsg, agentMsg := fallbackMessages(resp.UserMessage, resp.AgentMessage, fmt.Sprintf("Hook '%s' blocked execution", jobName))
		return adapter.block(userMsg, agentMsg)
	}

	switch strings.ToLower(resp.Permission) {
	case "deny":
		userMsg, agentMsg := fallbackMessages(resp.UserMessage, resp.AgentMessage, fmt.Sprintf("Hook '%s' denied permission", jobName))
		return adapter.block(userMsg, agentMsg)
	case "ask":
		if logger != nil && adapter.askEvent != "" {
			logger.LogHookEvent(adapter.askEvent, jobName, map[string]interface{}{
				"userMessage":  resp.UserMessage,
				"agentMessage": resp.AgentMessage,
				"note":         adapter.askNote,
			}, nil)
		}
		userMsg, agentMsg := fallbackMessages(resp.UserMessage, resp.AgentMessage, fmt.Sprintf("Hook '%s' requests confirmation", jobName))
		return adapter.ask(userMsg, agentMsg)
	case "allow", "":
		if resp.UserMessage != "" || resp.AgentMessage != "" {
			return adapter.approveWithMessages(resp.UserMessage, resp.AgentMessage)
		}
		return adapter.approve()
	default:
		userMsg := fmt.Sprintf("Hook '%s' returned unknown permission: %s", jobName, resp.Permission)
		agentMsg := fmt.Sprintf("Unknown permission '%s' in response", resp.Permission)
		return adapter.block(userMsg, agentMsg)
	}
}

func fallbackMessages(userMessage, agentMessage, defaultUserMessage string) (string, string) {
	userMsg := userMessage
	if userMsg == "" {
		userMsg = defaultUserMessage
	}
	agentMsg := agentMessage
	if agentMsg == "" {
		agentMsg = userMsg
	}
	return userMsg, agentMsg
}
