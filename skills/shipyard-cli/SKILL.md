---
name: Shipyard CLI
description: This skill should be used when the user asks to "create a consignment", "add a shipyard consignment", "bump version", "shipyard version", "create release", "publish release", "generate changelog", "check shipyard status", "convert changelog to history", "initialize shipyard", or mentions Shipyard semantic versioning, monorepo versioning, consignment-based release management, or changelog conversion.
version: 1.0.0
---

# Shipyard CLI

Shipyard is a semantic versioning and release management tool for monorepos and single-package repositories. It automates version bumps, changelog generation, and release management through a concept of "consignments" (change manifests).

## Core Concepts

### Consignments

Consignments are change manifests stored as markdown files in `.shipyard/consignments/`. Each consignment records:
- Change type (major/minor/patch)
- Affected packages
- Summary and metadata
- Unique ID format: `YYYYMMDD-HHMMSS-{random6}`

Think of consignments as shipping manifests - documenting what cargo is being shipped, which vessels carry it, and how it affects the voyage.

### Version History

Processed consignments are archived to `.shipyard/history.json` with version context. This creates a complete audit trail of all changes and their associated versions.

### Maritime Metaphors

Shipyard uses maritime metaphors throughout:
- **Consignment** - Cargo manifest (change entry)
- **Version** - Setting sail to the next port
- **Release** - Signaling arrival at port
- **Init** - Preparing the ship for voyage

## Core Workflow

The typical Shipyard workflow follows three steps:

### 1. Add Consignments

Record changes as they happen:

```bash
shipyard add --package core --type minor --summary "Add new feature"
```

Interactive mode available when flags are omitted:

```bash
shipyard add
```

### 2. Calculate and Apply Versions

Process pending consignments into version bumps:

```bash
# Preview changes
shipyard version --preview

# Apply versions
shipyard version
```

This command:
1. Calculates new versions from consignments
2. Updates version files (package.json, version.go, etc.)
3. Archives consignments to history
4. Generates changelogs
5. Creates git commit and tags

### 3. Publish Releases

Create GitHub releases from versions:

```bash
shipyard release --package my-api
```

Requires tags to be pushed first:

```bash
git push --tags
```

## Command Quick Reference

| Command | Aliases | Purpose |
|---------|---------|---------|
| `init` | `setup` | Initialize Shipyard in repository |
| `add` | `consign`, `log` | Create new consignment |
| `status` | - | View pending consignments |
| `version` | `bump`, `sail` | Apply version bumps |
| `release` | `publish` | Create GitHub release |
| `release-notes` | - | Generate release notes |
| `validate` | `check`, `lint` | Validate configuration |
| `remove` | `rm` | Remove pending consignment |
| `snapshot` | - | Create pre-release version |
| `promote` | - | Promote pre-release to stable |
| `prerelease` | `pre` | Create pre-release |
| `config-show` | - | Display configuration |
| `completion` | - | Generate shell completion |
| `upgrade` | - | Upgrade Shipyard CLI |

For detailed information on each command, consult **`references/commands.md`**.

## When to Use Shipyard

### Use Shipyard When:

- Managing semantic versions in monorepos
- Coordinating releases across multiple packages
- Automating changelog generation
- Tracking changes with detailed history
- Enforcing consistent versioning practices
- Managing dependency-aware version propagation
- Creating GitHub releases automatically

### Key Use Cases:

1. **Monorepo Version Management**: Multiple packages with dependencies between them
2. **Automated Release Notes**: Generate changelogs from structured change entries
3. **Semantic Versioning Enforcement**: Explicit change type declaration
4. **Audit Trail**: Complete history of what changed, when, and why
5. **Multi-Ecosystem Support**: Go, NPM, Python, Helm, Cargo, Deno

## Configuration

Shipyard is configured via `.shipyard/shipyard.yaml`:

```yaml
packages:
  - name: my-api
    path: packages/api
    ecosystem: npm
    dependencies:
      - package: shared-types
        strategy: linked

  - name: my-api-chart
    path: charts/my-api
    ecosystem: helm
    options:
      appDependency: my-api  # Sync appVersion to my-api's version
    dependencies:
      - package: my-api
        strategy: linked

templates:
  changelog:
    source: builtin:default
  tagName:
    source: builtin:npm
```

### Package Options

Some ecosystems support additional configuration through the `options` field:

**Helm Options:**
- `appDependency`: Package name whose version should be used for the chart's `appVersion` field

