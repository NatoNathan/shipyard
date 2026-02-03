# add - Log cargo in the ship's manifest

## Synopsis

```bash
shipyard add [OPTIONS]
```

## Description

The `add` command records a new consignment (change entry) in your repository. It:

1. Validates package names and change type
2. Validates metadata against configured fields
3. Generates a unique consignment ID
4. Writes a `.md` file to `.shipyard/consignments/`

Supports both interactive mode (prompts for input) and non-interactive mode (via flags).

**Maritime Metaphor**: Log cargo in the ship's manifest—documenting what's being shipped, which vessels carry it, and how it affects the voyage.

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

Package name(s) affected by this change. Can be repeated for multiple packages.

```bash
shipyard add --package core
shipyard add --package core --package api
```

### `--type <type>`, `-t`

Change type: `patch`, `minor`, or `major`.

```bash
shipyard add --type minor
```

### `--summary <text>`, `-s`

Summary of the change.

```bash
shipyard add --summary "Add new API endpoint"
```

### `--metadata <key=value>`, `-m`

Custom metadata in `key=value` format. Can be repeated. Keys must match fields defined in `shipyard.yaml`.

```bash
shipyard add --metadata author=dev@example.com
shipyard add --metadata author=dev@example.com --metadata issue=JIRA-123
```

## Examples

### Interactive Mode

When flags are omitted, prompts guide you through the process:

```bash
shipyard add
```

### Non-Interactive Mode

Provide all required flags:

```bash
shipyard add --package core --type minor --summary "Add new feature"
```

### Multiple Packages

```bash
shipyard add --package core --package api --type major --summary "Breaking API change"
```

### With Metadata

```bash
shipyard add --package core --type patch --summary "Fix null pointer" \
  --metadata author=dev@example.com --metadata issue=BUG-456
```

### Single-Package Repository

For repos with one package, `--package` can still be omitted in interactive mode:

```bash
shipyard add --type patch --summary "Quick fix"
```

## Output

On success:

```
✓ Created consignment: 20240130-120000-abc123.md

Packages: core, api
Type:     minor
Summary:  Add new feature
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success - consignment created |
| 1 | Error - validation failed or not a git repository |

## Behavior Details

### Interactive vs Non-Interactive

- **Interactive**: If `--package`, `--type`, or `--summary` is missing, prompts for input
- **Non-Interactive**: If all three are provided, runs without prompts

### Package Validation

Package names must exist in `shipyard.yaml`. Invalid packages return an error listing available options.

### Change Type Validation

Only `patch`, `minor`, and `major` are accepted.

### Metadata Validation

If metadata fields are configured in `shipyard.yaml`, provided values are validated:
- Keys must match configured field names
- Values must match allowed options (if defined)

### Consignment ID Format

Generated as `YYYYMMDD-HHMMSS-<random>` based on current UTC time.

### Git Requirement

Must be run inside a git repository.

## Related Commands

- [`status`](./status.md) - View pending consignments
- [`version`](./version.md) - Process consignments into versions

## See Also

- [Consignment Format](../consignment-format.md) - Structure of consignment files
- [Configuration Reference](../configuration.md) - Metadata field definitions