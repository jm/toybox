package main

import(
    "fmt"
    "os"

    "github.com/fatih/color"
)

func StartupBanner() {
	label := color.New(color.FgWhite, color.Bold, color.BgYellow)
	label.Printf("       ")
	
	message := color.New(color.Bold)
	message.Printf(" 🧸toybox v.%s\n", ToyboxVersion)
}

func Print(messageText string) {
	fmt.Printf("        %s\n", messageText)
}

func ReportDone() {
	label := color.New(color.FgWhite, color.Bold)
	label.Printf("\n DONE.  ")

	fmt.Println("Enjoy your packages. 👍\n")
}

func ReportProgress(messageText string) {
	label := color.New(color.FgWhite, color.Bold, color.BgGreen)
	label.Printf("       ")
	
	message := color.New(color.Bold)
	message.Printf(" %s\n", messageText)
}

func ReportInfo(messageText string) {
	label := color.New(color.FgWhite, color.Bold, color.BgBlue)
	label.Printf(" INFO. ")
	
	message := color.New(color.Bold)
	message.Printf(" %s\n", messageText)
}

func ReportError(messageText string, err error, exit bool) {
	label := color.New(color.FgWhite, color.Bold, color.BgRed)
	label.Printf(" ERROR ")
	
	message := color.New(color.Bold)

	if err != nil {
		message.Printf(" %s: \n", messageText)
		fmt.Println(err)
	} else {
		message.Printf(" %s\n", messageText)
	}

	if (exit == true) {
		os.Exit(1)
	}
}