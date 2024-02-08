package tools

var Npm = &BasicTool{
	ToolID:          "npm",
	ToolName:        "NPM",
	ToolCmd:         "npm",
	ToolDescription: "the package manager for JavaScript",
	UpdateArgs:      []string{"update --global --verbose"},
	VersionArgs:     []string{"--version"},
}
