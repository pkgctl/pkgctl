package tools

var Npm = &BasicTool{
	ToolName:        "NPM",
	ToolCmd:         "npm",
	ToolDescription: "the package manager for JavaScript",
	UpdateArgs:      []string{"update --quiet --global"},
	VersionArgs:     []string{"--version"},
}
