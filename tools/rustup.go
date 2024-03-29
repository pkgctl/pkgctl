package tools

import (
	"fmt"
	"regexp"
)

var RustUp = &BasicTool{
	ToolID:             "rustup",
	ToolName:           "rustup",
	ToolCmd:            "rustup",
	ToolDescription:    "The Rust toolchain installer",
	UpdateArgs:         []string{"update"},
	VersionArgs:        []string{"--version"},
	ParseUpdateLogFunc: parseRustUpLogForUpdates,
}

// stable-aarch64-apple-darwin updated - rustc 1.74.0 (79e9716c9 2023-11-13) (from rustc 1.73.0 (cc66ad468 2023-10-03))
// stable-x86_64-unknown-linux-gnu updated - (error reading rustc version) (from (error reading rustc version))
// nightly-aarch64-apple-darwin updated - rustc 1.76.0-nightly (6b771f6b5 2023-11-15) (from rustc 1.76.0-nightly (dd430bc8c 2023-11-14))

func parseRustUpLogForUpdates(logOutput string) ([]Update, error) {
	re := regexp.MustCompile(`(\S+) updated - (\w+) (.*\)) \(from \w+ (.*)\)`)
	matches := re.FindAllStringSubmatch(logOutput, -1)
	updates := []Update{}
	for _, match := range matches {
		updates = append(updates, Update{
			Name:        fmt.Sprintf("%s (%s)", match[2], match[1]),
			FromVersion: match[4],
			ToVersion:   match[3],
		})
	}
	return updates, nil
}
