package tools

var Brew = &BasicTool{
	ToolID:          "brew",
	ToolName:        "Homebrew",
	ToolCmd:         "brew",
	ToolDescription: "The missing package manager for macOS (or Linux)",
	UpdateArgs:      []string{"upgrade"},
	VersionArgs:     []string{"--version"},
}
