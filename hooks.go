package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/brads3290/cchooks"
)

// HookType represents available hook types
type HookType struct {
	Name        string
	Description string
	Runner      func() error
}

// GetHookTypes returns available hook types
func GetHookTypes() map[string]HookType {
	return map[string]HookType{
		"security": {
			Name:        "Security Hook",
			Description: "Blocks dangerous commands and provides security controls",
			Runner:      runSecurityHook,
		},
		"format": {
			Name:        "Format Hook",
			Description: "Enforces code formatting standards",
			Runner:      runFormatHook,
		},
		"debug": {
			Name:        "Debug Hook",
			Description: "Logs all tool usage for debugging purposes",
			Runner:      runDebugHook,
		},
		"audit": {
			Name:        "Audit Hook",
			Description: "Comprehensive audit logging with JSON output",
			Runner:      runAuditHook,
		},
	}
}

// runSecurityHook implements security blocking logic
func runSecurityHook() error {
	runner := &cchooks.Runner{
		PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface {
			// Block dangerous commands
			if event.ToolName == "Bash" {
				bash, err := event.AsBash()
				if err != nil {
					return cchooks.Block("failed to parse bash command")
				}

				// List of dangerous command patterns
				dangerousPatterns := []string{
					"rm -rf",
					"sudo rm",
					"dd if=",
					"mkfs",
					"format",
					"> /dev/",
				}

				for _, pattern := range dangerousPatterns {
					if strings.Contains(strings.ToLower(bash.Command), pattern) {
						return cchooks.Block(fmt.Sprintf("dangerous command pattern detected: %s", pattern))
					}
				}
			}

			return cchooks.Approve()
		},
	}

	runner.Run()
	return nil
}

// runFormatHook implements code formatting logic
func runFormatHook() error {
	runner := &cchooks.Runner{
		PostToolUse: func(ctx context.Context, event *cchooks.PostToolUseEvent) cchooks.PostToolUseResponseInterface {
			// Format code files after editing
			if event.ToolName == "Edit" || event.ToolName == "Write" {
				var filePath string

				if event.ToolName == "Edit" {
					edit, err := event.InputAsEdit()
					if err == nil {
						filePath = edit.FilePath
					}
				} else if event.ToolName == "Write" {
					write, err := event.InputAsWrite()
					if err == nil {
						filePath = write.FilePath
					}
				}

				if filePath != "" {
					ext := strings.ToLower(filepath.Ext(filePath))

					switch ext {
					case ".go":
						// TODO: Execute gofmt command
						fmt.Printf("Would format Go file: %s\n", filePath)
					case ".js", ".ts", ".jsx", ".tsx":
						// TODO: Execute prettier command
						fmt.Printf("Would format JS/TS file: %s\n", filePath)
					case ".py":
						// TODO: Execute black command
						fmt.Printf("Would format Python file: %s\n", filePath)
					}
				}
			}

			return cchooks.Allow()
		},
	}

	runner.Run()
	return nil
}

// runDebugHook implements debug logging logic
func runDebugHook() error {
	// Setup logging
	logFile, err := os.OpenFile("claude-hooks.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	defer logFile.Close()

	logger := log.New(logFile, "", log.LstdFlags)

	runner := &cchooks.Runner{
		PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface {
			logger.Printf("PRE-TOOL: %s", event.ToolName)

			// Log specific tool details
			switch event.ToolName {
			case "Bash":
				if bash, err := event.AsBash(); err == nil {
					logger.Printf("  Command: %s", bash.Command)
				}
			case "Edit":
				if edit, err := event.AsEdit(); err == nil {
					logger.Printf("  File: %s", edit.FilePath)
				}
			case "Write":
				if write, err := event.AsWrite(); err == nil {
					logger.Printf("  File: %s", write.FilePath)
				}
			case "Read":
				if read, err := event.AsRead(); err == nil {
					logger.Printf("  File: %s", read.FilePath)
				}
			}

			return cchooks.Approve()
		},

		PostToolUse: func(ctx context.Context, event *cchooks.PostToolUseEvent) cchooks.PostToolUseResponseInterface {
			logger.Printf("POST-TOOL: %s", event.ToolName)
			return cchooks.Allow()
		},
	}

	fmt.Println("Debug hook started - logging to claude-hooks.log")
	runner.Run()
	return nil
}

// AuditEntry represents an audit log entry
type AuditEntry struct {
	Timestamp string                 `json:"timestamp"`
	Event     string                 `json:"event"`
	ToolName  string                 `json:"tool_name"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

func logAuditEntry(entry AuditEntry) {
	entry.Timestamp = time.Now().Format(time.RFC3339)

	jsonData, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal audit entry: %v\n", err)
		return
	}

	fmt.Println(string(jsonData))
}

// runAuditHook implements comprehensive audit logging
func runAuditHook() error {
	runner := &cchooks.Runner{
		PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface {
			entry := AuditEntry{
				Event:    "pre_tool_use",
				ToolName: event.ToolName,
				Details:  make(map[string]interface{}),
			}

			// Add tool-specific details
			switch event.ToolName {
			case "Bash":
				if bash, err := event.AsBash(); err == nil {
					entry.Details["command"] = bash.Command
					entry.Details["description"] = bash.Description
				}
			case "Edit":
				if edit, err := event.AsEdit(); err == nil {
					entry.Details["file_path"] = edit.FilePath
					entry.Details["old_string_length"] = len(edit.OldString)
					entry.Details["new_string_length"] = len(edit.NewString)
				}
			case "Write":
				if write, err := event.AsWrite(); err == nil {
					entry.Details["file_path"] = write.FilePath
					entry.Details["content_length"] = len(write.Content)
				}
			case "Read":
				if read, err := event.AsRead(); err == nil {
					entry.Details["file_path"] = read.FilePath
				}
			}

			logAuditEntry(entry)
			return cchooks.Approve()
		},

		PostToolUse: func(ctx context.Context, event *cchooks.PostToolUseEvent) cchooks.PostToolUseResponseInterface {
			entry := AuditEntry{
				Event:    "post_tool_use",
				ToolName: event.ToolName,
				Details:  make(map[string]interface{}),
			}

			logAuditEntry(entry)
			return cchooks.Allow()
		},
	}

	runner.Run()
	return nil
}
