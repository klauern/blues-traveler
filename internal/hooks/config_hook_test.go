package hooks

import (
	"testing"
)

// assertCursorResponse is a helper function that verifies a CursorHookResponse matches expectations.
func assertCursorResponse(t *testing.T, resp *CursorHookResponse, want *CursorHookResponse) {
	t.Helper()

	if resp == nil {
		t.Error("parseCursorResponse() returned nil")
		return
	}

	// Compare fields
	if resp.Permission != want.Permission {
		t.Errorf("Permission = %q, want %q", resp.Permission, want.Permission)
	}

	if resp.UserMessage != want.UserMessage {
		t.Errorf("UserMessage = %q, want %q", resp.UserMessage, want.UserMessage)
	}

	if resp.AgentMessage != want.AgentMessage {
		t.Errorf("AgentMessage = %q, want %q", resp.AgentMessage, want.AgentMessage)
	}

	// Compare Continue field (pointer comparison)
	if (resp.Continue == nil) != (want.Continue == nil) {
		t.Errorf("Continue nil mismatch: got %v, want %v", resp.Continue, want.Continue)
	} else if resp.Continue != nil && want.Continue != nil {
		if *resp.Continue != *want.Continue {
			t.Errorf("Continue = %v, want %v", *resp.Continue, *want.Continue)
		}
	}
}

// TestParseCursorResponseValid tests parsing of valid Cursor JSON responses.
func TestParseCursorResponseValid(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		wantResp *CursorHookResponse
	}{
		{
			name:   "Complete JSON with all fields",
			output: `{"permission": "deny", "userMessage": "User msg", "agentMessage": "Agent msg", "continue": false}`,
			wantResp: &CursorHookResponse{
				Permission:   "deny",
				UserMessage:  "User msg",
				AgentMessage: "Agent msg",
				Continue:     boolPtr(false),
			},
		},
		{
			name:   "Permission only",
			output: `{"permission": "allow"}`,
			wantResp: &CursorHookResponse{
				Permission: "allow",
			},
		},
		{
			name:   "Ask permission",
			output: `{"permission": "ask", "userMessage": "Confirm?", "agentMessage": "Details here"}`,
			wantResp: &CursorHookResponse{
				Permission:   "ask",
				UserMessage:  "Confirm?",
				AgentMessage: "Details here",
			},
		},
		{
			name:   "Continue false",
			output: `{"continue": false, "userMessage": "Blocked"}`,
			wantResp: &CursorHookResponse{
				Continue:    boolPtr(false),
				UserMessage: "Blocked",
			},
		},
		{
			name:   "Continue true",
			output: `{"continue": true}`,
			wantResp: &CursorHookResponse{
				Continue: boolPtr(true),
			},
		},
		{
			name:   "Extra fields (forward compatibility)",
			output: `{"permission": "allow", "futureField": "value", "anotherField": 123}`,
			wantResp: &CursorHookResponse{
				Permission: "allow",
			},
		},
		{
			name:   "Empty permission value",
			output: `{"permission": ""}`,
			wantResp: &CursorHookResponse{
				Permission: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := parseCursorResponse(tt.output)
			if err != nil {
				t.Errorf("parseCursorResponse() unexpected error: %v", err)
				return
			}
			assertCursorResponse(t, resp, tt.wantResp)
		})
	}
}

