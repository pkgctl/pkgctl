package tools

import (
	"context"
	"debug/buildinfo"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

var GoBinaries = &BasicTool{
	ToolName:        "Go",
	ToolCmd:         "go",
	ToolDescription: "Commands installed with `go install`",
	VersionArgs:     []string{"version"},
	ListOutdated:    listGoOutdated,
	UpdateCommands:  goUpdateCommands,
	UpdateHint:      "go install <package>@latest",
}

// A goBinary is a command `go install` put in its directory.
type goBinary struct {
	Name    string
	Package string
	Module  string
	Version string
	Latest  string
}

func listGoOutdated(ctx context.Context) ([]Package, error) {
	binaries, err := outdatedGoBinaries(ctx)
	if err != nil {
		return nil, err
	}

	var packages []Package
	for _, binary := range binaries {
		packages = append(packages, Package{Name: binary.Package, DisplayName: binary.Name, Current: binary.Version, Latest: binary.Latest})
	}
	return packages, nil
}

// goUpdateCommands reinstalls packageName, or each outdated binary, at its latest version, one module at a time,
// since go has no command that updates them all and `go install` takes packages from only one module.
func goUpdateCommands(ctx context.Context, packageName string) ([]*exec.Cmd, error) {
	if packageName != "" {
		return []*exec.Cmd{ExecContext(ctx, "go", "install", packageName+"@latest")}, nil
	}

	binaries, err := outdatedGoBinaries(ctx)
	if err != nil {
		return nil, err
	}

	var cmds []*exec.Cmd
	for _, binary := range binaries {
		cmds = append(cmds, ExecContext(ctx, "go", "install", binary.Package+"@latest"))
	}
	return cmds, nil
}

// outdatedGoBinaries returns the binaries whose module has a newer version, with that version in [goBinary.Latest].
func outdatedGoBinaries(ctx context.Context) ([]goBinary, error) {
	binaries, err := goBinaries(ctx)
	if err != nil || len(binaries) == 0 {
		return nil, err
	}

	args := []string{"list", "-m", "-json"}
	for _, binary := range binaries {
		args = append(args, binary.Module+"@latest")
	}
	output, err := ExecContext(ctx, "go", args...).Output()
	if err != nil {
		return nil, err
	}
	latest, err := parseGoModules(string(output))
	if err != nil {
		return nil, err
	}

	var outdated []goBinary
	for _, binary := range binaries {
		if latest[binary.Module] != binary.Version {
			binary.Latest = latest[binary.Module]
			outdated = append(outdated, binary)
		}
	}
	return outdated, nil
}

// goBinaries returns the commands in `go install`'s directory that it can reinstall:
// Go binaries named for their package, unlike one renamed after installing, such as dlv-dap.
func goBinaries(ctx context.Context) ([]goBinary, error) {
	output, err := ExecContext(ctx, "go", "env", "-json", "GOBIN", "GOPATH").Output()
	if err != nil {
		return nil, err
	}
	var env struct {
		GOBIN  string
		GOPATH string
	}
	if err := json.Unmarshal(output, &env); err != nil {
		return nil, err
	}
	dir := env.GOBIN
	if dir == "" {
		dir = filepath.Join(env.GOPATH, "bin")
	}

	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var binaries []goBinary
	for _, entry := range entries {
		// The directory can hold other files, such as scripts, when GOBIN is a shared bin directory.
		info, err := buildinfo.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil || path.Base(info.Path) != entry.Name() {
			continue
		}
		binaries = append(binaries, goBinary{
			Name:    entry.Name(),
			Package: info.Path,
			Module:  info.Main.Path,
			Version: info.Main.Version,
		})
	}
	return binaries, nil
}

// parseGoModules parses `go list -m -json`, a stream of JSON objects, into each module's version, keyed by path.
func parseGoModules(output string) (map[string]string, error) {
	versions := map[string]string{}
	decoder := json.NewDecoder(strings.NewReader(output))
	for decoder.More() {
		var module struct {
			Path    string
			Version string
		}
		if err := decoder.Decode(&module); err != nil {
			return nil, err
		}
		versions[module.Path] = module.Version
	}
	return versions, nil
}
