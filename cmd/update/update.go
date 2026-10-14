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
}

const usage = "Usage: pkgctl update\n"

func NewUpdateCmd() *UpdateCmd {
	fs := flag.NewFlagSet("pkgctl update", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", usage)
	}

	return &UpdateCmd{
		fs: fs,
	}
}

func (c *UpdateCmd) Parse(args []string) error {
	return c.fs.Parse(args)
}

func (c *UpdateCmd) Run(ctx context.Context) error {
	var updated []string
	ran := 0
	for _, tool := range tools.List() {
		if ctx.Err() != nil {
			break
		}
		if !tool.Exists() {
			continue
		}

		if ran > 0 {
			fmt.Println()
		}
		if runUpdate(ctx, tool) == nil {
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
		delete(report.Packages, name)
	}
	if err := outdated.Save(report); err != nil {
		return err
	}
	return ctx.Err()
}

// runUpdate updates tool's packages in the terminal, between a header naming tool and a status line, and returns the error.
func runUpdate(ctx context.Context, tool tools.Tool) error {
	// Symbols: ✓ ✗ ⟳ ↻
	fmt.Println(colors.Stdout.Header(tool.Name()))

	start := time.Now()
	err := update(ctx, tool)
	duration := time.Since(start).Round(100 * time.Millisecond)

	if err != nil {
		fmt.Printf("%v %v\n", colors.Stdout.Red("✗"), tools.WithStderr(err))
		return err
	}
	fmt.Printf("%v finished in %v\n", colors.Stdout.Green("✓"), duration)
	return nil
}

// update runs the commands that update tool's packages, in order, stopping at the first failure.
func update(ctx context.Context, tool tools.Tool) error {
	cmds, err := tool.Update(ctx)
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
