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
   	"github.com/schollz/progressbar"
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
	c := HttpClient()
	resp, err := c.Get(fmt.Sprintf("https://api.github.com/repos/%s/tags", tb.Name))

	if (err != nil) || (resp.StatusCode != 200) {
		ReportError(fmt.Sprintf("Error fetching versions (status %s)", resp.StatusCode), err, true)
	} else {
		defer resp.Body.Close()
    	body, err := ioutil.ReadAll(resp.Body)

    	if err != nil {
			ReportError("Error reading versions response", err, true)
		}

    	var result []map[string]interface{}
    	if err = json.Unmarshal(body, &result); err != nil {
    		ReportError("Error parsing versions response", err, true)
    	}

    	tb.PossibleVersions = []string{}
    	for i := range result {
    		tb.PossibleVersions = append(tb.PossibleVersions, result[i]["name"].(string))
    	}

    	tb.PossibleVersions = append(tb.PossibleVersions, tb.DefaultRef)
	}
}

func (tb *Toybox) FetchDefaultRef() {
	c := HttpClient()
	resp, err := c.Get(fmt.Sprintf("https://api.github.com/repos/%s", tb.Name))

	if (err != nil) || (resp.StatusCode != 200) {
		ReportError(fmt.Sprintf("Error fetching default ref (status %d)", resp.StatusCode), err, true)
	} else {
		defer resp.Body.Close()
    	body, err := ioutil.ReadAll(resp.Body) // response body is []byte

    	if err != nil {
			ReportError("Error reading default ref response", err, true)
		}

    	var result map[string]interface{}
    	if err = json.Unmarshal(body, &result); err != nil {
    		ReportError("Error parsing default ref response", err, true)
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
			ReportError("Failed to resolve versions", nil, true)
		} else if !match {
			hasExistingPin = true

			if slices.Contains(tb.PossibleVersions, tb.Dependents[dependencyIndex].Requirement) {
				bestVersion = tb.Dependents[dependencyIndex].Requirement
			}
		}
	}

	if !hasExistingPin {
		for pvIndex := range tb.PossibleVersions {
			if bestVersion != "" {
				break
			}

			for dependencyIndex := range tb.Dependents {
				match, _ := regexp.MatchString("\\.", tb.PossibleVersions[pvIndex])
				if !match {
					continue
				}

				v, _ := version.NewVersion(tb.PossibleVersions[pvIndex])
				constraints, _ := version.NewConstraint(tb.Dependents[dependencyIndex].Requirement)

				if constraints.Check(v) {
					bestVersion = tb.PossibleVersions[pvIndex]
				} else {
					bestVersion = ""
				}
			}
		}
	} else if bestVersion == "" {
		ReportError("Failed to resolve versions", nil, true)
	}

	return bestVersion
}

func (tb *Toybox) Fetch() string {
	resp, err := http.Get(fmt.Sprintf("https://github.com/%s/zipball/%s", tb.Name, tb.CurrentlySelectedVersion))

	if err != nil {
        ReportError("Error fetching toybox file", err, true)
	}
	defer resp.Body.Close()

	out, err := ioutil.TempFile(os.TempDir(), "tb")
	if err != nil {
        ReportError("Error creating tempfile", err, true)
	}
	defer out.Close()

	bar := progressbar.DefaultBytes(
    	resp.ContentLength,
    	"       ",
	)

	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	if err != nil {
		ReportError("Error writing to tempfile", err, true)
	}

    return out.Name()
}

func (tb *Toybox) FetchBoxfile() string {
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/Boxfile", tb.Name, tb.CurrentlySelectedVersion)
	resp, err := http.Get(url)

	if err != nil {
		ReportError("Error fetching Boxfile", err, true)
	}

	if resp.StatusCode == 404 {
		return ""
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
    return string(body)
}
