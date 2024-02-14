package tools_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pkgctl/pkgctl/logs"
	"github.com/pkgctl/pkgctl/tools"
)

func testToolUpdates(t *testing.T, tool tools.Tool, expectedUpdates map[string][]tools.Update) {
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

			for i, update := range updates {
				if parsedUpdates[i].Name != update.Name {
					t.Errorf("Expected name `%s`, got `%s`", update.Name, parsedUpdates[i].Name)
				}

				if parsedUpdates[i].FromVersion != update.FromVersion {
					t.Errorf("Expected from version `%s`, got `%s`", update.FromVersion, parsedUpdates[i].FromVersion)
				}

				if parsedUpdates[i].ToVersion != update.ToVersion {
					t.Errorf("Expected to version `%s`, got `%s`", update.ToVersion, parsedUpdates[i].ToVersion)
				}
			}

		}
	})
}

func getTestLog(name string) logs.LogFile {

	wd, _ := os.Getwd()
	logFile := filepath.Join(wd, "..", "test", "logs", name)

	m, ok := logs.GetLogFile(logFile)
	if !ok {
		panic("Match failed")
	}

	return m
}
