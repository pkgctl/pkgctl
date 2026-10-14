package list

import (
	"context"
	"flag"
	"fmt"

	"github.com/pkgctl/pkgctl/tools"
)

type ListCommand struct {
	fs *flag.FlagSet
}

func NewListCommand() *ListCommand {
	fs := flag.NewFlagSet("pkgctl list", flag.ExitOnError)
	return &ListCommand{
		fs: fs,
	}
}

func (c *ListCommand) Parse(args []string) error {
	return c.fs.Parse(args)
}

func (c *ListCommand) Run(ctx context.Context) error {
	list()
	return nil
}

func list() {
	for _, tool := range tools.List() {
		fmt.Printf("%-8s %s — %s\n", tool.Command(), tool.Name(), tool.Description())
	}
}
