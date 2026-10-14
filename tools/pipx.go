package tools

import (
	"context"
	"encoding/json"
)

var Pipx = &BasicTool{
	ToolName:          "pipx",
	ToolCmd:           "pipx",
	ToolDescription:   "Install and run Python applications in isolated environments",
	UpdateArgs:        []string{"upgrade-all"},
	UpdatePackageArgs: []string{"upgrade"},
	VersionArgs:       []string{"--version"},
	ListOutdated:      listPipxOutdated,
}

func listPipxOutdated(ctx context.Context) ([]Package, error) {
	output, err := ExecContext(ctx, "pipx", "list", "--outdated", "--json").Output()
	if err != nil {
		return nil, err
	}
	return parsePipxOutdated(string(output))
}

// parsePipxOutdated parses `pipx list --outdated --json`, skipping pinned packages, which `pipx upgrade-all` leaves alone.
func parsePipxOutdated(output string) ([]Package, error) {
	var outdated struct {
		Data struct {
			Packages []struct {
				Package       string `json:"package"`
				Version       string `json:"version"`
				LatestVersion string `json:"latest_version"`
				Pinned        bool   `json:"pinned"`
			} `json:"packages"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(output), &outdated); err != nil {
		return nil, err
	}

	var packages []Package
	for _, entry := range outdated.Data.Packages {
		if entry.Pinned {
			continue
		}
		packages = append(packages, Package{Name: entry.Package, Current: entry.Version, Latest: entry.LatestVersion})
	}
	return packages, nil
}
