package update

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/pkgctl/pkgctl/ioutil/colors"
	"github.com/pkgctl/pkgctl/outdated"
	"github.com/pkgctl/pkgctl/tools"
)

type UpdateCmd struct {
	fs *flag.FlagSet
	// tool is the one tool to update, or nil to update every tool.
	tool        *tools.Tool
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
	candidates := tools.List()
	if c.tool != nil {
		candidates = []*tools.Tool{c.tool}
	}
	var toolsToUpdate []*tools.Tool
	for _, tool := range candidates {
		if tool.Exists() {
			toolsToUpdate = append(toolsToUpdate, tool)
		}
	}
	if c.tool != nil && len(toolsToUpdate) == 0 {
		return fmt.Errorf("%v isn't installed", c.tool.Name)
	}

	stopRenewing, sudoErr := c.askForSudo(ctx, toolsToUpdate)
	defer stopRenewing()

	var updated, failed []string
	for i, tool := range toolsToUpdate {
		if ctx.Err() != nil {
			break
		}

		if i > 0 {
			fmt.Println()
		}
		if runUpdate(ctx, tool, c.packageName, sudoErr) == nil {
			updated = append(updated, tool.Name)
		} else {
			failed = append(failed, tool.Name)
		}
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
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if len(failed) > 0 {
		return fmt.Errorf("updating %v failed", strings.Join(failed, ", "))
	}
	return nil
}

// How often an update renews the administrator password sudo caches, well within sudo's default of five minutes.
const sudoRenewal = time.Minute

// askForSudo asks for the administrator password up front when a tool will need it, so the update doesn't stop partway
// to ask, then renews it until the returned function is called. A tool needs it when it has outdated packages to update,
// and with a package named, when that package is one of them.
// When it can't get the password, it returns why, and the tools that need it fail with that error.
func (c *UpdateCmd) askForSudo(ctx context.Context, toolsToUpdate []*tools.Tool) (stopRenewing func(), err error) {
	var needing []string
	for _, tool := range toolsToUpdate {
		if !tool.NeedsSudo {
			continue
		}
		packages, err := tool.ListOutdated(ctx)
		if err != nil {
			return func() {}, fmt.Errorf("checking %v: %w", tool.Name, tools.WithStderr(err))
		}
		if c.packageName != "" {
			packages = slices.DeleteFunc(packages, func(outdatedPackage tools.Package) bool {
				return outdatedPackage.Name != c.packageName
			})
		}
		if len(packages) == 0 {
			continue
		}
		needing = append(needing, tool.Name)
	}
	if len(needing) == 0 {
		return func() {}, nil
	}

	fmt.Printf("%v needs the administrator password to update.\n", strings.Join(needing, ", "))
	sudo := exec.CommandContext(ctx, "sudo", "--validate")
	sudo.Stdin, sudo.Stdout, sudo.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := sudo.Run(); err != nil {
		return func() {}, fmt.Errorf("asking for the administrator password: %w", err)
	}
	fmt.Println()

	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(sudoRenewal)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				// A renewal that fails only means sudo asks again when a tool needs it.
				_ = exec.Command("sudo", "--non-interactive", "--validate").Run()
			}
		}
	}()
	return func() { close(done) }, nil
}

// runUpdate updates packageName, or all of tool's packages when it is empty, in the terminal,
// between a header naming tool and a status line, and returns the error.
func runUpdate(ctx context.Context, tool *tools.Tool, packageName string, sudoErr error) error {
	fmt.Println(colors.Stdout.Header(tool.Name))

	start := time.Now()
	err := update(ctx, tool, packageName, sudoErr)
	duration := time.Since(start).Round(100 * time.Millisecond)

	if err != nil {
		fmt.Printf("%v %v\n", colors.Stdout.Red("✗"), tools.WithStderr(err))
		return err
	}
	fmt.Printf("%v finished in %v\n", colors.Stdout.Green("✓"), duration)
	return nil
}

// update runs the commands that update packageName, or all of tool's packages, in order, stopping at the first failure.
// A tool that needs the administrator password fails with sudoErr when pkgctl couldn't get it.
func update(ctx context.Context, tool *tools.Tool, packageName string, sudoErr error) error {
	if tool.NeedsSudo && sudoErr != nil {
		return sudoErr
	}
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
