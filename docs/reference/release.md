# release - Signal arrival at port

## Synopsis

```bash
shipyard release [OPTIONS]
shipyard publish [OPTIONS]
```

**Aliases:** `publish`

## Description

The `release` command publishes a version release to GitHub. It:

1. Reads version history from `.shipyard/history.json`
2. Finds the entry for the specified package/tag
3. Generates release notes from the history entry
4. Creates a GitHub release using the existing git tag

**Prerequisites**: Run `shipyard version` first to create version tags, then push them with `git push --tags`.

**Maritime Metaphor**: Announce your arrival at port—declare the cargo delivered and the voyage complete.

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

Package to release. Required for multi-package repositories.

```bash
shipyard release --package my-api
```

### `--draft`

Create as a draft release (not published publicly).

```bash
shipyard release --draft
```

### `--prerelease`

Mark the release as a prerelease.

```bash
shipyard release --prerelease
```

### `--tag <tag>`

Use a specific tag instead of the latest for the package.

```bash
shipyard release --tag my-api/v1.2.0
```

## Configuration

GitHub settings must be configured in `.shipyard/shipyard.yaml`:

```yaml
github:
  owner: myorg
  repo: myrepo
```

The `GITHUB_TOKEN` environment variable must be set with appropriate permissions.

## Examples

### Basic Usage

```bash
shipyard release --package my-api
```

```
✓ Release published successfully
  Package: my-api
  Version: 1.3.0
  Tag: my-api/v1.3.0
  URL: https://github.com/myorg/myrepo/releases/tag/my-api/v1.3.0
```

### Draft Release

```bash
shipyard release --package my-api --draft
```

### Specific Version

```bash
shipyard release --package my-api --tag my-api/v1.2.0
```

### Single-Package Repository

For repos with one package, `--package` is auto-detected:

```bash
shipyard release
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success - release published |
| 1 | Error - missing config, missing token, tag not found, or API failure |

## Behavior Details

### Package Requirement

For multi-package repositories, `--package` is required. Single-package repos auto-detect.

### Tag Selection

Without `--tag`, uses the most recent release for the package from history.

### Release Notes

Generated automatically from the history entry using the `releaseNotes` template.

### Release Title

Extracted from the first line of the release notes. Falls back to `{package} v{version}` if:
- Release notes are empty
- First line is a markdown heading (`#`)

### GitHub Token

Requires `GITHUB_TOKEN` environment variable with `repo` scope permissions.

### Tag Must Exist

The git tag must already exist locally and be pushed to the remote. Run:

```bash
shipyard version
git push --tags
shipyard release --package my-api
```

## Related Commands

- [`version`](./version.md) - Create version tags
- [`release-notes`](./release-notes.md) - Generate release notes manually

## See Also

- [Configuration Reference](../configuration.md) - GitHub settings