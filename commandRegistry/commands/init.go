package commands

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sklair/commandRegistry"
	"sklair/logger"
	"sklair/sklairConfig"
	"strings"
)

var reader = bufio.NewReader(os.Stdin)

func askString(prompt, fallback string) string {
	fmt.Printf("%s (default: %s): ", prompt, fallback)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return fallback
	}
	return input
}

func askBool(prompt string, fallback bool) bool {
	defaultStr := "no"
	if fallback {
		defaultStr = "yes"
	}

	fmt.Printf("%s (default: %s): ", prompt, defaultStr)
	var input string
	_, _ = fmt.Scanln(&input)

	switch strings.ToLower(input) {
	case "y", "yes", "true":
		return true
	case "n", "no", "false":
		return false
	default:
		return fallback
	}
}

func yesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

func preeeent(a, b string) {
	fmt.Printf("%-22s %s\n", a, b)
}

func configurationSummary(cfg sklairConfig.ProjectConfig) {
	fmt.Println()
	fmt.Println(logger.Green + "Configuration summary:")
	fmt.Println("--------------------------------------------------" + logger.Cyan)
	preeeent("Input directory:", cfg.Input)
	preeeent("Components directory:", cfg.Components)

	preeeent("Output directory:", cfg.Output)

	preeeent("Minify output:", yesNo(cfg.Minify))
	if cfg.ObfuscateJS != nil && cfg.ObfuscateJS.Enabled {
		preeeent("Obfuscate JavaScript:", "enabled")
	} else {
		preeeent("Obfuscate JavaScript:", "disabled")
	}

	if cfg.PreventFOUC != nil && cfg.PreventFOUC.Enabled {
		preeeent("Prevent FOUC:", "enabled")
		preeeent("Prevent FOUC colour:", cfg.PreventFOUC.Colour)
	} else {
		preeeent("Prevent FOUC:", "disabled")
	}

	fmt.Println(logger.Reset)
}

func init() {
	commandRegistry.Registry.Register(&commandRegistry.Command{
		Name:        "init",
		Description: "Initialises a Sklair project in the current directory",
		Run: func(args []string) int {
			fmt.Println(logger.Cyan + "Welcome to Sklair! 🪶 Let's get you set up." + logger.Reset)
			fmt.Println(logger.Green + "This will guide you through creating a sklair.json configuration file for your Sklair project.")
			fmt.Println("Press Enter to accept the default, which is shown in brackets." + logger.Reset)
			fmt.Println()
			fmt.Println()

			// TODO: add --yes flag which automatically saves default sklair config
			cfg := sklairConfig.DefaultConfig

			cfg.Input = askString("Where is your site source located?", cfg.Input)
			cfg.Components = askString("Where are your components located?", cfg.Components)

			fmt.Println(logger.Yellow + "File/folder exclusions (exclude field) cannot be configured using sklair init. You can edit them yourself in sklair.json.")
			fmt.Println("Likewise with compiling exclusions (excludeCompile field)." + logger.Reset)

			cfg.Output = askString("Where should the built site be written?", cfg.Output)

			cfg.Minify = askBool("Do you want Sklair to minify your outputted HTML?", cfg.Minify)
			// TODO: in the future, when ObfuscateJS is extended as a bigger object (instead of a regular bool),
			// this question stays the same but notify the user that they can configure it in sklair.json themselves
			cfg.ObfuscateJS.Enabled = askBool("Do you want Sklair to obfuscate your outputted JS?", cfg.ObfuscateJS.Enabled)

			// TODO: add a "more info available at <docs link>" to this question because it is a bit vague
			cfg.PreventFOUC.Enabled = askBool("Do you want Sklair to help prevent FOUC (Flash Of Unstyled Content)?", cfg.PreventFOUC.Enabled)
			if cfg.PreventFOUC.Enabled {
				cfg.PreventFOUC.Colour = askString("What colour do you want the FOUC indicator to be, in hex format?", cfg.PreventFOUC.Colour)
				// TODO: validate hex colour here
			} else {
				cfg.PreventFOUC = nil
			}

			// --------------------------------------------------

			configurationSummary(cfg)

			if !askBool("Write sklair.json with this configuration?", true) {
				fmt.Println("Aborted. No files were written.")
				return 0
			}

			// write file
			data, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				fmt.Println("Failed to serialise configuration:", err)
				return 1
			}

			if err := os.WriteFile("sklair.json", data, 0644); err != nil {
				fmt.Println("Failed to write sklair.json:", err)
				return 1
			}

			fmt.Println("Created sklair.json")
			return 0
		},
	})
}
