package commandRegistry

import (
	"io"
	"os"

	commandregistry "github.com/numelon-oss/go-command-registry"
	"github.com/numelon-oss/go-logger"
)

type Environment struct{}

type Handler = commandregistry.Handler[*Environment]
type Configure = commandregistry.Configure[*Environment]
type Command = commandregistry.Command[*Environment]
type CommandRegistry = commandregistry.Registry[*Environment]

var UsageErrorf = commandregistry.UsageErrorf

func Simple(handler Handler) Configure {
	return commandregistry.Simple(handler)
}

func New() *CommandRegistry {
	return commandregistry.New(commandregistry.Config[*Environment]{
		Program:       "sklair",
		RootArguments: "[verbosity options] <command> [arguments]",
		GlobalOptions: []commandregistry.Option{
			{Usage: "--silent", Description: "Suppress all output except errors"},
			{Usage: "--verbose", Description: "Enable verbose output"},
			{Usage: "--debug", Description: "Enable debug output"},
			{Usage: "--help, -h", Description: "Show command help"},
		},
		EmptyResult: commandregistry.ResultUsage,
		Output: func(_ *Environment) io.Writer {
			return os.Stdout
		},
		ReportError: func(_ *Environment, err error) {
			logger.Error("%v", err)
		},
	})
}
