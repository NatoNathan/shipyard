# Tag Generation in Shipyard

## Overview

Shipyard generates **two types of git tags** for monorepo releases:

1. **Per-Package Tags**: One tag per versioned package (e.g., `core/v1.2.0`, `api/v2.0.0`)
2. **Release/Consignment Tag**: One tag marking the release event (e.g., `release-20260130-143000`)

## Per-Package Tags

Each package that receives a version bump gets its own tag. This is essential for:
- **Go modules**: `go get github.com/org/repo/core@v1.2.0`
- **NPM packages**: Package-specific version tracking
- **Changelog references**: Direct links to package versions

### Built-in Templates

Tag templates can output:
- **Single line**: Lightweight tag (name only)
- **Multi-line** (name + blank + message): Annotated tag with message

**`builtin:default`** - Simple version tag (lightweight)
```
v1.2.0
```

**`builtin:go`** - Go module monorepo tag (lightweight)
```
core/v1.2.0
```

**`builtin:npm`** - NPM package tag (lightweight)
```
@scope/package@1.2.0
```

**`builtin:go-annotated`** - Go module tag with release notes (annotated)
```
core/v1.2.0

# Release core v1.2.0

- Added OAuth2 support (alice@example.com)
- Fixed validation bug
```

**`builtin:detailed-annotated`** - Detailed annotated tag with metadata
```
core/v1.2.0

# Release core v1.2.0

Released on 2026-01-30

## Changes

### Minor

Added OAuth2 support

**Author**: alice@example.com
**Issue**: FEAT-123
```

Context for all per-package tags:
```go
{
  Package: "core",
  Version: "1.2.0",
  Consignments: [...], // Filtered to this package
  Date: time.Now(),
  Metadata: {...}
}
```

### API

```go
// Generate tag for a single package
// Returns: (tagName, message, error)
tagName, message, err := generator.GeneratePackageTag(
    consignments, // Filtered to package automatically
    "core",
    semver.Version{Major: 1, Minor: 2, Patch: 0},
    "builtin:go",
)
// Returns: ("core/v1.2.0", "", nil) - lightweight tag

// Or with annotated template
tagName, message, err := generator.GeneratePackageTag(
    consignments,
    "core",
    semver.Version{Major: 1, Minor: 2, Patch: 0},
    "builtin:go-annotated",
)
// Returns: ("core/v1.2.0", "# Release core v1.2.0\n...", nil) - annotated tag

// Generate tags for all packages
tags, err := generator.GenerateAllPackageTags(
    consignments,
    map[string]semver.Version{
        "core": {Major: 1, Minor: 2, Patch: 0},
        "api":  {Major: 2, Minor: 0, Patch: 0},
    },
    "builtin:go",
)
// Returns: {"core": PackageTag{Name: "core/v1.2.0", Message: ""}, ...}
```

## Release/Consignment Tags

A single tag marking the entire release/consignment. This is useful for:
- **Release tracking**: Mark the release event in history
- **Deployment**: Tag what was deployed together
- **Changelog**: Reference the release as a whole

Release tags use the `TemplateTypeRelease` template type and have access to the full consignment context.

### Built-in Templates

**`builtin:date`** - Timestamp-based release tag (lightweight)
```
release-20260130-143000
```

**`builtin:versions`** - Compound version tag (lightweight)
```
release-core-1.2.0-api-2.0.0
```

Context for all release tags:
```go
{
  Packages: [{Name: "core"}, {Name: "api"}],
  Versions: {"core": "1.2.0", "api": "2.0.0"},
  Consignments: [...], // ALL consignments in the release
  Date: time.Now(),
  Metadata: {...}       // Aggregated from all consignments
}
```

### API

```go
// Generate release tag
// Returns: (tagName, message, error)
tagName, message, err := generator.GenerateReleaseTag(
    consignments, // ALL consignments in release
    []string{"core", "api"},
    map[string]semver.Version{
        "core": {Major: 1, Minor: 2, Patch: 0},
        "api":  {Major: 2, Minor: 0, Patch: 0},
    },
    "builtin:date",
)
// Returns: ("release-20260130-143000", "", nil) - lightweight tag

// Or with custom annotated template
tagName, message, err := generator.GenerateReleaseTagWithContext(
    consignments,
    []string{"core", "api"},
    versions,
    `release-{{ .Date | date "20060102" }}

