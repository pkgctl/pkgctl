package tools

import (
	"context"
	"encoding/json"
	"os"
	"strings"
)

var Brew = &BasicTool{
	ToolName:        "Homebrew",
	ToolCmd:         "brew",
	ToolDescription: "The missing package manager for macOS (or Linux)",
	UpdateArgs:      []string{"upgrade", "--yes"},
	VersionArgs:     []string{"--version"},
	ListOutdated:    listBrewOutdated,
}

func listBrewOutdated(ctx context.Context) ([]Package, error) {
	// `brew outdated` can auto-update first and print the update report ahead of its JSON,
	// so update separately and keep that output out of the JSON. Output, unlike Run, keeps stderr in a failure's error.
	if _, err := ExecContext(ctx, "brew", "update", "--quiet").Output(); err != nil {
		return nil, err
	}

	outdatedCmd := ExecContext(ctx, "brew", "outdated", "--json=v2")
	outdatedCmd.Env = append(os.Environ(), "HOMEBREW_NO_AUTO_UPDATE=1")
	output, err := outdatedCmd.Output()
	if err != nil {
		return nil, err
	}
	return parseBrewOutdated(string(output))
}

type brewOutdated struct {
	Name              string   `json:"name"`
	InstalledVersions []string `json:"installed_versions"`
	CurrentVersion    string   `json:"current_version"`
	Pinned            bool     `json:"pinned"`
}

// parseBrewOutdated parses `brew outdated --json=v2`, skipping pinned formulae, which `brew upgrade` leaves alone.
func parseBrewOutdated(output string) ([]Package, error) {
	var outdated struct {
		Formulae []brewOutdated `json:"formulae"`
		Casks    []brewOutdated `json:"casks"`
	}
	if err := json.Unmarshal([]byte(output), &outdated); err != nil {
		return nil, err
	}

	var packages []Package
	for _, entry := range append(outdated.Formulae, outdated.Casks...) {
		if entry.Pinned {
			continue
		}
		packages = append(packages, Package{
			Name:    entry.Name,
			Current: strings.Join(entry.InstalledVersions, ", "),
			Latest:  entry.CurrentVersion,
		})
	}
	return packages, nil
}
