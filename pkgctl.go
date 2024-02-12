package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/pkgctl/pkgctl/cmd"
	"github.com/pkgctl/pkgctl/ioutil/colors"
)

const usage = `Usage: pkgctl <command> [options]

Commands:
  list      List all tools
  logs      List logs
  update    Update all tools
  version   Print the version
`

func main() {

	// Catch all panics, print the stack trace, and show a message to file a bug
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf(colors.RED+"Error: %v\n%s\n"+colors.END, r, string(debug.Stack()))
			fmt.Println("pkgctl has crashed! Please file a bug at https://github.com/pkgctl/pkgctl/issues")
		}
	}()

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", usage)
	}

	if len(os.Args) == 1 {
		flag.Usage()
		os.Exit(0)
	}

	var versionFlag bool

	flag.BoolVar(&versionFlag, "version", false, "print the version")
	flag.Parse() // Scans the arg list and sets up flags. We currently have no global flags

	if versionFlag {
		bi, ok := debug.ReadBuildInfo()
		if ok {
			fmt.Printf("%+v", bi)
		}

		fmt.Println("version unknown")
		os.Exit(0)
	}

	args := flag.Args()

	command := cmd.CommandList[args[0]]

	if command == nil {
		flag.Usage()
		os.Exit(0)
	}

	command.Parse(args[1:])
	command.Run()
}
