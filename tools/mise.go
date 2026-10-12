package tools

var Mise = &BasicTool{
	ToolName:        "mise",
	ToolCmd:         "mise",
	UpdateArgs:      []string{"upgrade"},
	VersionArgs:     []string{"version"},
	ToolDescription: "The front-end to your dev env",
}
