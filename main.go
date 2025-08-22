package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "hooks",
	Short: "CLI tool for managing Claude Code hooks",
	Long:  `A CLI tool to quickly create and manage Claude Code hooks using the cchooks library.`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}