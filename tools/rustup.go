package tools

import (
	"context"
	"regexp"
)

var RustUp = &BasicTool{
	ToolName:        "rustup",
	ToolCmd:         "rustup",
	ToolDescription: "The Rust toolchain installer",
	UpdateArgs:      []string{"update"},
	VersionArgs:     []string{"--version"},
	ListOutdated:    listRustupOutdated,
}

func listRustupOutdated(ctx context.Context) ([]Package, error) {
	// rustup check exits 100 when an update is available.
	output, err := outputAllowingExit(ExecContext(ctx, "rustup", "check"), 100)
	if err != nil {
		return nil, err
	}
	return parseRustupCheck(string(output)), nil
}

// A `rustup check` line for a toolchain with an update, such as
// "nightly-aarch64-apple-darwin - update available: 1.101.0-nightly (32dba69d6 2026-10-09) -> 1.101.0-nightly (080cc2714 2026-10-11)".
var rustupUpdateLine = regexp.MustCompile(`(?m)^(\S+) - update available: (.+) -> (.+)$`)

func parseRustupCheck(output string) []Package {
	var packages []Package
	for _, match := range rustupUpdateLine.FindAllStringSubmatch(output, -1) {
		packages = append(packages, Package{Name: match[1], Current: match[2], Latest: match[3]})
	}
	return packages
}
