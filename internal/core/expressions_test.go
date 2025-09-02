package core

import "testing"

func TestEvalExpression_Basics(t *testing.T) {
    env := map[string]string{
        "TOOL_NAME":     "Edit",
        "EVENT_NAME":    "PreToolUse",
        "FILES_CHANGED": "foo.go bar.txt",
    }

    cases := []struct{
        expr string
        want bool
    }{
        {"${TOOL_NAME} == \"Edit\"", true},
        {"${TOOL_NAME} != \"Write\"", true},
        {"${EVENT_NAME} == PreToolUse", true},
        {"${FILES_CHANGED} matches *.rb", false},
        {"${TOOL_NAME} == \"Write\" || ${TOOL_NAME} == \"Edit\"", true},
        {"!(${TOOL_NAME} == \"Write\") && ${EVENT_NAME} == PreToolUse", true},
    }

    for _, tc := range cases {
        got, err := EvalExpression(tc.expr, env)
        if err != nil {
            t.Fatalf("eval %q error: %v", tc.expr, err)
        }
        if got != tc.want {
            t.Fatalf("eval %q = %v, want %v", tc.expr, got, tc.want)
        }
    }
}

