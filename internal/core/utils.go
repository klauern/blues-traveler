package core

import "strings"

// IsInQuotedString performs a simple check if the pattern is within quotes in the command string.
// This is a basic implementation - a full parser would be more accurate but this covers most shell cases.
func IsInQuotedString(command, pattern string) bool {
	index := strings.Index(command, pattern)
	if index == -1 {
		return false
	}

	// Count quotes before the pattern
	beforePattern := command[:index]
	singleQuotes := strings.Count(beforePattern, "'")
	doubleQuotes := strings.Count(beforePattern, "\"")

	// If we have an odd number of quotes before the pattern, we're likely inside quotes
	// This assumes balanced quotes and doesn't handle escaped quotes perfectly, but is sufficient for basic checks
	return singleQuotes%2 == 1 || doubleQuotes%2 == 1
}
