package cmd

import (
	"context"

	"github.com/pkgctl/pkgctl/cmd/list"
	"github.com/pkgctl/pkgctl/cmd/update"
	"github.com/pkgctl/pkgctl/cmd/version"
)

type Command interface {
	Parse([]string) error
	Run(context.Context) error
}

var CommandList = map[string]Command{
	"list":    list.NewListCommand(),
	"update":  update.NewUpdateCmd(),
	"version": version.NewVersionCommand(),
}
