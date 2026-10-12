# pkgctl

A package manager for package managers.

## What is pkgctl?

`pkgctl` is a package manager that itself doesn't actually manage any individual packages. It is designed to be a single executable that can be used to, primarily, update the packages managed by other package managers.

No need to remember the exact command to update a package manager's packages. Just run `pkgctl update` and it will update all the package managers that it is aware of.

## Supported Package Managers

- [ASDF](https://asdf-vm.com/)
- [Fisher](https://github.com/jorgebucaran/fisher)
- [Homebrew](https://brew.sh/)
- [mise](https://mise.jdx.dev/)
- [RustUp](https://rustup.rs/)

## Installation

```shell
brew tap pkgctl/homebrew
brew install pkgctl
```

## Features

### Global Package Update

Update all packages managed by `pkgctl`'s supported package managers.

```shell
pkgctl update
```

### Outdated Packages

`pkgctl check` asks Homebrew, mise, and rustup which installed packages have a newer version, prints them, and saves the answer. `pkgctl status` lists the saved answer the same way without checking again. `pkgctl status --short` prints only a one-line summary, or nothing when everything is current, so it is quick enough to run whenever a shell starts:

```fish
# ~/.config/fish/config.fish
if status is-interactive
    pkgctl status --short
end
```

When the saved answer is more than a day old, `pkgctl status` also starts `pkgctl check` in the background, so the line stays current without running `check` yourself. If the last successful check is more than a week old, `status` says so; run `pkgctl check` to see what's failing.

`pkgctl update` clears each tool it updates from the saved answer.
