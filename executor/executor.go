package executor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/pkgctl/pkgctl/ioutil"
	"github.com/pkgctl/pkgctl/ioutil/colors"
	"github.com/pkgctl/pkgctl/tools"
)

type Executor struct {
	tool tools.Tool
	cmd  *exec.Cmd

	cmdWait chan error

	output lockedBuffer

	mutex     sync.Mutex
	startTime time.Time
	endTime   time.Time
}

// lockedBuffer is a bytes.Buffer that the render loop can read while a command writes to it.
type lockedBuffer struct {
	mutex  sync.Mutex
	buffer bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.buffer.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.buffer.String()
}

func newExecutor(tool tools.Tool, cmd *exec.Cmd, standardStreams bool) *Executor {
	e := &Executor{
		tool:    tool,
		cmd:     cmd,
		cmdWait: make(chan error),
	}

	if standardStreams {
		// cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = &e.output
		cmd.Stderr = &e.output
	}

	return e
}

func (e *Executor) Start() error {
	e.mutex.Lock()
	e.startTime = time.Now()
	e.mutex.Unlock()

	err := e.cmd.Start()
	if err != nil {
		return err
	}
	go func() {
		err := e.cmd.Wait()
		e.mutex.Lock()
		e.endTime = time.Now()
		e.mutex.Unlock()

		e.cmdWait <- err
		close(e.cmdWait)
	}()

	return nil
}

func (e *Executor) Wait() error {
	return <-e.cmdWait
}

func (e *Executor) Run() error {
	e.Start()
	return e.Wait()
}

func (e *Executor) Tool() tools.Tool {
	return e.tool
}

func (e *Executor) Duration() time.Duration {
	if e.startTime.IsZero() || e.endTime.IsZero() {
		panic(fmt.Sprintf("executor has not started OR finished: %v", e.tool.Name()))
	}
	e.mutex.Lock()
	defer e.mutex.Unlock()
	return e.endTime.Sub(e.startTime)
}

type Commander struct {
	executors []*Executor
	options   CommanderOptions
	spinner   ioutil.Spinner
}

type CommanderOptions struct {
	ParallelMode bool
	Prologue     func(*Executor) string
	Epilogue     func(*Executor) string
}

func NewCommander(options CommanderOptions) *Commander {
	return &Commander{
		executors: make([]*Executor, 0),
		options:   options,
	}
}

func (e *Commander) Add(tool tools.Tool, cmd *exec.Cmd) {
	if cmd == nil {
		panic(fmt.Sprintf("cmd is nil for %v", tool.Name()))
	}
	e.executors = append(e.executors, newExecutor(tool, cmd, !e.options.ParallelMode))
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

func (c *Commander) Run(ctx context.Context) {
	if c.options.ParallelMode {
		c.runInParallel(ctx)
	} else {
		c.runInSeries(ctx)
	}
}

func (c *Commander) runInSeries(_ context.Context) {
	for i, executor := range c.executors {
		if i > 0 {
			fmt.Println()
		}
		// Symbols: ✓ ✗ ⟳ ↻
		fmt.Printf(colors.BLUE + "↻ " + c.options.Prologue(executor) + colors.END + "\n")
		err := executor.Run()
		var status string
		if err == nil && executor.cmd.ProcessState.ExitCode() == 0 {
			status = colors.GREEN + "✓" + colors.END
		} else {
			status = colors.RED + "✗" + colors.END
		}
		fmt.Printf("%v %v\n", status, c.options.Epilogue(executor))
	}
}

func (c *Commander) runInParallel(_ context.Context) {
	lastLineCount := c.startCmds()

	loop(lastLineCount, func(lines, columns int, writer io.Writer) bool {
		spinner := c.spinner.Next()
		runningCmdCount := 0

		statusLines := len(c.executors) * 2
		maxLines := (lines - statusLines) / len(c.executors)

		for _, executor := range c.executors {
			outputLines := strings.Split(strings.TrimSpace(executor.output.String()), "\n")

			var statusLine string
			select {
			case err := <-executor.cmdWait:
				if err == nil && executor.cmd.ProcessState.ExitCode() == 0 {
					statusLine = colors.GREEN + "✓" + colors.END
				} else {
					statusLine = colors.RED + "✗" + colors.END
				}
			default:
				statusLine = spinner
				runningCmdCount += 1
			}

			fmt.Fprintf(writer, "%v %v\n", statusLine, colors.BLUE+executor.tool.Name()+colors.END)

			if maxLines > 0 && len(outputLines) > maxLines {
				prevLines := len(outputLines) - maxLines
				fmt.Fprintf(writer, colors.YELLOW+"... %v lines omitted ...\n"+colors.END, prevLines)
				outputLines = outputLines[len(outputLines)-maxLines:]
			}

			for _, line := range outputLines {
				fmt.Fprintln(writer, truncateLine(line, columns))
			}
		}

		return runningCmdCount == 0
	})

}

// truncateLine cuts line to fit a terminal columns wide, marking the cut with "...".
func truncateLine(line string, columns int) string {
	// Counts runes, not display cells, so a line with several wide characters such as emoji can still wrap.
	runes := []rune(line)
	if columns <= 0 || len(runes) <= columns {
		return line
	}
	return string(runes[:columns-4]) + "..."
}

func loop(initialLineCount int, fn func(int, int, io.Writer) bool) {
	lastLineCount := initialLineCount

	for {
		time.Sleep(100 * time.Millisecond)
		eraseLines(os.Stdout, lastLineCount)

		lines, columns, _ := ioutil.TerminalSize()

		var sb strings.Builder
		finished := fn(lines, columns, &sb)

		lastLineCount = strings.Count(sb.String(), "\n") + 1
		fmt.Fprintln(os.Stdout, sb.String())

		if finished {
			break
		}
	}
}

const ERASE_LINE_STR = "\x1b[1A\x1b[2K"

func eraseLines(w io.Writer, num int) {
	for i := 0; i < num; i++ {
		fmt.Fprint(w, ERASE_LINE_STR)
	}
}
