package cmd

import (
	"github.com/pkgctl/pkgctl/cmd/list"
	"github.com/pkgctl/pkgctl/cmd/logs"
	"github.com/pkgctl/pkgctl/cmd/update"
)

type Command interface {
	Parse([]string) error
	Run() error
}

var CommandList = map[string]Command{
	"list":   list.NewListCommand(),
	"logs":   logs.NewLogsCmd(),
	"update": update.NewUpdateCmd(),
}
