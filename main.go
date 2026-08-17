package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sklair/commandRegistry"
	_ "sklair/commandRegistry/commands"
	"syscall"

	"github.com/numelon-oss/go-logger"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	reg := commandRegistry.Registry

	global := flag.NewFlagSet("sklair", flag.ContinueOnError)
	global.SetOutput(io.Discard)

	silent := global.Bool("silent", false, "Suppress all output except errors")
	verbose := global.Bool("verbose", false, "Enable verbose output")
	debug := global.Bool("debug", false, "Enable debug output")

	help := global.Bool("help", false, "Show help")
	shortHelp := global.Bool("h", false, "Show help")

	if err := global.Parse(args); err != nil {
		logger.Error("%v", err)
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

	args = global.Args()
	if *help || *shortHelp {
		args = append([]string{"help"}, args...)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return int(reg.Execute(ctx, &commandRegistry.Environment{}, args))
}
