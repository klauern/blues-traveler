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
    any := false
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
        any = any || all
    }
    return any, nil
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
    // Handle unary !
    for strings.HasPrefix(s, "!") {
        inner, err := evalSimple(strings.TrimPrefix(s, "!"))
        if err != nil {
            return false, err
        }
        return !inner, nil
    }

    // Operators supported: ==, !=, matches
    for _, op := range []string{"==", "!=", "matches"} {
        if idx := indexOutsideQuotes(s, op); idx >= 0 {
            left := strings.TrimSpace(s[:idx])
            right := strings.Trim(strings.TrimSpace(s[idx+len(op):]), "\"'")
            switch op {
            case "==":
                return left == right, nil
            case "!=":
                return left != right, nil
            case "matches":
                // Glob match
                return filepath.Match(right, left)
            }
        }
    }
    // Bareword truthy if non-empty and not "false"/"0"
    l := strings.ToLower(strings.Trim(s, "\"'"))
    if l == "false" || l == "0" { //nolint:gocritic
        return false, nil
    }
    if l == "true" || l == "1" {
        return true, nil
    }
    // Non-empty literal considered true
    if l != "" {
        return true, nil
    }
    return false, fmt.Errorf("could not evaluate expression: %q", s)
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

