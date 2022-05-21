package main

import (
	"fmt"

    "github.com/fatih/color"
)

var commandHelpTexts = [][]string {
	[]string{"install", "Install all the dependencies.", "toybox install", "Install all the dependencies specified in your Boxfile to source/libraries."},
	[]string{"update", "Update a single dependency.", "toybox update <dependency>", "Update a single dependency's version."},
	[]string{"add", "Add a new dependency.", "toybox add <dependency>", "Add a new dependency to your Boxfile and resolves the dependency set again."},
	[]string{"remove", "Remove a dependency.", "toybox remove <dependency>", "Removes a dependency from your Boxfile and resolves the dependency set again."},
	[]string{"login", "Login to the thing.", "toybox login", "Store a GitHub personal access token to lift GitHub API limits."},
	[]string{"generate", "Generate a new Toybox-equipped Playdate project.", "toybox update <destination path>", "Generates a new, well-structured Playdate project that is pre-wired for Toybox."},
	[]string{"info", "Describe your dependency set.", "toybox info", "Provides simple information about your dependency set."},
	[]string{"version", "Get the Toybox version.", "toybox version", "Show a version message for the Toybox client itself."},
	[]string{"help", "Show a help message.", "toybox help", "See a very helpful message about how to use Toybox."},
}

func ShowHelp(query string) {
	if query == "help" {
		DefaultHelp()
		return
	} else {
		for i := range commandHelpTexts {
			currentCommand := commandHelpTexts[i]
			if (currentCommand[0] == query) {
				ExplainCommand(currentCommand)
				return
			}
		}
	}

	DefaultHelp()
}

func ExplainCommand(command []string) {
	bold := color.New(color.Bold).SprintFunc()
	fmt.Printf("\n%s - %s\n\n%s %s\n\n%s\n\n", bold(color.GreenString(command[0])), bold(command[1]), bold("Usage:"), command[2], command[3])
}

func DefaultHelp() {
	output := "\n"
	bold := color.New(color.Bold).SprintFunc()

	for commandIndex := range commandHelpTexts {
		command := commandHelpTexts[commandIndex]
		output += fmt.Sprintf("%s - %s\n", bold(command[2]), command[1])
	}

	fmt.Println(output)
}