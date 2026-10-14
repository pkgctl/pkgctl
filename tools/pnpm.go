package tools

import "context"

var Pnpm = &Tool{
	Name:              "pnpm",
	Command:           "pnpm",
	Description:       "Fast, disk space efficient package manager",
	UpdatePackageArgs: []string{"update", "--global"},
	VersionArgs:       []string{"--version"},
	ListOutdated:      listPnpmOutdated,
}

func listPnpmOutdated(ctx context.Context) ([]Package, error) {
	// pnpm outdated exits 1 when a package is outdated.
	output, err := outputAllowingExit(ExecContext(ctx, "pnpm", "outdated", "--global", "--format", "json"), 1)
	if err != nil {
		return nil, err
	}
	return parseNodeOutdated(string(output))
}
