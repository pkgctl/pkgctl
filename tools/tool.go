package tools

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/pkgctl/pkgctl/logs"
)

var toolList = []Tool{
	// Asdf,
	Brew,
	// Fisher,
	// Gem,
	// Npm,
	// Pip,
	RustUp,
}

type Tool interface {
	ID() string
	Name() string
	Cmd() string
	Exits() bool
	Update(ctx context.Context) *exec.Cmd
	Version() string
	Description() string

	ParseForUpdates(logs.LogFile) ([]Update, error)
}

func GetTool(id string) Tool {
	for _, tool := range toolList {
		if tool.ID() == id {
			return tool
		}
	}
	return nil
}

func List() []Tool {
	return toolList
}

type Update struct {
	Name        string
	FromVersion string
	ToVersion   string
}

type BasicTool struct {
	// ToolID is the unique identifier for the tool
	// eg. "brew"
	ToolID string
	// ToolName is the human readable name of the tool
	// eg. "Homebrew"
	ToolName string
	// ToolCmd is the command to run the tool
	// eg. "brew"
	ToolCmd string
	// ToolDescription is a short description of the tool
	// eg. "The missing package manager for macOS (or Linux)"
	ToolDescription string
	// UpdateArgs are the arguments to pass to the tool to update
	// eg. []string{"upgrade"}
	UpdateArgs []string
	// VersionArgs are the arguments to pass to the tool to get the version
	// eg. []string{"--version"}
	VersionArgs []string
	// UseShellCommand is a flag to indicate if the tool should be run as a shell command
	ShellCommand bool

	ParseUpdateLogFunc func(string) ([]Update, error)
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
	if _, err := exec.LookPath(b.ToolCmd); err == nil {
		return true
	}

	cmd := Exec(b.ToolCmd, b.VersionArgs...)
	err := cmd.Run()
	return err == nil && cmd.ProcessState.ExitCode() == 0
}

func (b *BasicTool) Update(ctx context.Context) *exec.Cmd {
	if b.UpdateArgs == nil {
		return nil
	}
	return ExecContext(ctx, b.ToolCmd, b.UpdateArgs...)
}

func (b *BasicTool) Version() string {
	cmd := Exec(b.ToolCmd, b.VersionArgs...)
	out, _ := cmd.Output()
	return string(out)
}

func Exec(name string, arg ...string) *exec.Cmd {
	return ExecContext(context.Background(), name, arg...)
}

func ExecContext(ctx context.Context, name string, arg ...string) *exec.Cmd {
	shell := os.Getenv("SHELL")

	if shell == "" {
		return exec.CommandContext(ctx, name, arg...)
	} else {
		shellArgs := fmt.Sprintf("%s %s", name, strings.Join(arg, " "))
		return exec.CommandContext(ctx, shell, "-c", shellArgs)
	}
}

func (b *BasicTool) ParseForUpdates(l logs.LogFile) ([]Update, error) {

	if b.ParseUpdateLogFunc == nil {
		return nil, nil
	}

	file, err := l.Open()

	if err != nil {
		return []Update{}, err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)

	if err != nil {
		return []Update{}, err
	}
	defer gzReader.Close()

	bytes, err := io.ReadAll(gzReader)
	if err != nil {
		return []Update{}, err
	}
	output := string(bytes)

	return b.ParseUpdateLogFunc(output)
}
