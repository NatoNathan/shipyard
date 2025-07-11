# Shipyard

> Where releases are built

Shipyard is a CLI tool for managing change notes, versions, and releases for both monorepos and single repositories.

## Features

- 🏗️ **Monorepo and Single-repo Support**: Configure for either repository type
- 📝 **Changelog Management**: Generate and manage changelogs with templates
- 🎯 **Multi-ecosystem Support**: Works with NPM, Go, Helm, Python, Docker, and more
- 🎪 **Interactive Setup**: User-friendly TUI for project initialization
- ⚙️ **Flexible Configuration**: YAML-based configuration with inheritance support

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

1. **Initialize a new project**:
   ```bash
   shipyard init
   ```

2. **Follow the interactive prompts** to configure your repository type, packages, and changelog settings.

3. **Your configuration will be saved** to `.shipyard/config.yaml`

## Usage

### Commands

- `shipyard init` - Initialize a new Shipyard project
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

- [ ] Changelog generation
- [ ] Version management
- [ ] Release automation
- [ ] Git integration
- [ ] CI/CD integration
- [ ] Plugin system
