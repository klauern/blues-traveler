package generator

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/klauern/blues-traveler/internal/constants"
)

//go:embed templates/*
var templates embed.FS

// HookType represents the type of hook to generate
type HookType string

const (
	PreToolHook  HookType = "pre_tool"
	PostToolHook HookType = "post_tool"
	BothHooks    HookType = "both"
)

// TemplateData holds data for template rendering
type TemplateData struct {
	Name        string // PascalCase name (e.g., "MyCustom")
	LowerName   string // snake_case name (e.g., "my_custom")
	Description string // Human readable description
	ModulePath  string // Module import path
}

// Generator handles hook code generation
type Generator struct {
	outputDir string
}

// NewGenerator creates a new generator instance
func NewGenerator(outputDir string) *Generator {
	if outputDir == "" {
		outputDir = constants.InternalHooksDir
	}
	return &Generator{outputDir: outputDir}
}

// GenerateHook generates a new hook file and test file
func (g *Generator) GenerateHook(name, description string, hookType HookType, includeTest bool) error {
	// Validate inputs
	if name == "" {
		return fmt.Errorf("hook name cannot be empty")
	}
	if description == "" {
		return fmt.Errorf("hook description cannot be empty")
	}

	// Prepare template data
	data := TemplateData{
		Name:        toPascalCase(name),
		LowerName:   toSnakeCase(name),
		Description: description,
		ModulePath:  constants.ModulePath,
	}

	// Generate main hook file
	templateName := string(hookType) + "_hook.go.tmpl"
	if err := g.generateFile(templateName, data, data.LowerName+".go"); err != nil {
		return fmt.Errorf("failed to generate hook file: %v", err)
	}

	// Generate test file if requested
	if includeTest {
		if err := g.generateFile("test_hook.go.tmpl", data, data.LowerName+"_test.go"); err != nil {
			return fmt.Errorf("failed to generate test file: %v", err)
		}
	}

	// Show registration instructions
	g.showRegistrationInstructions(data)

	return nil
}

func (g *Generator) generateFile(templateName string, data TemplateData, outputFileName string) error {
	// Read template
	templateContent, err := templates.ReadFile("templates/" + templateName)
	if err != nil {
		return fmt.Errorf("failed to read template %s: %v", templateName, err)
	}

	// Parse template
	tmpl, err := template.New(templateName).Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(g.outputDir, 0o750); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Create output file
	outputPath := filepath.Join(g.outputDir, outputFileName)
	file, err := os.Create(outputPath) // #nosec G304 - controlled output directory
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %v", outputPath, err)
	}
	defer func() { _ = file.Close() }()

	// Execute template
	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	fmt.Printf("Generated: %s\n", outputPath)
	return nil
}

func (g *Generator) showRegistrationInstructions(data TemplateData) {
	fmt.Println("\n📝 Registration Instructions:")
	fmt.Printf("Add the following line to %s/init.go in the init() function:\n", constants.InternalHooksDir)
	fmt.Printf("    MustRegisterHook(\"%s\", New%sHook)\n", data.LowerName, data.Name)
	fmt.Println("\n🧪 Testing:")
	fmt.Printf("    go test ./%s -run Test%sHook\n", constants.InternalHooksDir, data.Name)
	fmt.Println("\n🔧 Usage:")
	fmt.Printf("    ./%s run %s\n", constants.BinaryName, data.LowerName)
	fmt.Printf("    ./%s install %s\n", constants.BinaryName, data.LowerName)
}

// ValidateHookName checks if a hook name is valid
func ValidateHookName(name string) error {
	if name == "" {
		return fmt.Errorf("hook name cannot be empty")
	}
	if strings.Contains(name, " ") {
		return fmt.Errorf("hook name cannot contain spaces (use underscores or hyphens)")
	}
	// Check if it's a reserved name
	reserved := []string{"security", "format", "debug", "audit"}
	lowerName := strings.ToLower(name)
	for _, r := range reserved {
		if lowerName == r {
			return fmt.Errorf("hook name '%s' is reserved", name)
		}
	}
	return nil
}

// ListAvailableTypes returns available hook types
func ListAvailableTypes() []HookType {
	return []HookType{PreToolHook, PostToolHook, BothHooks}
}

// Helper functions for name conversion

func toPascalCase(s string) string {
	// Convert snake_case or kebab-case to PascalCase
	s = strings.ReplaceAll(s, "-", "_")
	parts := strings.Split(s, "_")
	result := ""
	for _, part := range parts {
		if len(part) > 0 {
			result += strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	return result
}

func toSnakeCase(s string) string {
	// Convert PascalCase or kebab-case to snake_case
	s = strings.ReplaceAll(s, "-", "_")

	var result []rune
	for i, r := range s {
		if i > 0 && 'A' <= r && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, rune(strings.ToLower(string(r))[0]))
	}
	return string(result)
}
