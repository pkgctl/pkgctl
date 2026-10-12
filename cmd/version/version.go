package version

import (
	"context"
	"flag"
	"fmt"

	"github.com/pkgctl/pkgctl/ioutil/colors"
	"github.com/pkgctl/pkgctl/tools"
)

const VERSION = "0.2.0"
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
	fmt.Println(colors.GREEN + "\ntools:" + colors.END)

	for _, tool := range tools.List() {
		if tool.Exists() {
			fmt.Printf(colors.BOLD+"%v: "+colors.END+"%v", tool.Name(), tool.Version())
		}
	}
}
