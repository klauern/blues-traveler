package hooks

import (
	"strings"
	"testing"

	"github.com/klauern/blues-traveler/internal/core"
)

func TestSecurityHook(t *testing.T) {
	ctx := core.TestHookContext(nil)
	hook := NewSecurityHook(ctx)

	// Test basic properties
	if hook.Key() != "security" {
		t.Errorf("Expected key 'security', got '%s'", hook.Key())
	}

	if hook.Name() != "Security Hook" {
		t.Errorf("Expected name 'Security Hook', got '%s'", hook.Name())
	}

	// Test that hook is enabled by default
	if !hook.IsEnabled() {
		t.Error("Expected hook to be enabled by default")
	}

	// Test running the hook (should not error)
	err := hook.Run()
	if err != nil {
		t.Errorf("Hook run failed: %v", err)
	}
}

func TestSecurityHookDisabled(t *testing.T) {
	ctx := core.TestHookContext(func(string) bool { return false })
	hook := NewSecurityHook(ctx)

	// Test that hook is disabled
	if hook.IsEnabled() {
		t.Error("Expected hook to be disabled")
	}

	// Running disabled hook should still work but skip execution
	err := hook.Run()
	if err != nil {
		t.Errorf("Disabled hook run failed: %v", err)
	}
}

func TestSecurityHookStaticPatterns(t *testing.T) {
	ctx := core.TestHookContext(nil)
	hook := NewSecurityHook(ctx).(*SecurityHook)

	testCases := []struct {
		command string
		blocked bool
		reason  string
	}{
		{"dd if=/dev/zero of=/dev/sda", true, "dd if="},
		{"mkfs.ext4 /dev/sda1", true, "mkfs"},
		{"echo hello > /dev/null", true, "> /dev/"},
		{"sudo rm -rf /", true, "sudo rm"},
		{"chmod -R 777 /", true, "chmod -r 777 /"},
		{"shutdown -h now", true, "shutdown -h now"},
		{"nvram -c", true, "nvram -c"},
		{"ls -la", false, ""},
		{"echo hello", false, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.command, func(t *testing.T) {
			blocked, reason := hook.checkStaticPatterns(strings.ToLower(tc.command))

			if blocked != tc.blocked {
				t.Errorf("Command '%s': expected blocked=%v, got blocked=%v", tc.command, tc.blocked, blocked)
			}

			if tc.blocked && !containsSubstring(reason, tc.reason) {
				t.Errorf("Command '%s': expected reason to contain '%s', got '%s'", tc.command, tc.reason, reason)
			}
		})
	}
}

func TestSecurityHookDangerousRm(t *testing.T) {
	ctx := core.TestHookContext(nil)
	hook := NewSecurityHook(ctx).(*SecurityHook)

	testCases := []struct {
		tokens  []string
		blocked bool
		reason  string
	}{
		{[]string{"rm", "-rf", "/"}, true, "targets filesystem root"},
		{[]string{"rm", "-rf", "/system"}, true, "targets critical path"},
		{[]string{"rm", "-rf", "/usr"}, true, "targets critical path"},
		{[]string{"rm", "-rf", "/*"}, true, "wildcard at root"},
		{[]string{"rm", "-rf", "/home/user/doc"}, false, ""},
		{[]string{"rm", "file.txt"}, false, ""}, // no -r flag
		{[]string{"ls", "-la"}, false, ""},      // not rm
	}

	for _, tc := range testCases {
		t.Run(tc.tokens[0]+" "+joinArgs(tc.tokens[1:]), func(t *testing.T) {
			blocked, reason := hook.detectDangerousRm(tc.tokens)

			if blocked != tc.blocked {
				t.Errorf("Tokens %v: expected blocked=%v, got blocked=%v", tc.tokens, tc.blocked, blocked)
			}

			if tc.blocked && !containsSubstring(reason, tc.reason) {
				t.Errorf("Tokens %v: expected reason to contain '%s', got '%s'", tc.tokens, tc.reason, reason)
			}
		})
	}
}

// Helper functions for tests

func containsSubstring(s, substr string) bool {
	return len(substr) == 0 || len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func joinArgs(args []string) string {
	if len(args) == 0 {
		return ""
	}
	result := args[0]
	for _, arg := range args[1:] {
		result += " " + arg
	}
	return result
}
