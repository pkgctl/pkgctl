package tools_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pkgctl/pkgctl/logs"
	"github.com/pkgctl/pkgctl/tools"
)

func testToolUpdates(t *testing.T, tool tools.CliTool, expectedUpdates map[string][]tools.Update) {
	t.Run(tool.ID(), func(t *testing.T) {
		for logFilePath, updates := range expectedUpdates {

			logFile := getTestLog(logFilePath)
			parsedUpdates, err := tool.ParseForUpdates(logFile)

			if err != nil {
				t.Error(err)
			}

			if len(parsedUpdates) != len(updates) {
				t.Errorf("Expected %d updates, got %d", len(updates), len(parsedUpdates))
			}

		}
	})
}

func getTestLog(name string) logs.LogFile {

	wd, _ := os.Getwd()
	logFile := filepath.Join(wd, "..", "test", "logs", name)

	fileInfo, err := os.Stat(logFile)
	if err != nil {
		panic(err)
	}

	m, ok := logs.Match(fileInfo)
	if !ok {
		panic("Match failed")
	}

	return m
}
