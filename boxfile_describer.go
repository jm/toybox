package main

import (
	"fmt"
)

func DescribeBoxfile() {
    root := Toybox{"root", []string{"default"}, "default", []*DependencyRelationship{}, []*DependencyRelationship{}, "default"}
	boxfile = &Boxfile{make(map[string]*Toybox), &root }

	boxfile.Load("./Boxfile")
	
	ReportInfo(fmt.Sprintf("Resolved with %d dependencies.\n", len(boxfile.Toyboxes)))

	for toybox := range boxfile.Toyboxes {
		Print(fmt.Sprintf("- %s@%s", toybox, boxfile.Toyboxes[toybox].CurrentlySelectedVersion))
	}

	fmt.Println()
}