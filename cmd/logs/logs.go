package logs

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
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
	w.Init(os.Stdout, 8, 8, 2, '\t', 0)

	// sort by time
	sort.Slice(logFiles, func(i, j int) bool {
		return logFiles[i].Time.After(logFiles[j].Time)
	})

	const COL_FORMAT = "%s\t%s\t%s\t%s\t%v\n"
	fmt.Fprintf(w, COL_FORMAT, "CMD", "TOOL", "TIME", "LOG FILE", "SIZE")

	for _, logFile := range logFiles {
		timeAgo := fmt.Sprintf("%vago", duration(time.Since(logFile.Time)))
		fmt.Fprintf(w, COL_FORMAT, logFile.PkgctlCmd, logFile.ToolID, timeAgo, logFile.Path, logFile.Size)
	}
	w.Flush()

	fmt.Println()

	// fmt.Println("view log file with `pkgctl logs view <log file>`")
	fmt.Println("view logs: `gzat <log file>`")

}

const (
	day   = time.Hour * 24
	week  = day * 7
	month = time.Hour * 24 * 365 / 12
	year  = month * 12
)

func duration(duration time.Duration) string {

	consume := func(cnt *int, sb *strings.Builder, d *time.Duration, m time.Duration) {
		const maxCnt = 1

		if *cnt == maxCnt {
			return
		}

		t := d.Truncate(m)
		v := int64(t / m)

		if v == 0 {
			return
		}

		var s string
		switch m {
		case year:
			s = "year"
		case month:
			s = "month"
		case week:
			s = "week"
		case day:
			s = "day"
		case time.Hour:
			s = "hour"
		case time.Minute:
			s = "minute"
		default:
			panic("unknown duration type")
		}
		*d -= t

		sb.WriteString(fmt.Sprintf("%v %v", v, s))
		if v > 1 {
			sb.WriteString("s")
		}
		if *cnt < maxCnt {
			sb.WriteString(" ")
		}
		*cnt++

	}

	sb := strings.Builder{}
	cnt := 0

	consume(&cnt, &sb, &duration, year)
	consume(&cnt, &sb, &duration, month)
	consume(&cnt, &sb, &duration, week)
	consume(&cnt, &sb, &duration, day)
	consume(&cnt, &sb, &duration, time.Hour)
	consume(&cnt, &sb, &duration, time.Minute)

	return sb.String()
}
