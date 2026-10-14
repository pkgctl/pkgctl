package status

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/pkgctl/pkgctl/config"
	"github.com/pkgctl/pkgctl/ioutil/colors"
	"github.com/pkgctl/pkgctl/outdated"
	"github.com/pkgctl/pkgctl/tools"
)

// How many check intervals can pass without a successful check before status says so,
// since with refreshes that often, it means checks keep failing.
const staleAfterIntervals = 7

type StatusCmd struct {
	fs    *flag.FlagSet
	short bool
}

const usage = `Usage: pkgctl status [options]

Options:
    --short Print only the one-line summary, for when a shell starts
`

func NewStatusCmd() *StatusCmd {
	fs := flag.NewFlagSet("pkgctl status", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", usage)
	}

	statusCmd := &StatusCmd{
		fs: fs,
	}

	fs.BoolVar(&statusCmd.short, "short", false, "Print only the one-line summary, for when a shell starts")

	return statusCmd
}

func (c *StatusCmd) Parse(args []string) error {
	if err := c.fs.Parse(args); err != nil {
		return err
	}
	if c.fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", c.fs.Arg(0))
	}
	return nil
}

// Run lists the outdated packages the last check found, or with --short only counts them in one line,
// and adds a line if that check is seven check intervals old. When a check is due, it starts one in the background.
func (c *StatusCmd) Run(ctx context.Context) error {
	settings, err := config.Load()
	if err != nil {
		return err
	}

	report, err := outdated.Load()
	if err != nil {
		return err
	}

	if settings.CheckDue(report.CheckedAt) {
		if err := startCheck(); err != nil {
			return err
		}
	}

	if c.short {
		if summary := report.Summary(); summary != "" {
			fmt.Println(summary)
		}
	} else {
		printDetails(report)
	}

	age := time.Since(report.CheckedAt)
	if !report.CheckedAt.IsZero() && age > staleAfterIntervals*settings.CheckInterval {
		fmt.Printf("Last checked %v ago. Run `pkgctl check` to see what's failing.\n", formatAge(age))
	}
	return nil
}

// printDetails lists each tool's outdated packages the way `pkgctl check` does, then the summary.
func printDetails(report outdated.Report) {
	for _, tool := range tools.List() {
		packages := report.Packages[tool.Name]
		if len(packages) == 0 {
			continue
		}
		fmt.Println(colors.Stdout.Header(tool.Name))
		fmt.Print(outdated.Describe(tool, packages))
		fmt.Println()
	}

	summary := report.Summary()
	switch {
	case summary != "":
		fmt.Println(summary)
	case report.CheckedAt.IsZero():
		fmt.Println("No check has run yet; one has started in the background.")
	default:
		fmt.Println(colors.Stdout.Green("Everything is up to date."))
	}
}

// formatAge renders age in whole hours below two days, and in whole days from there.
func formatAge(age time.Duration) string {
	if age < 48*time.Hour {
		return fmt.Sprintf("%v hours", int(age.Hours()))
	}
	return fmt.Sprintf("%v days", int(age.Hours()/24))
}

// startCheck starts `pkgctl check` with its output discarded, in its own session so it outlives the shell.
func startCheck() error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}

	cmd := exec.Command(executable, "check")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	return cmd.Start()
}
