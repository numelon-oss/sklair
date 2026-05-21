package commands

import (
	"flag"
	"fmt"
	"sklair/commandRegistry"
)

func init() {
	commandRegistry.Registry.Register(&commandRegistry.Command{
		Name:        "test",
		Description: "A test command",

		Flags: func() *flag.FlagSet {
			fs := flag.NewFlagSet("Test", flag.ExitOnError)
			return fs
		},
		Run: func(args []string) int {
			fmt.Println(args)

			return 0
		},
	})
}
