# release-notes - Tell the tale of your voyage

## Synopsis

```bash
shipyard release-notes [OPTIONS]
```

## Description

The `release-notes` command generates release notes from version history. It:

1. Reads entries from `.shipyard/history.json`
2. Filters by package, version, or metadata
3. Renders output using a template
4. Writes to stdout or a file

**Maritime Metaphor**: Recount the journey from the captain's log—tales of ports visited and cargo delivered.

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

Filter release notes by package name. Required for multi-package repositories.

```bash
shipyard release-notes --package my-api
```

### `--output <path>`, `-o`

Write output to a file instead of stdout.

```bash
shipyard release-notes --output RELEASE_NOTES.md
```

### `--version <version>`

Generate notes for a specific version only.

```bash
shipyard release-notes --version 1.2.0
```

### `--all-versions`

Show complete history instead of just the latest version. Automatically uses the changelog template.

```bash
shipyard release-notes --all-versions
```

### `--filter <key=value>`

Filter by custom metadata. Can be repeated for multiple filters.

```bash
shipyard release-notes --filter team=platform
shipyard release-notes --filter team=platform --filter scope=api
```

### `--template <name>`

Specify which template to use. Can be a builtin name or path to a `.tmpl` file.

```bash
shipyard release-notes --template builtin:grouped
shipyard release-notes --template .shipyard/templates/custom-notes.tmpl
```

## Examples

### Latest Version (Default)

```bash
shipyard release-notes --package my-api
```

```
# my-api v1.3.0

Released: 2024-01-30

## Changes

- **feat**: Add new API endpoint
- **fix**: Fix memory leak in handler
```

### Specific Version

```bash
shipyard release-notes --package my-api --version 1.2.0
```

### Complete History

```bash
shipyard release-notes --package my-api --all-versions
```

### Write to File

```bash
shipyard release-notes --package my-api --output docs/RELEASE_NOTES.md
```

### Filter by Metadata

```bash
shipyard release-notes --package my-api --filter team=backend --all-versions
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error - missing required flags, invalid filters, or file operation failed |

## Behavior Details

### Package Requirement

For multi-package repositories, `--package` is required. For single-package repos, the package is auto-detected.

### Default Behavior

Without `--version` or `--all-versions`, only the latest version is shown.

### Template Selection

- Default (single version): `release-notes` template
- With `--all-versions`: `changelog` template
- With `--template`: Uses specified template

### Metadata Validation

Filter keys and values are validated against metadata fields defined in `shipyard.yaml`. Invalid keys or values return an error.

### Empty History

If no releases are found in history:

```
No releases found in history
```

Exit code: 0 (success)

## Related Commands

- [`version`](./version.md) - Create new versions (populates history)
- [`release`](./release.md) - Publish releases to GitHub

## See Also

- [Tag Generation Guide](../tag-generation.md) - Template customization
- [Configuration Reference](../configuration.md) - Metadata field definitions
