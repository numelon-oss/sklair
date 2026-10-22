package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sklair/sklairConfig"
)

func cleanCommand() *Command {
	return &Command{
		Name:        "clean",
		Description: "Remove build output and the .sklair workspace, including hook caches",
		Examples:    []string{"sklair clean"},
		Configure: noArguments("clean", func(_ context.Context) error {
			config, configDir, err := sklairConfig.LoadProjectConfig()
			if err != nil {
				return fmt.Errorf("could not load sklair.json: %w", err)
			}
			outputDir := filepath.Join(configDir, config.Output)
			if err := os.RemoveAll(outputDir); err != nil {
				return fmt.Errorf("could not remove output directory %s: %w", outputDir, err)
			}
			sklairDir := filepath.Join(configDir, ".sklair")
			if err := os.RemoveAll(sklairDir); err != nil {
				return fmt.Errorf("could not remove Sklair's workspace directory %s: %w", sklairDir, err)
			}
			return nil
		}),
	}
}
