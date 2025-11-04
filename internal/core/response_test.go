package core

import (
	"testing"
)

// TestBlockWithMessagesSingleParam tests BlockWithMessages with a single parameter.
// The same message should be sent to both user and agent.
func TestBlockWithMessagesSingleParam(t *testing.T) {
	msg := "Operation blocked"
	resp := BlockWithMessages(msg)

	// Type assertion to access dual message fields
	dualResp, ok := resp.(*DualMessagePreToolResponse)
	if !ok {
		t.Fatal("BlockWithMessages should return *DualMessagePreToolResponse")
	}

	if dualResp.GetUserMessage() != msg {
		t.Errorf("Expected userMessage %q, got %q", msg, dualResp.GetUserMessage())
	}

	if dualResp.GetAgentMessage() != msg {
		t.Errorf("Expected agentMessage %q, got %q", msg, dualResp.GetAgentMessage())
	}

	// Verify embedded response is set
	if dualResp.PreToolUseResponse == nil {
		t.Error("Embedded PreToolUseResponse should not be nil")
	}
}

// TestBlockWithMessagesDualParams tests BlockWithMessages with separate user and agent messages.
func TestBlockWithMessagesDualParams(t *testing.T) {
	userMsg := "This command was blocked for security reasons."
	agentMsg := "Blocked dangerous pattern: sudo. Type: privilege_escalation"

	resp := BlockWithMessages(userMsg, agentMsg)

	dualResp, ok := resp.(*DualMessagePreToolResponse)
	if !ok {
		t.Fatal("BlockWithMessages should return *DualMessagePreToolResponse")
	}

	if dualResp.GetUserMessage() != userMsg {
		t.Errorf("Expected userMessage %q, got %q", userMsg, dualResp.GetUserMessage())
	}

	if dualResp.GetAgentMessage() != agentMsg {
		t.Errorf("Expected agentMessage %q, got %q", agentMsg, dualResp.GetAgentMessage())
	}
}

// TestApproveWithMessagesSingleParam tests ApproveWithMessages with a single parameter.
func TestApproveWithMessagesSingleParam(t *testing.T) {
	msg := "Security check passed"
	resp := ApproveWithMessages(msg)

	dualResp, ok := resp.(*DualMessagePreToolResponse)
	if !ok {
		t.Fatal("ApproveWithMessages should return *DualMessagePreToolResponse")
	}

	if dualResp.GetUserMessage() != msg {
		t.Errorf("Expected userMessage %q, got %q", msg, dualResp.GetUserMessage())
	}

	if dualResp.GetAgentMessage() != msg {
		t.Errorf("Expected agentMessage %q, got %q", msg, dualResp.GetAgentMessage())
	}

	if dualResp.PreToolUseResponse == nil {
		t.Error("Embedded PreToolUseResponse should not be nil")
	}
}

// TestApproveWithMessagesDualParams tests ApproveWithMessages with separate messages.
func TestApproveWithMessagesDualParams(t *testing.T) {
	userMsg := "Hook executed successfully"
	agentMsg := "Job completed in 245ms, 0 issues found"

	resp := ApproveWithMessages(userMsg, agentMsg)

	dualResp, ok := resp.(*DualMessagePreToolResponse)
	if !ok {
		t.Fatal("ApproveWithMessages should return *DualMessagePreToolResponse")
	}

	if dualResp.GetUserMessage() != userMsg {
		t.Errorf("Expected userMessage %q, got %q", userMsg, dualResp.GetUserMessage())
	}

	if dualResp.GetAgentMessage() != agentMsg {
		t.Errorf("Expected agentMessage %q, got %q", agentMsg, dualResp.GetAgentMessage())
	}
}

// TestPostBlockWithMessagesSingleParam tests PostBlockWithMessages with a single parameter.
func TestPostBlockWithMessagesSingleParam(t *testing.T) {
	msg := "Formatting failed"
	resp := PostBlockWithMessages(msg)

	dualResp, ok := resp.(*DualMessagePostToolResponse)
	if !ok {
		t.Fatal("PostBlockWithMessages should return *DualMessagePostToolResponse")
	}

	if dualResp.GetUserMessage() != msg {
		t.Errorf("Expected userMessage %q, got %q", msg, dualResp.GetUserMessage())
	}

	if dualResp.GetAgentMessage() != msg {
		t.Errorf("Expected agentMessage %q, got %q", msg, dualResp.GetAgentMessage())
	}

	if dualResp.PostToolUseResponse == nil {
		t.Error("Embedded PostToolUseResponse should not be nil")
	}
}

