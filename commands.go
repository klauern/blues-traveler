package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(showCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available hook templates",
	Long:  `List all available hook templates that can be created.`,
	Run: func(cmd *cobra.Command, args []string) {
		templates := GetHookTemplates()
		
		fmt.Println("Available hook templates:")
		fmt.Println()
		
		for name, template := range templates {
			fmt.Printf("  %s - %s\n", name, template.Description)
		}
		fmt.Println()
		fmt.Println("Use 'hooks create <template>' to create a hook from a template.")
		fmt.Println("Use 'hooks show <template>' to preview a template.")
	},
}

var createCmd = &cobra.Command{
	Use:   "create [template] [name]",
	Short: "Create a new hook from a template",
	Long:  `Create a new hook file from one of the available templates.`,
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		templateName := args[0]
		
		templates := GetHookTemplates()
		template, exists := templates[templateName]
		if !exists {
			fmt.Fprintf(os.Stderr, "Error: Template '%s' not found.\n", templateName)
			fmt.Fprintf(os.Stderr, "Available templates: %s\n", strings.Join(getTemplateNames(), ", "))
			os.Exit(1)
		}
		
		// Determine output filename
		var filename string
		if len(args) > 1 {
			filename = args[1]
		} else {
			filename = templateName + "-hook"
		}
		
		// Ensure .go extension
		if !strings.HasSuffix(filename, ".go") {
			filename += ".go"
		}
		
		// Check if file already exists
		if _, err := os.Stat(filename); err == nil {
			fmt.Fprintf(os.Stderr, "Error: File '%s' already exists.\n", filename)
			os.Exit(1)
		}
		
		// Write template to file
		err := os.WriteFile(filename, []byte(template.Code), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("Created %s hook: %s\n", template.Name, filename)
		fmt.Printf("Description: %s\n", template.Description)
		fmt.Println()
		fmt.Printf("Next steps:\n")
		fmt.Printf("  1. Review and customize the hook code in %s\n", filename)
		fmt.Printf("  2. Build the hook: go build -o %s %s\n", 
			strings.TrimSuffix(filename, ".go"), filename)
		fmt.Printf("  3. Configure Claude Code to use the hook\n")
	},
}

var showCmd = &cobra.Command{
	Use:   "show [template]",
	Short: "Show the code for a template",
	Long:  `Display the source code for a hook template without creating a file.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		templateName := args[0]
		
		templates := GetHookTemplates()
		template, exists := templates[templateName]
		if !exists {
			fmt.Fprintf(os.Stderr, "Error: Template '%s' not found.\n", templateName)
			fmt.Fprintf(os.Stderr, "Available templates: %s\n", strings.Join(getTemplateNames(), ", "))
			os.Exit(1)
		}
		
		fmt.Printf("Template: %s\n", template.Name)
		fmt.Printf("Description: %s\n", template.Description)
		fmt.Println()
		fmt.Println("Code:")
		fmt.Println("---")
		fmt.Println(template.Code)
	},
}

func getTemplateNames() []string {
	templates := GetHookTemplates()
	names := make([]string, 0, len(templates))
	for name := range templates {
		names = append(names, name)
	}
	return names
}

// Add build command
var buildCmd = &cobra.Command{
	Use:   "build [file]",
	Short: "Build a hook file into an executable",
	Long:  `Build a Go hook file into an executable that can be used with Claude Code.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sourceFile := args[0]
		
		// Check if source file exists
		if _, err := os.Stat(sourceFile); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: File '%s' does not exist.\n", sourceFile)
			os.Exit(1)
		}
		
		// Determine output name
		outputName := strings.TrimSuffix(filepath.Base(sourceFile), ".go")
		
		// Build the hook
		buildCmd := fmt.Sprintf("go build -o %s %s", outputName, sourceFile)
		fmt.Printf("Building: %s\n", buildCmd)
		
		// You would typically use exec.Command here, but for simplicity:
		fmt.Printf("Run this command to build your hook:\n")
		fmt.Printf("  %s\n", buildCmd)
		fmt.Println()
		fmt.Printf("Then configure Claude Code to use: ./%s\n", outputName)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}