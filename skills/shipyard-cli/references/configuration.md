# Shipyard Configuration Reference

This document provides a comprehensive reference for the `.shipyard/shipyard.yaml` configuration file.

## Configuration File Location

- **Default**: `.shipyard/shipyard.yaml`
- **Custom**: Use `--config <path>` flag to specify alternative location

## Configuration Structure

```yaml
# Package definitions
packages:
  - name: string              # Required: Package identifier
    path: string              # Required: Path to package directory
    ecosystem: string         # Required: go, npm, python, helm, cargo, deno
    versionFiles: []string    # Optional: Custom version file paths (or ["tag-only"] for git tags only)
    options:                  # Optional: Ecosystem-specific options (map[string]interface{})
      appDependency: string   # Helm only: Package name for appVersion sync
    dependencies:             # Optional: Package dependencies
      - package: string       # Required: Dependency package name
        strategy: string      # Required: linked, fixed
    templates:                # Optional: Package-specific templates
      changelog:
        source: string
      tagName:
        source: string
      releaseNotes:
        source: string

# Global templates
templates:
  changelog:
    source: string            # Template source (builtin:*, path, URL)
    inline: string            # Inline template (alternative to source)
  tagName:
    source: string
  tagMessage:
    source: string
    inline: string
  releaseNotes:
    source: string
    inline: string
  commitMessage:
    source: string
    inline: string

# Consignment configuration
consignments:
  path: string                # Default: .shipyard/consignments
  metadataFields:             # Optional: Custom metadata fields
    - name: string            # Required: Field name
      required: boolean       # Optional: Is field required
      options: []string       # Optional: Allowed values

# History configuration
history:
  path: string                # Default: .shipyard/history.json

# GitHub integration
github:
  owner: string               # Required for releases: GitHub org/user
  repo: string                # Required for releases: Repository name
```

## Package Configuration

### Required Fields

#### name

Unique identifier for the package. Used in consignments, tags, and commands.

```yaml
packages:
  - name: my-api
```

**Rules:**
- Must be unique across all packages
- Case-sensitive
- No spaces (use hyphens or underscores)
- Used in git tags (e.g., `my-api/v1.2.3`)

#### path

Relative path from repository root to package directory.

```yaml
packages:
  - name: my-api
    path: packages/api
```

**Rules:**
- Relative to repository root
- No leading or trailing slashes
- Must exist in filesystem

#### ecosystem

Package manager/ecosystem type.

```yaml
packages:
  - name: my-api
    path: packages/api
    ecosystem: npm
```

**Supported Ecosystems:**

| Ecosystem | Version Files | Format |
|-----------|---------------|--------|
| `go` | `version.go`, `go.mod` | `const Version = "X.Y.Z"` or `// version: X.Y.Z` |
| `npm` | `package.json` | `"version": "X.Y.Z"` |
| `python` | `pyproject.toml`, `setup.py`, `__version__.py` | Various |
| `helm` | `Chart.yaml` | `version: X.Y.Z` |
| `cargo` | `Cargo.toml` | `version = "X.Y.Z"` |
| `deno` | `deno.json`, `deno.jsonc` | `"version": "X.Y.Z"` |

### Optional Fields

#### versionFile

Custom path to version file (overrides ecosystem defaults).

```yaml
packages:
  - name: my-api
    path: packages/api
    ecosystem: go
    versionFile: internal/version/version.go
```

**Use Cases:**
- Non-standard version file location
- Multiple packages sharing filesystem but different version files
- Custom version file naming

#### dependencies

Define relationships between packages for version propagation.

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
        type: linked
```

**Dependency Types:**

1. **linked** - Dependent version follows dependency version
   ```yaml
   dependencies:
     - name: shared-lib
       type: linked
   ```
   - When `shared-lib` gets minor bump, dependent gets minor bump
   - When `shared-lib` gets major bump, dependent gets major bump
   - Keeps versions synchronized

2. **fixed** - Dependent uses exact dependency version, requires manual update
   ```yaml
   dependencies:
     - name: shared-lib
       type: fixed
   ```
   - When `shared-lib` changes, dependent gets patch bump
   - Signals that dependency version needs review

**Multiple Dependencies:**

```yaml
packages:
  - name: web-app
    path: apps/web
    ecosystem: npm
    dependencies:
      - name: api-client
        type: linked
      - name: ui-components
        type: linked
      - name: utils
        type: fixed
