package cmd

import (
	"context"
	"fmt"

	"github.com/klauern/blues-traveler/internal/generator"
	"github.com/urfave/cli/v3"
)

// parseHookType converts string to HookType
func parseHookType(hookTypeStr string) (generator.HookType, error) {
	switch hookTypeStr {
	case "pre", "pre_tool":
		return generator.PreToolHook, nil
	case "post", "post_tool":
		return generator.PostToolHook, nil
	case "both":
		return generator.BothHooks, nil
	default:
		return generator.PreToolHook, fmt.Errorf("invalid hook type '%s'. Valid types: pre, post, both", hookTypeStr)
	}
}

// executeGenerateCommand executes the generate command
func executeGenerateCommand(hookName, description, hookTypeStr, outputDir string, includeTest bool) error {
	// Validate hook name
	if err := generator.ValidateHookName(hookName); err != nil {
		return fmt.Errorf("invalid hook name '%s': %w\n  Suggestion: Hook names must be valid Go identifiers (alphanumeric and underscores only)", hookName, err)
	}

	// Set default description if not provided
	if description == "" {
		description = fmt.Sprintf("Custom %s hook implementation", hookName)
	}

	// Parse hook type
	hookType, err := parseHookType(hookTypeStr)
	if err != nil {
		return err
	}

	// Create generator
	gen := generator.NewGenerator(outputDir)

	// Generate hook
	if err := gen.GenerateHook(hookName, description, hookType, includeTest); err != nil {
		return fmt.Errorf("failed to generate hook '%s': %w\n  Suggestion: Check write permissions in the output directory '%s'", hookName, err, outputDir)
	}

	fmt.Printf("\n✅ Successfully generated hook '%s'\n", hookName)
	return nil
}

// NewGenerateCmd creates the 'generate' CLI command for generating new hooks from templates
func NewGenerateCmd() *cli.Command {
	return &cli.Command{
		Name:      "generate",
		Usage:     "Generate a new hook from template",
		ArgsUsage: "[hook-name]",
		Description: `Generate a new hook file from a template. This creates the hook implementation
and optionally a test file. The hook will need to be registered manually in the registry.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "description",
				Aliases: []string{"d"},
				Value:   "",
				Usage:   "Description of the hook",
			},
			&cli.StringFlag{
				Name:    "type",
				Aliases: []string{"t"},
				Value:   "both",
				Usage:   "Hook type: pre, post, or both",
			},
			&cli.BoolFlag{
				Name:  "test",
				Value: true,
				Usage: "Generate test file",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Value:   "",
				Usage:   "Output directory (default: internal/hooks)",
			},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			args := cmd.Args().Slice()
			if len(args) != 1 {
				return fmt.Errorf("exactly one argument required: [hook-name]")
			}
			hookName := args[0]

			return executeGenerateCommand(
				hookName,
				cmd.String("description"),
				cmd.String("type"),
				cmd.String("output"),
				cmd.Bool("test"),
			)
		},
	}
}
