package tools

import (
	"context"
	"encoding/json"
	"sort"
)

var Npm = &BasicTool{
	ToolName:          "npm",
	ToolCmd:           "npm",
	ToolDescription:   "The package manager for JavaScript",
	UpdateArgs:        []string{"update", "--global"},
	UpdatePackageArgs: []string{"update", "--global"},
	VersionArgs:       []string{"--version"},
	ListOutdated:      listNpmOutdated,
}

func listNpmOutdated(ctx context.Context) ([]Package, error) {
	// npm outdated exits 1 when a package is outdated.
	output, err := outputAllowingExit(ExecContext(ctx, "npm", "outdated", "--global", "--json"), 1)
	if err != nil {
		return nil, err
	}
	return parseNodeOutdated(string(output))
}

// parseNodeOutdated parses `npm outdated --json` or `pnpm outdated --format json`, an object keyed by package name.
func parseNodeOutdated(output string) ([]Package, error) {
	var outdated map[string]struct {
		Current string `json:"current"`
		Latest  string `json:"latest"`
	}
	if err := json.Unmarshal([]byte(output), &outdated); err != nil {
		return nil, err
	}

	var packages []Package
	for name, entry := range outdated {
		packages = append(packages, Package{Name: name, Current: entry.Current, Latest: entry.Latest})
	}
	sort.Slice(packages, func(i, j int) bool {
		return packages[i].Name < packages[j].Name
	})
	return packages, nil
}
