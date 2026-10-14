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
var toolList = []*Tool{
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

func List() []*Tool {
	return toolList
}

// Find returns the tool whose command is command, such as "brew".
func Find(command string) (*Tool, error) {
	for _, tool := range toolList {
		if tool.Command == command {
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

// A Tool is a package manager that pkgctl checks and updates.
type Tool struct {
	// Name is the human readable name of the tool
	// eg. "Homebrew"
	Name string
	// Command is the command to run the tool, which names the tool for `pkgctl check` and `pkgctl update`
	// eg. "brew"
	Command string
	// Shell is the shell that defines Command, for a tool that is a shell function rather than an executable
	// eg. "fish"; empty runs Command through $SHELL
	Shell string
	// Description is a short description of the tool
	// eg. "The missing package manager for macOS (or Linux)"
	Description string
	// UpdateArgs are the arguments to pass to the tool to update, or nil to pass the outdated packages
	// to UpdatePackageArgs instead, for a tool whose command can't update them all, or fails with none installed
	// eg. []string{"upgrade"}
	UpdateArgs []string
	// UpdatePackageArgs are the arguments to pass to the tool, followed by a package name, to update that package
	// eg. []string{"upgrade"}
	UpdatePackageArgs []string
	// UpdateCommands lists the commands that update a package, or every package when packageName is empty,
	// for a tool without one command that updates them all, in place of [Tool.UpdateArgs] and [Tool.UpdatePackageArgs]
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
	// NeedsSudo is whether updating the tool's packages asks for the administrator password through sudo
	NeedsSudo bool
}

func (t *Tool) Exists() bool {
	if _, err := exec.LookPath(t.Command); err == nil {
		return true
	}

	cmd := t.command(context.Background(), t.VersionArgs...)
	err := cmd.Run()
	return err == nil && cmd.ProcessState.ExitCode() == 0
}

// Update returns the commands that update packageName, as ListOutdated names it, or every package when it is empty,
// to run in order. It returns none if there is nothing to update.
func (t *Tool) Update(ctx context.Context, packageName string) ([]*exec.Cmd, error) {
	if t.UpdateCommands != nil {
		return t.UpdateCommands(ctx, packageName)
	}
	if packageName != "" {
		return []*exec.Cmd{t.command(ctx, append(slices.Clone(t.UpdatePackageArgs), packageName)...)}, nil
	}
	if t.UpdateArgs != nil {
		return []*exec.Cmd{t.command(ctx, t.UpdateArgs...)}, nil
	}

	packages, err := t.ListOutdated(ctx)
	if err != nil || len(packages) == 0 {
		return nil, err
	}
	args := slices.Clone(t.UpdatePackageArgs)
	for _, outdatedPackage := range packages {
		args = append(args, outdatedPackage.Name)
	}
	return []*exec.Cmd{t.command(ctx, args...)}, nil
}

// UpdateCommand returns the command to give the user for updating the tool's packages, such as "rustup update".
func (t *Tool) UpdateCommand() string {
	if t.UpdateHint != "" {
		return t.UpdateHint
	}
	if t.UpdateArgs == nil {
		return strings.Join(append([]string{t.Command}, t.UpdatePackageArgs...), " ") + " <packages>"
	}
	return strings.Join(append([]string{t.Command}, t.UpdateArgs...), " ")
}

func (t *Tool) Version() string {
	cmd := t.command(context.Background(), t.VersionArgs...)
	out, _ := cmd.Output()
	return string(out)
}

func (t *Tool) command(ctx context.Context, arg ...string) *exec.Cmd {
	if t.Shell == "" {
		return ExecContext(ctx, t.Command, arg...)
	}
	return exec.CommandContext(ctx, t.Shell, "-c", strings.Join(append([]string{t.Command}, arg...), " "))
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
