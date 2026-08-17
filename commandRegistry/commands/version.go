package commands

import (
	"context"
	"fmt"
	"sklair/commandRegistry"
	"sklair/constants"
)

func init() {
	commandRegistry.Registry.Register(&commandRegistry.Command{
		Name:        "version",
		Description: "Describes the current version of Sklair",
		Configure: commandRegistry.Simple(func(_ context.Context, _ *commandRegistry.Environment, args []string) error {
			if len(args) != 0 {
				return commandRegistry.UsageErrorf("Version does not accept arguments")
			}

			fmt.Printf(
				"sklair %s (%s) built %s\n",
				constants.Version,
				constants.Commit,
				constants.BuildDate,
			)
			return nil
		}),
	})
}
