# status - Check cargo and chart your course

## Synopsis

```bash
shipyard status [OPTIONS]
```

## Description

The `status` command shows pending consignments and their calculated version bumps. It:

1. Reads pending consignments from `.shipyard/consignments/`
2. Groups them by package
3. Calculates version bumps (including dependency propagation)
4. Displays results in table or JSON format

**Maritime Metaphor**: Review pending cargo and see which ports of call (versions) await.

## Global Options

These options are available for all shipyard commands:

| Option | Short | Description |
|--------|-------|-------------|
| `--config <path>` | `-c` | Config file (default: `.shipyard/shipyard.yaml`) |
| `--json` | `-j` | Output in JSON format |
| `--quiet` | `-q` | Suppress non-error output |
| `--verbose` | `-v` | Verbose output |

## Options

### `--package <name>`, `-p`

Filter by package name(s). Can be repeated.

```bash
shipyard status --package core
shipyard status --package core --package api
```

### `--quiet`, `-q`

Minimal output.

```bash
shipyard status --quiet
```

### `--verbose`, `-v`

Verbose output with timestamps and metadata.

```bash
shipyard status --verbose
```

## Examples

### Basic Usage

```bash
shipyard status
```

```
📦 Pending consignments

core (1.2.3 → 1.3.0)
  - 20240130-120000-abc123: Add new feature (minor)
  - 20240130-110000-def456: Fix null pointer (patch)

api (2.0.0 → 2.0.1)
  - 20240130-120000-abc123: Add new feature (minor)
```

### Filter by Package

```bash
shipyard status --package core
```

### JSON Output

```bash
shipyard status --json
```

### Verbose Mode

```bash
shipyard status --verbose
```

Shows additional details like timestamps, metadata, and propagation info.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success (even with no consignments) |
| 1 | Error - not initialized or failed to read consignments |

## Behavior Details

### No Consignments

```
No pending consignments
```

Exit code: 0 (success)

### Version Calculation

Shows what version each package would become if `shipyard version` were run now. Includes:
- Direct bumps from consignments affecting the package
- Propagated bumps from dependencies

### Package Filtering

With `--package`, only shows consignments affecting those packages.

## Related Commands

- [`add`](./add.md) - Create new consignments
- [`version`](./version.md) - Process consignments into versions

## See Also

- [Consignment Format](../consignment-format.md) - Structure of consignment files