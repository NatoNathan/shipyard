# validate - Inspect the hull before departure

## Synopsis

```bash
shipyard validate [OPTIONS]
shipyard check [OPTIONS]
shipyard lint [OPTIONS]
```

**Aliases:** `check`, `lint`

## Description

The `validate` command checks the health of your shipyard setup. It:

1. Loads and validates the configuration file
2. Validates dependency references between packages
3. Parses all pending consignment files for errors
4. Builds the dependency graph and checks for cycles

Reports errors and warnings found during validation.

**Maritime Metaphor**: Inspect the hull and rigging before departure—ensure everything is seaworthy.

## Global Options

These options are available for all shipyard commands:

| Option | Short | Description |
|--------|-------|-------------|
| `--config <path>` | `-c` | Config file (default: `.shipyard/shipyard.yaml`) |
| `--json` | `-j` | Output in JSON format |
| `--quiet` | `-q` | Suppress non-error output |
| `--verbose` | `-v` | Verbose output |

## Examples

### Basic Usage

```bash
shipyard validate
```

```
✓ Validation passed
```

### JSON Output

```bash
shipyard validate --json
```

```json
{
  "valid": true,
  "errors": [],
  "warnings": []
}
```

### Quiet Mode

```bash
shipyard validate --quiet
```

Exits silently with code 0 on success, or code 1 with "validation failed" on failure.

### With Validation Errors

```bash
shipyard validate
```

```
Errors:
  - config validation: package "core" references unknown dependency "missing-lib"
  - consignment 20240130-120000-abc123.md: unknown package "nonexistent"

Validation failed
```

### With Warnings

```bash
shipyard validate
```

```
Warnings:
  - dependency cycle detected: core -> api -> core

✓ Validation passed
```

Cycles are reported as warnings, not errors.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Validation passed (warnings may be present) |
| 1 | Validation failed - errors found in config, consignments, or dependencies |

## Behavior Details

### What Is Validated

| Check | Category | Severity |
|-------|----------|----------|
| Config file loads successfully | Config | Error |
| Config passes schema validation | Config | Error |
| Package dependency references exist | Dependencies | Error |
| Consignment files parse correctly | Consignments | Error |
| Dependency graph has no cycles | Graph | Warning |

### Quiet Mode

With `--quiet`, produces no output on success. On failure, outputs "validation failed" and exits with code 1.

### JSON Output

With `--json`, outputs a JSON object:

```json
{
  "valid": false,
  "errors": ["config validation: ..."],
  "warnings": ["dependency cycle detected: ..."]
}
```

### Warnings vs Errors

- **Errors** cause validation to fail (exit code 1)
- **Warnings** are informational only (validation still passes)

Currently, dependency cycles are the only condition that produces warnings.

## Related Commands

- [`config show`](./config-show.md) - Display resolved configuration
- [`status`](./status.md) - View pending consignments and version bumps

## See Also

- [Configuration Reference](../configuration.md) - Configuration file format
- [Consignment Format](../consignment-format.md) - Structure of consignment files
