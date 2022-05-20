package main

import (
	"fmt"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"net/http"
	"regexp"
	"sort"
	"strings"
	
	"golang.org/x/exp/slices"
	"golang.org/x/exp/maps"
	"golang.org/x/term"
	
	"github.com/fatih/color"
	version "github.com/hashicorp/go-version"
)

// Borrowed this approach from Stack Overflow, but I can't
// find the post now...
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
	DefaultRef string
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

    	tb.PossibleVersions = append(tb.PossibleVersions, tb.DefaultRef)
	}
}

func (tb *Toybox) FetchDefaultRef() {
	c := http.Client{
		Transport: &BasicAuthRoundTripper{Username: "", Password: "", RoundTripper: http.DefaultTransport},
	}
	resp, err := c.Get(fmt.Sprintf("https://api.github.com/repos/%s", tb.Name))

	if (err != nil) || (resp.StatusCode != 200) {
        fmt.Println("Error fetching default", resp.StatusCode, tb.Name)
        os.Exit(1)
	} else {
		defer resp.Body.Close()
    	body, err := ioutil.ReadAll(resp.Body) // response body is []byte

    	if err != nil {
        	fmt.Println(err)
        	os.Exit(1)
		}

    	var result map[string]interface{}
    	if err = json.Unmarshal(body, &result); err != nil {
    		fmt.Println(string(body))
        	fmt.Println(err)
        	os.Exit(1)
    	}

    	tb.DefaultRef = result["default_branch"].(string)
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

func (tb *Toybox) Fetch() string {
	resp, err := http.Get(fmt.Sprintf("https://github.com/%s/zipball/%s", tb.Name, tb.CurrentlySelectedVersion))

	if err != nil {
        fmt.Println(err)
        os.Exit(1)
	}
	defer resp.Body.Close()

	out, err := ioutil.TempFile("toyboxes", "tb")
	if err != nil {
		fmt.Println(err)
        os.Exit(1)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	fullPath, err := filepath.Abs(filepath.Dir(out.Name()))

    if err != nil {
        panic(err)
    }

    return fullPath
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
			newCandidate := Toybox{dependency, []string{}, "", []*DependencyRelationship{}, []*DependencyRelationship{}, "default"}
			newCandidate.FetchDefaultRef()
			newCandidate.FetchPossibleVersions()

			bf.Toyboxes[dependency] = &newCandidate
			candidate = &newCandidate

			if (requiredVersion == "default") || (requiredVersion == "*") {
				requiredVersion = candidate.DefaultRef
			}

			root.AddDependency(candidate, requiredVersion)
		}
    }
}

func (bf *Boxfile) Sort() []*Toybox {
	sortedBoxes := maps.Values(bf.Toyboxes)
	sort.Slice(sortedBoxes, func(i, j int) bool {
  		return len(sortedBoxes[i].Dependents) < len(sortedBoxes[j].Dependents)
	})

	return sortedBoxes
}

type Installer struct {

}

func (i *Installer) Install() {
	i.assertBoxfile()

    root := Toybox{"root", []string{"default"}, "default", []*DependencyRelationship{}, []*DependencyRelationship{}, "default"}
	boxfile = &Boxfile{make(map[string]*Toybox), &root }

	boxfile.Load("./Boxfile")
	fmt.Println("======= install")
	installList := boxfile.Sort()
	for tbi := range installList {	
		fmt.Println(installList[tbi].Name,installList[tbi].CurrentlySelectedVersion)
		installList[tbi].Fetch()
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

	switch os.Args[1] {
	case "install":
		fmt.Println("installing")
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
	case "show":
		fmt.Println("subcommand 'bar'")
	case "info":
		fmt.Println("subcommand 'bar'")
	case "help":
		fmt.Println("halp plz")
	default:
		fmt.Println("expected subcommand")
		os.Exit(1)
	}
}
