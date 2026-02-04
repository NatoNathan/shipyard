# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is Shipyard?

Shipyard is a semantic versioning and release management tool for monorepos and single-package repositories. It automates version bumps, changelog generation, and release management through a concept of "consignments" (change manifests).

## Development Commands

### Building and Testing
```bash
# Build the CLI
just build                 # Creates bin/shipyard

# Run all tests
just test                  # All tests with race detection and coverage
just test-unit             # Unit tests only
just test-integration      # Integration tests only

# Run a single test
go test -v ./internal/graph -run TestBuildGraph

# Run tests via Dagger (containerized)
just test-dagger           # Standard test run
just test-dagger-race      # With race detection (slower)
```

### Code Quality
```bash
# Lint code
just lint                  # Local linter
just lint-dagger           # Via Dagger

# Format code
just fmt                   # Format all Go files

# Check security
just security              # Run gosec scanner

# All CI checks (lint + test + verify)
just ci
just ci-dagger             # Via Dagger
```

### Development Setup
```bash
# Initial setup (installs dev tools)
just dev-setup

# Run CLI locally during development
just run status            # Runs: go run ./cmd/shipyard status
just run add --help

# Install to $GOPATH/bin
just install
```

### Coverage
```bash
# Generate coverage report
just coverage              # Creates coverage.html
just coverage-dagger       # Via Dagger
just coverage-check-dagger 80  # Check threshold
```

### Dependency Management
```bash
# Check for outdated dependencies
just check-deps

# Update dependencies
just update-deps

# Verify go.mod is clean
just verify
```

## Architecture Overview

### Core Workflow
1. **Consignment Management** (`internal/consignment/`) - Tracks changes as markdown files
2. **Version Calculation** (`internal/version/`) - Calculates semantic version bumps from consignments
3. **Graph Processing** (`internal/graph/`) - Handles package dependencies and propagation
4. **Ecosystem Support** (`internal/ecosystem/`) - Updates version files for different package managers

### Key Concepts

**Consignments**: Change manifests stored as markdown files in `.shipyard/consignments/`. Each consignment records:
- Change type (major/minor/patch)
- Affected packages
- Summary and metadata
- Unique ID format: `YYYYMMDD-HHMMSS-{random6}`

**Version Propagation**: When a package version changes, dependent packages can automatically bump their versions based on dependency strategies:
- `linked`: Dependent bumps with the same change type
- `fixed`: Dependent uses exact version, requires manual update

**Dependency Graph**: Built in `internal/graph/`, the graph:
- Detects cycles using Tarjan's algorithm
- Performs topological sorting for version application order
- Handles strongly connected components (SCCs) for cycle resolution

### Package Structure

```
cmd/shipyard/               # CLI entry point (main.go)
internal/
  ├── commands/             # Cobra command implementations
  │   ├── init.go          # Initialize repository
  │   ├── add.go           # Create consignments
  │   ├── version.go       # Calculate and apply versions
  │   ├── status.go        # Show pending changes
  │   ├── release.go       # Create GitHub releases
  │   └── upgrade.go       # Self-upgrade command
  ├── config/              # Configuration loading and validation
  │   ├── config.go        # Main config struct and types
  │   └── loader.go        # YAML loading with remote support
  ├── consignment/         # Change tracking
  │   ├── consignment.go   # Core data structure
  │   ├── read.go          # Load from filesystem
  │   ├── write.go         # Save to filesystem
  │   └── group.go         # Group by package
  ├── version/             # Version calculation engine
  │   ├── propagate.go     # Version propagation logic
  │   ├── apply.go         # Apply versions to files
  │   ├── direct.go        # Direct changes (from consignments)
  │   ├── conflict.go      # Conflict detection
  │   └── cycle.go         # Cycle handling
  ├── graph/               # Dependency graph
  │   ├── graph.go         # Graph data structure
  │   ├── build.go         # Build graph from config
  │   ├── tarjan.go        # SCC detection
  │   ├── topsort.go       # Topological sort
  │   ├── cycles.go        # Cycle detection
  │   └── cache.go         # Graph caching
  ├── ecosystem/           # Version file handlers
  │   ├── go.go            # Go (version.go, go.mod)
  │   ├── npm.go           # NPM (package.json)
  │   ├── python.go        # Python (pyproject.toml, etc.)
  │   ├── helm.go          # Helm (Chart.yaml)
  │   ├── cargo.go         # Rust (Cargo.toml)
  │   └── deno.go          # Deno (deno.json)
  ├── changelog/           # Changelog generation
  ├── template/            # Template engine
  ├── git/                 # Git operations (tags, commits, detection)
  ├── github/              # GitHub API integration
  ├── prompt/              # Interactive prompts (Bubble Tea)
  ├── ui/                  # Terminal UI components
  ├── detect/              # Auto-detect packages
  ├── editor/              # External editor integration
  ├── metadata/            # Custom metadata validation
  └── logger/              # Logging utilities
pkg/
  ├── types/               # Public types (ChangeType, etc.)
  └── semver/              # Semantic version parsing
test/integration/          # Integration tests
```

### Important Implementation Details

**Ecosystem Version Files**:
- Go: Looks for `version.go` with `const Version = "X.Y.Z"` or `go.mod` with `// version: X.Y.Z` comment
- NPM: Updates `version` field in `package.json`
- Python: Updates `pyproject.toml`, `__version__.py`, or `setup.py`
- Helm: Updates `version` and `appVersion` in `Chart.yaml`
- Cargo: Updates `version` in `[package]` section of `Cargo.toml`
- Deno: Updates `version` in `deno.json` or `deno.jsonc`

**Graph Algorithm**: Uses Tarjan's algorithm to detect strongly connected components (cycles). Cycles are handled by grouping packages in the same SCC and applying consistent version bumps.

**Consignment Format**: Markdown files with YAML frontmatter:
```markdown
---
id: 20240101-120000-abc123
timestamp: 2024-01-01T12:00:00Z
packages: [pkg1, pkg2]
changeType: minor
metadata:
  author: user@example.com
---

# Summary of changes

Description of the change...
```

**Template System**: Supports Go templates with Sprig functions. Built-in templates for changelogs, release notes, tag names, and commit messages. Can load templates from files or inline strings.

## Testing Guidelines

- All tests in `_test.go` files alongside implementation
- Integration tests in `test/integration/`
- Use table-driven tests where appropriate
- Mock external dependencies (git, GitHub API)
- Tests must pass with race detection: `go test -race`

## Pull Request Requirements

1. **Consignment Required**: All PRs that modify code must include a consignment:
   ```bash
   shipyard add --summary "Your change" --type patch
   ```
   This is enforced by CI (`.github/workflows/require-consignment.yml`)

2. **Tests**: Add tests for new functionality, ensure existing tests pass
3. **Linting**: Code must pass `golangci-lint`
4. **Commit Messages**: Follow conventional commits format

## Configuration Location

- Default config: `.shipyard/shipyard.yaml`
- Consignments directory: `.shipyard/consignments/`
- History tracking: `.shipyard/history.yaml`

## Supported Ecosystems

Go (1.21+), NPM, Python, Helm, Cargo (Rust), Deno

Each ecosystem has its own version file format and update logic in `internal/ecosystem/`.

## Release Process

Releases are automated via GitHub Actions (`.github/workflows/release.yml`):
1. Tag is pushed (format: `vX.Y.Z`)
2. Builds binaries for all platforms
3. Creates GitHub release with generated notes
4. Publishes to npm registry
5. Builds and pushes Docker images to GHCR

The Dagger pipeline is defined in `dagger/` for reproducible builds.
