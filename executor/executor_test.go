package executor

import "testing"

func TestTruncateLine(t *testing.T) {
	line := "🍺  /opt/homebrew/Cellar/wget/1.25.0: 92 files, 4.5MB"

	truncated := truncateLine(line, 20)

	if truncated != "🍺  /opt/homebrew..." {
		t.Errorf("Expected `🍺  /opt/homebrew...`, got `%s`", truncated)
	}
}
