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

    if Credential == nil {
        ReportInfo("It is highly recommended you log in to GitHub!")
        ReportInfo("Use `toybox login` to avoid GitHub API limits.")
    }

    ReportProgress("Loading Boxfile...")

    root := Toybox{"root", []string{"default"}, "default", []*DependencyRelationship{}, []*DependencyRelationship{}, "default"}
	boxfile = &Boxfile{make(map[string]*Toybox), &root }

	boxfile.Load("./Boxfile")
	
	ReportProgress("Installing")

	importList := []string{}
	installList := boxfile.Sort()
	for tbi := range installList {
		toybox := installList[tbi]
		
        _, err := os.Stat(fmt.Sprintf("source/libraries/%s", toybox.Name))

        if err != nil {
            Print(fmt.Sprintf("Fetching %s...", toybox.Name))
            zipFilePath := toybox.Fetch()
            Print("Extracting...")
            i.Extract(zipFilePath, toybox.Name)
        } else {
            Print(fmt.Sprintf("Using %s@%s", toybox.Name, toybox.CurrentlySelectedVersion))
        }

		mainFile := i.GenerateImportLine(toybox.Name)
		importList = append(importList, mainFile)
	}

    ReportProgress("Writing import file")
	i.WriteImportFile(importList)
}

func (i *Installer) WriteImportFile(importList []string) {
    err := os.WriteFile("source/toyboxes.lua", []byte(strings.Join(importList, "\n")), 0644)
    if err != nil {
    	ReportError("Error writing import file", err, true)
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
        ReportError(fmt.Sprintf("Error unzipping (%s)", zipFilePath), err, true)
    }
    defer archive.Close()

    for _, f := range archive.File {
    	relativeFilePathParts := strings.Split(f.Name, "/")
    	relativeFilePathParts = slices.Delete(relativeFilePathParts, 0, 1)
    	relativeFilePath := strings.Join(relativeFilePathParts, "/")

        filePath := filepath.Join(destination, relativeFilePath)

        if f.FileInfo().IsDir() {
            os.MkdirAll(filePath, os.ModePerm)
            continue
        }

        if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
            ReportError(fmt.Sprintf("Error creating path (%s)", filePath), err, true)
        }

        dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
        if err != nil {
            ReportError(fmt.Sprintf("Error extracting file (%s)", filePath), err, true)
        }

        fileInArchive, err := f.Open()
        if err != nil {
            ReportError(fmt.Sprintf("Error extracting file (%s)", fileInArchive), err, true)
        }

        if _, err := io.Copy(dstFile, fileInArchive); err != nil {
            ReportError(fmt.Sprintf("Error copying contents to file (%s)", dstFile), err, true)
        }

        dstFile.Close()
        fileInArchive.Close()
    }
}

func (i *Installer) assertBoxfile() {
	matches, err := filepath.Glob("./Boxfile")

    if err != nil {
        ReportError("Error locating Boxfile", err, true)
    }

    if len(matches) == 0 {
        ReportError("Boxfile not found", nil, true)
    }
}

func InstallDependencies() {
    Credential = LoadGitHubCredential()
    
    installer := Installer{}
    installer.Install()
    
    ReportDone()
}

func UpdateDependency(dependency string) {
    path := filepath.Join("source", "libraries", dependency)
    if _, err := os.Stat(path); !os.IsNotExist(err) {
        ReportInfo(fmt.Sprintf("Removing current version of %s...", dependency))
        os.RemoveAll(path)
    } else {
        ReportInfo("Dependency not installed, installing dependencies")
    }

    InstallDependencies()
}
