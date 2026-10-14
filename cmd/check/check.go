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

	"github.com/pkgctl/pkgctl/config"
	"github.com/pkgctl/pkgctl/ioutil/colors"
	"github.com/pkgctl/pkgctl/outdated"
	"github.com/pkgctl/pkgctl/tools"
)

type CheckCmd struct {
	fs    *flag.FlagSet
	ifDue bool
}

const usage = `Usage: pkgctl check [options]

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

func (c *CheckCmd) Parse(args []string) error {
	return c.fs.Parse(args)
}

// Run asks every installed tool that can report outdated packages for them, prints what each found,
// and saves the report for `pkgctl status`. If any tool fails, the previous report stays in place.
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
	for _, tool := range tools.List() {
		if !tool.ReportsOutdated() || !tool.Exists() {
			continue
		}

		fmt.Println(colors.Stdout.Header("↻ " + tool.Name()))
		packages, err := tool.Outdated(ctx)
		if err != nil {
			return fmt.Errorf("checking %v: %w", tool.Name(), withStderr(err))
		}

		if len(packages) == 0 {
			fmt.Println(colors.Stdout.Green("✓") + " up to date")
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
		fmt.Println(colors.Stdout.Green("Everything is up to date."))
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
