package version

import (
	"context"
	"flag"
	"fmt"

	"github.com/pkgctl/pkgctl/tools"
)

const VERSION = "0.1"
const VERSION_STRING = "pkgctl version " + VERSION

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
	return c.fs.Parse(args)
}

func (c *VersionCommand) Run(ctx context.Context) error {
	fmt.Println(VERSION_STRING)

	listToolVersions()

	return nil
}

func listToolVersions() {
	fmt.Println("\ntools:")

	for _, tool := range tools.List() {
		if tool.Exits() {
			fmt.Print(tool.Version())
		}
	}
}
