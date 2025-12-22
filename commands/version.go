package commands

import (
	"fmt"
	"sklair/commandRegistry"
	"sklair/constants"
)

func init() {
	commandRegistry.Registry.Register(&commandRegistry.Command{
		Name:        "version",
		Description: "Describes the current version of Sklair",
		Aliases:     []string{"v"},
		Run: func(args []string) int {
			fmt.Printf(
				"sklair %s (%s) built %s\n",
				constants.Version,
				constants.Commit,
				constants.BuildDate,
			)
			return 0
		},
	})
}
