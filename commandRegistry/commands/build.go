package commands

import (
	"context"
	"fmt"
	"sklair/building"
	"sklair/commandRegistry"
	"sklair/sklairConfig"
)

func init() {
	commandRegistry.Registry.Register(&commandRegistry.Command{
		Name:        "build",
		Description: "Builds a Sklair project",
		Configure: commandRegistry.Simple(func(_ context.Context, _ *commandRegistry.Environment, args []string) error {
			if len(args) != 0 {
				return commandRegistry.UsageErrorf("Build does not accept arguments")
			}

			config, configDir, err := sklairConfig.LoadProjectConfig()
			if err != nil {
				return fmt.Errorf("load sklair.json: %w", err)
			}

			if err := building.Build(config, configDir, ""); err != nil {
				return fmt.Errorf("build project: %w", err)
			}
			return nil
		}),
	})
}
