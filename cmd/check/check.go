package check

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkgctl/pkgctl/config"
	"github.com/pkgctl/pkgctl/ioutil/colors"
	"github.com/pkgctl/pkgctl/ioutil/terminal"
	"github.com/pkgctl/pkgctl/outdated"
	"github.com/pkgctl/pkgctl/tools"
)

type CheckCmd struct {
	fs    *flag.FlagSet
	ifDue bool
	// tool is the one tool to check, or nil to check every tool.
	tool *tools.Tool
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
	if c.ifDue {
		return errors.New("--if-due checks every tool, so it can't take a tool")
	}

	tool, err := tools.Find(c.fs.Arg(0))
	if err != nil {
		return err
	}
	if tool.ListOutdated == nil {
		return fmt.Errorf("%v can't report outdated packages", tool.Name)
	}
	c.tool = tool
	return nil
}

// Run asks every installed tool that can report outdated packages, or the one tool named, for them, all at once,
// prints what each found, and saves the report for `pkgctl status`. If any tool fails,
// the others still finish, and the previous report stays in place.
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
	var toolsToCheck []*tools.Tool
	if c.tool != nil {
		if !c.tool.Exists() {
			return fmt.Errorf("%v isn't installed", c.tool.Name)
		}
		// Checking one tool replaces only its part of the saved report, and keeps when every tool was last checked.
		report, err = outdated.Load()
		if err != nil {
			return err
		}
		report.Clear(c.tool.Name, "")
		toolsToCheck = []*tools.Tool{c.tool}
	} else {
		for _, tool := range tools.List() {
			if tool.ListOutdated != nil && tool.Exists() {
				toolsToCheck = append(toolsToCheck, tool)
			}
		}
	}
	if err := checkAll(ctx, toolsToCheck, report); err != nil {
		return err
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

var spinnerFrames = []rune("⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏")

const (
	spinnerInterval = 100 * time.Millisecond
	// cursorUp moves the cursor up the given number of lines.
	cursorUp = "\033[%dA"
)

// A checkResult is what one tool's check found.
type checkResult struct {
	tool     *tools.Tool
	packages []tools.Package
	err      error
}

// symbol marks whether the check found the tool up to date, outdated, or failed.
func (result checkResult) symbol() string {
	switch {
	case result.err != nil:
		return colors.Stdout.Red("✗")
	case len(result.packages) == 0:
		return colors.Stdout.Green("✓")
	default:
		return colors.Stdout.Yellow("↑")
	}
}

// checkAll asks each tool in checking for its outdated packages, all at once, and adds what each finds to report.
// It prints a row per tool, then what each outdated or failed tool found,
// and once every tool has finished, returns an error naming those that failed.
func checkAll(ctx context.Context, checking []*tools.Tool, report outdated.Report) error {
	results := make(chan checkResult, len(checking))
	for _, tool := range checking {
		go func(tool *tools.Tool) {
			packages, err := tool.ListOutdated(ctx)
			results <- checkResult{tool: tool, packages: packages, err: err}
		}(tool)
	}

	// On a terminal, redraw the rows in place until every tool finishes; elsewhere, print them once at the end.
	live := terminal.SupportsEscapes(os.Stdout)
	ticker := time.NewTicker(spinnerInterval)
	defer ticker.Stop()
	finished := map[*tools.Tool]checkResult{}
	for frame := 0; ; frame++ {
		if live {
			if frame > 0 {
				fmt.Printf(cursorUp, len(checking))
			}
			printRows(checking, finished, frame)
		}
		if len(finished) == len(checking) {
			break
		}
		select {
		case <-ticker.C:
		case result := <-results:
			finished[result.tool] = result
		}
	}
	// An interrupt kills the tools' commands, so their errors aren't worth printing.
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if !live {
		printRows(checking, finished, 0)
	}

	// List what each outdated or failed tool found, in the rows' order.
	var failed []string
	for _, tool := range checking {
		result := finished[tool]
		switch {
		case result.err != nil:
			failed = append(failed, tool.Name)
			fmt.Println()
			fmt.Println(colors.Stdout.Header(tool.Name))
			fmt.Println(colors.Stdout.Red(tools.WithStderr(result.err).Error()))
		case len(result.packages) > 0:
			report.Packages[tool.Name] = result.packages
			fmt.Println()
			fmt.Println(colors.Stdout.Header(tool.Name))
			fmt.Print(outdated.Describe(tool, result.packages))
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("checking %v failed", strings.Join(failed, ", "))
	}
	return nil
}

// printRows prints a row for each tool being checked: a spinner while it runs, then the symbol for what it found.
func printRows(checking []*tools.Tool, finished map[*tools.Tool]checkResult, frame int) {
	spinner := colors.Stdout.Cyan(string(spinnerFrames[frame%len(spinnerFrames)]))
	for _, tool := range checking {
		symbol := spinner
		if result, done := finished[tool]; done {
			symbol = result.symbol()
		}
		fmt.Println(symbol + " " + tool.Name)
	}
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
