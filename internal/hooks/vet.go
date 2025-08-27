package hooks

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/brads3290/cchooks"
	"github.com/klauern/klauer-hooks/internal/core"
)

// VetHook implements Python type checking logic using ty
type VetHook struct {
	*core.BaseHook
}

// NewVetHook creates a new vet hook instance
func NewVetHook(ctx *core.HookContext) core.Hook {
	base := core.NewBaseHook("vet", "Vet Hook", "Performs Python type checking using ty", ctx)
	return &VetHook{BaseHook: base}
}

// Run executes the vet hook.
func (h *VetHook) Run() error {
	if !h.IsEnabled() {
		fmt.Println("Vet plugin disabled - skipping")
		return nil
	}

	runner := h.Context().RunnerFactory(nil, h.postToolUseHandler, h.CreateRawHandler())
	runner.Run()
	return nil
}

func (h *VetHook) postToolUseHandler(ctx context.Context, event *cchooks.PostToolUseEvent) cchooks.PostToolUseResponseInterface {
	// Type check Python files after editing
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

		if filePath != "" && h.isPythonFile(filePath) {
			// Log detailed event data if logging is enabled
			if h.Context().LoggingEnabled {
				details := make(map[string]interface{})
				rawData := make(map[string]interface{})
				rawData["tool_name"] = event.ToolName
				details["file_path"] = filePath
				details["action"] = "vetting"

				h.LogHookEvent("vet_file", event.ToolName, rawData, details)
			}

			if err := h.typeCheckFile(filePath); err != nil {
				return cchooks.PostBlock(fmt.Sprintf("Vetting failed for %s: %v", filePath, err))
			}
		}
	}

	return cchooks.Allow()
}

func (h *VetHook) isPythonFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".py"
}

func (h *VetHook) typeCheckFile(filePath string) error {
	output, err := h.Context().CommandExecutor.ExecuteCommand("uvx", "ty", "check", filePath)
	if err != nil {
		log.Printf("ty check error on %s: %s", filePath, output)
		return fmt.Errorf("ty check failed: %s", output)
	}
	fmt.Printf("Vetted Python file: %s\n", filePath)
	return nil
}
