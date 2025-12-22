package commands

import (
	"fmt"
	"sklair/cliutil"
	"sklair/commandRegistry"
)

const docsURL = "https://sklair-docs.numelon.com/"

func init() {
	commandRegistry.Registry.Register(&commandRegistry.Command{
		Name:        "docs",
		Description: "Opens Sklair documentation in your browser",
		Run: func(args []string) int {
			if err := cliutil.OpenBrowser(docsURL); err != nil {
				fmt.Println(docsURL)
				return 1
			}

			return 0
		},
	})
}
