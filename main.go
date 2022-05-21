package main

import (
	"fmt"
	"os"
)

var Credential *GitHubCredential

func main() {
	if len(os.Args) < 2 {
		ReportError("Expected a subcommand!  Maybe start with `toybox help`?", nil, true)
	}

	StartupBanner()
	switch os.Args[1] {
	case "install":
		InstallDependencies()
	case "login":
		LoginOrRenewGitHubCredential()
	case "add":
		bfe := BoxfileEditor{}
		bfe.Add(os.Args[2])

		InstallDependencies()
	case "remove":
		bfe := BoxfileEditor{}
		bfe.Remove(os.Args[2])

		InstallDependencies()
	case "update":
		UpdateDependency(os.Args[2])
	case "generate":
		// Generate new tb equipped project
	case "info":
		bfd := BoxfileDescriber{}
		bfd.Describe()
	case "help":
		h := Helper{(os.Args[2:])}
		h.DispenseKnowledge()
	case "version":
		PrintVersion()
	default:
		fmt.Println("expected subcommand")
		os.Exit(1)
	}
}
