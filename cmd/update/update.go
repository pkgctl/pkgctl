package update

import (
	"flag"

	"github.com/pkgctl/pkgctl/executor"
	"github.com/pkgctl/pkgctl/tools"
)

const PKGCTL_CMD = "update"

type UpdateCmd struct {
	fs *flag.FlagSet
}

func NewUpdateCmd() *UpdateCmd {
	fs := flag.NewFlagSet("pkgctl update", flag.ExitOnError)

	return &UpdateCmd{
		fs: fs,
	}
}

func (c *UpdateCmd) Parse(args []string) error {
	return c.fs.Parse(args)
}

func (c *UpdateCmd) Run() error {
	update()
	return nil
}

func update() {
	commander := executor.NewCommander()

	for _, tool := range tools.List() {
		cmd := tool.Update()
		if tool.Exits() && cmd != nil {
			commander.Add(PKGCTL_CMD, tool, tool.Update)
		}
	}

	commander.Run()

}
