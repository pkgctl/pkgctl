package tools

import "os/exec"

// var Pip = &BasicTool{
// 	ToolName:        "pip",
// 	ToolCmd:         "pip3",
// 	ToolDescription: "The Rust toolchain installer",
// 	UpdateArgs:      nil,
// 	VersionArgs:     []string{"--version"},
// }

type PipTool struct {
	*BasicTool
}

var Pip = &PipTool{
	&BasicTool{
		ToolID:          "pip",
		ToolName:        "pip",
		ToolCmd:         "pip3",
		ToolDescription: "The Rust toolchain installer",
		VersionArgs:     []string{"--version"},
	},
}

func (b *PipTool) Update() *exec.Cmd {
	// Exec("pip3", "install", "--user", "--upgrade", "pip", "wheel", "setuptools", "build")
	return nil
}

// var fPip = &struct {
// 	ToolName string
// }{
// 	ToolName: "pip",
// }

// func (b *fPip) Name() string {
// 	return b.ToolName
// }

// python3 -m pip install --user --upgrade pip wheel setuptools build
// python3 -m pip install --user -r <(python3 -m pip list --user --outdated --format=json | jq '.[].name' ) --upgrade --quiet