For detailed configuration options, consult **`references/configuration.md`**.

## Common Workflows

### Single Package Workflow

```bash
# Initialize
shipyard init --yes

# Add changes
shipyard add --type minor --summary "Add user authentication"

# Check status
shipyard status

# Apply version
shipyard version

# Push and release
git push --tags
shipyard release
```

### Monorepo Workflow

```bash
# Initialize with multiple packages
shipyard init

# Add change affecting multiple packages
shipyard add --package api --package sdk --type major --summary "Breaking API change"

# Preview version propagation
shipyard version --preview

# Apply versions
shipyard version

# Release individual packages
git push --tags
shipyard release --package api
shipyard release --package sdk
```

For more workflow examples, consult **`references/workflows.md`**.

## Dependency Management

Shipyard supports dependency-aware version propagation:

### Dependency Types

- **linked**: Dependent bumps with the same change type as dependency
- **fixed**: Dependent requires manual update when dependency changes

### Example Configuration

```yaml
packages:
  - name: api
    path: packages/api
    ecosystem: npm

  - name: sdk
    path: packages/sdk
    ecosystem: npm
    dependencies:
      - name: api
        type: linked  # SDK version follows API version
```

When `api` gets a minor bump, `sdk` automatically gets a minor bump too.

## Ecosystem Support

Shipyard supports multiple package ecosystems:

| Ecosystem | Version File | Format |
|-----------|--------------|--------|
| Go | `version.go` or `go.mod` | `const Version = "X.Y.Z"` |
| NPM | `package.json` | `"version": "X.Y.Z"` |
| Python | `pyproject.toml`, `setup.py`, `__version__.py` | Various |
| Helm | `Chart.yaml` | `version: X.Y.Z`, `appVersion: "X.Y.Z"` |
| Cargo | `Cargo.toml` | `version = "X.Y.Z"` |
| Deno | `deno.json` | `"version": "X.Y.Z"` |

Each ecosystem has its own version file format and update logic. Shipyard detects the ecosystem automatically based on files present in the package directory.

### Helm-Specific Features

Helm charts support a special `appDependency` option that allows the chart's `appVersion` field to track a different package's version, while the chart's own `version` field tracks the chart's semantic version:

```yaml
packages:
  - name: myapp
    path: ./
    ecosystem: go

  - name: myapp-chart
    path: ./charts/myapp
    ecosystem: helm
    options:
      appDependency: myapp  # Chart's appVersion tracks myapp's version
    dependencies:
      - package: myapp
        strategy: linked    # Chart version bumps when myapp bumps
```

**How it works:**
- Chart's `version` field: Tracks the chart's own semantic version (bumps when chart changes or due to propagation)
- Chart's `appVersion` field: Syncs to the dependency package's version (e.g., the application being deployed)

**Example scenario:**
1. `myapp` is bumped to 1.2.0 → Chart.yaml gets `version: 0.2.0` (propagated) and `appVersion: "1.2.0"` (synced to myapp)
2. Chart-only change (e.g., update labels) → Chart.yaml gets `version: 0.2.1` (patch bump), `appVersion: "1.2.0"` (unchanged)

This pattern is ideal for Helm charts that deploy applications, allowing the chart version and application version to evolve independently.

## Template System

Shipyard uses Go templates for customizable output:

### Built-in Templates

- `builtin:default` - Standard format
- `builtin:grouped` - Grouped by change type
- `builtin:go` - Go module style tags (v-prefixed)
- `builtin:npm` - NPM style tags

### Custom Templates

```yaml
templates:
  changelog:
    source: ./templates/changelog.tmpl
  releaseNotes:
    inline: |
      # {{.Package}} {{.Version}}
      {{range .Consignments}}
      - {{.Summary}}
      {{end}}
```

For detailed template documentation, consult **`references/templates.md`**.

## Converting Existing Changelogs

If migrating from an existing project with a CHANGELOG.md, convert it to Shipyard's history format.

### Conversion Process

1. Parse existing CHANGELOG.md or git tags
2. Extract version, date, and changes for each release
3. Map changelog sections to change types (patch/minor/major)
4. Create consignment entries in history.json format
5. Validate with `shipyard validate`

For detailed conversion guidance including parsing strategies for Keep a Changelog format, conventional commits, and GitHub releases, consult **`references/history-conversion.md`**.

## Validation and Safety

### Before Applying Versions

