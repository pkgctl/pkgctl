package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/pkgctl/pkgctl/cmd"
)

const usage = `Usage: pkgctl <command> [options]

Commands:
  list      List all tools
  logs      List logs
  update    Update all tools
`

func main() {

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
