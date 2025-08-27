package cmd

import (
	"fmt"
	"os"

	"github.com/klauern/klauer-hooks/internal/generator"
	"github.com/spf13/cobra"
)

func NewGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate [hook-name]",
		Short: "Generate a new hook from template",
		Long: `Generate a new hook file from a template. This creates the hook implementation
and optionally a test file. The hook will need to be registered manually in the registry.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			hookName := args[0]

			// Get flags
			description, _ := cmd.Flags().GetString("description")
			hookTypeStr, _ := cmd.Flags().GetString("type")
			includeTest, _ := cmd.Flags().GetBool("test")
			outputDir, _ := cmd.Flags().GetString("output")

			// Validate hook name
			if err := generator.ValidateHookName(hookName); err != nil {
				fmt.Fprintf(os.Stderr, "Error validating hook name: %v\n", err)
				os.Exit(1)
			}

			// Set default description if not provided
			if description == "" {
				description = fmt.Sprintf("Custom %s hook implementation", hookName)
			}

			// Parse hook type
			var hookType generator.HookType
			switch hookTypeStr {
			case "pre", "pre_tool":
				hookType = generator.PreToolHook
			case "post", "post_tool":
				hookType = generator.PostToolHook
			case "both":
				hookType = generator.BothHooks
			default:
				fmt.Fprintf(os.Stderr, "Error: Invalid hook type '%s'. Valid types: pre, post, both\n", hookTypeStr)
				os.Exit(1)
			}

			// Create generator
			gen := generator.NewGenerator(outputDir)

			// Generate hook
			if err := gen.GenerateHook(hookName, description, hookType, includeTest); err != nil {
				fmt.Fprintf(os.Stderr, "Error generating hook: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("\n✅ Successfully generated hook '%s'\n", hookName)
		},
	}

	// Add flags for generate command
	cmd.Flags().StringP("description", "d", "", "Description of the hook")
	cmd.Flags().StringP("type", "t", "both", "Hook type: pre, post, or both")
	cmd.Flags().BoolP("test", "", true, "Generate test file")
	cmd.Flags().StringP("output", "o", "", "Output directory (default: internal/hooks)")

	return cmd
}
