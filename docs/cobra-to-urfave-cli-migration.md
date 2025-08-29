# Migration Plan: Cobra to urfave/cli v3

## Overview
Convert the blues-traveler CLI from `github.com/spf13/cobra` v1.9.1 to `github.com/urfave/cli/v3` while maintaining all existing functionality.

## Key Changes Required

### 1. Import and Module Changes
- Update `go.mod` to replace `github.com/spf13/cobra v1.9.1` with `github.com/urfave/cli/v3`
- Update all imports from `"github.com/spf13/cobra"` to `"github.com/urfave/cli/v3"`

### 2. Core Architecture Changes
- **Replace `cobra.Command`** → **`cli.Command`**
- **Replace `rootCmd.Execute()`** → **`rootCmd.Run(context.Background(), os.Args)`**
- **Function signatures**: All handlers now take `(context.Context, *cli.Command)` instead of `(*cobra.Command, []string)`

### 3. Command Structure Changes
- Root command definition changes from `&cobra.Command{}` to `&cli.Command{}`
- `rootCmd.AddCommand()` becomes adding to `Commands: []*cli.Command{}`
- Command execution changes from `Execute()` to `Run(ctx, args)`

### 4. Flag System Migration
**Cobra → urfave/cli v3:**
- `cmd.Flags().StringVarP(&var, "name", "n", "default", "usage")` → `&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Value: "default", Usage: "usage", Destination: &var}`
- `cmd.Flags().BoolVarP(&var, "name", "n", false, "usage")` → `&cli.BoolFlag{Name: "name", Aliases: []string{"n"}, Value: false, Usage: "usage", Destination: &var}`
- `cmd.Flags().IntVarP(&var, "name", "n", 0, "usage")` → `&cli.IntFlag{Name: "name", Aliases: []string{"n"}, Value: 0, Usage: "usage", Destination: &var}`

### 5. Handler Function Updates
All command handlers need signature changes:
```go
// Cobra
Run: func(cmd *cobra.Command, args []string) { ... }

// urfave/cli v3
Action: func(ctx context.Context, cmd *cli.Command) error { ... }
```

### 6. Argument Access Changes
- `args` parameter → `cmd.Args().Slice()`
- `cobra.ExactArgs(1)` → manual validation in Action function

### 7. Flag Access Changes
- `cmd.Flags().GetString("flag")` → `cmd.String("flag")`
- `cmd.Flags().GetBool("flag")` → `cmd.Bool("flag")`
- `cmd.Flags().GetInt("flag")` → `cmd.Int("flag")`

## Files to Modify

1. **main.go** - Root command setup and execution
2. **internal/cmd/list.go** - List commands with flags
3. **internal/cmd/run.go** - Run command with logging flags
4. **internal/cmd/install.go** - Install command with multiple flags
5. **internal/cmd/config.go** - Config command with various flag types
6. **internal/cmd/generate.go** - Generate command with validation
7. **go.mod** - Dependency updates

## Implementation Steps

1. **Update dependencies** in go.mod
2. **Convert main.go** root command structure
3. **Migrate each command file** individually
4. **Update all function signatures** and flag handling
5. **Test each command** to ensure functionality is preserved
6. **Update documentation** references from Cobra to urfave/cli

## Detailed Migration Examples

### Example: main.go Conversion

**Before (Cobra):**
```go
var rootCmd = &cobra.Command{
    Use:   "blues-traveler",
    Short: "Claude Code hook runner and manager",
    Long:  `A CLI tool that runs Claude Code hooks directly...`,
}

func main() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
        os.Exit(1)
    }
}
```

**After (urfave/cli v3):**
```go
func main() {
    cmd := &cli.Command{
        Name:  "blues-traveler",
        Usage: "Claude Code hook runner and manager",
        Description: `A CLI tool that runs Claude Code hooks directly...`,
        Commands: []*cli.Command{
            // Add subcommands here
        },
    }

    if err := cmd.Run(context.Background(), os.Args); err != nil {
        fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
        os.Exit(1)
    }
}
```

### Example: Flag Migration

**Before (Cobra):**
```go
func NewRunCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "run [plugin-key]",
        Short: "Run a specific hook plugin",
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
            key := args[0]
            logEnabled, _ := cmd.Flags().GetBool("log")
            logFormat, _ := cmd.Flags().GetString("log-format")
            // ...
        },
    }

    cmd.Flags().BoolP("log", "l", false, "Enable detailed logging")
    cmd.Flags().String("log-format", "jsonl", "Log output format")
    return cmd
}
```

**After (urfave/cli v3):**
```go
func NewRunCmd() *cli.Command {
    return &cli.Command{
        Name:      "run",
        Usage:     "Run a specific hook plugin",
        ArgsUsage: "[plugin-key]",
        Flags: []cli.Flag{
            &cli.BoolFlag{
                Name:    "log",
                Aliases: []string{"l"},
                Value:   false,
                Usage:   "Enable detailed logging",
            },
            &cli.StringFlag{
                Name:  "log-format",
                Value: "jsonl",
                Usage: "Log output format",
            },
        },
        Action: func(ctx context.Context, cmd *cli.Command) error {
            args := cmd.Args().Slice()
            if len(args) != 1 {
                return fmt.Errorf("exactly one argument required")
            }
            key := args[0]
            logEnabled := cmd.Bool("log")
            logFormat := cmd.String("log-format")
            // ...
            return nil
        },
    }
}
```

### Example: Subcommand Registration

**Before (Cobra):**
```go
func init() {
    rootCmd.AddCommand(cmd.NewListCmd(getPluginWrapper, compat.PluginKeys))
    rootCmd.AddCommand(cmd.NewRunCmd(getPluginWrapper, compat.IsPluginEnabled, compat.PluginKeys))
    // ...
}
```

**After (urfave/cli v3):**
```go
func main() {
    cmd := &cli.Command{
        Name: "blues-traveler",
        Commands: []*cli.Command{
            NewListCmd(getPluginWrapper, compat.PluginKeys),
            NewRunCmd(getPluginWrapper, compat.IsPluginEnabled, compat.PluginKeys),
            // ...
        },
    }
}
```

## Risk Mitigation
- Maintain exact same CLI interface for users
- Preserve all existing flags, commands, and behaviors
- Test thoroughly after each file conversion
- Keep error handling and output formatting consistent

## Testing Strategy
1. Build and test after each file conversion
2. Verify all existing commands work identically
3. Test all flag combinations
4. Ensure help output remains user-friendly
5. Validate error handling and exit codes

## Documentation Updates
After migration, update:
- README.md references to CLI framework
- CLAUDE.md instruction about not referencing Cobra
- Any developer documentation mentioning Cobra

## Migration Status: ✅ COMPLETED
This migration has been successfully completed. The blues-traveler CLI now uses urfave/cli v3 instead of Cobra.

### Summary of Changes Made:
1. ✅ Updated `go.mod` to use `github.com/urfave/cli/v3 v3.4.1`
2. ✅ Converted `main.go` from Cobra command structure to urfave/cli v3
3. ✅ Migrated all command files: `list.go`, `run.go`, `install.go`, `config.go`, `generate.go`
4. ✅ Updated all flag definitions and handlers
5. ✅ Improved error handling (no more `os.Exit(1)` calls in command handlers)
6. ✅ Verified functionality with comprehensive testing
7. ✅ Updated documentation to reflect the change

All commands work identically to before - no breaking changes for users.
