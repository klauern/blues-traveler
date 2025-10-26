package core

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// EvalExpression evaluates a minimal boolean expression used for skip/only conditions.
// Supported:
// - variable substitution: ${VAR}
// - operators: ==, !=, matches
// - boolean: &&, ||, ! (unary)
// - glob matching for right-hand side of matches
// This is intentionally simple; not a full parser. Expressions should be small.
func EvalExpression(expr string, vars map[string]string) (bool, error) {
	s := strings.TrimSpace(expr)
	if s == "" {
		return true, nil
	}

	// Expand variables: ${VAR}
	expanded := expandVars(s, vars)
	// Tokenize by || first (lowest precedence)
	orParts := splitRespectingQuotes(expanded, "||")
	anyMatch := false
	for _, orp := range orParts {
		andParts := splitRespectingQuotes(orp, "&&")
		all := true
		for _, ap := range andParts {
			v, err := evalSimple(strings.TrimSpace(ap))
			if err != nil {
				return false, err
			}
			all = all && v
		}
		anyMatch = anyMatch || all
	}
	return anyMatch, nil
}

func expandVars(s string, vars map[string]string) string {
	return varPattern.ReplaceAllStringFunc(s, func(m string) string {
		key := strings.TrimSuffix(strings.TrimPrefix(m, "${"), "}")
		if v, ok := vars[key]; ok {
			return v
		}
		return ""
	})
}

var varPattern = regexp.MustCompile(`\$\{[A-Za-z_][A-Za-z0-9_]*\}`)

func evalSimple(s string) (bool, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return true, nil
	}

	// Handle unary negation
	if result, handled, err := handleNegation(s); handled {
		return result, err
	}

	// Try operator-based evaluation
	if result, handled, err := evalOperator(s); handled {
		return result, err
	}

	// Bareword/literal evaluation
	return evalLiteral(s)
}

// handleNegation handles unary ! operator with recursion
func handleNegation(s string) (bool, bool, error) {
	negated := false
	for strings.HasPrefix(s, "!") {
		negated = !negated
		s = strings.TrimPrefix(s, "!")
	}

	if !negated {
		return false, false, nil // Not a negation case
	}

	inner, err := evalSimple(s)
	if err != nil {
		return false, true, err
	}
	return !inner, true, nil
}

// evalOperator evaluates expressions with operators ==, !=, matches, regex
func evalOperator(s string) (bool, bool, error) {
	operators := []string{"==", "!=", "matches", "regex"}
	for _, op := range operators {
		idx := indexOutsideQuotes(s, op)
		if idx < 0 {
			continue
		}

		left := strings.TrimSpace(s[:idx])
		right := strings.Trim(strings.TrimSpace(s[idx+len(op):]), "\"'")

		switch op {
		case "==":
			return left == right, true, nil
		case "!=":
			return left != right, true, nil
		case "matches":
			return globMatchAny(left, right), true, nil
		case "regex":
			result, err := regexMatchAny(left, right)
			return result, true, err
		}
	}
	return false, false, nil // No operator found
}

// evalLiteral evaluates a literal value as truthy/falsy
func evalLiteral(s string) (bool, error) {
	l := strings.ToLower(strings.Trim(s, "\"'"))

	if l == "false" || l == "0" {
		return false, nil
	}
	if l == "true" || l == "1" {
		return true, nil
	}
	if l != "" {
		return true, nil
	}

	return false, fmt.Errorf("could not evaluate expression: %q", s)
}

func globMatchAny(left string, pattern string) bool {
	// If left contains spaces, treat as multiple tokens
	tokens := strings.Fields(left)
	if len(tokens) == 0 {
		tokens = []string{left}
	}
	for _, t := range tokens {
		if ok, _ := filepath.Match(pattern, t); ok {
			return true
		}
	}
	return false
}

func regexMatchAny(left string, pattern string) (bool, error) {
	rx, err := regexp.Compile(pattern)
	if err != nil {
		return false, fmt.Errorf("invalid regex pattern: %v", err)
	}
	tokens := strings.Fields(left)
	if len(tokens) == 0 {
		tokens = []string{left}
	}
	for _, t := range tokens {
		if rx.MatchString(t) {
			return true, nil
		}
	}
	return false, nil
}

// splitRespectingQuotes splits on a delimiter while keeping quoted substrings intact
func splitRespectingQuotes(s, delim string) []string {
	var parts []string
	var cur strings.Builder
	inSingle, inDouble := false, false
	i := 0
	for i < len(s) {
		if s[i] == '\'' && !inDouble {
			inSingle = !inSingle
			cur.WriteByte(s[i])
			i++
			continue
		}
		if s[i] == '"' && !inSingle {
			inDouble = !inDouble
			cur.WriteByte(s[i])
			i++
			continue
		}
		if !inSingle && !inDouble && strings.HasPrefix(s[i:], delim) {
			parts = append(parts, cur.String())
			cur.Reset()
			i += len(delim)
			continue
		}
		cur.WriteByte(s[i])
		i++
	}
	parts = append(parts, cur.String())
	return parts
}

// indexOutsideQuotes finds index of substr outside quotes
func indexOutsideQuotes(s, sub string) int {
	inSingle, inDouble := false, false
	for i := 0; i+len(sub) <= len(s); i++ {
		c := s[i]
		if c == '\'' && !inDouble {
			inSingle = !inSingle
		} else if c == '"' && !inSingle {
			inDouble = !inDouble
		}
		if !inSingle && !inDouble && strings.HasPrefix(s[i:], sub) {
			return i
		}
	}
	return -1
}
