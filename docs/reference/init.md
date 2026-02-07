# init - Set sail - prepare your repository

## Synopsis

```bash
shipyard init [OPTIONS]
shipyard setup [OPTIONS]
```

**Aliases:** `setup`

## Description

The `init` command prepares a repository for versioning with Shipyard. It:

1. Verifies the current directory is a git repository
2. Creates the `.shipyard/` directory structure
3. Detects packages in the repository
4. Generates `shipyard.yaml` configuration
5. Initializes an empty `history.json`

Supports interactive mode (prompts for configuration) and non-interactive mode (`--yes`).

**Maritime Metaphor**: Prepare your repository for the versioning voyage ahead—set up cargo manifests, navigation charts, and the captain's log.

## Global Options

These options are available for all shipyard commands:

| Option | Short | Description |
|--------|-------|-------------|
| `--config <path>` | `-c` | Config file (default: `.shipyard/shipyard.yaml`) |
| `--json` | `-j` | Output in JSON format |
| `--quiet` | `-q` | Suppress non-error output |
| `--verbose` | `-v` | Verbose output |

## Options

### `--force`, `-f`

Force re-initialization if already initialized.

```bash
shipyard init --force
```

### `--remote <url>`, `-r`

Extend from a remote configuration URL.

```bash
shipyard init --remote https://example.com/shipyard-config.yaml
```

### `--yes`, `-y`

Skip all prompts and accept defaults. Uses auto-detected packages.

```bash
shipyard init --yes
```

## Examples

### Interactive Mode (Default)

```bash
shipyard init
```

Prompts for:
- Repository type (single package or monorepo)
- Package selection/configuration
- Package details (name, path, ecosystem)

### Non-Interactive Mode

```bash
shipyard init --yes
```

Uses auto-detected packages or creates a default package if none found.

### Re-initialize

```bash
shipyard init --force
```

### With Remote Config

```bash
shipyard init --remote https://example.com/shared-config.yaml
```

## Output

On success:

```
✓ Shipyard initialized successfully

Configuration:          .shipyard/shipyard.yaml
Consignments directory: .shipyard/consignments
History file:           .shipyard/history.json
```

## Created Files

| Path | Description |
|------|-------------|
| `.shipyard/shipyard.yaml` | Main configuration file |
| `.shipyard/consignments/` | Directory for pending consignments |
| `.shipyard/history.json` | Version history (empty array initially) |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success - repository initialized |
| 1 | Error - not a git repo, already initialized, or file operation failed |

## Behavior Details

### Package Detection

Automatically detects packages by looking for:
- `package.json` (npm)
- `go.mod` (Go)
- `Cargo.toml` (Cargo)
- `Chart.yaml` (Helm)
- `setup.py` / `pyproject.toml` (Python)
- `deno.json` (Deno)

### Already Initialized

Without `--force`, returns an error if `.shipyard/shipyard.yaml` exists.

### Git Requirement

Must be run inside a git repository.

### Default Package

If no packages are detected in `--yes` mode, creates a default package:

```yaml
packages:
  - name: default
    path: ./
    ecosystem: go
```

## Related Commands

- [`add`](./add.md) - Create consignments after initialization
- [`status`](./status.md) - View pending consignments

## See Also

- [Configuration Reference](../configuration.md) - Full shipyard.yaml format
- [Getting Started](../getting-started.md) - First-time setup guide