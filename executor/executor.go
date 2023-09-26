package executor

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/pkgctl/pkgctl/ioutil"
	"github.com/pkgctl/pkgctl/ioutil/colors"
	"github.com/pkgctl/pkgctl/logs"
	"github.com/pkgctl/pkgctl/tools"
)

type executor struct {
	pkgctlCmd string

	tool tools.CliTool
	cmd  *exec.Cmd

	cmdWait chan error

	logFile   *os.File
	logWriter *gzip.Writer

	logBuf bytes.Buffer

	logTime time.Time
}

func newExecutor(pkgctlCmd string, tool tools.CliTool, cmd *exec.Cmd) *executor {
	return &executor{
		pkgctlCmd: pkgctlCmd,
		tool:      tool,
		cmd:       cmd,
		cmdWait:   make(chan error),
	}
}

func (e *executor) Close() {
	if e.logWriter != nil {
		e.logWriter.Close()
	}
	if e.logFile != nil {
		e.logFile.Close()
	}
}

func (e *executor) openLogFiles() error {
	e.logTime = time.Now()

	logPath := fmt.Sprintf("%v/%v.%v.%v.log.gz", logs.LOG_DIR, e.pkgctlCmd, e.tool.ID(), e.logTime.Format(time.RFC3339))

	logFile, err := os.Create(logPath)
	if err != nil {
		return err
	}
	e.logFile = logFile

	e.logWriter = gzip.NewWriter(e.logFile)

	e.cmd.Stdout = io.MultiWriter(e.logWriter, &e.logBuf)
	e.cmd.Stderr = io.MultiWriter(e.logWriter, &e.logBuf)

	return nil
}

func (e *executor) Start() error {
	e.openLogFiles()

	err := e.cmd.Start()
	if err != nil {
		return err
	}
	go func() {
		err := e.cmd.Wait()
		if err != nil {

			headerExtras, err := json.Marshal(logs.GzipHeaderExtras{
				ExitCode: e.cmd.ProcessState.ExitCode(),
			})

			if err != nil {
				panic(err)
			}

			e.logWriter.Extra = headerExtras

		} else {
		}
		e.Close()

		e.cmdWait <- err
		close(e.cmdWait)
	}()

	return nil
}

type Commander struct {
	executors []*executor

	spinner ioutil.Spinner
}

func NewCommander() *Commander {
	return &Commander{
		executors: make([]*executor, 0),
	}
}

func (e *Commander) Add(pkgctlCmd string, tool tools.CliTool, cmdFn func() *exec.Cmd) {
	e.executors = append(e.executors, newExecutor(pkgctlCmd, tool, cmdFn()))
}

func (e *Commander) startCmds() int {
	started := 0
	for _, executor := range e.executors {
		println(e.spinner.Current(), colors.BLUE+executor.tool.Name()+colors.END)
		err := executor.Start()
		if err == nil {
			started += 1
		} else {
			fmt.Fprintf(os.Stderr, colors.RED+"— failed to start "+executor.tool.Name()+colors.END)
		}

	}
	return started
}

func (c *Commander) Run() error {

	lastLineCount := c.startCmds()

	for {
		time.Sleep(100 * time.Millisecond)
		eraseLines(lastLineCount)

		spinner := c.spinner.Next()
		runningCmdCount := 0
		currentLineCount := 0

		lines, columns, _ := ioutil.TerminalSize()

		statusLines := len(c.executors) * 2

		maxLines := (lines - statusLines) / len(c.executors)

		for _, executor := range c.executors {
			finished := executor.cmd.ProcessState != nil
			var status string
			if finished {
				status = "✓"
			} else {
				status = spinner
				runningCmdCount += 1
			}

			fmt.Printf("%v %v\n", status, colors.BLUE+executor.tool.Name()+colors.END)
			currentLineCount += 1

			lines := strings.Split(strings.TrimSpace(executor.logBuf.String()), "\n")

			if maxLines > 0 && len(lines) > maxLines {
				prevLines := len(lines) - maxLines
				fmt.Printf(colors.YELLOW+"... %v lines omitted ... log file: %v\n"+colors.END, prevLines, executor.logFile.Name())
				lines = lines[len(lines)-maxLines:]
				currentLineCount += 1
			}

			for _, line := range lines {
				if columns > 0 && utf8.RuneCountInString(line) > int(columns) {
					// TODO this is not correct and only works for ASCII
					line = line[:columns-4] + "..."
				}
				fmt.Println(line)
				currentLineCount += 1
			}

		}

		lastLineCount = currentLineCount

		if runningCmdCount == 0 {
			break
		}
	}

	return nil
}

const ERASE_LINE_STR = "\x1b[1A\x1b[2K"

func eraseLines(num int) {
	for i := 0; i < num; i++ {
		fmt.Print(ERASE_LINE_STR)
	}
}
