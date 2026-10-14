package version

import (
	"context"
	"flag"
	"fmt"
	"runtime/debug"

	"github.com/pkgctl/pkgctl/ioutil/colors"
	"github.com/pkgctl/pkgctl/tools"
)

const VERSION = "0.3.0-dev"

// String names the version, and for a build from a git checkout its commit,
// with "modified" if the checkout had uncommitted changes.
func String() string {
	version := "pkgctl version " + VERSION

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return version
	}
	var revision, modified string
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			revision = setting.Value[:min(7, len(setting.Value))]
		case "vcs.modified":
			modified = setting.Value
		}
	}
	if revision == "" {
		return version
	}
	if modified == "true" {
		revision += ", modified"
	}
	return version + " (" + revision + ")"
}

type VersionCommand struct {
	fs *flag.FlagSet
	// all bool
}

func NewVersionCommand() *VersionCommand {
	fs := flag.NewFlagSet("pkgctl version", flag.ExitOnError)

	vc := &VersionCommand{
		fs: fs,
	}

	// vc.fs.BoolVar(&vc.all, "all", false, "print versions of all installed tools")

	return vc
}

func (c *VersionCommand) Parse(args []string) error {
	if err := c.fs.Parse(args); err != nil {
		return err
	}
	if c.fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", c.fs.Arg(0))
	}
	return nil
}

func (c *VersionCommand) Run(ctx context.Context) error {
	fmt.Println(String())

	listToolVersions()

	return nil
}

func listToolVersions() {
	fmt.Println()
	fmt.Println(colors.Stdout.Green("tools:"))

	for _, tool := range tools.List() {
		if tool.Exists() {
			fmt.Printf("%v %v", colors.Stdout.Bold(tool.Name+":"), tool.Version())
		}
	}
}
