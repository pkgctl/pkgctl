package logs

import (
	"context"
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
	fs    *flag.FlagSet
	limit int
}

func NewLogsCmd() *LogsCmd {
	fs := flag.NewFlagSet("pkgctl logs", flag.ExitOnError)

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: pkgctl logs [--limit <n>]\n")
	}

	gc := &LogsCmd{
		fs: fs,
	}

	fs.IntVar(&gc.limit, "limit", 0, "Limit the number of logs to display. 0 for all logs.")

	return gc
}

func (lc *LogsCmd) Parse(args []string) error {
	err := lc.fs.Parse(args)

	if err != nil {
		return err
	}

	if lc.fs.NArg() > 0 {
		fmt.Printf("unexpected argument '%s'\n", lc.fs.Arg(0))
		lc.fs.Usage()
		os.Exit(1)
	}

	return nil
}

func (lc *LogsCmd) Run(ctx context.Context) error {
	lc.printLogs()
	return nil
}

func (lc *LogsCmd) printLogs() {
	logFiles := getLogs(lc.limit)
	w := new(tabwriter.Writer)
	// minwidth, tabwidth, padding, padchar, flags
	w.Init(os.Stdout, 8, 8, 2, '\t', 0)

	const COL_FORMAT = "%s\t%s\t%s\t%s\t%v\n"
	fmt.Fprintf(w, COL_FORMAT, "CMD", "TOOL", "TIME", "LOG FILE", "SIZE")

	for _, logFile := range logFiles {
		timeAgo := duration(time.Since(logFile.Time))
		fmt.Fprintf(w, COL_FORMAT, logFile.PkgctlCmd, logFile.ToolID, timeAgo, logFile.Path, logFile.Size)
	}
	w.Flush()
	fmt.Println("\nview logs: `gzat <log file>`")
}

func getLogs(limit int) []logs.LogFile {
	logFiles, err := logs.GetAll()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// sort by time
	sort.Slice(logFiles, func(i, j int) bool {
		return logFiles[i].Time.After(logFiles[j].Time)
	})

	// limit the number of logs to display
	// totalLogs := len(logFiles)
	if limit > 0 && len(logFiles) > limit {
		logFiles = logFiles[:limit]
	}

	return logFiles
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
		case time.Second:
			s = "second"
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
	consume(&cnt, &sb, &duration, time.Second)

	sb.WriteString("ago")

	return sb.String()
}
