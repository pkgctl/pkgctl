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
	if err := c.fs.Parse(args); err != nil {
		return err
	}
	if c.fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", c.fs.Arg(0))
	}
	return nil
}

func (c *ListCommand) Run(ctx context.Context) error {
	list()
	return nil
}

func list() {
	for _, tool := range tools.List() {
		fmt.Printf("%-8s %s — %s\n", tool.Command, tool.Name, tool.Description)
	}
}
