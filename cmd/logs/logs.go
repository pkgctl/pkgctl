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
	"github.com/pkgctl/pkgctl/tools"
)

type LogsCmd struct {
	fs    *flag.FlagSet
	limit int
	cmd   string
}

func NewLogsCmd() *LogsCmd {
	fs := flag.NewFlagSet("pkgctl logs", flag.ExitOnError)

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: pkgctl logs [update] [--limit <n>] \n")
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

	lc.cmd = lc.fs.Arg(0)

	switch lc.cmd {
	case "update":

	case "":
	default:
		fmt.Printf("invalid command '%s'\n", lc.cmd)
		lc.fs.Usage()
		os.Exit(1)
	}

	return nil
}

func (lc *LogsCmd) Run() error {

	switch lc.cmd {
	case "update":
		lc.printUpdateLogs()
	default:
		lc.printLogs()
	}

	return nil
}

func (lc *LogsCmd) printLogs() {
	logFiles := getLogs("", lc.limit)
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

func (lc *LogsCmd) printUpdateLogs() {
	logFiles := getLogs("update", -1)

	w := new(tabwriter.Writer)
	// minwidth, tabwidth, padding, padchar, flags
	w.Init(os.Stdout, 8, 8, 2, '\t', 0)

	const COL_FORMAT = "%s\t%s\t%s\t%s\t%s\n"
	fmt.Fprintf(w, COL_FORMAT, "TOOL", "PACKAGE", "FROM VERSION", "TO VERSION", "TIME")

	linesPrinted := 0

	for _, logFile := range logFiles {
		tool := tools.GetTool(logFile.ToolID)
		if tool == nil {
			fmt.Fprintf(os.Stderr, "unknown tool: %s\n", logFile.ToolID)
			continue
		}
		// Get the list of updated packages from the log
		updates, err := tool.ParseForUpdates(logFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to parse log %s: %s\n", logFile.Path, err.Error())
			// delete the log file
			os.Remove(logFile.Path)
			continue
		}

		for _, update := range updates {
			linesPrinted++
			fmt.Fprintf(w, COL_FORMAT, logFile.ToolID, update.Name, update.FromVersion, update.ToVersion, duration(time.Since(logFile.Time)))
		}

		if lc.limit > 0 && linesPrinted == lc.limit {
			break
		}
	}
	w.Flush()
}

func getLogs(cmd string, limit int) []logs.LogFile {
	logFiles, err := logs.GetAll()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// sort by time
	sort.Slice(logFiles, func(i, j int) bool {
		return logFiles[i].Time.After(logFiles[j].Time)
	})

	if cmd != "" {
		// filter by command
		var filteredLogs []logs.LogFile
		for _, logFile := range logFiles {
			if logFile.PkgctlCmd == cmd {
				filteredLogs = append(filteredLogs, logFile)
			}
		}
		logFiles = filteredLogs
	}

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
