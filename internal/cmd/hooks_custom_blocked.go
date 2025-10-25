package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/klauern/blues-traveler/internal/config"
	"github.com/urfave/cli/v3"
)

// newHooksCustomBlockedCommand creates the blocked URL management command
func newHooksCustomBlockedCommand() *cli.Command {
	return &cli.Command{
		Name:  "blocked",
		Usage: "Manage blocked URL prefixes used by fetch-blocker",
		Commands: []*cli.Command{
			createBlockedListCommand(),
			createBlockedAddCommand(),
			createBlockedRemoveCommand(),
			createBlockedClearCommand(),
		},
	}
}

// createBlockedListCommand creates the list subcommand
func createBlockedListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List blocked URL prefixes",
		Flags: []cli.Flag{&cli.BoolFlag{Name: "global", Aliases: []string{"g"}}},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			path, lc, err := loadLogConfigForBlockedURLs(cmd.Bool("global"))
			if err != nil {
				return err
			}
			displayBlockedURLs(lc, path, cmd.Bool("global"))
			return nil
		},
	}
}

// createBlockedAddCommand creates the add subcommand
func createBlockedAddCommand() *cli.Command {
	return &cli.Command{
		Name:  "add",
		Usage: "Add a blocked URL prefix",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "global", Aliases: []string{"g"}},
			&cli.StringFlag{Name: "suggestion", Aliases: []string{"s"}},
		},
		ArgsUsage: "<prefix>",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			prefix, err := validateSingleArgument(cmd.Args().Slice())
			if err != nil {
				return err
			}

			useGlobal := cmd.Bool("global")
			path, lc, err := loadLogConfigForBlockedURLs(useGlobal)
			if err != nil {
				return err
			}

			if !addBlockedURL(lc, prefix, cmd.String("suggestion")) {
				fmt.Println("Prefix already present; no change.")
				return nil
			}

			if err := config.SaveLogConfig(path, lc); err != nil {
				return err
			}

			fmt.Printf("Added blocked prefix to %s: %s\n", path, prefix)
			return nil
		},
	}
}

// createBlockedRemoveCommand creates the remove subcommand
func createBlockedRemoveCommand() *cli.Command {
	return &cli.Command{
		Name:      "remove",
		Usage:     "Remove a blocked URL prefix",
		Flags:     []cli.Flag{&cli.BoolFlag{Name: "global", Aliases: []string{"g"}}},
		ArgsUsage: "<prefix>",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			prefix, err := validateSingleArgument(cmd.Args().Slice())
			if err != nil {
				return err
			}

			useGlobal := cmd.Bool("global")
			path, lc, err := loadLogConfigForBlockedURLs(useGlobal)
			if err != nil {
				return err
			}

			if !removeBlockedURL(lc, prefix) {
				fmt.Println("Prefix not found; no change.")
				return nil
			}

			if err := config.SaveLogConfig(path, lc); err != nil {
				return err
			}

			fmt.Printf("Removed blocked prefix from %s: %s\n", path, prefix)
			return nil
		},
	}
}

// createBlockedClearCommand creates the clear subcommand
func createBlockedClearCommand() *cli.Command {
	return &cli.Command{
		Name:  "clear",
		Usage: "Clear all blocked URL prefixes",
		Flags: []cli.Flag{&cli.BoolFlag{Name: "global", Aliases: []string{"g"}}},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			useGlobal := cmd.Bool("global")
			path, lc, err := loadLogConfigForBlockedURLs(useGlobal)
			if err != nil {
				return err
			}

			if len(lc.BlockedURLs) == 0 {
				fmt.Println("Blocked URLs already empty; no change.")
				return nil
			}

			lc.BlockedURLs = nil
			if err := config.SaveLogConfig(path, lc); err != nil {
				return err
			}

			fmt.Printf("Cleared blocked URLs in %s\n", path)
			return nil
		},
	}
}

// validateSingleArgument validates that exactly one argument is provided and returns it trimmed
func validateSingleArgument(args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("exactly one argument required: <prefix>")
	}
	prefix := strings.TrimSpace(args[0])
	if prefix == "" {
		return "", fmt.Errorf("prefix cannot be empty")
	}
	return prefix, nil
}
