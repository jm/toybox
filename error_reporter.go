package main

import(
    "fmt"
    "os"

    "github.com/fatih/color"
)

func ReportError(messageText string, err error, exit bool) {
	label := color.New(color.FgWhite, color.Bold, color.BgRed)
	label.Printf(" ERROR ")
	
	message := color.New(color.Bold)

	if err != nil {
		message.Printf(" %s: ", messageText)
		fmt.Println(err)
	} else {
		message.Printf(" %s", messageText)
	}

	if (exit == true) {
		os.Exit(1)
	}
}