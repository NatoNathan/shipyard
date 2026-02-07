# upgrade - Refit the shipyard with latest provisions

## Synopsis

```bash
shipyard upgrade [OPTIONS]
shipyard update [OPTIONS]
shipyard self-update [OPTIONS]
```

**Aliases:** `update`, `self-update`

## Description

The `upgrade` command upgrades shipyard to the latest (or specified) version. It:

1. Detects how shipyard was installed
2. Fetches the latest release from GitHub
3. Compares versions
4. Upgrades using the appropriate method

Supports Homebrew, npm, Go install, and script installations. Docker installations must be upgraded manually.

## Global Options

These options are available for all shipyard commands:

| Option | Short | Description |
|--------|-------|-------------|
| `--config <path>` | `-c` | Config file (default: `.shipyard/shipyard.yaml`) |
| `--json` | `-j` | Output in JSON format |
| `--quiet` | `-q` | Suppress non-error output |
| `--verbose` | `-v` | Verbose output |

## Options

### `--yes`, `-y`

Skip confirmation prompt.

```bash
shipyard upgrade --yes
```

### `--version <version>`

Upgrade to a specific version instead of latest.

```bash
shipyard upgrade --version v1.2.0
```

### `--force`

Force upgrade even if already on the latest version.

```bash
shipyard upgrade --force
```

### `--dry-run`

Show upgrade plan without executing.

```bash
shipyard upgrade --dry-run
```

## Examples

### Basic Usage

```bash
shipyard upgrade
```

```
Checking installation... ✓
Checking for updates... ✓

Current version: 1.2.0
Latest version:  1.3.0

Upgrade shipyard from 1.2.0 to 1.3.0? [Y/n]
```

### Skip Confirmation

```bash
shipyard upgrade --yes
```

### Preview Upgrade

```bash
shipyard upgrade --dry-run
```

### Force Reinstall

```bash
shipyard upgrade --force
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success - upgraded or already on latest |
| 1 | Error - detection failed, network error, or upgrade failed |

## Behavior Details

### Installation Detection

Automatically detects:
- **Homebrew** - Uses `brew upgrade`
- **npm** - Uses `npm update -g`
- **Go install** - Uses `go install`
- **Script install** - Downloads binary from GitHub releases

### Docker

Docker installations cannot be auto-upgraded:

```
Cannot upgrade: Docker installations must be upgraded manually

To upgrade Docker installations:
  docker pull natonathan/shipyard:latest
```

### Already on Latest

```
✓ Already on latest version (1.3.0)
```

Use `--force` to reinstall anyway.

### Network Requirements

Requires internet access to fetch release information from GitHub.

## Related Commands

- [`completion`](./completion.md) - Shell completions

## See Also

- [GitHub Releases](https://github.com/NatoNathan/shipyard/releases) - All versions