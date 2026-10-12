package cmd

import (
	"context"

	"github.com/pkgctl/pkgctl/cmd/check"
	"github.com/pkgctl/pkgctl/cmd/list"
	"github.com/pkgctl/pkgctl/cmd/status"
	"github.com/pkgctl/pkgctl/cmd/update"
	"github.com/pkgctl/pkgctl/cmd/version"
)

type Command interface {
	Parse([]string) error
	Run(context.Context) error
}

var CommandList = map[string]Command{
	"check":   check.NewCheckCmd(),
	"list":    list.NewListCommand(),
	"status":  status.NewStatusCmd(),
	"update":  update.NewUpdateCmd(),
	"version": version.NewVersionCommand(),
}
