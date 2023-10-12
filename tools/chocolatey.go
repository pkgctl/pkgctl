package tools

var Chocolatey = &BasicTool{
	ToolID:          "chocolatey",
	ToolName:        "Chocolatey",
	ToolCmd:         "choco",
	ToolDescription: "The package manager for Windows",
	VersionArgs:     []string{"version"},
}
