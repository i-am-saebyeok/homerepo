package dogfood

import (
	"embed"
	"fmt"
	"sort"
	"strings"
)

const NotesFile = "dogfood-notes.md"

//go:embed scenarios/*.md
var scenarioFiles embed.FS

type Scenario struct {
	Name       string
	Repository string
}

var scenarios = []Scenario{
	{Name: "new-repository", Repository: "notes-app"},
}

func LookupScenario(name string) (Scenario, error) {
	for _, scenario := range scenarios {
		if scenario.Name == name {
			return scenario, nil
		}
	}
	return Scenario{}, fmt.Errorf("unknown scenario %q; available: %s", name, scenarioNames())
}

func scenarioNames() string {
	names := make([]string, 0, len(scenarios))
	for _, scenario := range scenarios {
		names = append(names, scenario.Name)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

func (s Scenario) Instructions(layout Layout) (string, error) {
	template, err := scenarioFiles.ReadFile("scenarios/" + s.Name + ".md")
	if err != nil {
		return "", err
	}
	filled := strings.NewReplacer(
		"{{repository}}", s.Repository,
		"{{remote}}", layout.Remote(s.Repository),
		"{{notes}}", NotesFile,
	).Replace(string(template))
	return strings.TrimSpace(filled), nil
}
