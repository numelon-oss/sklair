package commands

import (
	"context"
	"fmt"
	"sklair/commandRegistry"
	"sklair/util"
)

const docsURL = "https://sklair.numelon.com/docs"

func init() {
	commandRegistry.Registry.Register(&commandRegistry.Command{
		Name:        "docs",
		Description: "Opens the documentation in your browser",
		Configure: commandRegistry.Simple(func(_ context.Context, _ *commandRegistry.Environment, args []string) error {
			if len(args) != 0 {
				return commandRegistry.UsageErrorf("Docs does not accept arguments")
			}

			if err := util.OpenBrowser(docsURL); err != nil {
				fmt.Println(docsURL)
				return fmt.Errorf("open documentation: %w", err)
			}
			return nil
		}),
	})
}
