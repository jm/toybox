package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type BoxfileEditor struct {
	File Boxfile
}

func (e *BoxfileEditor) Add(dependencyToAdd string) {
	dependencies := e.loadBoxfile()

	version := "default"
	if strings.Contains(dependencyToAdd, "@") {
		pieces := strings.Split(dependencyToAdd, "@")
		
		dependencyToAdd = pieces[0]
		version = pieces[1]
	}
	Print(fmt.Sprintf("Adding %s@%s...", dependencyToAdd, version))

	dependencies[dependencyToAdd] = version

	e.writeBoxfile(dependencies)
}

func (e *BoxfileEditor) Remove(dependencyToRemove string) {
	dependencies := e.loadBoxfile()
	Print(fmt.Sprintf("Removing %s...", dependencyToRemove))

	delete(dependencies, dependencyToRemove)

	e.writeBoxfile(dependencies)
}

func (e *BoxfileEditor) loadBoxfile() map[string]string {
	jsonString, err := os.ReadFile("Boxfile")

	if err != nil {
		ReportError("Error reading Boxfile", err, true)
    }

	parsedData := map[string]string{}
    err = json.Unmarshal([]byte(jsonString), &parsedData)

    if err != nil {
		ReportError("Error parsing Boxfile", err, true)
    }

    return parsedData
}

func (e *BoxfileEditor) writeBoxfile(dependencyList map[string]string) {
	boxfileData, _ := json.MarshalIndent(dependencyList, "", " ")
	_ = ioutil.WriteFile("Boxfile", boxfileData, 0644)
}