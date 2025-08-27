package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/klauern/klauer-hooks/internal/config"
)

// LogEntry represents a detailed log entry for hook inspection (moved from base.go)
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	HookKey   string                 `json:"hook_key"`
	Event     string                 `json:"event"`
	ToolName  string                 `json:"tool_name"`
	RawData   map[string]interface{} `json:"raw_data,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// logHookEvent centralizes structured hook event logging.
// It is a no-op if LoggingEnabled is false.
func logHookEvent(ctx *HookContext, hookKey, event, toolName string,
	rawData map[string]interface{}, details map[string]interface{},
) {
	if ctx == nil || !ctx.LoggingEnabled {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		HookKey:   hookKey,
		Event:     event,
		ToolName:  toolName,
		RawData:   rawData,
		Details:   details,
	}

	// Ensure logging directory exists
	logDir := ctx.LoggingDir
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create log directory %s: %v\n", logDir, err)
		return
	}

	// Create log file path
	logFile := filepath.Join(logDir, fmt.Sprintf("%s.log", hookKey))

	// Marshal entry to JSON respecting format
	var jsonData []byte
	var err error
	if ctx.LoggingFormat == config.LoggingFormatPretty {
		jsonData, err = json.MarshalIndent(entry, "", "  ")
	} else {
		jsonData, err = json.Marshal(entry)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal log entry: %v\n", err)
		return
	}

	// Append to log file
	file, err := ctx.FileSystem.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file %s: %v\n", logFile, err)
		return
	}
	defer func() { _ = file.Close() }()

	if _, err := file.WriteString(string(jsonData) + "\n"); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write to log file: %v\n", err)
	}
}