```

**Rules:**
- Dependency package must be defined in configuration
- Circular dependencies are detected and reported
- Dependencies are resolved in topological order

#### templates (Package-Specific)

Override global templates for specific packages.

```yaml
packages:
  - name: go-cli
    path: packages/cli
    ecosystem: go
    templates:
      tagName:
        source: builtin:go        # v-prefixed tags

  - name: npm-lib
    path: packages/lib
    ecosystem: npm
    templates:
      tagName:
        source: builtin:npm       # package-prefixed tags
```

**Precedence:**
1. Package-specific templates (highest)
2. Global templates
3. Built-in defaults (lowest)

#### options

Ecosystem-specific configuration options.

```yaml
packages:
  - name: myapp
    path: ./
    ecosystem: go

  - name: myapp-chart
    path: ./charts/myapp
    ecosystem: helm
    options:
      appDependency: myapp  # Helm-specific option
    dependencies:
      - package: myapp
        strategy: linked
```

**Helm Options:**

##### appDependency

Synchronizes the chart's `appVersion` field to a dependency package's version.

```yaml
options:
  appDependency: myapp  # Package name to track for appVersion
```

**How It Works:**
- Chart's `version` field: Tracks the chart's own semantic version
- Chart's `appVersion` field: Syncs to the specified package's version

**Example Workflow:**

Initial state:
```yaml
# Chart.yaml
version: 0.1.0
appVersion: "1.0.0"
```

After `myapp` bumps to 1.2.0:
```yaml
# Chart.yaml
version: 0.2.0      # Propagated minor bump (linked dependency)
appVersion: "1.2.0"  # Synced from myapp
```

After chart-only change (update labels):
```yaml
# Chart.yaml
version: 0.2.1      # Chart's patch bump
appVersion: "1.2.0"  # Unchanged (still tracking myapp)
```

**Use Cases:**
- Helm charts that deploy applications
- Separating chart version from application version
- Tracking deployed application version in chart metadata
- Independent evolution of chart and application versions

**Validation:**
- appDependency package must exist in configuration
- Configuration validation enforces this at `shipyard validate`

## Template Configuration

Templates control output format for changelogs, tags, and release notes.

### Template Sources

#### Built-in Templates

```yaml
templates:
  changelog:
    source: builtin:default
```

**Available Built-ins:**

- `builtin:default` - Standard format
- `builtin:grouped` - Grouped by change type
- `builtin:go` - Go module style (v-prefixed tags)
- `builtin:npm` - NPM style (package-prefixed tags)

#### File Templates

```yaml
templates:
  changelog:
    source: ./templates/changelog.tmpl
```

**Rules:**
- Path relative to repository root
- Must be valid Go template
- File must exist

#### Remote Templates

```yaml
templates:
  changelog:
    source: https://raw.githubusercontent.com/org/repo/main/templates/changelog.tmpl
```

**Supported:**
- HTTPS URLs
- Git repository URLs
- GitHub raw URLs

#### Inline Templates

```yaml
templates:
  releaseNotes:
    inline: |
      # {{.Package}} {{.Version}}

      {{range .Consignments}}
      - {{.Summary}}
      {{end}}
```

**Use Cases:**
- Simple templates
- Quick customization
- No external file needed

### Template Types

#### changelog

Generates CHANGELOG.md content.

```yaml
templates:
  changelog:
    source: builtin:grouped
```

**Template Data:**
```go
{
  Package: string
  Version: string
  Consignments: []Consignment
  AllVersions: []HistoryEntry
}
```

#### tagName

Generates git tag names.

```yaml
templates:
  tagName:
    source: builtin:npm
