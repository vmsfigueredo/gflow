# gflow

Git Flow for modular repositories — manage feature, hotfix, bugfix, and release branches across multiple repos or submodules simultaneously.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8.svg)](https://golang.org)

## Installation

### One-liner (recommended)

```sh
curl -fsSL https://raw.githubusercontent.com/vmsfigueredo/gflow/main/install.sh | sh
```

Installs to `/usr/local/bin` by default. Override with `INSTALL_DIR`:

```sh
INSTALL_DIR=$HOME/.local/bin curl -fsSL https://raw.githubusercontent.com/vmsfigueredo/gflow/main/install.sh | sh
```

The script:
1. Detects your OS and architecture (macOS/Linux, amd64/arm64)
2. Fetches the latest release from GitHub
3. Downloads and extracts the binary
4. Installs it to `INSTALL_DIR`

### Manual download

Grab a binary from [Releases](https://github.com/vmsfigueredo/gflow/releases), extract, and place in your `$PATH`.

### Verify

```sh
gflow version
```

## Prerequisites

- Git
- [git-flow](https://github.com/petervanderdoes/gitflow-avh) (AVH edition recommended)

```sh
# macOS
brew install git-flow-avh

# Ubuntu/Debian
apt-get install git-flow

# Fedora
dnf install gitflow
```

## Quick start

```sh
# Start a feature across all modules
gflow feature start HP-2100

# Finish it
gflow feature finish HP-2100

# Start a hotfix
gflow hotfix start v1.2.3

# Show status across all modules
gflow status
```

## Configuration

Create `.gflow.conf` in your project root:

```yaml
modules:
  - path: api
    alias: api
  - path: services/web
    alias: web
```

Without a config file, gflow auto-detects modules from `.gitmodules`.

## Commands

| Command | Description |
|---|---|
| `gflow feature start [name]` | Start feature branch across modules |
| `gflow feature finish [name]` | Finish feature branch |
| `gflow feature publish [name]` | Publish to remote |
| `gflow hotfix start [name]` | Start hotfix (auto-syncs main + develop) |
| `gflow hotfix finish [name]` | Finish hotfix |
| `gflow bugfix start [name]` | Start bugfix branch |
| `gflow release start [name]` | Start release branch |
| `gflow pull [branch]` | Pull branch across modules |
| `gflow push [branch]` | Push branch across modules |
| `gflow checkout [branch]` | Checkout branch across modules |
| `gflow commit` | Commit across modules |
| `gflow status` | Show git status across modules |
| `gflow list` | List detected modules |
| `gflow pr` | Open pull requests |
| `gflow update` | Self-update binary |
| `gflow completion` | Generate shell completion |

### Common flags

| Flag | Description |
|---|---|
| `-i, --interactive` | Force interactive module picker |
| `--names api=v1.2,web=v1.3` | Per-module branch names |
| `-v, --verbose` | Stream raw git output |
| `-P, --project <alias>` | Filter by project alias (repeatable) |

## Shell completion

```sh
# bash
gflow completion bash >> ~/.bashrc

# zsh
gflow completion zsh >> ~/.zshrc

# fish
gflow completion fish > ~/.config/fish/completions/gflow.fish
```

## Update

```sh
gflow update
```

Auto-detects whether installed via Homebrew or binary and updates accordingly. Verifies SHA256 before replacing.

## License

MIT
