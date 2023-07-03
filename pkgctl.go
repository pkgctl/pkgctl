package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/pkgctl/pkgctl/tools"
)

var listCmd = flag.NewFlagSet("list", flag.ExitOnError)
var updateCmd = flag.NewFlagSet("update", flag.ExitOnError)

var help = flag.Bool("help", false, "Show help")

// var version = flag.String("version", "", "Print the version")

func main() {

	flag.Parse()

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(0)
	}

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	subcommand := os.Args[1]

	switch subcommand {
	case "list":
		listCmd.Parse(os.Args[2:])
		list()

	case "update":
		updateCmd.Parse(os.Args[2:])
		update()
	default:
		// flag.Usage()
	}

}

func list() {
	for _, tool := range tools.List() {
		fmt.Printf("%s — %s\n", tool.Name(), tool.Description())
	}
}

func update() {
	for _, tool := range tools.List() {
		cmd := tool.Update()
		if tool.Exits() && cmd != nil {
			println(BLUE + "— updating " + tool.Name() + " —" + END)

			// var stdout bytes.Buffer
			// var stderr bytes.Buffer

			// cmd.Stdout = &stdout
			// cmd.Stderr = &stderr

			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			// fmt.Println(cmd.String())

			err := cmd.Run()

			if err != nil {
				// fmt.Println(stderr.String())
				continue
			}

			// fmt.Println(stdout.String())
		}
	}
}

const (
	CLEAR  = "\033[H\033[2J"
	BLUE   = "\033[1;34m"
	RED    = "\033[1;31m"
	GREEN  = "\033[1;32m"
	YELLOW = "\033[1;33m"
	END    = "\033[0m"
)
