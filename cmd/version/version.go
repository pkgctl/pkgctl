package version

import (
	"flag"
	"fmt"

	"github.com/pkgctl/pkgctl/tools"
)

const VERSION = "0.1"

type VersionCommand struct {
	fs  *flag.FlagSet
	all bool
}

func NewVersionCommand() *VersionCommand {
	fs := flag.NewFlagSet("pkgctl version", flag.ExitOnError)

	vc := &VersionCommand{
		fs: fs,
	}

	vc.fs.BoolVar(&vc.all, "all", false, "print versions of all installed tools")

	return vc
}

func (c *VersionCommand) Parse(args []string) error {
	return c.fs.Parse(args)
}

func (c *VersionCommand) Run() error {
	fmt.Printf("pkgctl version %v\n", VERSION)

	if c.all {
		list()
	}

	return nil
}

func list() {
	fmt.Println("\nInstalled tool versions:")

	for _, tool := range tools.List() {
		if tool.Exits() {
			fmt.Printf("%s — %s", tool.Name(), tool.Version())
		}
	}
}
