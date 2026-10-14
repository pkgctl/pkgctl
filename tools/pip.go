package tools

import (
	"context"
	"encoding/json"
)

var Pip = &Tool{
	Name:              "pip",
	Command:           "pip",
	Description:       "The package installer for Python",
	VersionArgs:       []string{"--version"},
	UpdatePackageArgs: []string{"install", "--upgrade"},
	ListOutdated:      listPipOutdated,
}

// listPipOutdated lists the outdated packages that no other package requires, the ones installed on purpose,
// since upgrading a dependency past what its dependents require can break them.
func listPipOutdated(ctx context.Context) ([]Package, error) {
	output, err := ExecContext(ctx, "pip", "list", "--outdated", "--not-required", "--format=json").Output()
	if err != nil {
		return nil, err
	}
	return parsePipOutdated(string(output))
}

// parsePipOutdated parses `pip list --outdated --format=json`.
func parsePipOutdated(output string) ([]Package, error) {
	var outdated []struct {
		Name          string `json:"name"`
		Version       string `json:"version"`
		LatestVersion string `json:"latest_version"`
	}
	if err := json.Unmarshal([]byte(output), &outdated); err != nil {
		return nil, err
	}

	var packages []Package
	for _, entry := range outdated {
		packages = append(packages, Package{Name: entry.Name, Current: entry.Version, Latest: entry.LatestVersion})
	}
	return packages, nil
}
