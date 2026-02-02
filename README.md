# Shipyard

> Chart your project's version journey - manage cargo (changes), navigate to new ports (versions), and maintain your ship's log

[![Go Report](https://img.shields.io/badge/go%20report-A+-brightgreen.svg?style=flat)](https://goreportcard.com/report/github.com/natonathan/shipyard)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/natonathan/shipyard)](https://github.com/natonathan/shipyard/releases)

## What is Shipyard?

Shipyard automates semantic versioning, changelog generation, and release management for both monorepos and single-package repositories. Instead of manually updating version numbers and changelogs, you create "consignments" (cargo manifests describing changes), and Shipyard automatically calculates version bumps, navigates your fleet to new version ports, updates ship's logs (changelogs), and plants harbor markers (git tags).

### Workflow Overview

```
┌─────────────┐      ┌─────────────┐      ┌─────────────┐      ┌─────────────┐
│  shipyard   │─────▶│  shipyard   │─────▶│  shipyard   │─────▶│  shipyard   │
│     add     │      │   status    │      │   version   │      │release-notes│
└─────────────┘      └─────────────┘      └─────────────┘      └─────────────┘
Log cargo in the     Check cargo and      Sail to the next     Tell the tale of
ship's manifest      chart your course    port                 your voyage
```

## Key Features

- **🎯 Semantic Versioning** - Automatic version calculation based on change types (major/minor/patch)
- **📦 Monorepo Support** - Manage multiple packages with inter-package dependencies
- **🔗 Version Propagation** - Linked dependencies automatically propagate version changes
- **📝 Markdown Consignments** - Track cargo (changes) as reviewable markdown manifests in PRs
- **🎨 Custom Templates** - Fully customizable changelog and release note formats
- **🌐 Remote Config** - Share configuration across teams via Git or HTTP
- **🐙 GitHub Integration** - Optional automated GitHub release creation
- **🚀 Multi-Ecosystem** - Supports Go, NPM, Python, and Docker

## Quick Start

### Installation

**Go Install (Recommended)**:
```bash
go install github.com/natonathan/shipyard/cmd/shipyard@latest
```

**Download Binary**:
```bash
# macOS (Apple Silicon)
curl -LO https://github.com/natonathan/shipyard/releases/latest/download/shipyard-darwin-arm64
chmod +x shipyard-darwin-arm64
sudo mv shipyard-darwin-arm64 /usr/local/bin/shipyard

# macOS (Intel)
curl -LO https://github.com/natonathan/shipyard/releases/latest/download/shipyard-darwin-amd64
chmod +x shipyard-darwin-amd64
sudo mv shipyard-darwin-amd64 /usr/local/bin/shipyard

# Linux (x86_64)
curl -LO https://github.com/natonathan/shipyard/releases/latest/download/shipyard-linux-amd64
chmod +x shipyard-linux-amd64
sudo mv shipyard-linux-amd64 /usr/local/bin/shipyard
```

**Homebrew**:
```bash
brew install natonathan/tap/shipyard
```

**Verify Installation**:
```bash
shipyard --version
```

### Basic Usage

1. **Set sail** - prepare your repository:
   ```bash
   cd /path/to/your/project
   shipyard init
   ```

2. **Log cargo** when you make changes:
   ```bash
   shipyard add --summary "Fix authentication bug" --type patch
   ```

3. **Check cargo** and chart your course:
   ```bash
   shipyard status
   ```

4. **Sail to the next port** - apply version bumps:
   ```bash
   shipyard version
   ```

5. **Push your voyage** to remote:
   ```bash
   git push --follow-tags
   ```

That's it! Your version files, ship's logs (changelogs), and harbor markers (git tags) are all updated automatically.

## Documentation

**Getting Started**:
- [Quickstart Guide](docs/quickstart.md) - Step-by-step tutorial
- [Configuration Examples](examples/) - Real-world configuration examples

**Reference**:
- [Configuration Schema](https://shipyard.tamez.dev/docs/config) - Complete config reference
- [CLI Interface](https://shipyard.tamez.dev/docs/cli) - All commands and flags
- [Troubleshooting Guide](docs/troubleshooting.md) - Common errors and solutions

**Advanced**:
- [Contributing Guide](CONTRIBUTING.md) - Development setup and guidelines

## Use Cases

### Single Repository

Perfect for libraries, applications, or services with one package:

```yaml
packages:
  - name: "app"
    path: "./"
    ecosystem: "go"  # or npm, python, docker
```

**See**: [Single-repo examples](examples/single-repo/)

### Monorepo with Independent Packages

Multiple packages versioned independently:

```yaml
packages:
  - name: "web-app"
    path: "./apps/web"
    ecosystem: "npm"

  - name: "api-server"
    path: "./services/api"
    ecosystem: "go"
```

**See**: [Basic monorepo example](examples/monorepo/basic/)

### Monorepo with Dependencies

Packages with version propagation:

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

**See**: [Monorepo with dependencies](examples/monorepo/with-dependencies/)

### Custom Templates

Use your own changelog and release note formats:

```yaml
templates:
  changelog: |
    # Changelog

    ## Version {{.Version}}

    {{range .Consignments}}
    - [{{.ChangeType | upper}}] {{.Summary}}
    {{end}}
```

**See**: [Template examples](examples/templates/)

## Commands

### Core Commands

- **`shipyard init`** - Set sail - prepare your repository
- **`shipyard add`** - Log cargo in the ship's manifest
- **`shipyard status`** - Check cargo and chart your course
- **`shipyard version`** - Sail to the next port
- **`shipyard release-notes`** - Tell the tale of your voyage

### Flags

All commands support:
- `--config` - Custom config file path
- `--json` - JSON output for automation
- `--verbose` - Detailed logging
- `--quiet` - Suppress output

See [CLI Reference](https://shipyard.tamez.dev/docs/cli) for complete documentation.

## Configuration

Shipyard configuration lives in `.shipyard/shipyard.yaml`:

```yaml
# Define packages
packages:
  - name: "app"
    path: "./"
    ecosystem: "go"
    dependencies: []

# Optional: Custom templates
templates:
  changelog: "builtin:default"
  tagName: "v{{.Version}}"

# Optional: GitHub integration
github:
  enabled: false
  owner: "your-org"
  repo: "your-repo"
```

**Supported Ecosystems**:
- **Go**: `version.go` with `const Version = "X.Y.Z"`
- **NPM**: `package.json` with `"version": "X.Y.Z"`
- **Python**: `pyproject.toml`, `setup.py`, or `__version__.py`
- **Docker**: `Dockerfile` with `LABEL version="X.Y.Z"`

See [Configuration Schema](https://shipyard.tamez.dev/docs/config) for full details and [examples/](examples/) for real-world configurations.

## Development

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for:
- Development setup
- Testing guidelines
- Code style conventions
- Pull request process

### Quick Development Setup

```bash
# Clone repository
git clone https://github.com/natonathan/shipyard.git
cd shipyard

# Install dependencies and build
just dev-setup
just build

# Run tests
just test

# Run CLI locally
just run status

# Log cargo for your changes (required for PRs)
shipyard add --summary "Add new feature" --type minor
```

**Important**: All PRs that modify code must include a consignment (cargo manifest). This is enforced by CI and ensures proper version tracking.

## License

MIT License - see [LICENSE](LICENSE) for details

## Credits

Built with:
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Terminal UI
- [Glamour](https://github.com/charmbracelet/glamour) - Markdown rendering
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling

---

**Questions?** Open an issue at https://github.com/natonathan/shipyard/issues