```bash
# Preview changes
shipyard version --preview

# Validate configuration and consignments
shipyard validate

# Check current status
shipyard status
```

### Querying History

Use `jq` to query history.json:

```bash
# List all versions
jq -r '.[] | "\(.package) v\(.version) - \(.timestamp)"' .shipyard/history.json

# Get latest version for package
jq -r '.[] | select(.package == "my-api") | .version' .shipyard/history.json | head -n 1

# Filter by change type
jq '.[] | select(.consignments[].changeType == "major")' .shipyard/history.json
```

## Global Flags

All commands support these global flags:

| Flag | Short | Description |
|------|-------|-------------|
| `--config <path>` | `-c` | Config file (default: `.shipyard/shipyard.yaml`) |
| `--json` | `-j` | Output in JSON format |
| `--quiet` | `-q` | Suppress non-error output |
| `--verbose` | `-v` | Verbose output |

## Git Integration

Shipyard integrates deeply with git:

### Tags

Tags are created automatically during `shipyard version`:

```bash
# Go-style tags (v-prefixed)
v1.2.3

# NPM-style tags (package-prefixed)
my-api/v1.2.3

# Custom tag format via templates
```

### Commits

Version bumps create commits with detailed messages:

```bash
chore: release my-api v1.3.0, shared-types v0.5.1

- my-api: 1.2.3 → 1.3.0 (minor)
  - Add new API endpoint
  - Support for OAuth authentication

- shared-types: 0.5.0 → 0.5.1 (patch)
  - Fix type definitions
```

### GitHub Releases

Create releases automatically with `shipyard release`:

- Uses git tags
- Generates release notes from history
- Supports draft and prerelease flags
- Requires `GITHUB_TOKEN` environment variable

## Pre-releases

Shipyard supports pre-release versions:

### Create Pre-release

```bash
# Create snapshot (e.g., 1.2.3-20240315120000)
shipyard snapshot --package my-api

# Create named pre-release (e.g., 1.2.3-beta.1)
shipyard prerelease --package my-api --identifier beta
```

### Promote to Stable

```bash
# Promote pre-release to stable version
shipyard promote --package my-api
```

## Best Practices

### Consignment Creation

- Create consignments as changes are made (not batch at release time)
- Use descriptive summaries that explain the "why" not just the "what"
- Add metadata (author, issue tracker links) for traceability
- One consignment per logical change

### Change Type Selection

- **patch**: Bug fixes, documentation, performance improvements
- **minor**: New features, non-breaking changes
- **major**: Breaking changes, removed features, major refactors

### Version Management

- Preview changes before applying (`--preview`)
- Validate configuration regularly (`shipyard validate`)
- Check status frequently (`shipyard status`)
- Commit and push tags after versioning

### Monorepo Management

- Define dependencies explicitly in configuration
- Use `linked` dependencies for tightly coupled packages
- Test version propagation with `--preview`
- Release packages in dependency order

## Additional Resources

### Reference Files

For detailed information, consult:
- **`references/commands.md`** - Complete command reference for all 14 commands
- **`references/configuration.md`** - shipyard.yaml structure and options
- **`references/workflows.md`** - Common workflow patterns and CI/CD integration
- **`references/templates.md`** - Template system and customization
- **`references/history-conversion.md`** - Converting existing changelogs to history.json

## Troubleshooting

### Common Issues

**"Not a git repository"**
- Shipyard requires a git repository
- Initialize with `git init` if needed

**"Shipyard already initialized"**
- Use `--force` to reinitialize: `shipyard init --force`

**"Package not found"**
- Check package name in `.shipyard/shipyard.yaml`
- Packages are case-sensitive

**"Version file not found"**
- Ensure version file exists for ecosystem
- Check path configuration in shipyard.yaml

**"Validation failed"**
- Run `shipyard validate` for detailed errors
- Check consignment format and configuration

### Getting Help

```bash
# Command help
shipyard <command> --help

# Validate configuration
shipyard validate

# Check status
shipyard status --verbose
```

## Summary

To effectively use Shipyard:

1. **Initialize** repository with `shipyard init`
2. **Create consignments** as changes are made with `shipyard add`
3. **Check status** regularly with `shipyard status`
4. **Preview versions** before applying with `shipyard version --preview`
5. **Apply versions** when ready with `shipyard version`
6. **Push tags** to remote with `git push --tags`
7. **Create releases** with `shipyard release`

For detailed information, consult the reference files in the `references/` directory.
