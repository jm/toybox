package main

import (
	"fmt"
	"os"
	"github.com/fatih/color"
	"encoding/json"
	"path/filepath"
	"net/http"
	"io"
	"io/ioutil"
	"regexp"
	"golang.org/x/exp/slices"
	"golang.org/x/term"
	version "github.com/hashicorp/go-version"
)

type BasicAuthRoundTripper struct {
	Username string
	Password string
	
	RoundTripper http.RoundTripper
}

func (rt *BasicAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(rt.Username, rt.Password)
	return rt.RoundTripper.RoundTrip(req)
}

type Toybox struct {
	Name string
	PossibleVersions []string
	CurrentlySelectedVersion string
	Dependencies []*DependencyRelationship
	Dependents []*DependencyRelationship
}

func (tb *Toybox) FetchPossibleVersions() {
	c := http.Client{
		Transport: &BasicAuthRoundTripper{Username: "", Password: "", RoundTripper: http.DefaultTransport},
	}
	resp, err := c.Get(fmt.Sprintf("https://api.github.com/repos/%s/tags", tb.Name))

	if (err != nil) || (resp.StatusCode != 200) {
        fmt.Println("Error fetching versions", resp.StatusCode, fmt.Sprintf("https://api.github.com/repos/%s/tags", tb.Name))
        os.Exit(1)
	} else {
		defer resp.Body.Close()
    	body, err := ioutil.ReadAll(resp.Body) // response body is []byte

    	if err != nil {
        	fmt.Println(err)
        	os.Exit(1)
		}

    	var result []map[string]interface{}
    	if err = json.Unmarshal(body, &result); err != nil {
    		fmt.Println(string(body))
        	fmt.Println(err)
        	os.Exit(1)
    	}

    	tb.PossibleVersions = []string{}
    	for i := range result {
    		tb.PossibleVersions = append(tb.PossibleVersions, result[i]["name"].(string))
    	}

    	tb.PossibleVersions = append(tb.PossibleVersions, "main")
	}
}

type DependencyRelationship struct {
	From *Toybox
	To *Toybox
	Requirement string
}

func (tb *Toybox) AddDependency(newDependency *Toybox, requirement string) {
	// fmt.Println(tb.Name, "depends on", newDependency.Name)
	for dependencyIndex := range tb.Dependencies {
		if tb.Dependencies[dependencyIndex].From.Name == newDependency.Name {
			return
		}
	}

	relationship := &DependencyRelationship{tb, newDependency, requirement}
	tb.Dependencies = append(tb.Dependencies, relationship)
	newDependency.Dependents = append(newDependency.Dependents, relationship)
	
	version := newDependency.ResolveBestVersion()
	if version != newDependency.CurrentlySelectedVersion {
		newDependency.CurrentlySelectedVersion = version
		fmt.Println("new version", version)
		newDependency.ClearDependencies()

		fetchedBoxfile := newDependency.FetchBoxfile()
    	if (fetchedBoxfile != "") {
    		boxfile.parseAndLoadRequirements(newDependency, fetchedBoxfile)
    	}
	}
}

func (tb *Toybox) ClearDependencies() {
	for dependencyIndex := range tb.Dependencies {
		relationship := tb.Dependencies[dependencyIndex]
		dependency := boxfile.Toyboxes[relationship.From.Name]

		for dependentIndex := range dependency.Dependents {
			if dependency.Dependents[dependentIndex].From.Name == tb.Name {
				slices.Delete(dependency.Dependents, dependentIndex, dependentIndex + 1)
			}
		}
	}

	tb.Dependencies = []*DependencyRelationship{}
}

func (tb *Toybox) ResolveBestVersion() string {
	var bestVersion string

	hasExistingPin := false
	for dependencyIndex := range tb.Dependents {
		match, _ := regexp.MatchString("\\.", tb.Dependents[dependencyIndex].Requirement)
		if !match && hasExistingPin {
			fmt.Println("failed to resolve")
			os.Exit(1)
		} else if !match {
			hasExistingPin = true

			if slices.Contains(tb.PossibleVersions, tb.Dependents[dependencyIndex].Requirement) {
				bestVersion = tb.Dependents[dependencyIndex].Requirement
			}
		}
	}

	if !hasExistingPin {
		fmt.Println("----", tb.Name, "possible versions", tb.PossibleVersions)
		fmt.Println("---- dependents", tb.Dependents)
		for pvIndex := range tb.PossibleVersions {
			if bestVersion != "" {
				break
			}

			for dependencyIndex := range tb.Dependents {
				fmt.Println("++++", tb.PossibleVersions[pvIndex])
				match, _ := regexp.MatchString("\\.", tb.PossibleVersions[pvIndex])
				if !match {
					continue
				}

				v, _ := version.NewVersion(tb.PossibleVersions[pvIndex])
				constraints, _ := version.NewConstraint(tb.Dependents[dependencyIndex].Requirement)

				if constraints.Check(v) {
					fmt.Println(tb.Name, v, "satisfies constraints", constraints, "from", tb.Dependents[dependencyIndex].From.Name)
					bestVersion = tb.PossibleVersions[pvIndex]
				} else {
					fmt.Println(tb.Name, v, "DOES NOT satisfy constraints", constraints)
					bestVersion = ""
				}
			}
		}
	} else if bestVersion == "" {
		fmt.Println("failed to resolve due to existing pin")
		os.Exit(1)
	}

	return bestVersion
}

