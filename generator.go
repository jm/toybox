package main

import (
	"fmt"
	"path"
	"path/filepath"
	"os"
	"os/user"
	"runtime"
)

func GenerateProject(destination string) {
	gameName := path.Base(destination)
	currentUser, _ := user.Current()

	ReportProgress("Generating paths...")
	pathsToGenerate := []string{
		filepath.Join(destination, "source/libraries"),
		filepath.Join(destination, "source/images"),
		filepath.Join(destination, "source/sounds"),
		filepath.Join(destination, "source/fonts"),
	}

	for pathIndex := range pathsToGenerate {
		Print(fmt.Sprintf("Generating %s", pathsToGenerate[pathIndex]))
		if err := os.MkdirAll(pathsToGenerate[pathIndex], os.ModePerm); err != nil {
        	ReportError(fmt.Sprintf("Error generating path (%s)", pathsToGenerate[pathIndex]), nil, true)
    	}
	}

	ReportProgress("Adding files...")
	Print("Adding pdxinfo")
	pdxInfo := fmt.Sprintf("name=%s\nauthor=%s\ndescription=TODO: Describe this game.\nbundleID=com.%s.%s\nversion=0.0.1\nbuildNumber=1\nimagePath=images/launcher/\nlaunchSoundPath=path/to/launch/sound", gameName, currentUser.Name, currentUser.Username, gameName)

	err := os.WriteFile(filepath.Join(destination, "pdxinfo"), []byte(pdxInfo), 0644)
    if err != nil {
        ReportError("Error writing pdxinfo file", err, true)
    }

	Print("Adding main.lua")
	mainFile := fmt.Sprintf("import \"CoreLibs/object\"\nimport \"CoreLibs/graphics\"\n\nlocal gfx <const> = playdate.graphics\n\nfunction setupGame()\n    gfx.drawText(\"Hello, world, from *%s*!\", 20, 100)\nend\n\nsetupGame()\n\nfunction playdate.update()\n    gfx.sprite.update()\nend", gameName)
	
	err = os.WriteFile(filepath.Join(destination, "source", "main.lua"), []byte(mainFile), 0644)
    if err != nil {
        ReportError("Error writing main code file", err, true)
    }

	Print("Adding Boxfile")
	boxfileExample := "{\n  \"owner/dependency\":\"1.0\"\n}"
	
	err = os.WriteFile(filepath.Join(destination, "Boxfile"), []byte(boxfileExample), 0644)
    if err != nil {
        ReportError("Error writing example Boxfile", err, true)
    }

    makefileExample := ""
    operatingSystem := runtime.GOOS
    switch operatingSystem {
    case "darwin":
        makefileExample = ".PHONY: clean\n.PHONY: build\n.PHONY: run\n.PHONY: copy\n\nSDK = $(shell egrep '^\\s*SDKRoot' ~/.Playdate/config | head -n 1 | cut -c9-)\nSDKBIN=$(SDK)/bin\nGAME=$(notdir $(CURDIR))\nSIM=Playdate Simulator\n\nbuild: clean compile run\n\nrun: open\n\nclean:\n\trm -rf 'build/$(GAME).pdx'\n\ncompile:\n\tmkdir build ; \"$(SDKBIN)/pdc\" 'source' './build/$(GAME).pdx'\n\nopen: compile\n\topen -a '$(SDKBIN)/$(SIM).app/Contents/MacOS/$(SIM)' './build/$(GAME).pdx'"
    case "linux":
    	// TODO: Make sure this works...
        makefileExample = ".PHONY: clean\n.PHONY: build\n.PHONY: run\n.PHONY: copy\n\nSDK = $(shell egrep '^\\s*SDKRoot' ~/.Playdate/config | head -n 1 | cut -c9-)\nSDKBIN=$(SDK)/bin\nGAME=$(notdir $(CURDIR))\nSIM=Playdate Simulator\n\nbuild: clean compile run\n\nrun: open\n\nclean:\n\trm -rf 'build/$(GAME).pdx'\n\ncompile:\n\tmkdir build ; \"$(SDKBIN)/pdc\" 'source' './build/$(GAME).pdx'\n\nopen: compile\n\topen -a '$(SDKBIN)/$(SIM)' './build/$(GAME).pdx'"
    }

    if makefileExample != "" {
		Print("Adding Makefile")
	
		err = os.WriteFile(filepath.Join(destination, "Makefile"), []byte(makefileExample), 0644)
    	if err != nil {
        	ReportError("Error writing example Makefile", err, true)
    	}
    } else {
    	Print("Skipping Makefile on Windows")
    }

	Print("Adding README.md")
	readMe := fmt.Sprintf("# %s\n## A game by %s\n\nThis a fun game that you will enjoy very much.\n\n### Building the game\n\nTo build the game, use `make`.  Simply run `make build` to build it and `make run` to run it in the Simulator.", gameName, currentUser.Name)
	
	err = os.WriteFile(filepath.Join(destination, "README.md"), []byte(readMe), 0644)
    if err != nil {
        ReportError("Error writing example README.md", err, true)
    }

	ReportDone()
}