package cmd

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

// VersionInfo holds version information
type VersionInfo struct {
	Version string
	Commit  string
	Date    string
	GoVer   string
}

// NewVersionCmd creates a new version command
func NewVersionCmd(versionInfo VersionInfo) *cli.Command {
	return &cli.Command{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "Show version information",
		Action: func(_ context.Context, _ *cli.Command) error {
			fmt.Printf("blues-traveler version %s\n", versionInfo.Version)
			fmt.Printf("commit: %s\n", versionInfo.Commit)
			fmt.Printf("date: %s\n", versionInfo.Date)
			fmt.Printf("go: %s\n", versionInfo.GoVer)
			return nil
		},
	}
}
