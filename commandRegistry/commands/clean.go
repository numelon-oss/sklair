package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sklair/commandRegistry"
	"sklair/sklairConfig"
)

func init() {
	commandRegistry.Registry.Register(&commandRegistry.Command{
		Name:        "clean",
		Description: "Removes all temporary and generated files made by Sklair, including hook-created caches",
		Configure: commandRegistry.Simple(func(_ context.Context, _ *commandRegistry.Environment, args []string) error {
			if len(args) != 0 {
				return commandRegistry.UsageErrorf("Clean does not accept arguments")
			}

			config, configDir, err := sklairConfig.LoadProjectConfig()
			if err != nil {
				return fmt.Errorf("load sklair.json: %w", err)
			}

			sklairDir := filepath.Join(configDir, ".sklair")

			tempDir := filepath.Join(sklairDir, "temp")
			generatedDir := filepath.Join(sklairDir, "generated")

			outputDir := filepath.Join(configDir, config.Output)
			if err := os.RemoveAll(outputDir); err != nil {
				return fmt.Errorf("remove output directory %s: %w", outputDir, err)
			}
			if err := os.RemoveAll(tempDir); err != nil {
				return fmt.Errorf("remove Sklair temp directory %s: %w", tempDir, err)
			}
			if err := os.RemoveAll(generatedDir); err != nil {
				return fmt.Errorf("remove Sklair generated directory %s: %w", generatedDir, err)
			}
			return nil
		}),
	})
}
