package tools

var Fisher = &BasicTool{
	ToolID:          "fisher",
	ToolName:        "Fisher",
	ToolCmd:         "fisher",
	ToolDescription: "A plugin manager for Fish",
	VersionArgs:     []string{"--version"},
}
