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
	"github.com/klauern/blues-traveler/internal/core"
)

var (
	// Cache command availability to avoid repeated PATH lookups
	gofumptOnce       sync.Once
	gofumptAvailable  bool
	prettierOnce      sync.Once
	prettierAvailable bool
	uvxOnce           sync.Once
	uvxAvailable      bool
)

// checkGofumptAvailable checks if gofumpt is available in PATH (cached)
func checkGofumptAvailable() bool {
	gofumptOnce.Do(func() {
		_, err := exec.LookPath("gofumpt")
		gofumptAvailable = err == nil
	})
	return gofumptAvailable
}

// checkPrettierAvailable checks if prettier is available in PATH (cached)
func checkPrettierAvailable() bool {
	prettierOnce.Do(func() {
		_, err := exec.LookPath("prettier")
		prettierAvailable = err == nil
	})
	return prettierAvailable
}

// SetAvailabilityForTesting forces availability flags for testing
func SetAvailabilityForTesting(gofumpt, prettier, uvx bool) {
	gofumptOnce.Do(func() {})
	prettierOnce.Do(func() {})
	uvxOnce.Do(func() {})
	gofumptAvailable = gofumpt
	prettierAvailable = prettier
	uvxAvailable = uvx
}

// GetAvailabilityForTesting returns current availability flags for testing
func GetAvailabilityForTesting() (gofumpt, prettier, uvx bool) {
	return gofumptAvailable, prettierAvailable, uvxAvailable
}

// checkUvxAvailable checks if uvx is available in PATH (cached)
func checkUvxAvailable() bool {
	uvxOnce.Do(func() {
		_, err := exec.LookPath("uvx")
		uvxAvailable = err == nil
	})
	return uvxAvailable
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

// postToolUseHandler handles post-tool-use events and formats edited files
func (h *FormatHook) postToolUseHandler(ctx context.Context, event *cchooks.PostToolUseEvent) cchooks.PostToolUseResponseInterface {
	// Format code files after editing
	if event.ToolName == "Edit" || event.ToolName == "Write" {
		var filePath string

		switch event.ToolName {
		case "Edit":
			edit, err := event.InputAsEdit()
			if err != nil {
				if h.Context().LoggingEnabled {
					log.Printf("Failed to parse Edit input: %v", err)
				}
			} else {
				filePath = edit.FilePath
			}
		case "Write":
			write, err := event.InputAsWrite()
			if err != nil {
				if h.Context().LoggingEnabled {
					log.Printf("Failed to parse Write input: %v", err)
				}
			} else {
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

// formatFile formats a file based on its extension
func (h *FormatHook) formatFile(filePath string) error {
	// Validate file path
	if filePath == "" {
		return fmt.Errorf("empty file path")
	}

	// Check if file exists and is accessible
	if _, err := h.Context().FileSystem.Stat(filePath); err != nil {
		return fmt.Errorf("file not accessible: %w", err)
	}

	// Clean the path to prevent path traversal
	cleanPath := filepath.Clean(filePath)
	// Only reject paths that escape the workspace (start with ".." or "../")
	if cleanPath == ".." || strings.HasPrefix(cleanPath, ".."+string(filepath.Separator)) {
		return fmt.Errorf("invalid file path: path traversal attempt detected")
	}

	// Use cleanPath for all subsequent operations
	ext := strings.ToLower(filepath.Ext(cleanPath))

	switch ext {
	case ".go":
		return h.formatGoFile(cleanPath)
	case ".js", ".ts", ".jsx", ".tsx":
		return h.formatJSFile(cleanPath)
	case ".py":
		return h.formatPythonFile(cleanPath)
	case ".yml", ".yaml":
		return h.formatYAMLFile(cleanPath)
	}
	return nil
}

// formatGoFile formats a Go file using gofumpt or gofmt
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

// formatJSFile formats a JavaScript/TypeScript file using prettier
func (h *FormatHook) formatJSFile(filePath string) error {
	return h.formatWithPrettier(filePath, "JS/TS")
}

// formatWithPrettier formats a file using prettier (shared by JS and YAML)
func (h *FormatHook) formatWithPrettier(filePath string, fileType string) error {
	if !checkPrettierAvailable() {
		return fmt.Errorf("prettier not found in PATH (required for %s formatting)", fileType)
	}

	output, err := h.Context().CommandExecutor.ExecuteCommand("prettier", "--write", filePath)
	if err != nil {
		log.Printf("prettier error on %s: %s", filePath, output)
		return fmt.Errorf("prettier failed: %s", output)
	}
	fmt.Printf("Formatted %s file: %s\n", fileType, filePath)
	return nil
}

// formatPythonFile formats a Python file using ruff
func (h *FormatHook) formatPythonFile(filePath string) error {
	// Check if uvx is available (required for ruff)
	if !checkUvxAvailable() {
		return fmt.Errorf("uvx not found in PATH (required for ruff Python formatting)")
	}

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

// formatYAMLFile formats a YAML file using prettier
func (h *FormatHook) formatYAMLFile(filePath string) error {
	return h.formatWithPrettier(filePath, "YAML")
}
