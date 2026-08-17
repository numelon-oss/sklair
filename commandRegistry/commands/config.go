package commands

import (
	"context"
	"fmt"
	"sklair/commandRegistry"
	"sklair/sklairConfig"
	"sklair/util"
)

func init() {
	commandRegistry.Registry.Register(&commandRegistry.Command{
		Name:        "config",
		Description: "Opens the global Sklair configuration file in your editor",
		Configure: commandRegistry.Simple(func(_ context.Context, _ *commandRegistry.Environment, args []string) error {
			if len(args) != 0 {
				return commandRegistry.UsageErrorf("Config does not accept arguments")
			}

			path, err := sklairConfig.GlobalConfigPath()
			if err != nil {
				return fmt.Errorf("resolve global configuration path: %w", err)
			}

			// TODO: this should be done anyways in main.go!
			// all commands (in theory) rely on the global sklair configuration
			//if _, err := os.Stat(path); os.IsNotExist(err) {
			//	err = os.WriteFile(path, []byte("{}"), 0644)
			//	if err != nil {
			//		_, _ = fmt.Fprintln(os.Stderr, err)
			//		return 1
			//	}
			//}

			if err := util.OpenEditor(path); err != nil {
				return fmt.Errorf("open global configuration: %w", err)
			}
			return nil
		}),
	})
}
