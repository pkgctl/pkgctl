package status

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/pkgctl/pkgctl/ioutil/colors"
	"github.com/pkgctl/pkgctl/outdated"
	"github.com/pkgctl/pkgctl/tools"
)

// How old the last check can get before status starts a new one in the background.
const refreshAfter = 24 * time.Hour

// How old the last check can get before status says so; with daily refreshes, it means checks keep failing.
const staleAfter = 7 * 24 * time.Hour

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
	return c.fs.Parse(args)
}

// Run lists the outdated packages the last check found, or with --short only counts them in one line,
// and adds a line if that check is more than a week old. When the last check is more than a day old,
// it starts a new one in the background.
func (c *StatusCmd) Run(ctx context.Context) error {
	report, err := outdated.Load()
	if err != nil {
		return err
	}

	age := time.Since(report.CheckedAt)
	if age > refreshAfter {
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

	if !report.CheckedAt.IsZero() && age > staleAfter {
		days := int(age.Hours() / 24)
		fmt.Printf("Last checked %v days ago. Run `pkgctl check` to see what's failing.\n", days)
	}
	return nil
}

// printDetails lists each tool's outdated packages the way `pkgctl check` does, then the summary.
func printDetails(report outdated.Report) {
	for _, tool := range tools.List() {
		packages := report.Packages[tool.Name()]
		if len(packages) == 0 {
			continue
		}
		fmt.Println(colors.BLUE + tool.Name() + colors.END)
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
		fmt.Println(outdated.UpToDate)
	}
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
