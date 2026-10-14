package tools

import (
	"context"
	"encoding/json"
)

var DotnetTools = &BasicTool{
	ToolName:          ".NET tools",
	ToolCmd:           "dotnet",
	ToolDescription:   "Global tools for the .NET SDK",
	UpdateArgs:        []string{"tool", "update", "--all", "--global"},
	UpdatePackageArgs: []string{"tool", "update", "--global"},
	VersionArgs:       []string{"--version"},
	ListOutdated:      listDotnetOutdated,
}

// listDotnetOutdated asks the NuGet sources for each global tool's latest version,
// since dotnet can't list outdated tools.
func listDotnetOutdated(ctx context.Context) ([]Package, error) {
	output, err := ExecContext(ctx, "dotnet", "tool", "list", "--global", "--format", "json").Output()
	if err != nil {
		return nil, err
	}
	installed, err := parseDotnetTools(string(output))
	if err != nil {
		return nil, err
	}

	var packages []Package
	for _, tool := range installed {
		output, err := ExecContext(ctx, "dotnet", "package", "search", tool.Name, "--exact-match", "--format", "json").Output()
		if err != nil {
			return nil, err
		}
		latest, err := parseNugetLatest(string(output))
		if err != nil {
			return nil, err
		}
		if latest != tool.Current {
			packages = append(packages, Package{Name: tool.Name, Current: tool.Current, Latest: latest})
		}
	}
	return packages, nil
}

// parseDotnetTools parses `dotnet tool list --format json` into the installed tools, without a latest version.
func parseDotnetTools(output string) ([]Package, error) {
	var list struct {
		Data []struct {
			PackageID string `json:"packageId"`
			Version   string `json:"version"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(output), &list); err != nil {
		return nil, err
	}

	var tools []Package
	for _, tool := range list.Data {
		tools = append(tools, Package{Name: tool.PackageID, Current: tool.Version})
	}
	return tools, nil
}

// parseNugetLatest parses `dotnet package search --exact-match --format json`, which lists a package's stable versions
// oldest first, and returns the last.
func parseNugetLatest(output string) (string, error) {
	var search struct {
		SearchResult []struct {
			Packages []struct {
				Version string `json:"version"`
			} `json:"packages"`
		} `json:"searchResult"`
	}
	if err := json.Unmarshal([]byte(output), &search); err != nil {
		return "", err
	}

	latest := ""
	for _, source := range search.SearchResult {
		for _, version := range source.Packages {
			latest = version.Version
		}
	}
	return latest, nil
}
