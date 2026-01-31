# Shipyard Configuration Examples

This directory contains comprehensive examples of Shipyard configurations for various use cases. Each example includes inline comments explaining every configuration section.

## Quick Navigation

### Single Repository Examples

Configurations for projects with a single package:

- **[Go Module](./single-repo/go-module)** - Simple Go application with built-in templates
- **[NPM Package](./single-repo/npm-package)** - JavaScript/TypeScript package with custom metadata
- **[Python Package](./single-repo/python-package)** - Python library with pyproject.toml

### Monorepo Examples

Configurations for projects with multiple packages:

- **[Basic Monorepo](./monorepo/basic)** - Multiple independent packages
- **[With Dependencies](./monorepo/with-dependencies)** - Linked packages with version propagation
- **[Mixed Ecosystems](./monorepo/mixed-ecosystems)** - Multi-language monorepo (Go, Python, NPM, Docker)

### Template Examples

Custom template configurations:

- **[Changelog Template](./templates/changelog-keepachangelog.yaml)** - Keep a Changelog format
- **[Tag Template](./templates/tag-monorepo.yaml)** - Monorepo-style tag naming (package/version)
- **[Release Notes Template](./templates/release-notes-detailed.yaml)** - Enhanced GitHub release notes

### Remote Configuration

Extending and sharing configurations:

- **[Base Config](./remote-config/base-config.yaml)** - Shared configuration for teams
- **[Extending Config](./remote-config/extending-config.yaml)** - Local overrides of remote config

## How to Use These Examples

### 1. Copy Example Configuration

```bash
# For single-repo projects
cp examples/single-repo/go-module/.shipyard/shipyard.yaml .shipyard/

# For monorepos
cp examples/monorepo/with-dependencies/.shipyard/shipyard.yaml .shipyard/
```

### 2. Customize for Your Project

Edit the copied configuration to match your project structure:

- Update package names and paths
- Adjust ecosystem types
- Configure dependencies (for monorepos)
- Customize templates (optional)
- Add metadata fields (optional)

### 3. Verify Configuration

```bash
# Initialize or validate configuration
shipyard init --force

# Check package detection
shipyard status
```

## Common Patterns

### Pattern: Single Repository

Use when you have one package to version:

```yaml
packages:
  - name: "app"
    path: "./"
    ecosystem: "go"  # or npm, python, docker
```

**See**: [single-repo examples](./single-repo/)

### Pattern: Independent Monorepo

Use when packages are versioned independently:

```yaml
packages:
  - name: "service-a"
    path: "./services/a"
    ecosystem: "go"
    dependencies: []  # No dependencies

  - name: "service-b"
    path: "./services/b"
    ecosystem: "go"
    dependencies: []  # Independent
```

**See**: [basic monorepo example](./monorepo/basic/)

### Pattern: Linked Dependencies

Use when one package depends on another:

```yaml
packages:
  - name: "core"
    path: "./packages/core"
    ecosystem: "go"

  - name: "api-client"
    path: "./clients/api"
    ecosystem: "npm"
    dependencies:
      - package: "core"
        strategy: "linked"  # Version changes with core
```

**See**: [with-dependencies example](./monorepo/with-dependencies/)

### Pattern: Custom Bump Mapping

Use when dependency changes should have different impact:

```yaml
packages:
  - name: "web-app"
    path: "./apps/web"
    ecosystem: "npm"
    dependencies:
      - package: "api-client"
        strategy: "linked"
        bumpMapping:
          major: patch  # api-client major → web-app patch
          minor: patch  # api-client minor → web-app patch
          patch: patch  # api-client patch → web-app patch
```

**See**: [with-dependencies example](./monorepo/with-dependencies/)

### Pattern: Required Metadata

Use when you need custom fields in consignments:

```yaml
metadata:
  fields:
    - name: "author"
      required: true
      description: "Author email address"
      type: "string"

    - name: "issue"
      required: false
      description: "JIRA ticket reference"
      type: "string"
```

**See**: [npm-package example](./single-repo/npm-package/)

### Pattern: Custom Templates

Use when you want custom changelog/release note formats:

```yaml
templates:
  # Inline template (multiline YAML string)
  changelog: |
    # Changelog for {{.Package}}

    Version {{.Version}}

    {{range .Consignments}}
    - [{{.ChangeType | upper}}] {{.Summary}}
    {{end}}

  # Simple tag format
  tagName: "{{.Package}}/v{{.Version}}"
```

**See**: [template examples](./templates/)

### Pattern: Remote Configuration

Use when sharing configurations across teams:

```yaml
extends:
  # Public config
  - "https://configs.example.com/shipyard-base.yaml"

  # Private config (with auth)
  - url: "https://api.internal.com/config.yaml"
    auth: "env:CONFIG_TOKEN"

  # Git repository config
  - "git@github.com:org/configs.git#shipyard/base.yaml@v1.0.0"

# Local packages append to remote packages
packages:
  - name: "custom-service"
    path: "./services/custom"
    ecosystem: "go"
```

**See**: [remote-config examples](./remote-config/)

## Documentation References

For complete documentation, see:

- **[Quickstart Guide](../docs/quickstart.md)** - Getting started tutorial
- **[Configuration Schema](https://shipyard.tamez.dev/docs/config)** - Full schema reference
- **[CLI Interface](https://shipyard.tamez.dev/docs/cli)** - Command reference

## Getting Help

If you need help with configuration:

1. Check the [Troubleshooting Guide](../docs/troubleshooting.md)
2. Review similar examples in this directory
3. Read the [Configuration Schema](https://shipyard.tamez.dev/docs/config)
4. Open an issue at https://github.com/natonathan/shipyard/issues
