package tools

import "regexp"

var Brew = &BasicTool{
	ToolID:             "brew",
	ToolName:           "Homebrew",
	ToolCmd:            "brew",
	ToolDescription:    "The missing package manager for macOS (or Linux)",
	UpdateArgs:         []string{"upgrade"},
	VersionArgs:        []string{"--version"},
	ParseUpdateLogFunc: parseBrewLogForUpdates,
}

// EXAMPLE LOG OUTPUT
// ==> Upgrading python-setuptools
//   69.0.3 -> 69.1.0

func parseBrewLogForUpdates(logOutput string) ([]Update, error) {
	re := regexp.MustCompile(`(?m)^==>\s+Upgrading\s+(\S+)\s+(\S+)\s+->\s+(\S+)\s?`)

	matches := re.FindAllStringSubmatch(logOutput, -1)

	updates := []Update{}

	for _, match := range matches {
		updates = append(updates, Update{
			Name:        match[1],
			FromVersion: match[2],
			ToVersion:   match[3],
		})
	}
	return updates, nil
}
