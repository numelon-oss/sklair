package commands

import (
	"os"
	"path/filepath"
	"sklair/commandRegistry"
	"sklair/logger"
	"sklair/sklairConfig"
)

func init() {
	commandRegistry.Registry.Register(&commandRegistry.Command{
		Name:        "clean",
		Description: "Removes all temporary and generated files made by Sklair, including hook-created caches",
		Aliases:     []string{"c"},
		Run: func(args []string) int {
			config, configDir, err := sklairConfig.LoadProjectConfig()
			if err != nil {
				logger.Error("could not load sklair.json : %s", err.Error())
				return 1
			}

			sklairDir := filepath.Join(configDir, ".sklair")

			tempDir := filepath.Join(sklairDir, "temp")
			generatedDir := filepath.Join(sklairDir, "generated")

			if err == nil {
				outputDir := filepath.Join(configDir, config.Output)
				err = os.RemoveAll(outputDir)
				if err != nil {
					logger.Error("could not remove output directory %s : %s", outputDir, err.Error())
					return 1
				}
			}
			err = os.RemoveAll(tempDir)
			if err != nil {
				logger.Error("could not remove Sklair's temp directory %s : %s", tempDir, err.Error())
				return 1
			}
			err = os.RemoveAll(generatedDir)
			if err != nil {
				logger.Error("could not remove Sklair's generated directory %s : %s", generatedDir, err.Error())
				return 1
			}

			return 0
		},
	})
}
