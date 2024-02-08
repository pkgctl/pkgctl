package tools

var Gem = &BasicTool{
	ToolID:          "gem",
	ToolName:        "RubyGems",
	ToolCmd:         "gem",
	ToolDescription: "RubyGems is a sophisticated package manager for the Ruby programming language.",
	UpdateArgs:      []string{"update"},
	VersionArgs:     []string{"--version"},
}
