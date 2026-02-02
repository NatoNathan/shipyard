# Changelog

All notable changes to this project will be documented in this file.

## [0.1.0] - 2026-02-02

### Minor

Added repository initialization with shipyard init command

Implemented the init command to bootstrap Shipyard in any repository. Supports both single-repo and monorepo configurations with interactive prompts.

## Features

- Interactive setup wizard with package detection
- Automatic ecosystem detection (Go, NPM, Python, Docker)
- Monorepo vs single-repo selection
- Remote configuration support
- Default template initialization
- Creates .shipyard/ directory structure

## Commands

- shipyard init - Interactive initialization
- shipyard init --remote <url> - Initialize from remote config

## Details

- Detects existing package.json, go.mod, pyproject.toml
- Creates shipyard.yaml configuration file
- Sets up consignments and history directories
- Validates configuration on creation

### Minor

Added consignment management with shipyard add command

Implemented the add command for creating change notes (consignments) as markdown files with YAML frontmatter.

## Features

- Interactive mode with prompts for package, type, and summary
- Non-interactive mode with flags for automation
- Multi-package support for changes affecting multiple components
- Custom metadata support (author, issue, etc.)
- Automatic ID and timestamp generation
- Editor integration for detailed summaries

## File Format

- YAML frontmatter with id, timestamp, packages, changeType
- Markdown body for change description
- Stored in .shipyard/consignments/
- Unique ID format: YYYYMMDD-HHMMSS-random6

## Validation

- Package name validation against config
- Change type validation (patch, minor, major)
- Summary required and non-empty

### Minor

Added status command for viewing pending changes

Implemented the status command to preview pending consignments and calculated version bumps before applying them.

## Features

- Lists all pending consignments grouped by package
- Shows current version and proposed new version for each package
- Displays change type (patch, minor, major) for each consignment
- Interactive table view with color-coded change types
- JSON output mode for automation
- Package filtering support

## Output

- Package name and current version
- Proposed version bump with semantic versioning
- List of pending consignments with summaries
- Change type indicators (🔴 major, 🟡 minor, 🟢 patch)
- Total count of changes per package

## Use Cases

- Review changes before version bump
- Verify correct change types
- Check version propagation effects
- Generate status reports for CI/CD

### Minor

Added version command for applying semantic version bumps

Implemented the version command to calculate and apply version bumps based on pending consignments.

## Features

- Automatic semantic version calculation from consignments
- Dependency-aware version propagation
- Updates version files across all affected packages
- Generates changelog entries
- Archives processed consignments to history before committing
- Deletes processed consignment files before committing
- Creates git commit with all changes (versions, changelogs, history, deletions)
- Creates git tags for each version
- Dry-run mode with --preview flag

## Version Bumping

- Patch: x.y.Z increments (bug fixes)
- Minor: x.Y.0 increments (new features)
- Major: X.0.0 increments (breaking changes)
- Highest change type wins per package

## Operation Order

1. Reads pending consignments from .shipyard/consignments/
2. Calculates version bumps for each package (including dependency propagation)
3. Updates version files in ecosystem-specific formats
4. Generates changelogs using configured templates
5. Archives processed consignments to history.json
6. Deletes processed consignment files
7. Stages all changes (versions, changelogs, history.json, deleted consignments)
8. Creates git commit with detailed message
9. Creates annotated git tags for each package

## Git Integration

- Creates annotated tags (e.g., v1.2.3 or core-v1.2.3)
- Commits all version-related changes atomically
- Includes history archival and consignment cleanup in commit
- Preserves git history with detailed messages

## Safety

- Preview mode shows changes without applying
- Validates all consignments before processing
- Atomic operations with rollback on failure
- Warns about uncommitted changes
- No-commit and no-tag flags for incremental workflows

### Minor

Added release-notes command for changelog generation

Implemented the release-notes command to generate formatted release notes from version history.

## Features

- Generates release notes from archived consignments
- Template-based formatting with customization
- Multiple output formats (markdown, JSON)
- Package-specific or repository-wide notes
- Version filtering and date ranges
- GitHub release integration (optional)

## Output Options

- Print to stdout for display
- Write to file for documentation
- Publish to GitHub releases API
- Multiple format support

## Template Support

- Default built-in template
- Custom templates via configuration
- Access to full consignment metadata
- Group by change type or package
- Custom formatting with Go templates

## GitHub Integration

- Automatic release creation
- Markdown formatting for GitHub
- Attachment support
- Draft and prerelease options

### Minor

Added monorepo support with dependency graph management

Implemented comprehensive monorepo support with dependency tracking and topological ordering.

## Features

