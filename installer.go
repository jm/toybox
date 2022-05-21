package main

import(
    "archive/zip"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"

    "golang.org/x/exp/slices"
)

type Installer struct {

}

func (i *Installer) Install() {
	i.assertBoxfile()

    root := Toybox{"root", []string{"default"}, "default", []*DependencyRelationship{}, []*DependencyRelationship{}, "default"}
	boxfile = &Boxfile{make(map[string]*Toybox), &root }

	boxfile.Load("./Boxfile")
	fmt.Println("======= install")
	
	importList := []string{}
	installList := boxfile.Sort()
	for tbi := range installList {
		toybox := installList[tbi]
		
		zipFilePath := toybox.Fetch()
		i.Extract(zipFilePath, toybox.Name)

		mainFile := i.GenerateImportLine(toybox.Name)
		importList = append(importList, mainFile)
	}

	i.WriteImportFile(importList)
}

func (i *Installer) WriteImportFile(importList []string) {
    err := os.WriteFile("source/toyboxes.lua", []byte(strings.Join(importList, "\n")), 0644)
    if err != nil {
    	fmt.Println("Error writing import file:", err)
        os.Exit(1)
    }
}

func (i *Installer) GenerateImportLine(toyboxName string) string {
	result := ""

	possibilities := []string{
		fmt.Sprintf("source/libraries/%s/source/import.lua", toyboxName),
		fmt.Sprintf("source/libraries/%s/source/main.lua", toyboxName),
		fmt.Sprintf("source/libraries/%s/import.lua", toyboxName),
		fmt.Sprintf("source/libraries/%s/main.lua", toyboxName),
	}

	for possibility := range possibilities {
		fmt.Println("**** locating", possibilities[possibility])
		if _, err := os.Stat(possibilities[possibility]); err == nil {
  			result = fmt.Sprintf("import(\"%s\")", strings.TrimPrefix(strings.TrimSuffix(possibilities[possibility], ".lua"), "source/"))
  			break
		} else {
			result = fmt.Sprintf("-- Couldn't locate the right import for %s", toyboxName)
		}
	}

	return result
}

func (i *Installer) Extract(zipFilePath string, toyboxName string) {
	destination := fmt.Sprintf("source/libraries/%s", toyboxName)

    archive, err := zip.OpenReader(zipFilePath)
    if err != nil {
    	fmt.Println("Error unzipping (", zipFilePath, "):", err)
        os.Exit(1)
    }
    defer archive.Close()

    for _, f := range archive.File {
    	relativeFilePathParts := strings.Split(f.Name, "/")
    	relativeFilePathParts = slices.Delete(relativeFilePathParts, 0, 1)
    	relativeFilePath := strings.Join(relativeFilePathParts, "/")

        filePath := filepath.Join(destination, relativeFilePath)
        fmt.Println("unzipping file ", filePath)

        if f.FileInfo().IsDir() {
            fmt.Println("creating directory...")
            os.MkdirAll(filePath, os.ModePerm)
            continue
        }

        if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
            panic(err)
        }

        dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
        if err != nil {
            panic(err)
        }

        fileInArchive, err := f.Open()
        if err != nil {
            panic(err)
        }

        if _, err := io.Copy(dstFile, fileInArchive); err != nil {
            panic(err)
        }

        dstFile.Close()
        fileInArchive.Close()
    }
}

func (i *Installer) assertBoxfile() {
	matches, err := filepath.Glob("./Boxfile")

    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    if len(matches) == 0 {
        fmt.Println("Boxfile not found")
        os.Exit(1)
    }
}