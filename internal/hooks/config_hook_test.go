package hooks

import (
	"testing"
)

// TestParseCursorResponse tests the parseCursorResponse function with various inputs.
func TestParseCursorResponse(t *testing.T) {
	tests := []struct {
		name        string
		output      string
		wantResp    *CursorHookResponse
		wantErr     bool
		description string
	}{
		{
			name:   "Valid JSON with all fields",
			output: `{"permission": "deny", "userMessage": "User msg", "agentMessage": "Agent msg", "continue": false}`,
			wantResp: &CursorHookResponse{
				Permission:   "deny",
				UserMessage:  "User msg",
				AgentMessage: "Agent msg",
				Continue:     boolPtr(false),
			},
			wantErr:     false,
			description: "Should parse complete JSON response",
		},
		{
			name:   "Valid JSON with permission only",
			output: `{"permission": "allow"}`,
			wantResp: &CursorHookResponse{
				Permission: "allow",
			},
			wantErr:     false,
			description: "Should parse minimal JSON with permission field",
		},
		{
			name:   "Valid JSON with ask permission",
			output: `{"permission": "ask", "userMessage": "Confirm?", "agentMessage": "Details here"}`,
			wantResp: &CursorHookResponse{
				Permission:   "ask",
				UserMessage:  "Confirm?",
				AgentMessage: "Details here",
			},
			wantErr:     false,
			description: "Should parse ask permission correctly",
		},
		{
			name:   "Valid JSON with continue false",
			output: `{"continue": false, "userMessage": "Blocked"}`,
			wantResp: &CursorHookResponse{
				Continue:    boolPtr(false),
				UserMessage: "Blocked",
			},
			wantErr:     false,
			description: "Should parse continue field correctly",
		},
		{
			name:   "Valid JSON with continue true",
			output: `{"continue": true}`,
			wantResp: &CursorHookResponse{
				Continue: boolPtr(true),
			},
			wantErr:     false,
			description: "Should parse continue true",
		},
		{
			name:        "Empty string",
			output:      "",
			wantResp:    nil,
			wantErr:     false,
			description: "Empty string should return nil (fallback to exit code)",
		},
		{
			name:        "Non-JSON output",
			output:      "This is plain text output",
			wantResp:    nil,
			wantErr:     false,
			description: "Non-JSON should return nil (fallback to exit code)",
		},
		{
			name:        "Whitespace only",
			output:      "   \n\t  ",
			wantResp:    nil,
			wantErr:     false,
			description: "Whitespace should return nil",
		},
		{
			name:        "Invalid JSON",
			output:      `{"permission": "deny", invalid}`,
			wantErr:     true,
			description: "Invalid JSON should return error",
		},
		{
			name:        "JSON array instead of object",
			output:      `["permission", "deny"]`,
			wantErr:     true,
			description: "JSON array should return error",
		},
		{
			name:   "Valid JSON with extra fields (forward compatibility)",
			output: `{"permission": "allow", "futureField": "value", "anotherField": 123}`,
			wantResp: &CursorHookResponse{
				Permission: "allow",
			},
			wantErr:     false,
			description: "Should ignore unknown fields for forward compatibility",
		},
		{
			name:   "Valid JSON with escaped characters",
			output: `{"userMessage": "Line 1\nLine 2\tTabbed", "agentMessage": "Path: \"C:\\Users\""}`,
			wantResp: &CursorHookResponse{
				UserMessage:  "Line 1\nLine 2\tTabbed",
				AgentMessage: "Path: \"C:\\Users\"",
			},
			wantErr:     false,
			description: "Should handle escaped characters correctly",
		},
		{
			name:   "Valid JSON with Unicode",
			output: `{"userMessage": "操作がブロックされました", "agentMessage": "Détails techniques 🔒"}`,
			wantResp: &CursorHookResponse{
				UserMessage:  "操作がブロックされました",
				AgentMessage: "Détails techniques 🔒",
			},
			wantErr:     false,
			description: "Should handle Unicode characters",
		},
		{
			name:   "Valid JSON with whitespace",
			output: `  {"permission": "deny"}  `,
			wantResp: &CursorHookResponse{
				Permission: "deny",
			},
			wantErr:     false,
			description: "Should trim whitespace before parsing",
		},
		{
			name:   "Empty permission value",
			output: `{"permission": ""}`,
			wantResp: &CursorHookResponse{
				Permission: "",
			},
			wantErr:     false,
			description: "Empty permission should be allowed (interpreted as allow)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := parseCursorResponse(tt.output)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseCursorResponse() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("parseCursorResponse() unexpected error: %v", err)
				return
			}

			if tt.wantResp == nil {
				if resp != nil {
					t.Errorf("parseCursorResponse() expected nil but got %+v", resp)
				}
				return
			}

			if resp == nil {
				t.Errorf("parseCursorResponse() expected response but got nil")
				return
			}

			// Compare fields
			if resp.Permission != tt.wantResp.Permission {
				t.Errorf("Permission = %q, want %q", resp.Permission, tt.wantResp.Permission)
			}

			if resp.UserMessage != tt.wantResp.UserMessage {
				t.Errorf("UserMessage = %q, want %q", resp.UserMessage, tt.wantResp.UserMessage)
			}

			if resp.AgentMessage != tt.wantResp.AgentMessage {
				t.Errorf("AgentMessage = %q, want %q", resp.AgentMessage, tt.wantResp.AgentMessage)
			}

			// Compare Continue field (pointer comparison)
			if (resp.Continue == nil) != (tt.wantResp.Continue == nil) {
				t.Errorf("Continue nil mismatch: got %v, want %v", resp.Continue, tt.wantResp.Continue)
			} else if resp.Continue != nil && tt.wantResp.Continue != nil {
				if *resp.Continue != *tt.wantResp.Continue {
					t.Errorf("Continue = %v, want %v", *resp.Continue, *tt.wantResp.Continue)
				}
			}
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
		name        string
		output      string
		wantUser    string
		wantAgent   string
		description string
	}{
		{
			name:        "Both messages present",
			output:      `{"userMessage": "User", "agentMessage": "Agent"}`,
			wantUser:    "User",
			wantAgent:   "Agent",
			description: "Both messages should be preserved",
		},
		{
			name:        "Only user message",
			output:      `{"userMessage": "User"}`,
			wantUser:    "User",
			wantAgent:   "",
			description: "Agent message should be empty string",
		},
		{
			name:        "Only agent message",
			output:      `{"agentMessage": "Agent"}`,
			wantUser:    "",
			wantAgent:   "Agent",
			description: "User message should be empty string",
		},
		{
			name:        "No messages",
			output:      `{"permission": "allow"}`,
			wantUser:    "",
			wantAgent:   "",
			description: "Both messages should be empty strings",
		},
		{
			name:        "Empty string messages",
			output:      `{"userMessage": "", "agentMessage": ""}`,
			wantUser:    "",
			wantAgent:   "",
			description: "Explicit empty strings should be preserved",
		},
		{
			name:        "Very long messages",
			output:      `{"userMessage": "` + longString(1000) + `", "agentMessage": "` + longString(2000) + `"}`,
			wantUser:    longString(1000),
			wantAgent:   longString(2000),
			description: "Long messages should be handled",
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
