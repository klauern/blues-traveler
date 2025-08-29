package hooks

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/brads3290/cchooks"
	"github.com/klauern/blues-traveler/internal/core"
)

// FormatHook implements code formatting logic
type FormatHook struct {
	*core.BaseHook
}

// NewFormatHook creates a new format hook instance
func NewFormatHook(ctx *core.HookContext) core.Hook {
	base := core.NewBaseHook("format", "Format Hook", "Enforces code formatting standards", ctx)
	return &FormatHook{BaseHook: base}
}

// Run executes the format hook.
func (h *FormatHook) Run() error {
	if !h.IsEnabled() {
		fmt.Println("Format plugin disabled - skipping")
		return nil
	}

	runner := h.Context().RunnerFactory(nil, h.postToolUseHandler, h.CreateRawHandler())
	runner.Run()
	return nil
}

func (h *FormatHook) postToolUseHandler(ctx context.Context, event *cchooks.PostToolUseEvent) cchooks.PostToolUseResponseInterface {
	// Format code files after editing
	if event.ToolName == "Edit" || event.ToolName == "Write" {
		var filePath string

		switch event.ToolName {
		case "Edit":
			edit, err := event.InputAsEdit()
			if err == nil {
				filePath = edit.FilePath
			}
		case "Write":
			write, err := event.InputAsWrite()
			if err == nil {
				filePath = write.FilePath
			}
		}

		if filePath != "" {
			// Log detailed event data if logging is enabled
			if h.Context().LoggingEnabled {
				details := make(map[string]interface{})
				rawData := make(map[string]interface{})
				rawData["tool_name"] = event.ToolName
				details["file_path"] = filePath
				details["action"] = "formatting"

				h.LogHookEvent("format_file", event.ToolName, rawData, details)
			}

			if err := h.formatFile(filePath); err != nil {
				return cchooks.PostBlock(fmt.Sprintf("Formatting failed for %s: %v", filePath, err))
			}
		}
	}

	return cchooks.Allow()
}

func (h *FormatHook) formatFile(filePath string) error {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".go":
		return h.formatGoFile(filePath)
	case ".js", ".ts", ".jsx", ".tsx":
		return h.formatJSFile(filePath)
	case ".py":
		return h.formatPythonFile(filePath)
	case ".yml", ".yaml":
		return h.formatYAMLFile(filePath)
	}
	return nil
}

func (h *FormatHook) formatGoFile(filePath string) error {
	output, err := h.Context().CommandExecutor.ExecuteCommand("gofmt", "-w", filePath)
	if err != nil {
		log.Printf("gofmt error on %s: %s", filePath, output)
		return fmt.Errorf("gofmt failed: %s", output)
	}
	fmt.Printf("Formatted Go file: %s\n", filePath)
	return nil
}

func (h *FormatHook) formatJSFile(filePath string) error {
	output, err := h.Context().CommandExecutor.ExecuteCommand("prettier", "--write", filePath)
	if err != nil {
		log.Printf("prettier error on %s: %s", filePath, output)
		return fmt.Errorf("prettier failed: %s", output)
	}
	fmt.Printf("Formatted JS/TS file: %s\n", filePath)
	return nil
}

func (h *FormatHook) formatPythonFile(filePath string) error {
	// Run ruff format first
	output, err := h.Context().CommandExecutor.ExecuteCommand("uvx", "ruff", "format", filePath)
	if err != nil {
		log.Printf("ruff format error on %s: %s", filePath, output)
		return fmt.Errorf("ruff format failed: %s", output)
	}

	// Run ruff check --fix second
	output, err = h.Context().CommandExecutor.ExecuteCommand("uvx", "ruff", "check", "--fix", filePath)
	if err != nil {
		log.Printf("ruff check --fix error on %s: %s", filePath, output)
		return fmt.Errorf("ruff check --fix failed: %s", output)
	}

	fmt.Printf("Formatted Python file: %s\n", filePath)
	return nil
}

func (h *FormatHook) formatYAMLFile(filePath string) error {
	output, err := h.Context().CommandExecutor.ExecuteCommand("prettier", "--write", filePath)
	if err != nil {
		log.Printf("prettier error on %s: %s", filePath, output)
		return fmt.Errorf("prettier failed: %s", output)
	}
	fmt.Printf("Formatted YAML file: %s\n", filePath)
	return nil
}
