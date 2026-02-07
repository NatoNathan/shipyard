# remove - Jettison cargo from the manifest

## Synopsis

```bash
shipyard remove [OPTIONS]
shipyard rm [OPTIONS]
shipyard delete [OPTIONS]
```

**Aliases:** `rm`, `delete`

## Description

The `remove` command removes one or more pending consignments from the manifest. It:

1. Validates that `--id` or `--all` is specified
2. Locates consignment files in `.shipyard/consignments/`
3. Deletes the specified files

Use `--id` to remove specific consignments by ID, or `--all` to clear all pending consignments.

**Maritime Metaphor**: Unload cargo from the manifest before setting sail—discard changes that are no longer needed.

## Global Options

These options are available for all shipyard commands:

| Option | Short | Description |
|--------|-------|-------------|
| `--config <path>` | `-c` | Config file (default: `.shipyard/shipyard.yaml`) |
| `--json` | `-j` | Output in JSON format |
| `--quiet` | `-q` | Suppress non-error output |
| `--verbose` | `-v` | Verbose output |

## Options

### `--id <consignment-id>`

Consignment ID(s) to remove. Can be repeated.

```bash
shipyard remove --id 20240101-120000-abc123
shipyard remove --id c1 --id c2
```

### `--all`

Remove all pending consignments.

```bash
shipyard remove --all
```

## Examples

### Remove Specific Consignment

```bash
shipyard remove --id 20240130-120000-abc123
```

```
✓ Removed 1 consignment(s)
  - 20240130-120000-abc123
```

### Remove Multiple by ID

```bash
shipyard remove --id 20240130-120000-abc123 --id 20240131-090000-def456
```

### Remove All Pending

```bash
shipyard remove --all
```

```
✓ Removed 3 consignment(s)
  - 20240130-120000-abc123
  - 20240131-090000-def456
  - 20240201-140000-ghi789
```

### JSON Output

```bash
shipyard remove --all --json
```

```json
{
  "removed": ["20240130-120000-abc123", "20240131-090000-def456"],
  "count": 2
}
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success - consignment(s) removed (or none to remove with `--all`) |
| 1 | Error - missing flags, consignment not found, or failed to load config |

## Behavior Details

### Flag Requirement

Either `--id` or `--all` must be specified. Running `remove` without flags returns an error.

### Not Found

If a consignment ID specified with `--id` does not exist, the command returns an error.

### Empty Directory

Running `remove --all` with no pending consignments exits successfully with a message:

```
No pending consignments to remove
```

### JSON Output

With `--json`, outputs a JSON object with `removed` (array of IDs) and `count` (number removed).

## Related Commands

- [`add`](./add.md) - Create new consignments
- [`status`](./status.md) - View pending consignments
- [`version`](./version.md) - Process consignments into versions

## See Also

- [Consignment Format](../consignment-format.md) - Structure of consignment files
