package update

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/pkgctl/pkgctl/ioutil/colors"
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
	ran := 0
	for _, tool := range tools.List() {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		cmd := tool.Update(ctx)
		if !tool.Exists() || cmd == nil {
			continue
		}

		if ran > 0 {
			fmt.Println()
		}
		runUpdate(tool, cmd)
		ran++
	}
	return nil
}

// runUpdate runs cmd in the terminal, between a header naming tool and a status line.
func runUpdate(tool tools.Tool, cmd *exec.Cmd) {
	// Symbols: ✓ ✗ ⟳ ↻
	fmt.Println(colors.BLUE + "↻ " + tool.Name() + colors.END)

	// cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start).Round(100 * time.Millisecond)

	status := colors.GREEN + "✓" + colors.END
	if err != nil {
		status = colors.RED + "✗" + colors.END
	}
	fmt.Printf("%v finished in %v\n", status, duration)
}
