package main

import(
	"fmt"
	"github.com/fatih/color"
)

var ToyboxVersion = "0.0.3"

func PrintVersion() {
	title := color.New(color.FgCyan, color.Bold)
	title.Printf("🧸toybox ")
	fmt.Printf("v.%s\n", ToyboxVersion)
}