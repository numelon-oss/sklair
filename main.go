package main

import (
	"context"
	"flag"
	"io"
	"os"
	"sklair/cli"

	commandregistry "github.com/numelon-oss/go-command-registry"
	"github.com/numelon-oss/go-logger"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	global := flag.NewFlagSet("sklair", flag.ContinueOnError)
	global.SetOutput(io.Discard)
	silent := global.Bool("silent", false, "Suppress all output except errors")
	verbose := global.Bool("verbose", false, "Enable verbose output")
	debug := global.Bool("debug", false, "Enable debug output")
	help := global.Bool("help", false, "Show command help")
	shortHelp := global.Bool("h", false, "Show command help")
	if err := global.Parse(args); err != nil {
		logger.Error("%s", err.Error())
		return int(commandregistry.ResultUsage)
	}
	if *silent && (*verbose || *debug) {
		logger.Error("Cannot use --silent with --verbose or --debug")
		return int(commandregistry.ResultUsage)
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
	args = global.Args()
	// forward root help after parsing global options. the registry renders all command help
	if *help || *shortHelp {
		args = append([]string{"help"}, args...)
	}
	return int(cli.New().Execute(context.Background(), struct{}{}, args))
}
