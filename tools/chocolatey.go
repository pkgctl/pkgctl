package tools

var Chocolatey = &BasicTool{
	ToolID:          "chocolatey",
	ToolName:        "Chocolatey",
	ToolCmd:         "choco",
	ToolDescription: "The package manager for Windows",
	UpdateArgs:      []string{"upgrade", "all", "--yes"},
	VersionArgs:     []string{"version"},
}
