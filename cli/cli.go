package cli

import (
	"context"
	"io"
	"os"

	commandregistry "github.com/numelon-oss/go-command-registry"
	"github.com/numelon-oss/go-logger"
)

type Command = commandregistry.Command[struct{}]
type Handler = commandregistry.Handler[struct{}]
type Registry = commandregistry.Registry[struct{}]

func New() *Registry {
	registry := commandregistry.New(commandregistry.Config[struct{}]{
		Program:         "sklair",
		RootArguments:   "[global options] <command> [options]",
		RootDescription: "A HTML compiler for components, Lua, and document generation.",
		EmptyResult:     commandregistry.ResultUsage,
		GlobalOptions: []commandregistry.Option{
			{Usage: "--silent", Description: "Suppress all output except errors"},
			{Usage: "--verbose", Description: "Include informational output"},
			{Usage: "--debug", Description: "Include debug output"},
			{Usage: "--help, -h", Description: "Show command help"},
		},
		Output: func(_ struct{}) io.Writer { return os.Stdout },
		ReportError: func(_ struct{}, err error) {
			logger.Error("%s", err.Error())
		},
	})
	registry.Register(buildCommand(), cleanCommand(), docsCommand(), initCommand(), serveCommand(), versionCommand())
	return registry
}

func noArguments(name string, run func(context.Context) error) commandregistry.Configure[struct{}] {
	return commandregistry.Simple(func(ctx context.Context, _ struct{}, args []string) error {
		if len(args) != 0 {
			return commandregistry.UsageErrorf("%s does not accept positional arguments", name)
		}
		return run(ctx)
	})
}
