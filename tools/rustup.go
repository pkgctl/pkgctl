package tools

var RustUp = &BasicTool{
	ToolName:        "rustup",
	ToolCmd:         "rustup",
	ToolDescription: "The Rust toolchain installer",
	UpdateArgs:      []string{"update"},
	VersionArgs:     []string{"--version"},
}
