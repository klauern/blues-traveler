// Package cursor provides Cursor IDE integration, including JSON protocol types,
// event models, and hook configuration management for transforming between
// Cursor's stdin/stdout format and Claude Code's event model.
package cursor

// Event names for Cursor hooks
const (
	BeforeShellExecution = "beforeShellExecution"
	BeforeMCPExecution   = "beforeMCPExecution"
	AfterFileEdit        = "afterFileEdit"
	BeforeReadFile       = "beforeReadFile"
	BeforeSubmitPrompt   = "beforeSubmitPrompt"
	Stop                 = "stop"
)

// HookInput represents the JSON input received from Cursor
type HookInput struct {
	ConversationID string   `json:"conversation_id"`
	GenerationID   string   `json:"generation_id"`
	HookEventName  string   `json:"hook_event_name"`
	WorkspaceRoots []string `json:"workspace_roots"`

	// beforeShellExecution
	Command string `json:"command,omitempty"`
	CWD     string `json:"cwd,omitempty"`

	// beforeMCPExecution
	ToolName  string `json:"tool_name,omitempty"`
	ToolInput string `json:"tool_input,omitempty"`
	URL       string `json:"url,omitempty"`

	// afterFileEdit, beforeReadFile
	FilePath string `json:"file_path,omitempty"`
	Content  string `json:"content,omitempty"`

	// afterFileEdit
	Edits []Edit `json:"edits,omitempty"`

	// beforeSubmitPrompt
	Prompt      string       `json:"prompt,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`

	// stop
	Status string `json:"status,omitempty"`
}

// Edit represents a file edit operation
type Edit struct {
	OldString string `json:"old_string"`
	NewString string `json:"new_string"`
}

// Attachment represents a prompt attachment
type Attachment struct {
	Type     string `json:"type"` // "file" | "rule"
	FilePath string `json:"file_path"`
}

// HookOutput represents the JSON response to Cursor
type HookOutput struct {
	Permission   string `json:"permission,omitempty"`   // "allow" | "deny" | "ask"
	UserMessage  string `json:"userMessage,omitempty"`  // Message shown to user
	AgentMessage string `json:"agentMessage,omitempty"` // Message sent to agent
	Continue     *bool  `json:"continue,omitempty"`     // For beforeSubmitPrompt
}

// Config represents the Cursor hooks.json configuration file
type Config struct {
	Version int                  `json:"version"`
	Hooks   map[string][]HookDef `json:"hooks"`
}

// HookDef represents a single hook definition in the config
type HookDef struct {
	Command string `json:"command"`
}

// NewConfig creates a new Cursor hooks config with version 1
func NewConfig() *Config {
	return &Config{
		Version: 1,
		Hooks:   make(map[string][]HookDef),
	}
}

// AddHook adds a hook command to the specified event
func (c *Config) AddHook(event, command string) {
	if c.Hooks == nil {
		c.Hooks = make(map[string][]HookDef)
	}
	c.Hooks[event] = append(c.Hooks[event], HookDef{Command: command})
}

// RemoveHook removes a hook command from the specified event
func (c *Config) RemoveHook(event, command string) bool {
	hooks, exists := c.Hooks[event]
	if !exists {
		return false
	}

	for i, hook := range hooks {
		if hook.Command == command {
			c.Hooks[event] = append(hooks[:i], hooks[i+1:]...)
			if len(c.Hooks[event]) == 0 {
				delete(c.Hooks, event)
			}
			return true
		}
	}
	return false
}

// HasHook checks if a hook command exists for the specified event
func (c *Config) HasHook(event, command string) bool {
	hooks, exists := c.Hooks[event]
	if !exists {
		return false
	}

	for _, hook := range hooks {
		if hook.Command == command {
			return true
		}
	}
	return false
}
