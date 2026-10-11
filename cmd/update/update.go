package update

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/pkgctl/pkgctl/executor"
	"github.com/pkgctl/pkgctl/tools"
)

const PKGCTL_CMD = "update"

type UpdateCmd struct {
	fs       *flag.FlagSet
	parallel bool
}

const usage = `Usage: pkgctl update [options]

Options:
    --parallel Run updates in parallel
`

func NewUpdateCmd() *UpdateCmd {
	fs := flag.NewFlagSet("pkgctl update", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", usage)
	}

	updateCmd := &UpdateCmd{
		fs: fs,
	}

	fs.BoolVar(&updateCmd.parallel, "parallel", false, "Run updates in parallel")

	return updateCmd
}

func (c *UpdateCmd) Parse(args []string) error {
	return c.fs.Parse(args)
}

func (c *UpdateCmd) Run(ctx context.Context) error {
	commander := executor.NewCommander(executor.CommanderOptions{
		ParallelMode: c.parallel,
		Prologue:     prologue,
		Epilogue:     epilogue,
	})

	for _, tool := range tools.List() {
		cmd := tool.Update(ctx)
		if tool.Exits() && cmd != nil {
			commander.Add(PKGCTL_CMD, tool, cmd)
		}
	}

	commander.Run(ctx)
	return nil
}

func prologue(e *executor.Executor) string {
	return e.Tool().Name()
}

func epilogue(e *executor.Executor) string {
	return fmt.Sprintf("finished in %v", e.Duration().Round(100*time.Millisecond))
}
