package tools_test

import (
	"testing"

	"github.com/pkgctl/pkgctl/tools"
)

func TestRustUpdate(t *testing.T) {
	expectedUpdates := map[string][]tools.Update{
		"update.rustup.2024-02-08T10:15:53-05:00.log.gz": {
			{
				Name:        "rustc (stable-aarch64-apple-darwin)",
				FromVersion: "1.75.0 (82e1608df 2023-12-21)",
				ToVersion:   "1.76.0 (07dca489a 2024-02-04)",
			},
		},
	}
	testToolUpdates(t, tools.RustUp, expectedUpdates)
}
