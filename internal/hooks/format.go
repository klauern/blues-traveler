package hooks

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/brads3290/cchooks"
	"github.com/klauern/blues-traveler/internal/constants"
	"github.com/klauern/blues-traveler/internal/core"
)

var (
	// Cache command availability to avoid repeated PATH lookups
	gofumptOnce      sync.Once
	gofumptAvailable bool
)

// checkGofumptAvailable checks if gofumpt is available in PATH (cached)
func checkGofumptAvailable() bool {
	gofumptOnce.Do(func() {
		_, err := exec.LookPath("gofumpt")
		gofumptAvailable = err == nil
	})
	return gofumptAvailable
}

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

func (h *FormatHook) postToolUseHandler(_ context.Context, event *cchooks.PostToolUseEvent) cchooks.PostToolUseResponseInterface {
	// Format code files after editing
	if event.ToolName != constants.ToolEdit && event.ToolName != constants.ToolWrite {
		return cchooks.Allow()
	}

	filePath := h.extractFilePath(event)
	if filePath == "" {
		return cchooks.Allow()
	}

	h.logFormatEvent(event.ToolName, filePath)

	if err := h.formatFile(filePath); err != nil {
		// User-friendly message + technical details for agent
		userMsg := fmt.Sprintf("Code formatting failed for %s", filepath.Base(filePath))
		agentMsg := fmt.Sprintf("Formatting failed for %s: %v", filePath, err)
		return core.PostBlockWithMessages(userMsg, agentMsg)
	}

	return cchooks.Allow()
}

func (h *FormatHook) extractFilePath(event *cchooks.PostToolUseEvent) string {
	switch event.ToolName {
	case constants.ToolEdit:
		if edit, err := event.InputAsEdit(); err == nil {
			return edit.FilePath
		}
	case constants.ToolWrite:
		if write, err := event.InputAsWrite(); err == nil {
			return write.FilePath
		}
	}
	return ""
}

func (h *FormatHook) logFormatEvent(toolName, filePath string) {
	if !h.Context().LoggingEnabled {
		return
	}

	details := map[string]interface{}{
		"file_path": filePath,
		"action":    "formatting",
	}
	rawData := map[string]interface{}{
		"tool_name": toolName,
	}

	h.LogHookEvent("format_file", toolName, rawData, details)
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
	var output []byte
	var err error
	var formatter string

	// Prefer gofumpt over gofmt if available
	if checkGofumptAvailable() {
		output, err = h.Context().CommandExecutor.ExecuteCommand("gofumpt", "-w", filePath)
		formatter = "gofumpt"
	} else {
		output, err = h.Context().CommandExecutor.ExecuteCommand("gofmt", "-w", filePath)
		formatter = "gofmt"
	}

	if err != nil {
		log.Printf("%s error on %s: %s", formatter, filePath, output)
		return fmt.Errorf("%s failed: %s", formatter, output)
	}
	fmt.Printf("Formatted Go file with %s: %s\n", formatter, filePath)
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
