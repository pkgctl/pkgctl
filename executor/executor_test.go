package executor_test

import (
	"fmt"
	"testing"

	"github.com/pkgctl/pkgctl/executor"
	"github.com/pkgctl/pkgctl/tools"
)

func TestPassingCmds(t *testing.T) {
	tl := GetTool(3)

	job := "testUpdateJob"
	commander := executor.NewCommander()

	for _, tool := range tl {
		commander.Add(job, tool, tool.Update)
	}

	commander.Run()

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
