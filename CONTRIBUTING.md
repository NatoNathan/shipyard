# Contributing to Shipyard

Thank you for your interest in contributing to Shipyard! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Getting Started](#getting-started)
- [Project Structure](#project-structure)
- [Development Workflow](#development-workflow)
- [Testing](#testing)
- [Code Style](#code-style)
- [Adding New Ecosystem Support](#adding-new-ecosystem-support)
- [Adding New Commands](#adding-new-commands)
- [Commit Conventions](#commit-conventions)
- [Pull Request Process](#pull-request-process)
- [Releasing](#releasing)
- [Useful Commands](#useful-commands)
- [Architecture Overview](#architecture-overview)
- [Questions?](#questions)

## Getting Started

### Prerequisites

- **Go 1.21+** - [Install Go](https://go.dev/doc/install)
- **Git** - [Install Git](https://git-scm.com/downloads)
- **Just** - [Install Just](https://github.com/casey/just#installation) (task runner)
- **golangci-lint** - [Install golangci-lint](https://golangci-lint.run/usage/install/)

### Development Setup

1. **Clone the repository**:
   ```bash
   git clone https://github.com/natonathan/shipyard.git
   cd shipyard
   ```

2. **Initialize development environment**:
   ```bash
   just dev-setup
   ```

   This will:
   - Download Go dependencies
   - Install golangci-lint (linter)
   - Install gosec (security scanner)

3. **Build the CLI**:
   ```bash
   just build
   ```

   Binary will be at `bin/shipyard`

4. **Verify setup**:
   ```bash
   just test
   ```

   All tests should pass.

**Total setup time**: ~5 minutes

## Project Structure

```
shipyard/
├── cmd/
│   └── shipyard/          # CLI entry point
│       └── main.go
├── internal/              # Private application code
│   ├── commands/          # CLI command implementations (add, version, etc.)
│   ├── config/            # Configuration loading and validation
│   ├── consignment/       # Consignment (change note) management
│   ├── changelog/         # Changelog generation
│   ├── detect/            # Ecosystem auto-detection
│   ├── ecosystem/         # Ecosystem-specific version handling
│   ├── editor/            # Editor integration for markdown
│   ├── errors/            # Error types and handling
│   ├── fileutil/          # File system utilities
│   ├── git/               # Git operations (tags, commits)
│   ├── graph/             # Dependency graph and topological sorting
│   ├── history/           # Version history tracking
│   ├── logger/            # Logging utilities
│   ├── metadata/          # Metadata field validation
│   ├── prompt/            # Interactive prompts (Bubble Tea)
│   ├── template/          # Template engine and builtin templates
│   ├── ui/                # Terminal UI components
│   └── version/           # Version calculation and bumping
├── pkg/                   # Public library code
│   ├── types/             # Shared data structures
│   └── semver/            # Semantic versioning utilities
├── test/                  # Tests organized by type
│   ├── unit/              # Unit tests
│   ├── integration/       # Integration tests
│   └── contract/          # Contract tests
├── examples/              # Example configurations
├── docs/                  # Documentation
├── justfile               # Task runner commands
├── go.mod                 # Go dependencies
└── README.md              # Project readme
```

## Development Workflow

### Making Changes

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following the [Code Style](#code-style) guidelines

3. **Run tests**:
   ```bash
   just test
   ```

4. **Run linter**:
   ```bash
   just lint
   ```

5. **Format code**:
   ```bash
   just fmt
   ```

6. **Commit your changes** following [Commit Conventions](#commit-conventions)

7. **Push and create a pull request**

### Local Testing

Test your changes with the local CLI:

```bash
# Run CLI with arguments
just run status
just run add --summary "Test change" --bump patch

# Or build and run binary directly
just build
./bin/shipyard status
```

## Testing

Shipyard uses three types of tests:

### Unit Tests

Test individual functions and methods in isolation:

```bash
# Run all unit tests
just test-unit

# Run tests for specific package
go test -v ./internal/version/...
```

**Location**: Alongside source files (`*_test.go`)

**Example**:
```go
func TestCalculateVersionBump(t *testing.T) {
    result := version.CalculateBump("1.0.0", "minor")
    assert.Equal(t, "1.1.0", result)
}
```

### Integration Tests

Test interactions between components:

```bash
# Run all integration tests
just test-integration
```

**Location**: `test/integration/`

**Example**: Testing full workflow from add → status → version

### Contract Tests

Verify CLI behavior matches specifications:

```bash
# Run all contract tests
just test-contract
```

**Location**: `test/contract/`

**Example**: Testing command output formats, exit codes

### Test Coverage

Generate coverage report:

```bash
just coverage
```

Opens `coverage.html` in your browser.

**Target**: 80%+ coverage for new code

## Code Style

### Go Conventions

Follow standard Go conventions:

- Use `gofmt` for formatting (automatically applied by `just fmt`)
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

### Naming

- **Packages**: Short, lowercase, single-word names (e.g., `config`, `version`)
- **Interfaces**: Verb-based names (e.g., `Reader`, `Writer`, `Validator`)
- **Functions**: Clear, descriptive names using camelCase
- **Variables**: Short names for local scope, descriptive for package scope

### Error Handling

- Return errors, don't panic (except for truly unrecoverable situations)
- Wrap errors with context: `fmt.Errorf("failed to load config: %w", err)`
- Use custom error types for specific error conditions

### Documentation

- Public functions must have doc comments
- Doc comments start with the function name
- Keep comments concise and meaningful

**Example**:
```go
// LoadConfig reads and validates the Shipyard configuration file.
// Returns an error if the file doesn't exist or is invalid.
func LoadConfig(path string) (*Config, error) {
    // ...
}
```

### Testing

- Table-driven tests for multiple cases
- Use `testify/assert` for assertions
- Test both success and error cases
- Use descriptive test names: `TestLoadConfig_WhenFileNotFound_ReturnsError`

## Adding New Ecosystem Support

To add support for a new package ecosystem (e.g., Rust, Ruby):

### 1. Create Ecosystem Handler

Create `internal/ecosystem/rust.go`:

```go
package ecosystem

import (
    "fmt"
    "os"
    "path/filepath"
    "regexp"
)

type RustEcosystem struct{}

func (e *RustEcosystem) Name() string {
    return "rust"
}

func (e *RustEcosystem) VersionFilePath(packagePath string) (string, error) {
    cargoPath := filepath.Join(packagePath, "Cargo.toml")
    if _, err := os.Stat(cargoPath); err != nil {
        return "", fmt.Errorf("Cargo.toml not found: %w", err)
    }
    return cargoPath, nil
}

func (e *RustEcosystem) ReadVersion(filePath string) (string, error) {
    content, err := os.ReadFile(filePath)
    if err != nil {
        return "", fmt.Errorf("failed to read Cargo.toml: %w", err)
    }

    // Match: version = "X.Y.Z"
    re := regexp.MustCompile(`version\s*=\s*"([^"]+)"`)
    matches := re.FindSubmatch(content)
    if matches == nil {
        return "", fmt.Errorf("version not found in Cargo.toml")
    }

    return string(matches[1]), nil
}

func (e *RustEcosystem) WriteVersion(filePath, newVersion string) error {
    content, err := os.ReadFile(filePath)
    if err != nil {
        return fmt.Errorf("failed to read Cargo.toml: %w", err)
    }

    re := regexp.MustCompile(`(version\s*=\s*")[^"]+("`)
    updated := re.ReplaceAll(content, []byte("${1}"+newVersion+"${2}"))

    if err := os.WriteFile(filePath, updated, 0644); err != nil {
        return fmt.Errorf("failed to write Cargo.toml: %w", err)
    }

    return nil
}
```

### 2. Register Ecosystem

In `internal/ecosystem/registry.go`:

```go
func init() {
    Register("rust", &RustEcosystem{})
}
```

### 3. Add Detection

In `internal/detect/detect.go`:

```go
func DetectEcosystem(path string) (string, error) {
    // ... existing checks ...

    if fileExists(filepath.Join(path, "Cargo.toml")) {
        return "rust", nil
    }

    return "", errors.New("could not detect ecosystem")
}
```

### 4. Add Tests

Create `internal/ecosystem/rust_test.go`:

```go
func TestRustEcosystem_ReadVersion(t *testing.T) {
    // Test version reading
}

func TestRustEcosystem_WriteVersion(t *testing.T) {
    // Test version writing
}
```

### 5. Update Documentation

- Add to `README.md` "Supported Ecosystems" section
- Add example in `examples/single-repo/rust-crate/`
- Update online documentation at shipyard.tamez.dev

## Adding New Commands

To add a new CLI command:

### 1. Create Command Handler

Create `internal/commands/yourcommand.go`:

```go
package commands

import (
    "github.com/spf13/cobra"
)

func NewYourCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "yourcommand [args]",
        Short: "Brief description",
        Long:  `Detailed description of what the command does.`,
        RunE:  runYourCommand,
    }

    // Add flags
    cmd.Flags().StringP("option", "o", "", "Option description")

    return cmd
}

func runYourCommand(cmd *cobra.Command, args []string) error {
    // Implementation
    return nil
}
```

### 2. Register Command

In `cmd/shipyard/main.go`:

```go
rootCmd.AddCommand(commands.NewYourCommand())
```

### 3. Add Tests

Create `internal/commands/yourcommand_test.go`:

```go
func TestYourCommand(t *testing.T) {
    // Test command behavior
}
```

### 4. Update Documentation

- Update online documentation at shipyard.tamez.dev
- Update README.md command list
- Add examples to examples/ directory

## Commit Conventions

We follow [Conventional Commits](https://www.conventionalcommits.org/):

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, no logic change)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks (dependencies, build, etc.)

### Examples

```
feat(ecosystem): add Rust ecosystem support

Implement RustEcosystem handler for Cargo.toml version management.
Includes version reading, writing, and detection.

Closes #123
```

```
fix(version): handle missing version file gracefully

Previously crashed when version file didn't exist.
Now returns clear error message.
```

```
docs: update CONTRIBUTING.md with ecosystem guide

Add step-by-step guide for adding new ecosystem support.
```

## Pull Request Process

### Before Submitting

1. **Create a consignment** (required for code changes):
   ```bash
   # Add a change note describing your changes
   shipyard add --summary "Brief description of change" --bump [major|minor|patch]

   # Commit the consignment with your changes
   git add .shipyard/consignments/
   git commit -m "feat: your feature description"
   ```

   **Note**: PRs that modify code must include a consignment. The `require-consignment` workflow will block merging without one. You can skip this requirement by adding the `skip-consignment` label for:
   - CI/workflow changes only
   - Test-only changes
   - Development tooling updates
   - Documentation-only PRs (automatically detected)

2. **Run all checks**:
   ```bash
   just ci
   ```

3. **Update tests** for new functionality

4. **Update documentation**:
   - README.md (if user-facing changes)
   - Code comments
   - Online documentation at shipyard.tamez.dev

5. **Self-review** your changes

### PR Template

```markdown
## Description

Brief description of changes.

## Type of Change

- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing

How was this tested?

- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed

## Checklist

- [ ] Consignment added (if code changes)
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex code
- [ ] Documentation updated
- [ ] Tests pass locally
- [ ] No new warnings
```

### Review Process

1. **Automated checks** must pass (CI/CD)
2. **Code review** by maintainer(s)
3. **Address feedback** and update PR
4. **Approval** and merge

## Useful Commands

All commands use [Just](https://github.com/casey/just) task runner:

### Build & Run

```bash
just build                 # Build binary to bin/shipyard
just install               # Install to $GOPATH/bin
just run [ARGS]            # Run CLI with arguments (e.g., just run status)
just build-all             # Build for all platforms
just release VERSION       # Release build with version info
```

### Testing

```bash
just test                  # Run all tests
just test-unit             # Run unit tests only
just test-integration      # Run integration tests only
just test-contract         # Run contract tests only
just coverage              # Generate coverage report
just bench                 # Run benchmarks
```

### Code Quality

```bash
just lint                  # Run linters
just fmt                   # Format code
just security              # Run security scanner (gosec)
just verify                # Verify go.mod is tidy
just ci                    # Run all CI checks (lint, test, verify)
```

### Dependencies

```bash
just check-deps            # Check for outdated dependencies
just update-deps           # Update all dependencies
```

### Development

```bash
just dev-setup             # Initialize development environment
just clean                 # Clean build artifacts
just mocks                 # Generate mocks (if using mockgen)
just watch                 # Watch for changes and run tests
```

### Examples

```bash
# Build and test
just build && just test

# Run CLI locally
just run init
just run add --summary "Test" --bump patch
just run status

# Full CI check before PR
just ci

# Generate coverage and open in browser
just coverage
```

## Releasing

Shipyard dogfoods itself for version management and release automation.

### Automated Release Process (Recommended)

1. **Create consignments for your changes**:
   ```bash
   # Add a change note
   shipyard add --summary "Add new feature" --bump minor

   # Or open PR with consignments committed
   ```

2. **Merge to main**:
   - When PR merges to main, GitHub Actions automatically:
     - Checks for pending consignments
     - Runs `shipyard version` to create tag
     - Pushes tag to trigger release build
     - GoReleaser builds binaries
     - Shipyard generates release notes

3. **Release is published**:
   - GitHub release created with binaries
   - Release notes generated from Shipyard
   - All platforms built and uploaded

### Manual Release Process

If you need to create a release manually:

1. **Ensure main branch is ready**:
   ```bash
   just ci  # Run all checks
   ```

2. **Use Shipyard to create version**:
   ```bash
   # Review pending changes
   shipyard status

   # Create version bump and git tag
   shipyard version
   ```

3. **Push to trigger release**:
   ```bash
   git push --follow-tags
   ```

4. **GitHub Actions will**:
   - Build binaries for all platforms
   - Generate and attach release notes
   - Publish GitHub release

### Testing Releases Locally

```bash
# Test GoReleaser configuration
just goreleaser-check

# Build snapshot (no tag needed)
just goreleaser-snapshot

# Binaries will be in dist/
# Test a binary:
./dist/shipyard_linux_amd64_v1/shipyard version
```

### Version Format

- Releases: `v1.0.0`, `v1.2.3`
- Pre-releases: `v1.0.0-rc.1`, `v1.0.0-beta.1`

### Dogfooding: Why Shipyard Manages Its Own Releases

Shipyard uses itself for version management:
- ✅ Consignments track changes
- ✅ `shipyard version` creates tags and updates changelogs
- ✅ GoReleaser builds binaries and publishes to GitHub
- ✅ `shipyard release-notes` generates release descriptions

This ensures Shipyard's features are battle-tested on real releases.

## Architecture Overview

### Key Components

1. **CLI Layer** (`internal/commands/`)
   - Command implementations (init, add, status, version, release-notes)
   - Flag parsing and validation
   - User interaction

2. **Config Layer** (`internal/config/`)
   - Configuration loading and parsing
   - Remote config fetching
   - Validation and merging

3. **Consignment Layer** (`internal/consignment/`)
   - Change note management
   - Markdown parsing and generation
   - Metadata validation

4. **Version Layer** (`internal/version/`)
   - Semantic version calculation
   - Dependency graph traversal
   - Version file updates

5. **Ecosystem Layer** (`internal/ecosystem/`)
   - Ecosystem-specific version handling
   - File format parsing/writing
   - Auto-detection

6. **Template Layer** (`internal/template/`)
   - Template rendering (changelog, tags, release notes)
   - Builtin template definitions
   - Custom template loading

7. **Git Layer** (`internal/git/`)
   - Git operations (commit, tag, push)
   - Repository state checking

### Data Flow

```
User Input
    ↓
CLI Commands (internal/commands)
    ↓
Config Loading (internal/config)
    ↓
Business Logic (internal/version, internal/consignment)
    ↓
Ecosystem Handlers (internal/ecosystem)
    ↓
File System / Git Operations
```

### Key Design Patterns

- **Strategy Pattern**: Ecosystem handlers implement common interface
- **Template Method**: Command execution follows standard flow
- **Factory Pattern**: Ecosystem registry for handler creation
- **Repository Pattern**: Config and history persistence

## Questions?

- **Bug reports**: [Open an issue](https://github.com/natonathan/shipyard/issues)
- **Feature requests**: [Open an issue](https://github.com/natonathan/shipyard/issues)
- **Questions**: [Open a discussion](https://github.com/natonathan/shipyard/discussions)
- **Security issues**: Email security@shipyard.tamez.dev

---

Thank you for contributing to Shipyard! 🚀
