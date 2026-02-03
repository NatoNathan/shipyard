# completion - Teach your shell to speak Shipyard

## Synopsis

```bash
shipyard completion [bash|zsh|fish|powershell]
```

## Description

The `completion` command generates shell completion scripts. It enables your shell to suggest commands, flags, and arguments as you type.

Supports Bash, Zsh, Fish, and PowerShell.

**Maritime Metaphor**: Train your shell to understand the shipyard's language—let your navigator suggest the course.

## Global Options

These options are available for all shipyard commands:

| Option | Short | Description |
|--------|-------|-------------|
| `--config <path>` | `-c` | Config file (default: `.shipyard/shipyard.yaml`) |
| `--json` | `-j` | Output in JSON format |
| `--quiet` | `-q` | Suppress non-error output |
| `--verbose` | `-v` | Verbose output |

## Arguments

The shell name is required. Valid values: `bash`, `zsh`, `fish`, `powershell`.

## Installation

### Bash

Linux:

```bash
shipyard completion bash > /etc/bash_completion.d/shipyard
```

macOS:

```bash
shipyard completion bash > /usr/local/etc/bash_completion.d/shipyard
```

Or add to `~/.bashrc`:

```bash
echo 'source <(shipyard completion bash)' >> ~/.bashrc
```

### Zsh

```bash
shipyard completion zsh > "${fpath[1]}/_shipyard"
```

Or add to `~/.zshrc`:

```bash
echo 'source <(shipyard completion zsh)' >> ~/.zshrc
echo 'compdef _shipyard shipyard' >> ~/.zshrc
```

### Fish

```bash
shipyard completion fish > ~/.config/fish/completions/shipyard.fish
```

### PowerShell

Add to your profile:

```powershell
shipyard completion powershell | Out-String | Invoke-Expression
```

Or save to profile:

```powershell
shipyard completion powershell >> $PROFILE
```

After installing, restart your shell or source the completion file.

## Features

Completions include:

- Command names (`init`, `add`, `version`, etc.)
- Flag names and values
- Package names from `shipyard.yaml`
- Change types (`patch`, `minor`, `major`)

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success - completion script generated |
| 1 | Error - invalid shell name |

## Related Commands

- [`upgrade`](./upgrade.md) - Upgrade shipyard

## See Also

- [Getting Started](../getting-started.md) - Initial setup