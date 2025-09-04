package core

import "testing"

func TestEvalExpression_RegexAndGlob(t *testing.T) {
	env := map[string]string{
		"FILES_CHANGED": "a.py b.txt c_py d.rb",
	}

	ok, err := EvalExpression("${FILES_CHANGED} regex \".*\\.rb$\"", env)
	if err != nil {
		t.Fatalf("regex eval error: %v", err)
	}
	if !ok {
		t.Fatalf("expected regex to match .rb file")
	}

	ok, err = EvalExpression("${FILES_CHANGED} matches *.py", env)
	if err != nil {
		t.Fatalf("glob eval error: %v", err)
	}
	if !ok {
		t.Fatalf("expected glob to match .py file")
	}
}