// TestPostBlockWithMessagesDualParams tests PostBlockWithMessages with separate messages.
func TestPostBlockWithMessagesDualParams(t *testing.T) {
	userMsg := "Code formatting failed for example.py"
	agentMsg := "Black formatter failed: example.py\nStderr: invalid syntax on line 42"

	resp := PostBlockWithMessages(userMsg, agentMsg)

	dualResp, ok := resp.(*DualMessagePostToolResponse)
	if !ok {
		t.Fatal("PostBlockWithMessages should return *DualMessagePostToolResponse")
	}

	if dualResp.GetUserMessage() != userMsg {
		t.Errorf("Expected userMessage %q, got %q", userMsg, dualResp.GetUserMessage())
	}

	if dualResp.GetAgentMessage() != agentMsg {
		t.Errorf("Expected agentMessage %q, got %q", agentMsg, dualResp.GetAgentMessage())
	}
}

// TestAllowWithMessagesSingleParam tests AllowWithMessages with a single parameter.
func TestAllowWithMessagesSingleParam(t *testing.T) {
	msg := "Operation completed successfully"
	resp := AllowWithMessages(msg)

	dualResp, ok := resp.(*DualMessagePostToolResponse)
	if !ok {
		t.Fatal("AllowWithMessages should return *DualMessagePostToolResponse")
	}

	if dualResp.GetUserMessage() != msg {
		t.Errorf("Expected userMessage %q, got %q", msg, dualResp.GetUserMessage())
	}

	if dualResp.GetAgentMessage() != msg {
		t.Errorf("Expected agentMessage %q, got %q", msg, dualResp.GetAgentMessage())
	}

	if dualResp.PostToolUseResponse == nil {
		t.Error("Embedded PostToolUseResponse should not be nil")
	}
}

// TestAllowWithMessagesDualParams tests AllowWithMessages with separate messages.
func TestAllowWithMessagesDualParams(t *testing.T) {
	userMsg := "Format completed"
	agentMsg := "Formatted 3 files: example.go, test.go, main.go"

	resp := AllowWithMessages(userMsg, agentMsg)

	dualResp, ok := resp.(*DualMessagePostToolResponse)
	if !ok {
		t.Fatal("AllowWithMessages should return *DualMessagePostToolResponse")
	}

	if dualResp.GetUserMessage() != userMsg {
		t.Errorf("Expected userMessage %q, got %q", userMsg, dualResp.GetUserMessage())
	}

	if dualResp.GetAgentMessage() != agentMsg {
		t.Errorf("Expected agentMessage %q, got %q", agentMsg, dualResp.GetAgentMessage())
	}
}

// TestAskWithMessagesSingleParam tests AskWithMessages with a single parameter.
func TestAskWithMessagesSingleParam(t *testing.T) {
	msg := "This operation requires confirmation"
	resp := AskWithMessages(msg)

	dualResp, ok := resp.(*DualMessagePreToolResponse)
	if !ok {
		t.Fatal("AskWithMessages should return *DualMessagePreToolResponse")
	}

	if dualResp.GetUserMessage() != msg {
		t.Errorf("Expected userMessage %q, got %q", msg, dualResp.GetUserMessage())
	}

	if dualResp.GetAgentMessage() != msg {
		t.Errorf("Expected agentMessage %q, got %q", msg, dualResp.GetAgentMessage())
	}

	if dualResp.PreToolUseResponse == nil {
		t.Error("Embedded PreToolUseResponse should not be nil")
	}
}

// TestAskWithMessagesDualParams tests AskWithMessages with separate messages.
func TestAskWithMessagesDualParams(t *testing.T) {
	userMsg := "This operation requires your confirmation"
	agentMsg := "Attempting to execute: sudo rm -rf /tmp/cache"

	resp := AskWithMessages(userMsg, agentMsg)

	dualResp, ok := resp.(*DualMessagePreToolResponse)
	if !ok {
		t.Fatal("AskWithMessages should return *DualMessagePreToolResponse")
	}

	if dualResp.GetUserMessage() != userMsg {
		t.Errorf("Expected userMessage %q, got %q", userMsg, dualResp.GetUserMessage())
	}

	if dualResp.GetAgentMessage() != agentMsg {
		t.Errorf("Expected agentMessage %q, got %q", agentMsg, dualResp.GetAgentMessage())
	}
}

// TestPreToolUseInterfaceCompliance verifies that DualMessagePreToolResponse
// implements cchooks.PreToolUseResponseInterface.
func TestPreToolUseInterfaceCompliance(_ *testing.T) {
	// This should compile - if it doesn't, we're not implementing the interface correctly
	_ = BlockWithMessages("test")
	_ = ApproveWithMessages("test")
	_ = AskWithMessages("test")
}

// TestPostToolUseInterfaceCompliance verifies that DualMessagePostToolResponse
// implements cchooks.PostToolUseResponseInterface.
func TestPostToolUseInterfaceCompliance(_ *testing.T) {
	// This should compile - if it doesn't, we're not implementing the interface correctly
	_ = PostBlockWithMessages("test")
	_ = AllowWithMessages("test")
}

