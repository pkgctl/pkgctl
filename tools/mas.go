package tools

import (
	"context"
	"encoding/json"
	"strings"
)

var Mas = &BasicTool{
	ToolName:          "mas",
	ToolCmd:           "mas",
	ToolDescription:   "Mac App Store command-line interface",
	UpdateArgs:        []string{"update"},
	UpdatePackageArgs: []string{"update"},
	VersionArgs:       []string{"version"},
	ListOutdated:      listMasOutdated,
	UpdateNeedsSudo:   true,
}

func listMasOutdated(ctx context.Context) ([]Package, error) {
	output, err := ExecContext(ctx, "mas", "outdated", "--json").Output()
	if err != nil {
		return nil, err
	}
	return parseMasOutdated(string(output))
}

// parseMasOutdated parses `mas outdated --json`, a stream of JSON objects, one per app.
// It names each app by its bundle ID, which `mas update` takes, with the app's own name to display.
func parseMasOutdated(output string) ([]Package, error) {
	var packages []Package
	decoder := json.NewDecoder(strings.NewReader(output))
	for decoder.More() {
		var app struct {
			BundleID   string `json:"bundleID"`
			Name       string `json:"name"`
			Version    string `json:"version"`
			NewVersion string `json:"newVersion"`
		}
		if err := decoder.Decode(&app); err != nil {
			return nil, err
		}
		packages = append(packages, Package{Name: app.BundleID, DisplayName: app.Name, Current: app.Version, Latest: app.NewVersion})
	}
	return packages, nil
}
