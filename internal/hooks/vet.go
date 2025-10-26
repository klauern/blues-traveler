package hooks

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/brads3290/cchooks"
	"github.com/klauern/blues-traveler/internal/constants"
	"github.com/klauern/blues-traveler/internal/core"
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

func (h *VetHook) postToolUseHandler(_ context.Context, event *cchooks.PostToolUseEvent) cchooks.PostToolUseResponseInterface {
	// Type check Python files after editing
	if event.ToolName != constants.ToolEdit && event.ToolName != constants.ToolWrite {
		return cchooks.Allow()
	}

	filePath := h.extractFilePath(event)
	if filePath == "" || !h.isPythonFile(filePath) {
		return cchooks.Allow()
	}

	h.logVetEvent(event.ToolName, filePath)

	if err := h.typeCheckFile(filePath); err != nil {
		// User-friendly message + technical details for agent
		userMsg := fmt.Sprintf("Code quality check failed for %s", filepath.Base(filePath))
		agentMsg := fmt.Sprintf("Type checking failed for %s: %v", filePath, err)
		return core.PostBlockWithMessages(userMsg, agentMsg)
	}

	return cchooks.Allow()
}

func (h *VetHook) extractFilePath(event *cchooks.PostToolUseEvent) string {
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

func (h *VetHook) logVetEvent(toolName, filePath string) {
	if !h.Context().LoggingEnabled {
		return
	}

	details := map[string]interface{}{
		"file_path": filePath,
		"action":    "vetting",
	}
	rawData := map[string]interface{}{
		"tool_name": toolName,
	}

	h.LogHookEvent("vet_file", toolName, rawData, details)
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
