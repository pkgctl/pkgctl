package tools

import (
	"context"
	"encoding/json"
	"sort"
)

var Mise = &BasicTool{
	ToolName:          "mise",
	ToolCmd:           "mise",
	UpdateArgs:        []string{"upgrade", "--global"},
	UpdatePackageArgs: []string{"upgrade", "--global"},
	VersionArgs:       []string{"version"},
	ToolDescription:   "The front-end to your dev env",
	ListOutdated:      listMiseOutdated,
}

func listMiseOutdated(ctx context.Context) ([]Package, error) {
	output, err := ExecContext(ctx, "mise", "outdated", "--global", "--json").Output()
	if err != nil {
		return nil, err
	}
	return parseMiseOutdated(string(output))
}

// parseMiseOutdated parses `mise outdated --json`, an object keyed by tool name.
func parseMiseOutdated(output string) ([]Package, error) {
	var outdated map[string]struct {
		Current string `json:"current"`
		Latest  string `json:"latest"`
	}
	if err := json.Unmarshal([]byte(output), &outdated); err != nil {
		return nil, err
	}

	var packages []Package
	for name, tool := range outdated {
		packages = append(packages, Package{Name: name, Current: tool.Current, Latest: tool.Latest})
	}
	sort.Slice(packages, func(i, j int) bool {
		return packages[i].Name < packages[j].Name
	})
	return packages, nil
}