// TestParseCursorResponseInvalid tests parsing of invalid JSON.
func TestParseCursorResponseInvalid(t *testing.T) {
	tests := []struct {
		name   string
		output string
	}{
		{
			name:   "Invalid JSON syntax",
			output: `{"permission": "deny", invalid}`,
		},
		{
			name:   "JSON array instead of object",
			output: `["permission", "deny"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseCursorResponse(tt.output)
			if err == nil {
				t.Error("parseCursorResponse() expected error but got none")
			}
		})
	}
}

// TestParseCursorResponseEdgeCases tests edge cases like empty strings, whitespace, and special characters.
func TestParseCursorResponseEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		wantNil  bool
		wantResp *CursorHookResponse
	}{
		{
			name:    "Empty string",
			output:  "",
			wantNil: true,
		},
		{
			name:    "Non-JSON plain text",
			output:  "This is plain text output",
			wantNil: true,
		},
		{
			name:    "Whitespace only",
			output:  "   \n\t  ",
			wantNil: true,
		},
		{
			name:   "JSON with whitespace",
			output: `  {"permission": "deny"}  `,
			wantResp: &CursorHookResponse{
				Permission: "deny",
			},
		},
		{
			name:   "Escaped characters",
			output: `{"userMessage": "Line 1\nLine 2\tTabbed", "agentMessage": "Path: \"C:\\Users\""}`,
			wantResp: &CursorHookResponse{
				UserMessage:  "Line 1\nLine 2\tTabbed",
				AgentMessage: "Path: \"C:\\Users\"",
			},
		},
		{
			name:   "Unicode characters",
			output: `{"userMessage": "操作がブロックされました", "agentMessage": "Détails techniques 🔒"}`,
			wantResp: &CursorHookResponse{
				UserMessage:  "操作がブロックされました",
				AgentMessage: "Détails techniques 🔒",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := parseCursorResponse(tt.output)
			if err != nil {
				t.Errorf("parseCursorResponse() unexpected error: %v", err)
				return
			}

			if tt.wantNil {
				if resp != nil {
					t.Errorf("parseCursorResponse() expected nil but got %+v", resp)
				}
				return
			}

			assertCursorResponse(t, resp, tt.wantResp)
		})
	}
}

// TestCursorResponsePermissionValues tests different permission values.
func TestCursorResponsePermissionValues(t *testing.T) {
	permissions := []string{"allow", "deny", "ask", "ALLOW", "DENY", "ASK", ""}

	for _, perm := range permissions {
		t.Run("permission_"+perm, func(t *testing.T) {
			output := `{"permission": "` + perm + `"}`
			resp, err := parseCursorResponse(output)
			if err != nil {
				t.Errorf("parseCursorResponse() error = %v for permission %q", err, perm)
				return
			}

			if resp == nil {
				t.Errorf("parseCursorResponse() returned nil for permission %q", perm)
				return
			}

			if resp.Permission != perm {
				t.Errorf("Permission = %q, want %q", resp.Permission, perm)
			}
		})
	}
}

// TestCursorResponseContinueField tests the continue field handling.
func TestCursorResponseContinueField(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   *bool
	}{
		{
			name:   "continue: true",
			output: `{"continue": true}`,
			want:   boolPtr(true),
		},
		{
			name:   "continue: false",
			output: `{"continue": false}`,
			want:   boolPtr(false),
		},
		{
			name:   "continue field missing",
			output: `{"permission": "allow"}`,
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := parseCursorResponse(tt.output)
			if err != nil {
				t.Errorf("parseCursorResponse() error = %v", err)
				return
			}

			if resp == nil {
				t.Error("parseCursorResponse() returned nil")
				return
			}

			if (resp.Continue == nil) != (tt.want == nil) {
				t.Errorf("Continue nil mismatch: got %v, want %v", resp.Continue, tt.want)
			} else if resp.Continue != nil && tt.want != nil {
				if *resp.Continue != *tt.want {
					t.Errorf("Continue = %v, want %v", *resp.Continue, *tt.want)
				}
			}
		})
	}
}

// TestCursorResponseMessageFields tests message field handling.
func TestCursorResponseMessageFields(t *testing.T) {
	tests := []struct {
		name      string
		output    string
		wantUser  string
		wantAgent string
	}{
		{
			name:      "Both messages present",
			output:    `{"userMessage": "User", "agentMessage": "Agent"}`,
			wantUser:  "User",
			wantAgent: "Agent",
		},
		{
			name:      "Only user message",
			output:    `{"userMessage": "User"}`,
			wantUser:  "User",
			wantAgent: "",
		},
		{
			name:      "Only agent message",
			output:    `{"agentMessage": "Agent"}`,
			wantUser:  "",
			wantAgent: "Agent",
		},
		{
			name:      "No messages",
			output:    `{"permission": "allow"}`,
			wantUser:  "",
			wantAgent: "",
		},
		{
			name:      "Empty string messages",
			output:    `{"userMessage": "", "agentMessage": ""}`,
			wantUser:  "",
			wantAgent: "",
		},
		{
			name:      "Very long messages",
			output:    `{"userMessage": "` + longString(1000) + `", "agentMessage": "` + longString(2000) + `"}`,
			wantUser:  longString(1000),
			wantAgent: longString(2000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := parseCursorResponse(tt.output)
			if err != nil {
				t.Errorf("parseCursorResponse() error = %v", err)
				return
			}

			if resp == nil {
				t.Error("parseCursorResponse() returned nil")
				return
			}

			if resp.UserMessage != tt.wantUser {
				t.Errorf("UserMessage = %q, want %q", resp.UserMessage, tt.wantUser)
			}

			if resp.AgentMessage != tt.wantAgent {
				t.Errorf("AgentMessage = %q, want %q", resp.AgentMessage, tt.wantAgent)
			}
		})
	}
}

// Helper function to create a bool pointer
func boolPtr(b bool) *bool {
	return &b
}

// Helper function to generate long strings for testing
func longString(length int) string {
	s := ""
	for i := 0; i < length; i++ {
		s += "a"
	}
	return s
}

// BenchmarkParseCursorResponse benchmarks the JSON parsing performance.
func BenchmarkParseCursorResponse(b *testing.B) {
	output := `{"permission": "deny", "userMessage": "User message", "agentMessage": "Agent message", "continue": false}`

	for i := 0; i < b.N; i++ {
		_, _ = parseCursorResponse(output)
	}
}

// BenchmarkParseCursorResponseMinimal benchmarks parsing minimal JSON.
func BenchmarkParseCursorResponseMinimal(b *testing.B) {
	output := `{"permission": "allow"}`

	for i := 0; i < b.N; i++ {
		_, _ = parseCursorResponse(output)
	}
}

// BenchmarkParseCursorResponseNonJSON benchmarks the non-JSON path.
func BenchmarkParseCursorResponseNonJSON(b *testing.B) {
	output := "This is plain text output"

	for i := 0; i < b.N; i++ {
		_, _ = parseCursorResponse(output)
	}
}
