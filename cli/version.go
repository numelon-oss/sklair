package cli

import (
	"context"
	"fmt"
	"sklair/constants"
)

func versionCommand() *Command {
	return &Command{
		Name:        "version",
		Description: "Print the Sklair version, commit, and build date",
		Examples:    []string{"sklair version"},
		Configure: noArguments("version", func(_ context.Context) error {
			fmt.Printf("sklair %s (%s) built %s\n", constants.Version, constants.Commit, constants.BuildDate)
			return nil
		}),
	}
}
