# Shipyard Template System

Shipyard uses Go templates for customizable output formats. This document explains the template system and available options.

## Template Types

Shipyard supports templates for:
- **changelog** - CHANGELOG.md generation
- **tagName** - Git tag names
- **tagMessage** - Git tag annotation messages
- **releaseNotes** - GitHub release notes
- **commitMessage** - Version bump commit messages

## Built-in Templates

### builtin:default

Standard format for changelogs and release notes.

**Changelog Example:**
```markdown
# Changelog

## 1.2.3 - 2024-01-15

- Add new feature
- Fix bug in authentication
- Update dependencies
```

### builtin:grouped

Groups changes by type (major/minor/patch).

**Changelog Example:**
```markdown
# Changelog

## 1.2.3 - 2024-01-15

### Features (Minor)
- Add new API endpoint
- Support OAuth authentication

### Fixes (Patch)
- Fix memory leak
- Correct validation logic
```

### builtin:go

Go module-style tags (v-prefixed).

**Tag Example:**
```
v1.2.3
```

### builtin:npm

NPM-style tags (package-prefixed).

**Tag Example:**
```
my-package/v1.2.3
```

## Template Configuration

### Global Templates

```yaml
templates:
  changelog:
    source: builtin:grouped
  tagName:
    source: builtin:npm
  releaseNotes:
    source: ./templates/release.tmpl
```

### Package-Specific Templates

```yaml
packages:
  - name: my-api
    path: packages/api
    ecosystem: npm
    templates:
      changelog:
        source: ./templates/api-changelog.tmpl
```

### Inline Templates

```yaml
templates:
  releaseNotes:
    inline: |
      # {{.Package}} {{.Version}}

      ## Changes
      {{range .Consignments}}
      - [{{.ChangeType}}] {{.Summary}}
      {{end}}
```

## Template Data

### Changelog Template Data

```go
{
  Package: string              // Package name
  Version: string              // Version number
  Consignments: []Consignment  // Changes for this version
  AllVersions: []HistoryEntry  // Complete version history
}
```

### Tag Name Template Data

```go
{
  Package: string  // Package name
  Version: string  // Version number
}
```

### Tag Message Template Data

```go
{
  Package: string              // Package name
  Version: string              // Version number
  Consignments: []Consignment  // Changes for this version
}
```

### Release Notes Template Data

```go
{
  Package: string              // Package name
  Version: string              // Version number
  Tag: string                  // Git tag name
  Consignments: []Consignment  // Changes for this version
}
```

### Commit Message Template Data

```go
{
  Packages: string                  // Comma-separated list
  PackageVersions: []PackageVersion // All versioned packages
}
```

## Template Functions

Shipyard templates have access to Go template functions plus Sprig functions.

### String Functions

```
{{.Summary | upper}}        # UPPERCASE
{{.Summary | lower}}        # lowercase
{{.Summary | title}}        # Title Case
{{.Summary | trim}}         # Remove whitespace
{{.Package | replace "-" "_"}}  # String replacement
```

### Collection Functions

```
{{range .Consignments}}     # Iterate
{{len .Consignments}}       # Count
{{index .Consignments 0}}   # Access by index
```

### Conditional Functions

```
{{if eq .ChangeType "major"}}
  ⚠️ BREAKING CHANGE
{{else if eq .ChangeType "minor"}}
  ✨ New Feature
{{else}}
  🐛 Bug Fix
{{end}}
```

### Date Functions

```
{{now | date "2006-01-02"}}          # Current date
{{.Timestamp | date "January 2, 2006"}}  # Format timestamp
```

## Custom Template Examples

### Grouped by Change Type

```go
# {{.Package}} Changelog

{{range $version := .AllVersions}}
## {{$version.Version}} - {{$version.Timestamp | date "2006-01-02"}}

{{$major := list}}
{{$minor := list}}
{{$patch := list}}
{{range $version.Consignments}}
  {{if eq .ChangeType "major"}}
    {{$major = append $major .}}
  {{else if eq .ChangeType "minor"}}
    {{$minor = append $minor .}}
  {{else}}
    {{$patch = append $patch .}}
  {{end}}
{{end}}

{{if $major}}
### ⚠️ Breaking Changes
{{range $major}}
- {{.Summary}}
{{end}}
{{end}}

{{if $minor}}
### ✨ Features
{{range $minor}}
- {{.Summary}}
{{end}}
{{end}}

{{if $patch}}
### 🐛 Fixes
{{range $patch}}
- {{.Summary}}
{{end}}
{{end}}

{{end}}
```

### With Metadata

```go
# Release Notes: {{.Package}} {{.Version}}

{{range .Consignments}}
## {{.Summary}}

**Change Type:** {{.ChangeType}}
{{if .Metadata.author}}
**Author:** {{.Metadata.author}}
{{end}}
{{if .Metadata.issue}}
**Issue:** {{.Metadata.issue}}
{{end}}

---
{{end}}
```

### Semantic Release Format

```go
# [{{.Version}}](https://github.com/org/repo/compare/v{{.PrevVersion}}...v{{.Version}}) ({{.Timestamp | date "2006-01-02"}})

{{$features := list}}
{{$fixes := list}}
{{$breaking := list}}

{{range .Consignments}}
  {{if eq .ChangeType "major"}}
    {{$breaking = append $breaking .}}
  {{else if eq .ChangeType "minor"}}
    {{$features = append $features .}}
  {{else}}
    {{$fixes = append $fixes .}}
  {{end}}
{{end}}

{{if $breaking}}
### ⚠ BREAKING CHANGES

{{range $breaking}}
* {{.Summary}}
{{end}}
{{end}}

{{if $features}}
### Features

{{range $features}}
* {{.Summary}}
{{end}}
{{end}}

{{if $fixes}}
### Bug Fixes

{{range $fixes}}
* {{.Summary}}
{{end}}
{{end}}
```

## Testing Templates

### Preview with --preview

```bash
shipyard version --preview
```

Shows how templates will render without applying changes.

### Generate Release Notes

```bash
shipyard release-notes --package my-api
```

Tests release notes template.

### Validate Configuration

```bash
shipyard validate
```

Checks template syntax and accessibility.

## Best Practices

- Start with built-in templates
- Customize only when needed
- Store custom templates in `templates/` directory
- Test templates with `--preview`
- Use meaningful section headers
- Include metadata when available
- Keep templates maintainable

## Troubleshooting

### Template Parse Error

```
Error: template: parse error at line 5
```

Check Go template syntax:
- Matching `{{` and `}}`
- Valid function names
- Correct variable references

### Missing Data

```
Error: can't evaluate field X
```

Check template data structure - field may not exist in context.

### Empty Output

Template may have logic that filters out all content - verify conditionals.