```

**Output Examples:**
- `builtin:go` → `v1.2.3`
- `builtin:npm` → `my-package/v1.2.3`
- Custom → `release/my-package/1.2.3`

**Template Data:**
```go
{
  Package: string
  Version: string
}
```

#### tagMessage

Generates git tag annotation messages.

```yaml
templates:
  tagMessage:
    inline: |
      Release {{.Package}} {{.Version}}

      {{range .Consignments}}
      - {{.Summary}}
      {{end}}
```

**Template Data:**
```go
{
  Package: string
  Version: string
  Consignments: []Consignment
}
```

#### releaseNotes

Generates GitHub release notes.

```yaml
templates:
  releaseNotes:
    source: builtin:default
```

**Template Data:**
```go
{
  Package: string
  Version: string
  Tag: string
  Consignments: []Consignment
}
```

#### commitMessage

Generates version bump commit messages.

```yaml
templates:
  commitMessage:
    inline: |
      chore: release {{.Packages}}

      {{range .PackageVersions}}
      - {{.Package}}: {{.OldVersion}} → {{.Version}} ({{.ChangeType}})
      {{end}}
```

**Template Data:**
```go
{
  Packages: string              // Comma-separated list
  PackageVersions: []PackageVersion
}
```

### Template Functions

Templates have access to these functions:

**String Functions:**
- `upper`, `lower`, `title` - Case conversion
- `trim`, `trimPrefix`, `trimSuffix` - Whitespace/string trimming
- `replace` - String replacement
- `split`, `join` - String splitting/joining

**Collection Functions:**
- `range` - Iterate over collections
- `len` - Collection length
- `index` - Access by index

**Date Functions:**
- `now` - Current timestamp
- `date` - Format dates

**Conditional Functions:**
- `if`, `else` - Conditionals
- `eq`, `ne`, `lt`, `gt` - Comparisons

## Consignment Configuration

### path

Directory for consignment files.

```yaml
consignments:
  path: .shipyard/consignments
```

**Default:** `.shipyard/consignments`

### metadataFields

Define custom metadata fields for consignments.

```yaml
consignments:
  metadataFields:
    - name: author
      required: true
    - name: issue
      required: false
    - name: reviewer
      required: false
    - name: priority
      options:
        - high
        - medium
        - low
```

**Field Properties:**

- `name` (required) - Field identifier
- `required` (optional) - Is field mandatory
- `options` (optional) - Allowed values (enum)

**Usage:**

```bash
shipyard add --metadata author=dev@example.com --metadata issue=JIRA-123
```

**Validation:**
- Fields not in configuration are rejected
- Required fields must be provided
- Values must match options if defined

## History Configuration

### path

Location of history archive file.

```yaml
history:
  path: .shipyard/history.json
```

**Default:** `.shipyard/history.json`

**Format:**

```json
[
  {
    "version": "1.2.3",
    "package": "my-api",
    "tag": "my-api/v1.2.3",
    "timestamp": "2024-01-15T10:30:00Z",
    "consignments": [
      {
        "id": "20240115-103000-abc123",
        "summary": "Add new feature",
        "changeType": "minor"
      }
    ]
  }
]
```

## GitHub Configuration

### owner

GitHub organization or username.

```yaml
github:
  owner: myorg
```

**Required for:** `shipyard release` command

### repo

GitHub repository name.

```yaml
github:
  repo: myrepo
```

**Required for:** `shipyard release` command

**Authentication:**

Set `GITHUB_TOKEN` environment variable with `repo` scope:

```bash
export GITHUB_TOKEN=ghp_xxxxxxxxxxxxx
shipyard release --package my-api
```

## Configuration Examples

### Single Package Repository

```yaml
packages:
  - name: my-app
    path: .
    ecosystem: go

templates:
  tagName:
    source: builtin:go
  changelog:
    source: builtin:default

history:
  path: .shipyard/history.json

github:
  owner: myorg
  repo: my-app
```

### Monorepo with Dependencies

```yaml
packages:
  - name: shared-types
    path: packages/types
    ecosystem: npm

  - name: api
    path: packages/api
    ecosystem: npm
    dependencies:
      - name: shared-types
        type: linked

  - name: sdk
    path: packages/sdk
    ecosystem: npm
    dependencies:
      - name: api
        type: linked
      - name: shared-types
        type: linked

  - name: web
    path: apps/web
    ecosystem: npm
    dependencies:
      - name: sdk
        type: linked

