package update

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/pkgctl/pkgctl/ioutil/colors"
	"github.com/pkgctl/pkgctl/outdated"
	"github.com/pkgctl/pkgctl/tools"
)

type UpdateCmd struct {
	fs *flag.FlagSet
	// tool is the one tool to update, or nil to update every tool.
	tool        tools.Tool
	packageName string
}

const usage = `Usage: pkgctl update [tool [package]]

Update every tool, or one tool named by its command, as ` + "`pkgctl list`" + ` shows it, such as brew or go,
or one of its packages, as ` + "`pkgctl status`" + ` names it.
`

func NewUpdateCmd() *UpdateCmd {
	fs := flag.NewFlagSet("pkgctl update", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", usage)
	}

	return &UpdateCmd{
		fs: fs,
	}
}

// Parse takes the tool and package to update from the arguments, every tool when there are none.
func (c *UpdateCmd) Parse(args []string) error {
	if err := c.fs.Parse(args); err != nil {
		return err
	}
	if c.fs.NArg() > 2 {
		return fmt.Errorf("unexpected argument %q", c.fs.Arg(2))
	}
	if c.fs.NArg() == 0 {
		return nil
	}

	tool, err := tools.Find(c.fs.Arg(0))
	if err != nil {
		return err
	}
	c.tool = tool
	c.packageName = c.fs.Arg(1)
	return nil
}

func (c *UpdateCmd) Run(ctx context.Context) error {
	toolsToUpdate := tools.List()
	if c.tool != nil {
		toolsToUpdate = []tools.Tool{c.tool}
	}

	var updated []string
	ran := 0
	for _, tool := range toolsToUpdate {
		if ctx.Err() != nil {
			break
		}
		if !tool.Exists() {
			continue
		}

		if ran > 0 {
			fmt.Println()
		}
		if runUpdate(ctx, tool, c.packageName) == nil {
			updated = append(updated, tool.Name())
		}
		ran++
	}

	// A successful update leaves nothing for `pkgctl status` to report until the next check.
	// Load the report only now, so a check saved during the updates is kept.
	report, err := outdated.Load()
	if err != nil {
		return err
	}
	for _, name := range updated {
		report.Clear(name, c.packageName)
	}
	if err := outdated.Save(report); err != nil {
		return err
	}
	return ctx.Err()
}

// runUpdate updates packageName, or all of tool's packages when it is empty, in the terminal,
// between a header naming tool and a status line, and returns the error.
func runUpdate(ctx context.Context, tool tools.Tool, packageName string) error {
	// Symbols: ✓ ✗ ⟳ ↻
	fmt.Println(colors.Stdout.Header(tool.Name()))

	start := time.Now()
	err := update(ctx, tool, packageName)
	duration := time.Since(start).Round(100 * time.Millisecond)

	if err != nil {
		fmt.Printf("%v %v\n", colors.Stdout.Red("✗"), tools.WithStderr(err))
		return err
	}
	fmt.Printf("%v finished in %v\n", colors.Stdout.Green("✓"), duration)
	return nil
}

// update runs the commands that update packageName, or all of tool's packages, in order, stopping at the first failure.
func update(ctx context.Context, tool tools.Tool, packageName string) error {
	cmds, err := tool.Update(ctx, packageName)
	if err != nil {
		return err
	}
	for _, cmd := range cmds {
		// cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}
