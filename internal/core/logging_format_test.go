package core

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/klauern/klauer-hooks/internal/config"
)

// helper to read file lines trimming trailing newline
func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}

func TestLogHookEvent_JSONLFormat(t *testing.T) {
	ctx := DefaultHookContext()
	ctx.LoggingEnabled = true
	ctx.LoggingDir = t.TempDir()
	ctx.LoggingFormat = config.LoggingFormatJSONL

	logHookEvent(ctx, "testhook", "test_event", "ToolX",
		map[string]interface{}{"k": "v"},
		map[string]interface{}{"d": 1},
	)

	logFile := filepath.Join(ctx.LoggingDir, "testhook.log")
	lines, err := readLines(logFile)
	if err != nil {
		t.Fatalf("failed reading log file: %v", err)
	}
	if len(lines) != 1 {
		t.Fatalf("expected exactly 1 line for jsonl, got %d", len(lines))
	}
	line := lines[0]
	if strings.HasPrefix(line, "  ") {
		t.Errorf("jsonl line should not start with indentation")
	}
	// Should be valid JSON
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(line), &obj); err != nil {
		t.Fatalf("jsonl line not valid JSON: %v", err)
	}
	if obj["event"] != "test_event" {
		t.Errorf("expected event 'test_event', got %v", obj["event"])
	}
}

func TestLogHookEvent_PrettyFormat(t *testing.T) {
	ctx := DefaultHookContext()
	ctx.LoggingEnabled = true
	ctx.LoggingDir = t.TempDir()
	ctx.LoggingFormat = config.LoggingFormatPretty

	logHookEvent(ctx, "prettyhook", "pretty_event", "ToolY",
		map[string]interface{}{"a": "b"},
		map[string]interface{}{"x": 42},
	)

	logFile := filepath.Join(ctx.LoggingDir, "prettyhook.log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed reading log file: %v", err)
	}
	text := string(content)
	// Expect multiple lines due to indentation
	if !strings.Contains(text, "\n  ") {
		t.Errorf("expected indented pretty JSON containing newline + two spaces")
	}
	// Remove trailing newline for parsing (json.Indent already produced)
	raw := strings.TrimSpace(text)
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &obj); err != nil {
		t.Fatalf("pretty JSON not valid: %v", err)
	}
	if obj["event"] != "pretty_event" {
		t.Errorf("expected event 'pretty_event', got %v", obj["event"])
	}
}
