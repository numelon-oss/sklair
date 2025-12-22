package main

import (
	"flag"
	"fmt"
	"os"
	"sklair/commandRegistry"
	"sklair/logger"

	_ "sklair/commands"
)

func main() {
	os.Exit(run())
}

func run() int {
	reg := *commandRegistry.Registry

	global := flag.NewFlagSet("sklair", flag.ContinueOnError)

	// TODO: add these verbosity flags to the main help command
	silent := global.Bool("silent", false, "Suppress all output except errors")
	verbose := global.Bool("verbose", false, "Enable verbose output")
	debug := global.Bool("debug", false, "Enable debug output")

	help := global.Bool("help", false, "Show help")
	if *help {
		reg.PrintHelp()
		return 0
	}

	// TODO: ensure that help also shows the silent, verbose, debug flags
	// replace flag libs -help and --help with our own help command

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

	return cmd.Run(args[1:])
}
