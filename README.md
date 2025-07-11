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
- 🎯 **Multi-ecosystem**: Supports Go, NPM, Python, Docker, Helm, and more
- 🎪 **Interactive CLI**: User-friendly prompts for all operations
- ⚙️ **Flexible Configuration**: YAML-based with inheritance support

## Installation

### From Source

```bash
git clone https://github.com/NatoNathan/shipyard.git
cd shipyard
just build
```

### Using Go

```bash
go install github.com/NatoNathan/shipyard/cmd/shipyard@latest
```

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

3. **Your configuration is saved** to `.shipyard/config.yaml` and consignments are stored in `.shipyard/consignments/`

## Core Concepts

### Consignments
A consignment is a record of changes made to your packages. Each consignment contains:
- **Packages affected**: Which parts of your project changed
- **Change type**: Patch (bug fixes), Minor (new features), or Major (breaking changes)
- **Summary**: Description of what changed

### Workflow
1. Make changes to your code
2. Create a consignment to document the changes
3. Commit both your code and the consignment
4. Shipyard handles version calculation and changelog generation

## Usage

### Commands

- `shipyard init` - Initialize a new Shipyard project
- `shipyard add` - Create a new consignment to track changes
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

## Development

### Prerequisites

- Go 1.23 or later
- [Just](https://github.com/casey/just) (optional, for build scripts)

### Building

```bash
# Using just (recommended)
just build

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
go test ./...
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[Add your license here]

## Roadmap

- [x] Project initialization
- [x] Consignment creation and management
- [x] Multi-ecosystem package support
- [ ] Changelog generation from consignments
- [ ] Automatic version calculation
- [ ] Release automation
- [ ] Git integration and tagging
- [ ] CI/CD integration
- [ ] Plugin system for custom workflows
