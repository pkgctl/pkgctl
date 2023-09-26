package logs

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/pkgctl/pkgctl/logs"
)

type LogsCmd struct {
	fs *flag.FlagSet
}

func NewLogsCmd() *LogsCmd {
	fs := flag.NewFlagSet("pkgctl logs", flag.ExitOnError)

	var helpFlag bool

	fs.BoolVar(&helpFlag, "help", false, "Show help")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: pkgctl logs [options]\n")
	}

	gc := &LogsCmd{
		fs: fs,
	}

	return gc
}

func (l *LogsCmd) Parse(args []string) error {
	return l.fs.Parse(args)
}

func (l *LogsCmd) Run() error {
	getLogs()
	return nil
}

func getLogs() {
	logFiles, err := logs.GetAll()

	if err != nil {
		panic(err)
	}

	w := new(tabwriter.Writer)
	// minwidth, tabwidth, padding, padchar, flags
	w.Init(os.Stdout, 8, 8, 0, '\t', 0)

	// sort by time
	sort.Slice(logFiles, func(i, j int) bool {
		return logFiles[i].Time.After(logFiles[j].Time)
	})

	const COL_FORMAT = "%s\t%s\t%s\t%s\t%v\n"
	fmt.Fprintf(w, COL_FORMAT, "CMD", "TOOL", "TIME", "LOG FILE", "SIZE")

	for _, logFile := range logFiles {
		fmt.Fprintf(w, COL_FORMAT, logFile.PkgctlCmd, logFile.ToolID, logFile.Time.Format(time.Stamp), logFile.Path, logFile.Size)
	}
	w.Flush()

	fmt.Println()

	// fmt.Println("view log file with `pkgctl logs view <log file>`")
	fmt.Println("view logs: `gzat <log file>`")

}
