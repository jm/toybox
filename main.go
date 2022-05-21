package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("expected subcommand")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "install":
		installer := Installer{}
		installer.Install()
	case "login":
		ReportError("Error writing import file", nil, true)
		if credential := LoadGitHubCredential(); credential != nil {
			fmt.Println("Logged in to GitHub as", credential.User)
			fmt.Printf("Update stored credentials? (y/n) ")
			
			answer := "n"
			fmt.Scanln(&answer)

			if answer == "y" {
				fmt.Println()
				RequestGitHubCredential()	
			}
		} else {
			RequestGitHubCredential()
		}
	case "add":
		// barCmd.Parse(os.Args[2:])
		fmt.Println("subcommand 'bar'")
	case "remove":
		fmt.Println("subcommand 'bar'")
	case "info":
		fmt.Println("subcommand 'bar'")
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