- Multiple package management in single repository
- Dependency graph construction and validation
- Cycle detection with detailed error reporting
- Topological sorting for correct build order
- Package-specific version tracking
- Independent or linked versioning strategies

## Dependency Strategies

- **fixed**: Pin to specific version (e.g., 1.2.3)
- **linked**: Follow parent package version exactly
- **independent**: Version independently

## Graph Operations

- Build dependency graph from configuration
- Detect circular dependencies
- Compute topological sort for processing order
- Find strongly connected components (Tarjan algorithm)
- Cache graph for performance

## Validation

- Validates package references exist
- Prevents circular dependencies
- Checks version strategy compatibility
- Reports detailed error locations

### Minor

Added version propagation for dependency-aware versioning

Implemented automatic version propagation through dependency chains.

## Features

- Automatic version bumps for dependent packages
- Cascade propagation through dependency chains
- Strategy-aware propagation (linked vs fixed vs independent)
- Conflict detection and resolution
- Minimum version calculation across multiple dependencies
- Direct vs transitive dependency handling

## Propagation Rules

- **Linked dependencies**: Always match parent version
- **Fixed dependencies**: Update when parent changes
- **Independent**: No automatic propagation
- Multiple sources: Highest version wins

## Conflict Resolution

- Detects incompatible version requirements
- Reports conflict chains with full path
- Suggests resolution strategies
- Validates propagation results

## Use Cases

- Shared library updates propagate to consumers
- Breaking changes cascade appropriately
- Internal dependencies stay synchronized
- Prevent version skew in monorepos

### Minor

Added template system for customizable output formats

Implemented flexible template system for changelogs and release notes with built-in and custom template support.

## Features

- Go template-based rendering engine
- Built-in default templates for common formats
- Custom template support via files or remote URLs
- Template validation and error reporting
- Rich template functions for formatting
- Access to full consignment metadata

## Built-in Templates

- **builtin:default** - Standard changelog format
- **builtin:github** - GitHub-optimized markdown
- **builtin:keepachangelog** - Keep a Changelog format
- Customizable tag name format

## Custom Templates

- Load from local files
- Load from remote URLs (Git, HTTP)
- Inline templates in configuration
- Template inheritance and includes

## Template Data

- Package information (name, version, ecosystem)
- Consignment list with metadata
- Version history
- Timestamp and date functions
- Custom metadata access

## Functions

- String manipulation (upper, lower, title, trim)
- Date formatting
- Collection operations (join, range, filter)
- Markdown rendering

### Minor

Added multi-ecosystem support for Go, NPM, Python, Helm, and Cargo

Implemented ecosystem-specific version file handling for multiple programming languages and platforms.

## Supported Ecosystems

### Go
- Version file: version.go or any .go file
- Format: const Version = "X.Y.Z"
- Module-aware updates

### NPM
- Version file: package.json
- Format: "version": "X.Y.Z"
- Preserves JSON formatting and other fields

### Python
- Multiple file support:
  - pyproject.toml: version = "X.Y.Z"
  - setup.py: version="X.Y.Z"
  - __version__.py: __version__ = "X.Y.Z"
- PEP 440 compliant versioning

### Helm
- Version file: Chart.yaml
- Format: version: X.Y.Z
- Updates both version and appVersion fields
- Preserves chart metadata

### Cargo (Rust)
- Version file: Cargo.toml
- Format: version = "X.Y.Z"
- Preserves TOML structure and dependencies

## Features

- Automatic ecosystem detection
- Custom version file paths
- Multiple version files per package
- Preserves file formatting
- Validates version format per ecosystem
- Idempotent updates
- GetVersionFiles() returns relative paths for correct staging
- Comprehensive test coverage for all ecosystems

### Minor

Added configuration system with local and remote config support

Implemented flexible configuration system with local files and remote config loading.

## Configuration File

- Default: .shipyard/shipyard.yaml
- YAML format with full schema validation
- Package definitions with dependencies
- Template configuration
- GitHub integration settings

## Remote Configuration

- Load configuration from Git repositories
- Load from HTTP/HTTPS URLs
- Configuration inheritance and overrides
- Local config overrides remote config
- Caching for performance

## Configuration Schema

### Packages
- name: Package identifier
- path: Package directory path
- ecosystem: Language/platform type
- dependencies: Package dependency list with strategies

### Templates
- changelog: Changelog template source
- tagName: Git tag format template
- releaseNotes: Release notes template

### Paths
- consignments: Pending changes directory
- history: Archived changes file

### GitHub Integration
- enabled: Enable/disable GitHub features
- owner: Repository owner
- repo: Repository name

## Validation

- Schema validation on load
- Package reference validation
- Dependency cycle detection
- Template source validation
- Path existence checks

