package cli

import (
	"context"
	"fmt"
	"sklair/building"
	"sklair/sklairConfig"
)

func buildCommand() *Command {
	return &Command{
		Name:        "build",
		Description: "Compile the configured project into its output directory",
		Examples:    []string{"sklair build", "sklair --verbose build"},
		Configure: noArguments("build", func(_ context.Context) error {
			config, configDir, err := sklairConfig.LoadProjectConfig()
			if err != nil {
				return fmt.Errorf("could not load sklair.json: %w", err)
			}
			return building.Build(config, configDir, "")
		}),
	}
}
