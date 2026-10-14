package tools

import (
	"context"
	"regexp"
)

var Uv = &BasicTool{
	ToolName:          "uv",
	ToolCmd:           "uv",
	ToolDescription:   "An extremely fast Python package and project manager",
	UpdateArgs:        []string{"tool", "upgrade", "--all"},
	UpdatePackageArgs: []string{"tool", "upgrade"},
	VersionArgs:       []string{"--version"},
	ListOutdated:      listUvOutdated,
}

func listUvOutdated(ctx context.Context) ([]Package, error) {
	output, err := ExecContext(ctx, "uv", "tool", "list", "--outdated").Output()
	if err != nil {
		return nil, err
	}
	return parseUvOutdated(string(output)), nil
}

// A `uv tool list --outdated` line for a tool, such as "ruff v0.5.0 [latest: 0.17.0]".
var uvOutdatedLine = regexp.MustCompile(`(?m)^(\S+) v(\S+) \[latest: (\S+)\]$`)

func parseUvOutdated(output string) []Package {
	var packages []Package
	for _, match := range uvOutdatedLine.FindAllStringSubmatch(output, -1) {
		packages = append(packages, Package{Name: match[1], Current: match[2], Latest: match[3]})
	}
	return packages
}
