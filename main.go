package main

import (
	"flag"
	"fmt"
	"os"
	"sklair/commandRegistry"
	_ "sklair/commandRegistry/commands"

	"github.com/numelon-oss/go-logger"
)

func main() {
	os.Exit(run())
}

func run() int {
	reg := commandRegistry.Registry

	global := flag.NewFlagSet("sklair", flag.ContinueOnError)

	silent := global.Bool("silent", false, "Suppress all output except errors")
	verbose := global.Bool("verbose", false, "Enable verbose output")
	debug := global.Bool("debug", false, "Enable debug output")

	help := global.Bool("help", false, "Show help")
	if *help {
		reg.PrintHelp()
		return 0
	}

	// wrong usage
	if err := global.Parse(os.Args[1:]); err != nil {
		return 2
	}

	if *silent && (*verbose || *debug) {
		_, _ = fmt.Fprintf(os.Stderr, "%sCannot use --silent with --verbose or --debug%s\n", logger.Red, logger.Reset)
		return 2
	}

	level := logger.LevelWarning
	switch {
	case *silent:
		level = logger.LevelError
	case *debug:
		level = logger.LevelDebug
	case *verbose:
		level = logger.LevelInfo
	}
	logger.SetLevel(level)

	// --------------------------------------------------

	args := global.Args()
	if len(args) == 0 {
		reg.PrintHelp()
		return 2
	}

	cmdName := args[0]
	cmd, ok := reg.Get(cmdName)
	if !ok {
		_, _ = fmt.Fprintf(os.Stderr, "%sUnknown command: %s%s\n\n", logger.Red, cmdName, logger.Reset)
		reg.PrintHelp()
		return 2
	}

	// TODO: set up the sklair dir inside the users home directory here along with the default app config

	remainingArgs := args[1:]
	if cmd.Flags != nil {
		fs := cmd.Flags()

		if err := fs.Parse(remainingArgs); err != nil {
			return 2
		}

		// fs is the flagset, fs.Args() gives the remaining args
		// in this case, we send the remaining args
		// as flags are assumed to be handled already in most command init() funcs
		return cmd.Run(fs.Args())
	}

	return cmd.Run(remainingArgs)
}
