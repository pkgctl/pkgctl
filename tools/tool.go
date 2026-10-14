package tools

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
)

// The tools that install runtimes come first, so the package managers after them update the runtimes' new installs.
var toolList = []Tool{
	Brew,
	Asdf,
	Mise,
	RustUp,
	DotnetTools,
	Fisher,
	Gem,
	GoBinaries,
	Mas,
	Npm,
	Pip,
	Pipx,
	Pnpm,
	Uv,
}

type Tool interface {
	Name() string
	// Command returns the tool's command, such as "brew", which names the tool for `pkgctl update`.
	Command() string
	Exists() bool
	// Update returns the commands that update packageName, as Outdated names it, or every package when it is empty,
	// to run in order. It returns none if there is nothing to update.
	Update(ctx context.Context, packageName string) ([]*exec.Cmd, error)
	// UpdateCommand returns the command to give the user for updating the tool's packages, such as "rustup update".
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

// Find returns the tool whose command is command, such as "brew".
func Find(command string) (Tool, error) {
	for _, tool := range toolList {
		if tool.Command() == command {
			return tool, nil
		}
	}
	return nil, fmt.Errorf("unknown tool %q; run `pkgctl list` to see them", command)
}

// Package is an installed package that has a newer version available.
// Name is what the tool's update command takes, such as a Go binary's package path.
// DisplayName is the package's name for people when Name isn't one, such as an app's name for its bundle ID, or empty.
// Current and Latest are empty for a tool that doesn't track versions, such as Fisher.
type Package struct {
	Name        string
	DisplayName string
	Current     string
	Latest      string
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
	// UpdateArgs are the arguments to pass to the tool to update, or nil to pass the outdated packages
	// to UpdatePackageArgs instead, for a tool whose command can't update them all, or fails with none installed
	// eg. []string{"upgrade"}
	UpdateArgs []string
	// UpdatePackageArgs are the arguments to pass to the tool, followed by a package name, to update that package
	// eg. []string{"upgrade"}
	UpdatePackageArgs []string
	// UpdateCommands lists the commands that update a package, or every package when packageName is empty,
	// for a tool without one command that updates them all, in place of [BasicTool.UpdateArgs] and [BasicTool.UpdatePackageArgs]
	UpdateCommands func(ctx context.Context, packageName string) ([]*exec.Cmd, error)
	// UpdateHint is the command to give the user for a tool with UpdateCommands
	// eg. "go install <package>@latest"
	UpdateHint string
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

func (b *BasicTool) Command() string {
	return b.ToolCmd
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

func (b *BasicTool) Update(ctx context.Context, packageName string) ([]*exec.Cmd, error) {
	if b.UpdateCommands != nil {
		return b.UpdateCommands(ctx, packageName)
	}
	if packageName != "" {
		return []*exec.Cmd{b.command(ctx, append(slices.Clone(b.UpdatePackageArgs), packageName)...)}, nil
	}
	if b.UpdateArgs != nil {
		return []*exec.Cmd{b.command(ctx, b.UpdateArgs...)}, nil
	}

	packages, err := b.Outdated(ctx)
	if err != nil || len(packages) == 0 {
		return nil, err
	}
	args := slices.Clone(b.UpdatePackageArgs)
	for _, outdatedPackage := range packages {
		args = append(args, outdatedPackage.Name)
	}
	return []*exec.Cmd{b.command(ctx, args...)}, nil
}

func (b *BasicTool) UpdateCommand() string {
	if b.UpdateHint != "" {
		return b.UpdateHint
	}
	if b.UpdateArgs == nil {
		return strings.Join(append([]string{b.ToolCmd}, b.UpdatePackageArgs...), " ") + " <packages>"
	}
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

// WithStderr adds a failed command's stderr to its error, which otherwise says only "exit status 1".
func WithStderr(err error) error {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
		return fmt.Errorf("%w\n%s", err, bytes.TrimSpace(exitErr.Stderr))
	}
	return err
}

// outputAllowingExit returns cmd's output, taking exit status code as success,
// for a tool that signals outdated packages with it.
func outputAllowingExit(cmd *exec.Cmd, code int) ([]byte, error) {
	output, err := cmd.Output()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == code {
		err = nil
	}
	return output, err
}