# Release {{ .Date | date "2006-01-02" }}

{{ range .Consignments -}}
- {{ .Summary }}
{{ end -}}`,
)
// Returns: ("release-20260130", "# Release 2026-01-30\n...", nil) - annotated tag
```

## Version Command Flow

When running `shipyard version`:

1. **Read consignments** from `.shipyard/consignments/`
2. **Calculate version bumps** for each affected package
3. **Generate changelog** for each package
4. **Generate package tags** for each versioned package:
   ```
   core/v1.2.0
   api/v2.0.0
   ```
5. **Generate release tag**:
   ```
   release-20260130-143000
   ```
6. **Create all git tags**:
   - **Lightweight tags** (single-line template output):
     ```bash
     git tag core/v1.2.0
     git tag api/v2.0.0
     ```
   - **Annotated tags** (multi-line template output with message):
     ```bash
     git tag -a core/v1.2.0 -m "# Release core v1.2.0\n\n- Added feature\n- Fixed bug"
     git tag -a api/v2.0.0 -m "# Release api v2.0.0\n\n- Breaking change"
     ```
   - **Release tag** (always lightweight for releases):
     ```bash
     git tag release-20260130-143000
     ```
7. **Archive consignments** to `.shipyard/history/`

## Template Output Format

Tag templates determine whether to create lightweight or annotated tags based on their output:

### Lightweight Tags (Single Line)
Template outputs a single line containing only the tag name:
```
core/v1.2.0
```
Result: `git tag core/v1.2.0` (lightweight tag)

### Annotated Tags (Multi-Line)
Template outputs tag name, blank line, then message:
```
core/v1.2.0

# Release core v1.2.0

- Added feature
- Fixed bug
```
Result: `git tag -a core/v1.2.0 -m "# Release core v1.2.0\n\n- Added feature\n- Fixed bug"`

**Important**: The blank line (line 2) is required to separate tag name from message.

## Custom Templates

### Per-Package Tag Template

Create a custom package tag template. Templates can output:
- **Single line**: Lightweight tag
- **Multi-line** (name\n\nmessage): Annotated tag

#### Lightweight Tag Example
```go
// custom-package-tag.tmpl
{{ .Package }}-v{{ .Version }}
```

#### Annotated Tag Example
```go
// custom-annotated-tag.tmpl
{{ .Package }}/v{{ .Version }}

# Release {{ .Package }} v{{ .Version }}

{{ range .Consignments -}}
- {{ .Summary }}
{{ end -}}
```

Context available:
- `Package` (string): Package name (e.g., "core")
- `Version` (string): Semantic version (e.g., "1.2.0")
- `Consignments` ([]Consignment): Filtered consignments affecting this package
  - Each has: `ID`, `Timestamp`, `Packages`, `ChangeType`, `Summary`, `Metadata`
- `Date` (time.Time): Current timestamp
- `Metadata` (map): Aggregated metadata from all consignments

### Release Tag Template

Create a custom release tag template. Release tags have access to the full consignment context.

#### Lightweight Release Tag Example
```go
// custom-release-tag.tmpl
release-{{ range $i, $pkg := .Packages }}{{ if $i }}-{{ end }}{{ $pkg.Name }}{{ end }}-{{ .Date | date "20060102" }}
```

#### Annotated Release Tag Example
```go
// custom-annotated-release.tmpl
release-{{ .Date | date "20060102-150405" }}

# Release {{ .Date | date "2006-01-02 15:04:05" }}

## Packages Released
{{ range $pkg, $ver := .Versions -}}
- {{ $pkg }}: v{{ $ver }}
{{ end }}

## All Changes
{{ range .Consignments -}}
- [{{ .ChangeType }}] {{ .Summary }} ({{ join .Packages ", " }})
{{ end -}}
```

Context available:
- `Packages` ([]Package): List of packages (each with `.Name`)
- `Versions` (map[string]string): Map of package name to version string
- `Consignments` ([]Consignment): ALL consignments in the release
  - Each has: `ID`, `Timestamp`, `Packages`, `ChangeType`, `Summary`, `Metadata`
- `Date` (time.Time): Current timestamp
- `Metadata` (map): Aggregated metadata from all consignments

## Type Safety

Template types are enforced by the context. Each template type has its own directory and context:

```go
// ✅ WORKS: keepachangelog exists in builtin/changelog/
generator.GenerateForPackage(consignments, "core", version, "builtin:keepachangelog")

// ✅ WORKS: go exists in builtin/tag/
tagName, message, err := generator.GeneratePackageTag(
    consignments, "core", version, "builtin:go")

// ✅ WORKS: date exists in builtin/release/
tagName, message, err := generator.GenerateReleaseTag(
    consignments, packages, versions, "builtin:date")

// ❌ FAILS: keepachangelog doesn't exist in builtin/tag/
tagName, message, err := generator.GeneratePackageTag(
    consignments, "core", version, "builtin:keepachangelog")
// Error: builtin template not found: tag/keepachangelog

// ❌ FAILS: date doesn't exist in builtin/tag/ (it's in builtin/release/)
tagName, message, err := generator.GeneratePackageTag(
    consignments, "core", version, "builtin:date")
// Error: builtin template not found: tag/date
```

**Template Directory Structure:**
- `builtin/changelog/` - Per-package changelog templates
- `builtin/tag/` - Per-package tag templates (single package context)
- `builtin/release/` - Release tag templates (all packages context)
- `builtin/releasenotes/` - Release notes templates

## Configuration

In `.shipyard/config.yaml`:

```yaml
templates:
  # Per-package changelog template (TemplateTypeChangelog)
  changelog: builtin:keepachangelog

  # Per-package tag template (TemplateTypeTag)
  # Can be lightweight or annotated
  packageTag: builtin:go-annotated  # or builtin:go for lightweight

  # Release tag template (TemplateTypeRelease)
  # Marks the entire release event
  releaseTag: builtin:date  # or builtin:versions

  # Release notes template (TemplateTypeReleaseNotes)
  releaseNotes: builtin:default

# Per-package overrides
packages:
  core:
    templates:
      packageTag: builtin:go-annotated  # Annotated tags with release notes

  api:
    templates:
      packageTag: builtin:go  # Lightweight tags only

  "@scope/package":
    templates:
      packageTag: builtin:npm  # NPM-style lightweight tags
```

## Examples

### Single-Package Repository (Lightweight Tags)

```bash
# Package: core
# Version: v1.5.0
# Template: builtin:default

# Generated tags:
v1.5.0                    # Lightweight package tag
release-20260130-143000   # Release tag
```

### Single-Package Repository (Annotated Tags)

```bash
# Package: core
# Version: v1.5.0
# Template: builtin:detailed-annotated

# Generated tags:
v1.5.0 (annotated)        # Tag with message:
                          # # Release core v1.5.0
                          #
                          # Released on 2026-01-30
                          #
                          # ## Changes
                          # ...
release-20260130-143000   # Release tag
```

### Monorepo with Go Modules (Lightweight)

```bash
# Packages: core v1.2.0, api v2.0.0
# Template: builtin:go

# Generated tags:
core/v1.2.0              # Lightweight package tag
api/v2.0.0               # Lightweight package tag
release-20260130-143000  # Release tag
```

### Monorepo with Go Modules (Annotated)

```bash
# Packages: core v1.2.0, api v2.0.0
# Template: builtin:go-annotated

# Generated tags:
core/v1.2.0 (annotated)  # Tag with release notes
api/v2.0.0 (annotated)   # Tag with release notes
release-20260130-143000  # Release tag
```

### Monorepo with NPM Packages

```bash
# Packages: @scope/ui v0.5.0, @scope/utils v1.3.0
# Template: builtin:npm

# Generated tags:
@scope/ui@0.5.0          # Lightweight package tag
@scope/utils@1.3.0       # Lightweight package tag
release-20260130-143000  # Release tag
```

## Best Practices

1. **Use consistent tag formats** within a repository
2. **Choose between lightweight and annotated tags**:
   - **Lightweight** (`builtin:go`, `builtin:npm`): Faster, simpler, good for frequent releases
   - **Annotated** (`builtin:go-annotated`, `builtin:detailed-annotated`): Include release notes in git history, better for major releases
3. **Use `builtin:go` or `builtin:go-annotated` for Go monorepos** to enable `go get` with specific packages
4. **Use `builtin:npm` for NPM monorepos** to match NPM's tagging convention
5. **Use `builtin:release-date` for release tags** for unique, sortable release markers
6. **Don't mix tagging strategies** - pick one format and stick with it
7. **Per-package overrides** allow mixing lightweight/annotated tags by package (e.g., critical packages get annotated, utilities get lightweight)
