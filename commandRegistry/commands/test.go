package commands

import (
	"context"
	"fmt"
	"sklair/commandRegistry"
)

func init() {
	commandRegistry.Registry.Register(&commandRegistry.Command{
		Name:        "test",
		Description: "A test command",
		Configure: commandRegistry.Simple(func(_ context.Context, _ *commandRegistry.Environment, args []string) error {
			fmt.Println(args)
			return nil
		}),
	})
}
