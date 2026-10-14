package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"

	"github.com/pkgctl/pkgctl/cmd"
	"github.com/pkgctl/pkgctl/cmd/version"
	"github.com/pkgctl/pkgctl/ioutil/colors"
	"github.com/pkgctl/pkgctl/tools"
)

const usage = `Usage: pkgctl <command> [options]

Commands:
  check     Check for outdated packages, in every tool or one
  list      List all tools
  status    Show the outdated packages found by the last check
  update    Update all tools, or one tool or package
  version   Print the version
`

func main() {

	ctx := context.Background()

	// // Catch all signals and cancel the context
	// ctx, cancel := context.WithCancel(ctx)
	// defer cancel()

	// signals := make(chan os.Signal, 1)
	// signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Catch all panics, print the stack trace, and show a message to file a bug
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(colors.Stdout.Red(fmt.Sprintf("Error: %v\n%s", r, debug.Stack())))
			fmt.Println("pkgctl has crashed! Please file a bug at https://github.com/pkgctl/pkgctl/issues")
		}
	}()

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n%s\n", usage, toolsUsage())
	}

	var versionFlag bool

	flag.BoolVar(&versionFlag, "version", false, "print the version")
	flag.Parse() // Scans the arg list and sets up flags.

	if versionFlag {
		fmt.Println(version.String())
		os.Exit(0)
	}

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(0)
	}

	// Parse the subcommand
	args := flag.Args()
	command := cmd.CommandList[args[0]]

	if command == nil {
		fmt.Fprintln(os.Stderr, colors.Stderr.Red(fmt.Sprintf("pkgctl: unknown command %q", args[0])))
		flag.Usage()
		os.Exit(2)
	}

	// Bad arguments exit 2, like a bad flag does.
	if err := command.Parse(args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, colors.Stderr.Red("pkgctl: "+err.Error()))
		os.Exit(2)
	}

	err := command.Run(ctx)
	if ctx.Err() != nil {
		// Interrupted: stop quietly, with the shell's usual status for Ctrl-C.
		os.Exit(130)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, colors.Stderr.Red("pkgctl: "+err.Error()))
		os.Exit(1)
	}
}

// toolsUsage lists the installed tools by the command that check and update take, then the supported ones that aren't installed.
func toolsUsage() string {
	var installed, missing []string
	for _, tool := range tools.List() {
		if tool.Exists() {
			installed = append(installed, tool.Command())
		} else {
			missing = append(missing, tool.Command())
		}
	}

	text := "Tools:\n  " + strings.Join(installed, ", ") + "\n"
	if len(missing) > 0 {
		text += "\nNot installed:\n  " + strings.Join(missing, ", ") + "\n"
	}
	return text
}
