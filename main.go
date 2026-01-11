package main

import (
	"flag"
	"fmt"
	"os"
	"sklair/commandRegistry"
	_ "sklair/commands"
	"sklair/logger"
)

func main() {
	os.Exit(run())
}

func run() int {
	reg := *commandRegistry.Registry

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
		_, _ = fmt.Fprintln(os.Stderr, "Cannot use --silent with --verbose or --debug")
		return 2
	}

	level := logger.LevelWarning
	switch {
	case *silent:
		level = logger.LevelError
	case *debug:
		level = logger.LevelDebug

		// TODO: this is temporary!!
		if exePath, err := os.Executable(); err == nil {
			fmt.Println(exePath)
		} else {
			panic(err)
		}
	case *verbose:
		level = logger.LevelInfo
	}

	logger.InitShared(level)

	// --------------------------------------------------

	args := global.Args()
	if len(args) == 0 {
		reg.PrintHelp()
		return 2
	}

	cmdName := args[0]
	cmd, ok := reg.Get(cmdName)
	if !ok {
		_, _ = fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmdName)
		reg.PrintHelp()
		return 2
	}

	// TODO: set up the sklair dir inside the users home directory here along with the default app config

	return cmd.Run(args[1:])
}
