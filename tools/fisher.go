package tools

var Fisher = &BasicTool{
	ToolName:        "Fisher",
	ToolCmd:         "fisher",
	ToolDescription: "A plugin manager for Fish",
	VersionArgs:     []string{"--version"},
}
