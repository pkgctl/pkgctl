package tools

var Brew = &BasicTool{
	ToolName:        "Homebrew",
	ToolCmd:         "brew",
	ToolDescription: "The missing package manager for macOS (or Linux)",
	UpdateArgs:      []string{"update"},
	VersionArgs:     []string{"--version"},
}
