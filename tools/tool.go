package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var toolList = []Tool{
	Asdf,
	Brew,
	Fisher,
	Mise,
	RustUp,
}

type Tool interface {
	Name() string
	Exists() bool
	Update(ctx context.Context) *exec.Cmd
	// UpdateCommand returns the command line Update runs, for telling the user, such as "rustup update".
	UpdateCommand() string
	// ReportsOutdated reports whether Outdated can list the tool's outdated packages.
	ReportsOutdated() bool
	// Outdated returns the installed packages that have a newer version available.
	// Call it only when ReportsOutdated is true.
	Outdated(ctx context.Context) ([]Package, error)
	Version() string
	Description() string
}

func List() []Tool {
	return toolList
}

// Package is an installed package that has a newer version available.
// Current and Latest are empty for a tool that doesn't track versions, such as Fisher.
type Package struct {
	Name    string
	Current string
	Latest  string
}

type BasicTool struct {
	// ToolName is the human readable name of the tool
	// eg. "Homebrew"
	ToolName string
	// ToolCmd is the command to run the tool
	// eg. "brew"
	ToolCmd string
	// Shell is the shell that defines ToolCmd, for a tool that is a shell function rather than an executable
	// eg. "fish"; empty runs ToolCmd through $SHELL
	Shell string
	// ToolDescription is a short description of the tool
	// eg. "The missing package manager for macOS (or Linux)"
	ToolDescription string
	// UpdateArgs are the arguments to pass to the tool to update
	// eg. []string{"upgrade"}
	UpdateArgs []string
	// VersionArgs are the arguments to pass to the tool to get the version
	// eg. []string{"--version"}
	VersionArgs []string
	// ListOutdated lists the installed packages that have a newer version available,
	// or is nil if the tool can't report them
	ListOutdated func(ctx context.Context) ([]Package, error)
}

func (b *BasicTool) Name() string {
	return b.ToolName
}

func (b *BasicTool) Description() string {
	return b.ToolDescription
}

func (b *BasicTool) Exists() bool {
	if _, err := exec.LookPath(b.ToolCmd); err == nil {
		return true
	}

	cmd := b.command(context.Background(), b.VersionArgs...)
	err := cmd.Run()
	return err == nil && cmd.ProcessState.ExitCode() == 0
}

func (b *BasicTool) Update(ctx context.Context) *exec.Cmd {
	if b.UpdateArgs == nil {
		return nil
	}
	return b.command(ctx, b.UpdateArgs...)
}

func (b *BasicTool) UpdateCommand() string {
	return strings.Join(append([]string{b.ToolCmd}, b.UpdateArgs...), " ")
}

func (b *BasicTool) ReportsOutdated() bool {
	return b.ListOutdated != nil
}

func (b *BasicTool) Outdated(ctx context.Context) ([]Package, error) {
	return b.ListOutdated(ctx)
}

func (b *BasicTool) Version() string {
	cmd := b.command(context.Background(), b.VersionArgs...)
	out, _ := cmd.Output()
	return string(out)
}

func (b *BasicTool) command(ctx context.Context, arg ...string) *exec.Cmd {
	if b.Shell == "" {
		return ExecContext(ctx, b.ToolCmd, arg...)
	}
	return exec.CommandContext(ctx, b.Shell, "-c", strings.Join(append([]string{b.ToolCmd}, arg...), " "))
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
