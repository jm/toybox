package main

import(
	"encoding/json"
	"golang.org/x/exp/maps"
	"os"
	"sort"
)

type Boxfile struct {
	Toyboxes map[string]*Toybox
	Root *Toybox
}

var boxfile *Boxfile

func (bf *Boxfile) Load(path string) {
	jsonString, err := os.ReadFile(path)
    
    if err != nil {
		ReportError("Error reading Boxfile", err, true)
    }

    Print("Resolving dependencies...")
   	bf.parseAndLoadRequirements(bf.Root, string(jsonString))
}

func (bf *Boxfile) parseAndLoadRequirements(root *Toybox, jsonString string) {
    parsedData := map[string]string{}
    err := json.Unmarshal([]byte(jsonString), &parsedData)

    if err != nil {
		ReportError("Error parsing Boxfile", err, true)
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