# Shipyard

> Where releases are built

Shipyard is a modern CLI tool for managing change notes, versions, and releases across both monorepos and single repositories. It uses a "consignment" system to track changes and automatically manages version bumps and changelog generation.

## Overview

Shipyard simplifies release management by:
- **Tracking Changes**: Create consignments to document what changed and why
- **Managing Versions**: Automatically calculate semantic version bumps based on change types
- **Generating Changelogs**: Build release notes from your consignments
- **Supporting Any Ecosystem**: Works with Go, NPM, Python, Docker, and more

Perfect for teams who want structured release management without the complexity of traditional changelog workflows.

## Features

- 🏗️ **Universal Support**: Works with monorepos and single repositories
- 📦 **Consignment System**: Track changes with structured metadata
- 🔄 **Semantic Versioning**: Automatic version calculation (patch/minor/major)
- 📝 **Changelog Automation**: Generate release notes from consignments
- 🎯 **Multi-ecosystem**: Supports Go, NPM, Helm, and more
- 🎪 **Interactive CLI**: User-friendly prompts and confirmations
- ⚙️ **Flexible Configuration**: YAML-based with inheritance support
- 🔍 **Status Monitoring**: View consignment status and version previews
- 📄 **Template System**: Customizable changelog templates (Keep a Changelog, etc.)
- 🎨 **Pretty Output**: Markdown rendering with syntax highlighting
- 🚀 **Dry Run Mode**: Preview changes before applying them

## Installation

### Using Go

```bash
go install github.com/NatoNathan/shipyard/cmd/shipyard@latest
```

### From Source

```bash
git clone https://github.com/NatoNathan/shipyard.git
cd shipyard
just build
# Binary will be in ./dist/shipyard
```

### From Releases

Download the latest release from [GitHub Releases](https://github.com/NatoNathan/shipyard/releases) (coming soon).

## Quick Start

1. **Initialize your project**:
   ```bash
   shipyard init
   ```
   Configure your repository type, packages, and changelog settings through the interactive prompts.

2. **Create your first consignment**:
   ```bash
   shipyard add
   ```
   Document a change with its type (patch/minor/major) and summary.

3. **Check project status**:
   ```bash
   shipyard status
   ```
   View current consignments, see what new versions would be calculated, and preview release notes.

4. **Generate changelog and apply versions**:
   ```bash
   shipyard version
   ```
   Generate changelog from consignments and update package versions.

## Core Concepts

### Consignments
A consignment is a record of changes made to your packages. Each consignment contains:
- **Packages affected**: Which parts of your project changed
- **Change type**: Patch (bug fixes), Minor (new features), or Major (breaking changes)
- **Summary**: Description of what changed

## Workflow Example

Here's a complete example of using Shipyard in a project:

```bash
# 1. Initialize your project
shipyard init

# 2. Make code changes to your project
# ... (edit files, add features, fix bugs)

# 3. Document your changes
shipyard add
# Follow prompts to select packages and change type

# 4. Check status anytime
shipyard status

# 5. Preview your changelog
shipyard version --preview

# 6. Generate changelog and apply versions
shipyard version

# 7. Commit everything
git add .
git commit -m "Release v1.2.0"

# 8. (Future) Tag and release
git tag v1.2.0
git push origin v1.2.0
```

## Usage

### Commands

- `shipyard init` - Initialize a new Shipyard project
- `shipyard add` - Create a new consignment to track changes
- `shipyard status` - Show consignment status and version information
- `shipyard version` - Generate changelogs and apply version updates
- `shipyard release-notes [version]` - Get release notes for a specific version
- `shipyard --version` or `shipyard -V` - Show version information
- `shipyard --help` - Show available commands and options

### Configuration

Shipyard uses a YAML configuration file (`.shipyard/config.yaml`) to store project settings:

```yaml
type: monorepo  # or "single-repo"
repo: github.com/your-org/your-repo
changelog:
  template: keepachangelog
packages:
  - name: api
    path: packages/api
    ecosystem: npm
    manifest: packages/api/package.json
  - name: frontend
    path: packages/frontend
    ecosystem: npm
    manifest: packages/frontend/package.json
```

### Global Options

- `--config, -c` - Path to configuration file (default: .shipyard/config.yaml)
- `--verbose, -v` - Enable verbose logging
- `--log-level` - Set log level (debug, info, warn, error)
- `--log-file` - Set log file path

### Advanced Usage

#### Status Command Options

```bash
# Show status for all packages
shipyard status

# Show status for specific package (monorepo only)
shipyard status --package api

# Generate and display release notes
shipyard status --release-notes

# Show raw markdown instead of pretty output
shipyard status --release-notes --raw

# Use a different changelog template
shipyard status --release-notes --template keepachangelog
```

#### Version Command Options

```bash
# Preview the changelog without applying changes
shipyard version --preview

# Dry run - show changelog and version info without applying
shipyard version --dry-run

# Skip confirmation prompts
shipyard version --yes

# Generate for specific package only (monorepo)
shipyard version --package api

# Custom output file
shipyard version --output RELEASE_NOTES.md

# Use different changelog template
shipyard version --template keepachangelog
```

#### Release Notes Command Options

```bash
# Get release notes for a specific version
shipyard release-notes 1.2.3

# Get release notes with 'v' prefix
shipyard release-notes v1.2.3

# Get release notes for specific package (monorepo only)
shipyard release-notes 1.2.3 --package api

# Show raw markdown instead of pretty output
shipyard release-notes 1.2.3 --raw

# Use a different changelog template
shipyard release-notes 1.2.3 --template simple
```

## Development

### Prerequisites

- Go 1.23 or later
- [Just](https://github.com/casey/just) (optional, for build scripts)

### Building

```bash
# Using just (recommended)
just build

# Build with specific version
just build v1.0.0

# Or using go directly
go build -o shipyard ./cmd/shipyard/main.go
```

### Running

```bash
# Using just
just run [OPTIONS]

# Or using go directly
go run ./cmd/shipyard/main.go [OPTIONS]
```

### Testing

```bash
# Run all tests
just test

# Run specific test
just test ./pkg/changelog

# Or using go directly
go test ./...
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Roadmap

- [x] Project initialization
- [x] Consignment creation and management
- [x] Multi-ecosystem package support (Go, NPM, Helm)
- [x] Changelog generation from consignments
- [x] Automatic version calculation
- [x] Status command with version preview
- [x] Interactive CLI with confirmation prompts
- [x] Template-based changelog generation
- [x] Package filtering for monorepos
- [x] Pretty-printed markdown output
- [ ] Release automation
- [ ] Git integration and tagging
- [ ] CI/CD integration
- [ ] Plugin system for custom workflows
- [ ] Additional ecosystem support (Python, Docker, etc.)
