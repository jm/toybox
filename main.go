package main

import (
	"fmt"
	"os"
	
	"golang.org/x/term"
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
		passwd, _ := term.ReadPassword(0)
		fmt.Println(passwd)
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
