package commands

import (
	"sklair/building"
	"sklair/commandRegistry"
	"sklair/logger"
	"sklair/sklairConfig"
)

func init() {
	commandRegistry.Registry.Register(&commandRegistry.Command{
		Name:        "build",
		Description: "Builds a Sklair project",
		Aliases:     []string{"b"},
		Run: func(args []string) int {
			config, configDir, err := sklairConfig.LoadProjectConfig()
			if err != nil {
				logger.Error("could not load sklair.json : %s", err.Error())
				return 1
			}

			err = building.Build(config, configDir, "")
			if err != nil {
				logger.Error(err.Error())
				return 1
			}

			return 0
		},
	})
}
