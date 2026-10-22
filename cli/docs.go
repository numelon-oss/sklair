package cli

import (
	"context"
	"fmt"
	"sklair/util"
)

const docsURL = "https://sklair.numelon.com/docs/"

func docsCommand() *Command {
	return &Command{
		Name:        "docs",
		Description: "Open the documentation in your browser",
		Examples:    []string{"sklair docs"},
		Configure: noArguments("docs", func(_ context.Context) error {
			if err := util.OpenBrowser(docsURL); err != nil {
				return fmt.Errorf("could not open documentation at %s: %w", docsURL, err)
			}
			return nil
		}),
	}
}
