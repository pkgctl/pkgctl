package tools_test

import (
	"testing"

	"github.com/pkgctl/pkgctl/tools"
)

func TestBrew(t *testing.T) {
	expectedUpdates := map[string][]tools.Update{
		"update.brew.2024-02-12T09:04:06-05:00.log.gz": {
			{
				Name:        "python-setuptools",
				FromVersion: "69.0.3",
				ToVersion:   "69.1.0",
			},
			{
				Name:        "python-cryptography",
				FromVersion: "42.0.1",
				ToVersion:   "42.0.2",
			},
		},
	}
	testToolUpdates(t, tools.Brew, expectedUpdates)
}
