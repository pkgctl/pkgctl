package check

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/pkgctl/pkgctl/config"
	"github.com/pkgctl/pkgctl/ioutil/colors"
	"github.com/pkgctl/pkgctl/outdated"
	"github.com/pkgctl/pkgctl/tools"
)

type CheckCmd struct {
	fs    *flag.FlagSet
	ifDue bool
	// tool is the one tool to check, or nil to check every tool.
	tool tools.Tool
}

const usage = `Usage: pkgctl check [options] [tool]

Check every tool, or one tool named by its command, as ` + "`pkgctl list`" + ` shows it, such as brew or go.

Options:
    --if-due Check only if check_interval has passed since the last check
`

func NewCheckCmd() *CheckCmd {
	fs := flag.NewFlagSet("pkgctl check", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", usage)
	}

	checkCmd := &CheckCmd{
		fs: fs,
	}

	fs.BoolVar(&checkCmd.ifDue, "if-due", false, "Check only if check_interval has passed since the last check")

	return checkCmd
}

// Parse takes the tool to check from the arguments, every tool when there is none.
func (c *CheckCmd) Parse(args []string) error {
	if err := c.fs.Parse(args); err != nil {
		return err
	}
	if c.fs.NArg() > 1 {
		return fmt.Errorf("unexpected argument %q", c.fs.Arg(1))
	}
	if c.fs.NArg() == 0 {
		return nil
	}

	tool, err := tools.Find(c.fs.Arg(0))
	if err != nil {
		return err
	}
	c.tool = tool
	return nil
}

// Run asks every installed tool that can report outdated packages, or the one tool named, for them,
// prints what each found, and saves the report for `pkgctl status`. If any tool fails, the previous report stays in place.
// If another check is running, or --if-due is set and no check is due, Run leaves the report alone.
func (c *CheckCmd) Run(ctx context.Context) error {
	if c.ifDue {
		due, err := checkDue()
		if err != nil {
			return err
		}
		if !due {
			return nil
		}
	}

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
	toolsToCheck := tools.List()
	if c.tool != nil {
		// Checking one tool replaces only its part of the saved report, and keeps when every tool was last checked.
		report, err = outdated.Load()
		if err != nil {
			return err
		}
		report.Clear(c.tool.Name(), "")
		toolsToCheck = []tools.Tool{c.tool}
	}
	for _, tool := range toolsToCheck {
		if !tool.ReportsOutdated() || !tool.Exists() {
			continue
		}

		packages, err := tool.Outdated(ctx)
		if err != nil {
			return fmt.Errorf("checking %v: %w", tool.Name(), tools.WithStderr(err))
		}

		if len(packages) == 0 {
			fmt.Println(colors.Stdout.Green("✓") + " " + tool.Name())
			continue
		}
		report.Packages[tool.Name()] = packages
		fmt.Println(colors.Stdout.Header(tool.Name()))
		fmt.Print(outdated.Describe(tool, packages))
	}

	if err := outdated.Save(report); err != nil {
		return err
	}

	if summary := report.Summary(); summary != "" {
		fmt.Println()
		fmt.Println(summary)
	}
	return nil
}

func checkDue() (bool, error) {
	settings, err := config.Load()
	if err != nil {
		return false, err
	}

	report, err := outdated.Load()
	if err != nil {
		return false, err
	}
	return settings.CheckDue(report.CheckedAt), nil
}
