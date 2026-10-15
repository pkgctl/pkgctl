# pkgctl

A package manager for package managers.

## What is pkgctl?

`pkgctl` is a package manager that itself doesn't actually manage any individual packages. It is designed to be a single executable that can be used to, primarily, update the packages managed by other package managers.

No need to remember the exact command to update a package manager's packages. Just run `pkgctl update` and it will update every package manager that has outdated packages.

## Supported Package Managers

- [.NET global tools](https://learn.microsoft.com/dotnet/core/tools/global-tools)
- [ASDF](https://asdf-vm.com/)
- [Fisher](https://github.com/jorgebucaran/fisher)
- [Go](https://go.dev/ref/mod#go-install), for commands installed with `go install`
- [Homebrew](https://brew.sh/)
- [mas](https://github.com/mas-cli/mas)
- [mise](https://mise.jdx.dev/)
- [npm](https://www.npmjs.com/)
- [pip](https://pip.pypa.io/)
- [pipx](https://pipx.pypa.io/)
- [pnpm](https://pnpm.io/)
- [RubyGems](https://rubygems.org/), for the gems that don't ship with Ruby
- [RustUp](https://rustup.rs/)
- [uv](https://docs.astral.sh/uv/)

## Installation

```shell
brew install pkgctl/tap/pkgctl
```

To check for outdated packages in the background, also start the service, which runs `pkgctl check --if-due` every 15 minutes:

```shell
brew services start pkgctl
```

## Features

### Global Package Update

Update the package managers that the last check found outdated packages for, as `pkgctl status` lists them. ASDF, which pkgctl can't check, is always updated, and so is every package manager until the first check.

```shell
pkgctl update
```

Add `--all` to update every package manager, whether or not the last check found outdated packages:

```shell
pkgctl update --all
```

Name a tool by its command, as `pkgctl list` shows it, to update only that tool, and add a package, as `pkgctl status` names it, to update only that package:

```shell
pkgctl update go golang.org/x/tools/gopls
```

When mas has apps to update, `pkgctl update` asks for the administrator password once, before it starts.

### Outdated Packages

`pkgctl check` asks every supported package manager except ASDF which installed packages have a newer version, prints them, and saves the answer. `pkgctl check go` asks only the one named, as `pkgctl list` shows it. `pkgctl status` lists the saved answer the same way without checking again. `pkgctl status --short` prints only a one-line summary, or nothing when everything is current, so it is quick enough to run whenever a shell starts. Add it to your shell's startup:

```fish
# ~/.config/fish/config.fish, after PATH is set
status is-interactive; and pkgctl status --short
```

```zsh
# ~/.zshrc
pkgctl status --short
```

```bash
# ~/.bashrc
[[ $- == *i* ]] && pkgctl status --short
```

For fish, keep the `is-interactive` check: pkgctl runs tools through `fish -c`, which loads `config.fish` as well.

When `check_interval` has passed since the last check, `pkgctl status` also starts `pkgctl check` in the background, so the line stays current without running `check` yourself. `pkgctl check --if-due` checks only in that case, for running on a schedule. If seven intervals pass without a successful check, `status` says so; run `pkgctl check` to see what's failing.

`pkgctl update` clears each tool it updates from the saved answer.

## Settings

Settings live in `~/.config/pkgctl/config.json`. The file is optional, and each setting has a default.

```json
{
  "check_interval": "1h"
}
```

- `check_interval`: how often to check for outdated packages in the background, as a duration such as `30m` or `6h`. Defaults to `24h`.

pkgctl colors its output only on a terminal. Set `NO_COLOR` to any value to turn color off there too.
