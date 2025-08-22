package main

import (
	"fmt"
	"strings"
)

// HookTemplate represents a template for generating hooks
type HookTemplate struct {
	Name        string
	Description string
	Code        string
}

// GetHookTemplates returns available hook templates
func GetHookTemplates() map[string]HookTemplate {
	return map[string]HookTemplate{
		"security": {
			Name:        "Security Hook",
			Description: "Blocks dangerous commands and provides security controls",
			Code: `package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/brads3290/cchooks"
)

func main() {
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
}`,
		},
		"format": {
			Name:        "Format Hook",
			Description: "Enforces code formatting standards",
			Code: `package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/brads3290/cchooks"
)

func main() {
	runner := &cchooks.Runner{
		PostToolUse: func(ctx context.Context, event *cchooks.PostToolUseEvent) cchooks.PostToolUseResponseInterface {
			// Format code files after editing
			if event.ToolName == "Edit" || event.ToolName == "Write" {
				var filePath string
				
				if event.ToolName == "Edit" {
					edit, err := event.AsEdit()
					if err == nil {
						filePath = edit.FilePath
					}
				} else if event.ToolName == "Write" {
					write, err := event.AsWrite()
					if err == nil {
						filePath = write.FilePath
					}
				}

				if filePath != "" {
					ext := strings.ToLower(filepath.Ext(filePath))
					
					switch ext {
					case ".go":
						return cchooks.RunCommand("gofmt -w " + filePath)
					case ".js", ".ts", ".jsx", ".tsx":
						return cchooks.RunCommand("prettier --write " + filePath)
					case ".py":
						return cchooks.RunCommand("black " + filePath)
					}
				}
			}

			return cchooks.Continue()
		},
	}

	runner.Run()
}`,
		},
		"debug": {
			Name:        "Debug Hook",
			Description: "Logs all tool usage for debugging purposes",
			Code: `package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/brads3290/cchooks"
)

func main() {
	// Setup logging
	logFile, err := os.OpenFile("claude-hooks.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file:", err)
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
			logger.Printf("POST-TOOL: %s (Success: %t)", event.ToolName, event.Success)
			if !event.Success && event.Error != "" {
				logger.Printf("  Error: %s", event.Error)
			}
			return cchooks.Continue()
		},
	}

	fmt.Println("Debug hook started - logging to claude-hooks.log")
	runner.Run()
}`,
		},
		"audit": {
			Name:        "Audit Hook",
			Description: "Comprehensive audit logging with JSON output",
			Code: `package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/brads3290/cchooks"
)

type AuditEntry struct {
	Timestamp string                 ` + "`json:\"timestamp\"`" + `
	Event     string                 ` + "`json:\"event\"`" + `
	ToolName  string                 ` + "`json:\"tool_name\"`" + `
	Success   *bool                  ` + "`json:\"success,omitempty\"`" + `
	Error     string                 ` + "`json:\"error,omitempty\"`" + `
	Details   map[string]interface{} ` + "`json:\"details,omitempty\"`" + `
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

func main() {
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
				Success:  &event.Success,
				Error:    event.Error,
				Details:  make(map[string]interface{}),
			}
			
			if event.Output != "" {
				entry.Details["output_length"] = len(event.Output)
			}
			
			logAuditEntry(entry)
			return cchooks.Continue()
		},
	}

	runner.Run()
}`,
		},
	}
}