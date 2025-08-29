package hooks

import (
	"testing"
)

func TestIsFindCommand(t *testing.T) {
	hook := &FindBlockerHook{}

	testCases := []struct {
		name        string
		command     string
		shouldBlock bool
		description string
	}{
		{
			name:        "simple find command",
			command:     "find . -name '*.go'",
			shouldBlock: true,
			description: "should block basic find command",
		},
		{
			name:        "find with type flag",
			command:     "find . -type f -name '*.txt'",
			shouldBlock: true,
			description: "should block find with type flag",
		},
		{
			name:        "find in pipeline",
			command:     "find . -name '*.go' | xargs grep 'func'",
			shouldBlock: true,
			description: "should block find used in pipeline",
		},
		{
			name:        "find in command substitution",
			command:     "echo $(find . -name '*.md')",
			shouldBlock: true,
			description: "should block find in command substitution",
		},
		{
			name:        "grep with find in pipeline",
			command:     "grep -r 'pattern' $(find . -name '*.go')",
			shouldBlock: true,
			description: "should block commands that use find in substitution",
		},
		{
			name:        "ls command",
			command:     "ls -la",
			shouldBlock: false,
			description: "should not block non-find commands",
		},
		{
			name:        "fd command",
			command:     "fd '*.go'",
			shouldBlock: false,
			description: "should not block fd commands",
		},
		{
			name:        "grep command",
			command:     "grep -r 'pattern' .",
			shouldBlock: false,
			description: "should not block grep commands",
		},
		{
			name:        "command with 'find' in string",
			command:     "echo 'I need to find something'",
			shouldBlock: false,
			description: "should not block commands that just contain 'find' in quoted strings",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			blocked, suggestion := hook.isFindCommand(tc.command)

			if blocked != tc.shouldBlock {
				t.Errorf("Expected blocked=%v, got blocked=%v for command: %s", tc.shouldBlock, blocked, tc.command)
			}

			if tc.shouldBlock && suggestion == "" {
				t.Errorf("Expected suggestion for blocked command: %s", tc.command)
			}

			if !tc.shouldBlock && suggestion != "" {
				t.Errorf("Unexpected suggestion for allowed command: %s", tc.command)
			}

			t.Logf("Command: %s", tc.command)
			t.Logf("Blocked: %v", blocked)
			if blocked {
				t.Logf("Suggestion: %s", suggestion)
			}
		})
	}
}

func TestGenerateFdSuggestion(t *testing.T) {
	hook := &FindBlockerHook{}

	testCases := []struct {
		name                 string
		findCommand          string
		expectedInSuggestion []string
	}{
		{
			name:                 "basic name search",
			findCommand:          "find . -name '*.go'",
			expectedInSuggestion: []string{"fd", "*.go", "better performance"},
		},
		{
			name:                 "type file search",
			findCommand:          "find . -type f -name '*.txt'",
			expectedInSuggestion: []string{"fd", "--type f", "*.txt"},
		},
		{
			name:                 "case insensitive search",
			findCommand:          "find . -iname '*.PDF'",
			expectedInSuggestion: []string{"fd", "--ignore-case", "case-insensitive"},
		},
		{
			name:                 "max depth search",
			findCommand:          "find . -maxdepth 2 -name 'README*'",
			expectedInSuggestion: []string{"fd", "--max-depth", "limit search depth"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suggestion := hook.generateFdSuggestion(tc.findCommand)

			for _, expected := range tc.expectedInSuggestion {
				if !contains(suggestion, expected) {
					t.Errorf("Expected suggestion to contain '%s', but it didn't.\nSuggestion: %s", expected, suggestion)
				}
			}

			t.Logf("Find command: %s", tc.findCommand)
			t.Logf("Suggestion: %s", suggestion)
		})
	}
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) &&
		(haystack == needle ||
			len(needle) == 0 ||
			indexIgnoreCase(haystack, needle) >= 0)
}

func indexIgnoreCase(s, substr string) int {
	s, substr = toLower(s), toLower(substr)
	return indexString(s, substr)
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if 'A' <= c && c <= 'Z' {
			c = c + ('a' - 'A')
		}
		result[i] = c
	}
	return string(result)
}

func indexString(s, substr string) int {
	n := len(substr)
	if n == 0 {
		return 0
	}
	if n > len(s) {
		return -1
	}
	for i := 0; i <= len(s)-n; i++ {
		if s[i:i+n] == substr {
			return i
		}
	}
	return -1
}
