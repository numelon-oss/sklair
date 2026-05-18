package commandRegistry

// adapted from CommandRegistry intended for a bot in another numelon-proprietary project

import (
	"flag"
	"fmt"
)

type Command struct {
	Name        string
	Description string
	Flags       func() *flag.FlagSet
	Run         func(args []string) int
}

type CommandRegistry struct {
	commands map[string]*Command
}

func New() *CommandRegistry {
	r := &CommandRegistry{
		commands: make(map[string]*Command),
	}

	r.Register(&Command{
		Name:        "help",
		Description: "Shows available commands",
		Run: func(args []string) int {
			r.PrintHelp()
			return 0
		},
	})

	return r
}

func (r *CommandRegistry) Register(cmd *Command) {
	r.commands[cmd.Name] = cmd
}

func (r *CommandRegistry) Get(name string) (*Command, bool) {
	cmd, ok := r.commands[name]
	return cmd, ok
}

func (r *CommandRegistry) PrintHelp() {
	fmt.Println("Usage:")
	fmt.Println("\tsklair [verbosity flags] <command> [arguments]")
	fmt.Println()
	fmt.Println("Available verbosity flags:") // TODO: instead of just putting all the verbosity flags here, maybe add them to the registry somehow?
	fmt.Println("\t--silent: Suppress all output except errors")
	fmt.Println("\t--verbose: Enable verbose output")
	fmt.Println("\t--debug: Enable debug output")
	fmt.Println()
	fmt.Println()
	fmt.Println("Available commands:")

	seen := make(map[*Command]bool)
	for _, cmd := range r.commands {
		if seen[cmd] {
			continue
		}
		seen[cmd] = true

		fmt.Printf("\t%-12s %s\n", cmd.Name, cmd.Description)
	}
	//fmt.Println()
	//fmt.Println("\tUse 'sklair help <command>' to get help for a specific command.") // TODO: todo!!
}
