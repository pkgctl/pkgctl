package update

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
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

		cmd := tool.Update(ctx)
		if !tool.Exists() || cmd == nil {
			continue
		}

		if ran > 0 {
			fmt.Println()
		}
		if runUpdate(tool, cmd) == nil {
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

// runUpdate runs cmd in the terminal, between a header naming tool and a status line, and returns its error.
func runUpdate(tool tools.Tool, cmd *exec.Cmd) error {
	// Symbols: ✓ ✗ ⟳ ↻
	fmt.Println(colors.Stdout.Header(tool.Name()))

	// cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start).Round(100 * time.Millisecond)

	status := colors.Stdout.Green("✓")
	if err != nil {
		status = colors.Stdout.Red("✗")
	}
	fmt.Printf("%v finished in %v\n", status, duration)
	return err
}
