package hooks

import (
	"testing"

	"github.com/klauern/blues-traveler/internal/config"
	"github.com/klauern/blues-traveler/internal/core"
)

func TestParseCursorResponse(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		output  string
		wantErr bool
		wantNil bool
		expect  cursorResponseExpectation
	}{
		{
			name:   "allow minimal payload",
			output: `{"permission":"allow"}`,
			expect: cursorResponseExpectation{permission: "allow"},
		},
		{
			name:   "deny with explicit messages and continue false",
			output: `{"permission":"deny","userMessage":"Action blocked","agentMessage":"Reason: policy","continue":false}`,
			expect: cursorResponseExpectation{
				permission:  "deny",
				continueVal: boolPtr(false),
				messages:    messages("Action blocked", "Reason: policy"),
			},
		},
		{
			name:   "ask requires approval with continue true",
			output: `{"permission":"ask","userMessage":"Need confirmation","agentMessage":"Provide justification","continue":true}`,
			expect: cursorResponseExpectation{
				permission:  "ask",
				continueVal: boolPtr(true),
				messages:    messages("Need confirmation", "Provide justification"),
			},
		},
		{
			name:   "messages without permission fall back to zero values",
			output: `{"userMessage":"Heads up","agentMessage":"log-only"}`,
			expect: cursorResponseExpectation{
				messages: messages("Heads up", "log-only"),
			},
		},
		{
			name:    "plain text is treated as non JSON",
			output:  "plain text output",
			wantNil: true,
		},
		{
			name:    "whitespace is ignored",
			output:  "   \n\t  ",
			wantNil: true,
		},
		{
			name:    "invalid JSON surfaces an error",
			output:  `{"permission":"deny", invalid`,
			wantErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			resp, err := parseCursorResponse(tc.output)
			switch {
			case tc.wantErr:
				if err == nil {
					t.Fatalf("parseCursorResponse() expected error but got none")
				}
				return
			case err != nil:
				t.Fatalf("parseCursorResponse() unexpected error: %v", err)
			}

			if tc.wantNil {
				if resp != nil {
					t.Fatalf("parseCursorResponse() expected nil response, got %+v", resp)
				}
				return
			}

			if resp == nil {
				t.Fatal("parseCursorResponse() returned nil response")
			}

			assertCursorResponse(t, resp, tc.expect)
		})
	}
}

// cursorResponseExpectation captures the expected Cursor response fields for assertion.
type cursorResponseExpectation struct {
	permission  string
	continueVal *bool
	messages    messageExpectation
}

// assertCursorResponse validates the parsed response matches the expected fields.
func assertCursorResponse(t *testing.T, resp *CursorHookResponse, expect cursorResponseExpectation) {
	t.Helper()

	if resp.Permission != expect.permission {
		t.Fatalf("Permission = %q, want %q", resp.Permission, expect.permission)
	}

	expect.messages.assert(t, resp)

	switch {
	case resp.Continue == nil && expect.continueVal == nil:
		// Both unset, nothing to assert.
	case resp.Continue == nil || expect.continueVal == nil:
		t.Fatalf("Continue mismatch: got %v, want %v", resp.Continue, expect.continueVal)
	case *resp.Continue != *expect.continueVal:
		t.Fatalf("Continue = %v, want %v", *resp.Continue, *expect.continueVal)
	}
}

func newTestConfigHookWithPlatform(t *testing.T, platform core.Platform) *ConfigHook {
	t.Helper()

	ctx := &core.HookContext{
		FileSystem:      core.NewMockFileSystem(),
		CommandExecutor: core.NewMockCommandExecutor(),
		RunnerFactory:   core.MockRunnerFactory,
		SettingsChecker: func(string) bool { return true },
		Platform:        platform,
	}

	job := config.HookJob{Name: "ask-job"}
	hook := NewConfigHook("group", job.Name, job, string(core.PreToolUseEvent), ctx)
	cfgHook, ok := hook.(*ConfigHook)
	if !ok {
		t.Fatal("NewConfigHook should return *ConfigHook")
	}
	return cfgHook
}

// messageExpectation holds the expected user and agent messages for a response.
type messageExpectation struct {
	user  string
	agent string
}

func messages(user, agent string) messageExpectation {
	return messageExpectation{user: user, agent: agent}
}

func (m messageExpectation) assert(t *testing.T, resp *CursorHookResponse) {
	t.Helper()

	if resp.UserMessage != m.user {
		t.Fatalf("UserMessage = %q, want %q", resp.UserMessage, m.user)
	}

	if resp.AgentMessage != m.agent {
		t.Fatalf("AgentMessage = %q, want %q", resp.AgentMessage, m.agent)
	}
}

// boolPtr returns a pointer to the provided bool literal for concise test expectations.
func boolPtr(v bool) *bool {
	return &v
}
