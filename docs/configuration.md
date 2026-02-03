# Configuration Reference

The `shipyard.yaml` file defines how Shipyard manages versioning for your repository.

**Location**: `.shipyard/shipyard.yaml`

## Full Example

```yaml
extends:
  - url: https://example.com/shared-config.yaml

packages:
  - name: my-api
    path: packages/api
    ecosystem: npm
    versionFiles:
      - package.json
    templates:
      changelog:
        source: builtin:grouped
      tagName:
        source: builtin:npm
    dependencies:
      - package: shared-types
        strategy: linked

  - name: shared-types
    path: packages/types
    ecosystem: npm

templates:
  changelog:
    source: builtin:default
  tagName:
    source: builtin:default
  releaseNotes:
    source: builtin:default
  commitMessage:
    inline: "chore: release {{.Packages}}"

metadata:
  fields:
    - name: author
      required: true
      type: string
      pattern: "^[a-z]+@example\\.com$"
    - name: issue
      type: string
      allowedValues: []
    - name: priority
      type: string
      allowedValues: [low, medium, high]
      default: medium

consignments:
  path: .shipyard/consignments

history:
  path: .shipyard/history.json

github:
  owner: myorg
  repo: myrepo
```

## Top-Level Fields

### `extends`

Extend from remote configuration sources.

```yaml
extends:
  - url: https://example.com/config.yaml
  - git: github.com/org/repo
    path: .shipyard/shared.yaml
    ref: main
```

| Field | Description |
|-------|-------------|
| `url` | HTTP(S) URL to fetch config from |
| `git` | Git repository to clone |
| `path` | Path within git repo (default: `.shipyard/shipyard.yaml`) |
| `ref` | Git ref to checkout (branch, tag, commit) |

### `packages`

List of versionable packages in the repository.

```yaml
packages:
  - name: my-package
    path: ./
    ecosystem: go
```

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Unique package identifier |
| `path` | Yes | Path to package root (relative to repo root) |
| `ecosystem` | No | Package ecosystem (auto-detected if omitted) |
| `versionFiles` | No | Files to update with version (auto-detected) |
| `dependencies` | No | Other packages this depends on |
| `templates` | No | Package-specific template overrides |

#### Ecosystems

| Value | Version File | Description |
|-------|--------------|-------------|
| `go` | `VERSION` file | Go modules (or tag-only) |
| `npm` | `package.json` | Node.js packages |
| `python` | `__init__.py` or `setup.py` | Python packages |
| `helm` | `Chart.yaml` | Helm charts |
| `cargo` | `Cargo.toml` | Rust crates |
| `deno` | `deno.json` | Deno modules |

#### Tag-Only Mode

For packages that don't need version files updated (e.g., Go modules):

```yaml
packages:
  - name: my-lib
    path: ./
    ecosystem: go
    versionFiles:
      - tag-only
```

#### Dependencies

```yaml
dependencies:
  - package: shared-types
    strategy: linked
    bumpMapping:
      major: major
      minor: minor
      patch: patch
```

| Field | Description |
|-------|-------------|
| `package` | Name of the dependency package |
| `strategy` | `linked` (same bump) or `fixed` (patch bump) |
| `bumpMapping` | Custom mapping of dependency bumps to this package |

### `templates`

Global template configuration. Package-level templates override these.

```yaml
templates:
  changelog:
    source: builtin:default
  tagName:
    source: builtin:go
  releaseNotes:
    inline: |
      # {{.Package}} {{.Version}}
      {{range .Consignments}}- {{.Summary}}
      {{end}}
  commitMessage:
    source: .shipyard/templates/commit.tmpl
```

Each template accepts either:
- `source`: Path to `.tmpl` file or builtin name
- `inline`: Template content directly in YAML

#### Builtin Templates

| Template | Builtins Available |
|----------|-------------------|
| `changelog` | `builtin:default`, `builtin:grouped` |
| `tagName` | `builtin:default`, `builtin:go`, `builtin:npm` |
| `releaseNotes` | `builtin:default` |
| `commitMessage` | `builtin:default` |

### `metadata`

Define custom metadata fields for consignments.

```yaml
metadata:
  fields:
    - name: author
      required: true
      type: string
    - name: priority
      type: string
      allowedValues: [low, medium, high]
      default: medium
```

#### Field Properties

| Property | Description |
|----------|-------------|
| `name` | Field name (used in `--metadata key=value`) |
| `required` | Whether field must be provided |
| `type` | `string`, `int`, `list`, or `map` |
| `default` | Default value if not provided |
| `description` | Help text for interactive prompts |
| `allowedValues` | List of valid values |

#### String Validation

| Property | Description |
|----------|-------------|
| `pattern` | Regex pattern to match |
| `minLength` | Minimum string length |
| `maxLength` | Maximum string length |

#### Integer Validation

| Property | Description |
|----------|-------------|
| `min` | Minimum value |
| `max` | Maximum value |

#### List Validation

| Property | Description |
|----------|-------------|
| `itemType` | Type of list items (`string` or `int`) |
| `minItems` | Minimum number of items |
| `maxItems` | Maximum number of items |

### `consignments`

Configure consignment storage.

```yaml
consignments:
  path: .shipyard/consignments
```

| Field | Default | Description |
|-------|---------|-------------|
| `path` | `.shipyard/consignments` | Directory for pending consignments |

### `history`

Configure version history storage.

```yaml
history:
  path: .shipyard/history.json
```

| Field | Default | Description |
|-------|---------|-------------|
| `path` | `.shipyard/history.json` | Path to history file |

### `github`

GitHub integration settings for the `release` command.

```yaml
github:
  owner: myorg
  repo: myrepo
```

| Field | Description |
|-------|-------------|
| `owner` | GitHub organization or username |
| `repo` | Repository name |

**Note**: The `GITHUB_TOKEN` environment variable must be set for GitHub operations.

## Minimal Configuration

For a single-package repository:

```yaml
packages:
  - name: my-app
    path: ./
    ecosystem: npm
```

Everything else uses sensible defaults.

## See Also

- [`init`](./reference/init.md) - Generate configuration interactively
- [Consignment Format](./consignment-format.md) - Consignment file structure
- [Tag Generation Guide](./tag-generation.md) - Template customization