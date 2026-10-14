package tools

var Asdf = &Tool{
	Name:              "asdf",
	Command:           "asdf",
	UpdateArgs:        []string{"plugin", "update", "--all"},
	UpdatePackageArgs: []string{"plugin", "update"},
	VersionArgs:       []string{"version"},
	Description:       "Extendable version manager with support for Ruby, Node.js, Elixir, Erlang & more",
}

// TODO Updating the plugins is more work

// # for plugin in $(asdf plugin list); do
// #     log Updating ASDF plugin $plugin
// #     set -l latest (asdf latest $plugin)
// #     asdf install $plugin latest
// #     asdf global $plugin $latest
// # done
