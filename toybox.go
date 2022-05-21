package main

import(
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"

    "golang.org/x/exp/slices"
   	version "github.com/hashicorp/go-version"
)

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

	out, err := ioutil.TempFile(os.TempDir(), "tb")
	if err != nil {
		fmt.Println("Error creating tempfile:", err)
        os.Exit(1)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Println("Error writing to tempfile:", err)
        os.Exit(1)
	}

    return out.Name()
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