func (tb *Toybox) Fetch() {
	resp, err := http.Get(fmt.Sprintf("https://github.com/%s/zipball/%s", tb.Name, tb.CurrentlySelectedVersion))

	if err != nil {
        fmt.Println(err)
        os.Exit(1)
	}
	defer resp.Body.Close()

	out, err := os.Create(fmt.Sprintf("%s.zip", tb.CurrentlySelectedVersion))
	if err != nil {
		fmt.Println(err)
        os.Exit(1)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
}

func (tb *Toybox) FetchBoxfile() string {
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/Boxfile", tb.Name, tb.CurrentlySelectedVersion)
		fmt.Println("fetching", url)
	resp, err := http.Get(url)

	if err != nil {
		fmt.Println("Error fetching Boxfile:")
        fmt.Println(err)
        os.Exit(1)
	}

	if resp.StatusCode == 404 {
		return ""
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
    return string(body)
}

type Boxfile struct {
	Toyboxes map[string]*Toybox
	Root *Toybox
}

var boxfile *Boxfile

func (bf *Boxfile) Load(path string) {
	jsonString, err := os.ReadFile(path)
    
    if err != nil {
		fmt.Println("Error loading Boxfile:")
        fmt.Println(err)
        os.Exit(1)
    }

   	bf.parseAndLoadRequirements(bf.Root, string(jsonString))
}

func (bf *Boxfile) parseAndLoadRequirements(root *Toybox, jsonString string) {
    parsedData := map[string]string{}
    err := json.Unmarshal([]byte(jsonString), &parsedData)

    if err != nil {
		fmt.Println("Error parsing and loading :", jsonString)
        fmt.Println(err)
        os.Exit(1)
    }

    bf.Resolve(root, parsedData)
}

func (bf *Boxfile) Resolve(root *Toybox, requirementsMap map[string]string) {
	for dependency, requiredVersion := range requirementsMap {

		var candidate *Toybox

		if bf.Toyboxes[dependency] != nil {
			root.AddDependency(bf.Toyboxes[dependency], requiredVersion)
		} else {
			newCandidate := Toybox{dependency, []string{}, "", []*DependencyRelationship{}, []*DependencyRelationship{}}
			newCandidate.FetchPossibleVersions()

			bf.Toyboxes[dependency] = &newCandidate
			candidate = &newCandidate

			root.AddDependency(candidate, requiredVersion)
		}
    }
}

type Installer struct {

}

func (i *Installer) Install() {
	i.assertBoxfile()

    root := Toybox{"root", []string{"master"}, "master", []*DependencyRelationship{}, []*DependencyRelationship{}}
	boxfile = &Boxfile{make(map[string]*Toybox), &root }

	boxfile.Load("./Boxfile")
	fmt.Println("======= install")
	for tbi := range boxfile.Toyboxes {	
		fmt.Println(boxfile.Toyboxes[tbi].Name,boxfile.Toyboxes[tbi].CurrentlySelectedVersion)
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

func main() {
	if len(os.Args) < 2 {
		fmt.Println("expected subcommand")
		os.Exit(1)
	}

	color.Cyan("TOYBOX")

	passwd, _ := term.ReadPassword(0)
	fmt.Println(passwd)

	switch os.Args[1] {
	case "install":
		fmt.Println("installing")
		installer := Installer{}
		installer.Install()
	case "add":
		// barCmd.Parse(os.Args[2:])
		fmt.Println("subcommand 'bar'")
	case "remove":
		fmt.Println("subcommand 'bar'")
	case "show":
		fmt.Println("subcommand 'bar'")
	case "info":
		fmt.Println("subcommand 'bar'")
	default:
		fmt.Println("expected subcommand")
		os.Exit(1)
	}
}