// TestEmptyMessages tests that empty strings are handled correctly.
func TestEmptyMessages(t *testing.T) {
	tests := []struct {
		name      string
		userMsg   string
		agentMsg  []string
		wantUser  string
		wantAgent string
	}{
		{
			name:      "Empty single message",
			userMsg:   "",
			agentMsg:  nil,
			wantUser:  "",
			wantAgent: "",
		},
		{
			name:      "Empty user, non-empty agent",
			userMsg:   "",
			agentMsg:  []string{"agent message"},
			wantUser:  "",
			wantAgent: "agent message",
		},
		{
			name:      "Non-empty user, empty agent",
			userMsg:   "user message",
			agentMsg:  []string{""},
			wantUser:  "user message",
			wantAgent: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := BlockWithMessages(tt.userMsg, tt.agentMsg...)
			dualResp := resp.(*DualMessagePreToolResponse)

			if dualResp.GetUserMessage() != tt.wantUser {
				t.Errorf("Expected userMessage %q, got %q", tt.wantUser, dualResp.GetUserMessage())
			}

			if dualResp.GetAgentMessage() != tt.wantAgent {
				t.Errorf("Expected agentMessage %q, got %q", tt.wantAgent, dualResp.GetAgentMessage())
			}
		})
	}
}

// TestUnicodeMessages tests that Unicode characters are handled correctly.
func TestUnicodeMessages(t *testing.T) {
	userMsg := "操作がブロックされました"                       // "Operation was blocked" in Japanese
	agentMsg := "Commande bloquée: motif détecté 🔒" // "Command blocked: pattern detected" in French with emoji

	resp := BlockWithMessages(userMsg, agentMsg)
	dualResp := resp.(*DualMessagePreToolResponse)

	if dualResp.GetUserMessage() != userMsg {
		t.Errorf("Unicode userMessage not preserved: expected %q, got %q", userMsg, dualResp.GetUserMessage())
	}

	if dualResp.GetAgentMessage() != agentMsg {
		t.Errorf("Unicode agentMessage not preserved: expected %q, got %q", agentMsg, dualResp.GetAgentMessage())
	}
}

// TestVeryLongMessages tests handling of very long messages.
func TestVeryLongMessages(t *testing.T) {
	// Create a message longer than typical command output
	longMsg := ""
	for i := 0; i < 1000; i++ {
		longMsg += "This is a very long message that tests buffer handling. "
	}

	resp := PostBlockWithMessages(longMsg)
	dualResp := resp.(*DualMessagePostToolResponse)

	if dualResp.GetUserMessage() != longMsg {
		t.Error("Long message not preserved correctly")
	}

	if dualResp.GetAgentMessage() != longMsg {
		t.Error("Long message not preserved correctly for agent")
	}
}

// TestEmbeddedResponseNotNil verifies that the embedded cchooks response is properly set.
func TestEmbeddedResponseNotNil(t *testing.T) {
	tests := []struct {
		name string
		resp interface{}
	}{
		{"BlockWithMessages", BlockWithMessages("test")},
		{"ApproveWithMessages", ApproveWithMessages("test")},
		{"PostBlockWithMessages", PostBlockWithMessages("test")},
		{"AllowWithMessages", AllowWithMessages("test")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the embedded response is not nil
			switch r := tt.resp.(type) {
			case *DualMessagePreToolResponse:
				if r.PreToolUseResponse == nil {
					t.Error("Embedded PreToolUseResponse should not be nil")
				}
			case *DualMessagePostToolResponse:
				if r.PostToolUseResponse == nil {
					t.Error("Embedded PostToolUseResponse should not be nil")
				}
			default:
				t.Errorf("Unexpected type: %T", tt.resp)
			}
		})
	}
}

// TestMultipleAgentMessages verifies that only the first agent message is used
// when multiple are provided.
func TestMultipleAgentMessages(t *testing.T) {
	userMsg := "Error"
	agentMsg1 := "First message"
	agentMsg2 := "Second message (should be ignored)"
	agentMsg3 := "Third message (should be ignored)"

	resp := BlockWithMessages(userMsg, agentMsg1, agentMsg2, agentMsg3)
	dualResp := resp.(*DualMessagePreToolResponse)

	if dualResp.GetUserMessage() != userMsg {
		t.Errorf("Expected userMessage %q, got %q", userMsg, dualResp.GetUserMessage())
	}

	// Only the first agent message should be used
	if dualResp.GetAgentMessage() != agentMsg1 {
		t.Errorf("Expected agentMessage %q (first one), got %q", agentMsg1, dualResp.GetAgentMessage())
	}
}

// BenchmarkBlockWithMessages benchmarks the overhead of creating a BlockWithMessages response.
func BenchmarkBlockWithMessages(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = BlockWithMessages("test message", "agent message")
	}
}

// BenchmarkBlockWithMessagesSingle benchmarks creating a single-message response.
func BenchmarkBlockWithMessagesSingle(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = BlockWithMessages("test message")
	}
}

// BenchmarkPostBlockWithMessages benchmarks the PostToolUse wrapper.
func BenchmarkPostBlockWithMessages(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = PostBlockWithMessages("test message", "agent message")
	}
}
