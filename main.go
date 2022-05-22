package main

import (
	"os"
)

var Credential *GitHubCredential
var DependencyManager = Manager{}

func main() {
	if len(os.Args) < 2 {
		ReportError("Expected a subcommand!  Maybe start with `toybox help`?", nil, true)
	}

	StartupBanner()
    Credential = LoadGitHubCredential()

	switch os.Args[1] {
	case "install":
		DependencyManager.Install()
	case "login":
		LoginOrRenewGitHubCredential()
	case "add":
		bfe := BoxfileEditor{}
		bfe.Add(os.Args[2])

		DependencyManager.Install()
	case "remove":
		bfe := BoxfileEditor{}
		bfe.Remove(os.Args[2])

		DependencyManager.Install()
	case "update":
		UpdateDependency(os.Args[2])
	case "generate":
		GenerateProject(os.Args[2])
	case "info":
		DescribeBoxfile()
	case "help":
		ShowHelp(os.Args[len(os.Args)-1])
	case "version":
		PrintVersion()
	default:
		ShowHelp(os.Args[len(os.Args)-1])
	}
}
