package check

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/pkgctl/pkgctl/ioutil/colors"
	"github.com/pkgctl/pkgctl/outdated"
	"github.com/pkgctl/pkgctl/tools"
)

type CheckCmd struct {
	fs *flag.FlagSet
}

const usage = "Usage: pkgctl check\n"

func NewCheckCmd() *CheckCmd {
	fs := flag.NewFlagSet("pkgctl check", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", usage)
	}

	return &CheckCmd{
		fs: fs,
	}
}

func (c *CheckCmd) Parse(args []string) error {
	return c.fs.Parse(args)
}

// Run asks every installed tool that can report outdated packages for them, prints what each found,
// and saves the report for `pkgctl status`. If any tool fails, the previous report stays in place.
// If another check is running, Run leaves the report to it.
func (c *CheckCmd) Run(ctx context.Context) error {
	lock, err := outdated.LockCheck()
	if errors.Is(err, outdated.ErrCheckRunning) {
		fmt.Println("Another check is already running.")
		return nil
	}
	if err != nil {
		return err
	}
	defer lock.Close()

	report := outdated.Report{
		CheckedAt: time.Now(),
		Packages:  map[string][]tools.Package{},
	}
	for _, tool := range tools.List() {
		if !tool.ReportsOutdated() || !tool.Exists() {
			continue
		}

		fmt.Println(colors.BLUE + "↻ " + tool.Name() + colors.END)
		packages, err := tool.Outdated(ctx)
		if err != nil {
			return fmt.Errorf("checking %v: %w", tool.Name(), withStderr(err))
		}

		if len(packages) == 0 {
			fmt.Println(colors.GREEN + "✓" + colors.END + " up to date")
			continue
		}
		report.Packages[tool.Name()] = packages
		fmt.Print(outdated.Describe(tool, packages))
	}

	if err := outdated.Save(report); err != nil {
		return err
	}

	fmt.Println()
	if summary := report.Summary(); summary != "" {
		fmt.Println(summary)
	} else {
		fmt.Println(outdated.UpToDate)
	}
	return nil
}

// withStderr adds a failed command's stderr to its error, which otherwise says only "exit status 1".
func withStderr(err error) error {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
		return fmt.Errorf("%w\n%s", err, bytes.TrimSpace(exitErr.Stderr))
	}
	return err
}
