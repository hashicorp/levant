// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"os"

	"github.com/mitchellh/cli"
)

func main() {
	os.Exit(Run(os.Args[1:]))
}

// Run sets up the commands and triggers RunCustom which inacts the correct
// run of Levant.
func Run(args []string) int {
	return RunCustom(args, Commands(nil))
}

// RunCustom is the main function to trigger a run of Levant.
func RunCustom(args []string, commands map[string]cli.CommandFactory) int {
	// Get the command line args. We shortcut "--version" and "-v" to
	// just show the version.
	for _, arg := range args {
		if arg == "-v" || arg == "-version" || arg == "--version" {
			newArgs := make([]string, len(args)+1)
			newArgs[0] = "version"
			copy(newArgs[1:], args)
			args = newArgs
			break
		}
	}

	// Build the commands to include in the help now.
	commandsInclude := make([]string, 0, len(commands))
	for k := range commands {
		switch k {
		default:
			commandsInclude = append(commandsInclude, k)
		}
	}

	cli := &cli.CLI{
		Args:     args,
		Commands: commands,
		HelpFunc: cli.FilteredHelpFunc(commandsInclude, cli.BasicHelpFunc("levant")),
	}

	exitCode, err := cli.Run()
	if err != nil {
		_, pErr := fmt.Fprintf(os.Stderr, "Error executing CLI: %s\n", err.Error())

		// If we are unable to log to stderr; try just printing the error to
		// provide some insight.
		if pErr != nil {
			fmt.Print(pErr)
		}

		return 1
	}

	return exitCode
}
