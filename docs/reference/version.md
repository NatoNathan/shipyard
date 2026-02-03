# version - Set sail to the next port

## Synopsis

```bash
shipyard version [OPTIONS]
```

## Description

The `version` command processes pending consignments and creates new package versions. It:

1. Calculates new version numbers based on change types
2. Updates ecosystem-specific version files
3. Archives consignments to history
4. Generates changelogs from complete history
5. Creates a git commit and tags

**Maritime Metaphor**: The ship leaves port with its cargo (consignments), and each package reaches its next destination (version).

## Global Options

These options are available for all shipyard commands:

| Option | Short | Description |
|--------|-------|-------------|
| `--config <path>` | `-c` | Config file (default: `.shipyard/shipyard.yaml`) |
| `--json` | `-j` | Output in JSON format |
| `--quiet` | `-q` | Suppress non-error output |
| `--verbose` | `-v` | Verbose output |

## Options

### `--preview`

Show what changes would be made without applying them.

```bash
shipyard version --preview
```

### `--no-commit`

Apply version changes but skip creating a git commit. Tags are also skipped.

```bash
shipyard version --no-commit
```

### `--no-tag`

Create the commit but skip creating git tags.

```bash
shipyard version --no-tag
```

### `--package <name>`

Process consignments only for specified package(s). Can be repeated.

```bash
shipyard version --package my-api
shipyard version --package cli --package sdk
```



## Workflow

The command executes these phases:

1. **Validation** - Read and validate pending consignments
2. **Dependency Graph** - Build package dependency map
3. **Version Calculation** - Determine new versions based on change types
4. **Preview** (if `--preview`) - Display changes and exit
5. **Update Version Files** - Write new versions to ecosystem files
6. **Generate Tags** - Render tag names and messages from templates
7. **Archive Consignments** - Append to `history.json` with version context
8. **Generate Changelogs** - Regenerate from complete history (including new version)
9. **Delete Consignments** - Remove processed `.md` files
10. **Git Operations** - Create commit and tags (unless `--no-commit`)

**Note**: Changelogs are generated *after* archiving so the new version appears in the output.

## Configuration

Configuration is read from `.shipyard/shipyard.yaml`:

```yaml
packages:
  - name: my-api
    path: packages/api
    ecosystem: npm
    templates:                           # optional, overrides global
      changelog:
        source: builtin:grouped
      tagName:
        source: builtin:npm
    dependencies:                        # optional
      - name: shared-types
        type: linked  # or: fixed

templates:
  changelog:
    source: builtin:default              # or path to .tmpl file
  tagName:
    source: builtin:go                   # or path to .tmpl file
  releaseNotes:
    source: builtin:default              # or inline (see below)
    inline: |                            # alternative to source
      # {{.Package}} {{.Version}}
      {{range .Consignments}}
      - {{.Summary}}
      {{end}}

consignments:
  path: .shipyard/consignments

history:
  path: .shipyard/history.json
```

### Supported Ecosystems

- **go** - `VERSION` file (or tag-only)
- **npm** - `package.json`
- **python** - `__version__` in `__init__.py` or `setup.py`
- **helm** - `Chart.yaml`
- **cargo** - `Cargo.toml`
- **deno** - `deno.json`

### Template Options

Templates are configured globally under `templates:` with `source:` (file path or builtin) or `inline:` (embedded template).

**Builtins**:
- `builtin:default` - Available for changelog, tagName, and releaseNotes
- `builtin:go` - Go module style tags (v-prefixed)
- `builtin:npm` - NPM style tags
- `builtin:grouped` - Changelog grouped by change type

See [Tag Generation Guide](../tag-generation.md) for template details.

## Examples

### Basic Usage

```bash
shipyard version
```

```
📦 Versioning packages...
  - my-api: 1.2.3 → 1.3.0 (minor)
  - shared-types: 0.5.0 → 0.5.1 (patch)
✓ Created commit: "chore: release my-api v1.3.0, shared-types v0.5.1"
✓ Created tags: my-api/v1.3.0, shared-types/v0.5.1
```

### Preview Changes

```bash
shipyard version --preview
```

```
📦 Preview: Version changes
  - my-api: 1.2.3 → 1.3.0 (minor)
    - 20240130-120000-abc123: Add new API endpoint

ℹ Preview mode: no changes made
```

### Manual Review Before Commit

```bash
shipyard version --no-commit
git diff
git add -A && git commit -m "chore: release my-api v1.3.0"
git tag -a my-api/v1.3.0 -m "Release my-api v1.3.0"
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success (or no consignments to process) |
| 1 | Error - validation, file, or git operation failed |

## Behavior Details

### No Consignments

Exits successfully with no-op message.

### Package Filtering

With `--package`, unmatched consignments remain in `.shipyard/consignments/`.

### Version Propagation

When a dependency is versioned, dependents are also bumped:
- **linked**: Same change type as the dependency
- **fixed**: Patch bump

### Tag Format

Tags follow git commit message format:
- Single line → lightweight tag
- Multiple lines (blank line separator) → annotated tag with message body

### Git Requirements

- Repository must be initialized
- Working directory must be clean
- `user.name` and `user.email` must be configured

## Related Commands

- [`consign`](./consign.md) - Record a new change
- [`releasenotes`](./releasenotes.md) - Generate release notes from history
- [`changelog`](./changelog.md) - Generate changelog from history

## See Also

- [Tag Generation Guide](../tag-generation.md)
- [Configuration Reference](../configuration.md)
- [Consignment Format](../consignment-format.md)