templates:
  tagName:
    source: builtin:npm
  changelog:
    source: builtin:grouped

github:
  owner: myorg
  repo: monorepo
```

### Multi-Ecosystem Monorepo

```yaml
packages:
  - name: go-api
    path: services/api
    ecosystem: go
    templates:
      tagName:
        source: builtin:go

  - name: python-worker
    path: services/worker
    ecosystem: python
    versionFiles:
      - worker/__version__.py

  - name: go-api-chart
    path: deploy/charts/api
    ecosystem: helm
    options:
      appDependency: go-api  # Chart appVersion tracks go-api version
    dependencies:
      - package: go-api
        strategy: linked      # Chart version bumps when go-api bumps

  - name: rust-cli
    path: cli
    ecosystem: cargo

templates:
  changelog:
    source: ./templates/changelog.tmpl
  releaseNotes:
    inline: |
      # {{.Package}} {{.Version}}
      {{range .Consignments}}
      - [{{.ChangeType}}] {{.Summary}}
      {{end}}

consignments:
  metadataFields:
    - name: author
      required: true
    - name: issue
      required: false

github:
  owner: myorg
  repo: platform
```

### Custom Metadata Fields

```yaml
packages:
  - name: my-api
    path: .
    ecosystem: npm

consignments:
  metadataFields:
    - name: author
      required: true
    - name: issue
      required: true
    - name: reviewer
      required: false
    - name: category
      required: false
      options:
        - feature
        - bugfix
        - security
        - performance
    - name: breaking
      required: false
      options:
        - "true"
        - "false"

templates:
  releaseNotes:
    inline: |
      # {{.Package}} {{.Version}}

      {{range .Consignments}}
      ## {{.Metadata.category | upper}}

      {{.Summary}}

      - Author: {{.Metadata.author}}
      - Issue: {{.Metadata.issue}}
      {{if .Metadata.breaking}}⚠️ BREAKING CHANGE{{end}}
      {{end}}
```

## Configuration Validation

Validate configuration with:

```bash
shipyard validate
```

**Checks:**
- YAML syntax
- Required fields present
- Package names unique
- Dependency references valid
- No circular dependencies
- Template sources accessible
- Metadata field definitions valid

**Common Errors:**

- **"Package not found"** - Dependency references non-existent package
- **"Circular dependency detected"** - Dependency cycle exists
- **"Duplicate package name"** - Multiple packages with same name
- **"Template not found"** - Template file doesn't exist
- **"Invalid ecosystem"** - Unsupported ecosystem type

## Remote Configuration

Load base configuration from remote URL:

```bash
shipyard init --remote https://github.com/myorg/shipyard-config/main/base.yaml
```

**Local configuration extends remote:**

```yaml
# Loaded from remote
packages:
  - name: shared-types
    path: packages/types
    ecosystem: npm

# Extended locally
packages:
  - name: my-app
    path: .
    ecosystem: npm
    dependencies:
      - name: shared-types
        type: linked
```

**Use Cases:**
- Organization-wide standards
- Shared templates
- Common metadata fields
- Base configurations

## Best Practices

### Package Organization

- Use consistent naming (lowercase with hyphens)
- Group related packages in monorepo
- Define dependencies explicitly
- Use semantic paths (packages/, apps/, services/)

### Template Management

- Start with built-in templates
- Customize only when needed
- Store custom templates in version control
- Test templates with `--preview`

### Metadata Fields

- Define organization-wide fields in remote config
- Make critical fields required (author, issue)
- Use options for enum fields
- Keep field names simple and consistent

### Dependency Strategy

- Use `linked` for tightly coupled packages
- Use `fixed` for loosely coupled packages
- Document dependency relationships
- Test propagation with `--preview`

### GitHub Integration

- Store token in CI/CD secrets
- Use fine-grained personal access tokens
- Limit token scope to `repo`
- Set owner/repo once in configuration
