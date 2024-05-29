package executor_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/pkgctl/pkgctl/executor"
	"github.com/pkgctl/pkgctl/tools"
)

func TestPassingCmds(t *testing.T) {
	tl := GetTool(3)

	ctx := context.Background()

	job := "testUpdateJob"
	commander := executor.NewCommander(executor.CommanderOptions{ParallelMode: false})

	for _, tool := range tl {
		commander.Add(job, tool, tool.Update(ctx))
	}

	commander.Run(ctx)

}

func GetTool(num int) []*tools.BasicTool {
	list := make([]*tools.BasicTool, num)

	for i := 0; i < num; i++ {
		list[i] = &tools.BasicTool{
			ToolName:        fmt.Sprintf("Tool %v", i),
			ToolCmd:         "sleep",
			ToolDescription: "Sleeps for a bit",
			UpdateArgs:      []string{"1000"},
			VersionArgs:     []string{"--version"},
		}
	}

	return list
}
