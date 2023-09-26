package tools

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var toolList = []CliTool{
	Asdf,
	Brew,
	Fisher,
	Npm,
	Pip,
	RustUp,
}

type CliTool interface {
	ID() string
	Name() string
	Cmd() string
	Exits() bool
	Update() *exec.Cmd
	Version() string
	Description() string
}

func List() []CliTool {
	return toolList
}

type BasicTool struct {
	ToolID          string
	ToolName        string
	ToolCmd         string
	ToolDescription string
	UpdateArgs      []string
	VersionArgs     []string
}

func (b *BasicTool) ID() string {
	return b.ToolID
}

func (b *BasicTool) Name() string {
	return b.ToolName
}

func (b *BasicTool) Cmd() string {
	return b.ToolCmd
}

func (b *BasicTool) Description() string {
	return b.ToolDescription
}

func (b *BasicTool) Exits() bool {
	_, err := exec.LookPath(b.ToolCmd)
	return err == nil
}

func (b *BasicTool) Update() *exec.Cmd {
	if b.UpdateArgs == nil {
		return nil
	}
	return Exec(b.ToolCmd, b.UpdateArgs...)
}

func (b *BasicTool) Version() string {
	cmd := Exec(b.ToolCmd, b.VersionArgs...)
	out, _ := cmd.Output()
	return string(out)
}

func Exec(name string, arg ...string) *exec.Cmd {
	shell := os.Getenv("SHELL")

	if shell == "" {
		return exec.Command(name, arg...)
	} else {
		shellArgs := fmt.Sprintf("%s %s", name, strings.Join(arg, " "))
		return exec.Command(shell, "-c", shellArgs)
	}
}
