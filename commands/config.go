package commands

import (
	"fmt"
	"os"
	"sklair/cliutil"
	"sklair/commandRegistry"
)

func init() {
	commandRegistry.Registry.Register(&commandRegistry.Command{
		Name:        "config",
		Description: "Opens the global Sklair configuration file in your editor",
		Run: func(args []string) int {
			path, err := cliutil.GlobalConfigPath()
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
				return 1
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

			if err := cliutil.OpenEditor(path); err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
				return 1
			}

			return 0
		},
	})
}
