package update

import "testing"

func TestParseAllAfterTool(t *testing.T) {
	updateCmd := NewUpdateCmd()

	err := updateCmd.Parse([]string{"go", "--all"})

	expected := "--all updates every tool, so it can't take a tool"
	if err == nil || err.Error() != expected {
		t.Errorf("Expected %q, got %v", expected, err)
	}
}
