# Shipyard Command Reference

Shipyard is a semantic versioning and release management tool for monorepos and single-package repositories. This comprehensive reference guide documents all 14 commands available in the Shipyard CLI. Each command includes detailed usage information, examples, and integration patterns to help you manage versions, track changes, and automate releases.

## Table of Contents

1. [add](#add---log-cargo-in-the-ships-manifest) - Log cargo in the ship's manifest
2. [completion](#completion---teach-your-shell-to-speak-shipyard) - Teach your shell to speak Shipyard
3. [config show](#config-show---read-the-ships-charter) - Read the ship's charter
4. [init](#init---set-sail---prepare-your-repository) - Set sail - prepare your repository
5. [prerelease](#prerelease---create-or-increment-a-pre-release-version-at-the-current-stage) - Create or increment a pre-release version
6. [promote](#promote---advance-through-the-harbor-channel) - Advance through the harbor channel
7. [release](#release---signal-arrival-at-port) - Signal arrival at port
8. [release-notes](#release-notes---tell-the-tale-of-your-voyage) - Tell the tale of your voyage
9. [remove](#remove---jettison-cargo-from-the-manifest) - Jettison cargo from the manifest
10. [snapshot](#snapshot---create-a-timestamped-snapshot-pre-release-version) - Create a timestamped snapshot pre-release version
11. [status](#status---check-cargo-and-chart-your-course) - Check cargo and chart your course
12. [upgrade](#upgrade---refit-the-shipyard-with-latest-provisions) - Refit the shipyard with latest provisions
13. [validate](#validate---inspect-the-hull-before-departure) - Inspect the hull before departure
14. [version](#version---set-sail-to-the-next-port) - Set sail to the next port

---

## add - Log cargo in the ship's manifest

### Synopsis

```bash
shipyard add [OPTIONS]
shipyard consign [OPTIONS]
shipyard log [OPTIONS]
```

**Aliases:** `consign`, `log`

### Description

The `add` command records a new consignment (change entry) in your repository. It:

1. Validates package names and change type
2. Validates metadata against configured fields
3. Generates a unique consignment ID
4. Writes a `.md` file to `.shipyard/consignments/`

Supports both interactive mode (prompts for input) and non-interactive mode (via flags).

**Maritime Metaphor**: Log cargo in the ship's manifest—documenting what's being shipped, which vessels carry it, and how it affects the voyage.

### Global Options

These options are available for all shipyard commands:

| Option | Short | Description |
|--------|-------|-------------|
| `--config <path>` | `-c` | Config file (default: `.shipyard/shipyard.yaml`) |
| `--json` | `-j` | Output in JSON format |
| `--quiet` | `-q` | Suppress non-error output |
| `--verbose` | `-v` | Verbose output |

### Options

#### `--package <name>`, `-p`

Package name(s) affected by this change. Can be repeated for multiple packages.

```bash
shipyard add --package core
shipyard add --package core --package api
```

#### `--type <type>`, `-t`

Change type: `patch`, `minor`, or `major`.

```bash
shipyard add --type minor
```

#### `--summary <text>`, `-s`

Summary of the change.

```bash
shipyard add --summary "Add new API endpoint"
```

#### `--metadata <key=value>`, `-m`

Custom metadata in `key=value` format. Can be repeated. Keys must match fields defined in `shipyard.yaml`.

```bash
shipyard add --metadata author=dev@example.com
shipyard add --metadata author=dev@example.com --metadata issue=JIRA-123
```

### Examples

#### Interactive Mode

When flags are omitted, prompts guide you through the process:

```bash
shipyard add
```

#### Non-Interactive Mode

Provide all required flags:

```bash
shipyard add --package core --type minor --summary "Add new feature"
```

#### Multiple Packages

```bash
shipyard add --package core --package api --type major --summary "Breaking API change"
```

#### With Metadata

```bash
shipyard add --package core --type patch --summary "Fix null pointer" \
  --metadata author=dev@example.com --metadata issue=BUG-456
```

#### Single-Package Repository

For repos with one package, `--package` can still be omitted in interactive mode:

```bash
shipyard add --type patch --summary "Quick fix"
```

### Output

On success:

```
✓ Created consignment: 20240130-120000-abc123.md

Packages: core, api
Type:     minor
Summary:  Add new feature
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success - consignment created |
| 1 | Error - validation failed or not a git repository |

### Behavior Details

#### Interactive vs Non-Interactive

- **Interactive**: If `--package`, `--type`, or `--summary` is missing, prompts for input
- **Non-Interactive**: If all three are provided, runs without prompts

#### Package Validation

Package names must exist in `shipyard.yaml`. Invalid packages return an error listing available options.

#### Change Type Validation

Only `patch`, `minor`, and `major` are accepted.

#### Metadata Validation

If metadata fields are configured in `shipyard.yaml`, provided values are validated:
- Keys must match configured field names
- Values must match allowed options (if defined)

#### Consignment ID Format

Generated as `YYYYMMDD-HHMMSS-<random>` based on current UTC time.

#### Git Requirement

Must be run inside a git repository.

### Related Commands

- [`status`](./status.md) - View pending consignments
- [`version`](./version.md) - Process consignments into versions

### See Also

- [Consignment Format](../consignment-format.md) - Structure of consignment files
- [Configuration Reference](../configuration.md) - Metadata field definitions

---

## completion - Teach your shell to speak Shipyard

### Synopsis

```bash
shipyard completion [bash|zsh|fish|powershell]
```

### Description

The `completion` command generates shell completion scripts. It enables your shell to suggest commands, flags, and arguments as you type.

Supports Bash, Zsh, Fish, and PowerShell.

**Maritime Metaphor**: Train your shell to understand the shipyard's language—let your navigator suggest the course.

### Global Options

These options are available for all shipyard commands:

| Option | Short | Description |
|--------|-------|-------------|
| `--config <path>` | `-c` | Config file (default: `.shipyard/shipyard.yaml`) |
| `--json` | `-j` | Output in JSON format |
| `--quiet` | `-q` | Suppress non-error output |
| `--verbose` | `-v` | Verbose output |

### Arguments

The shell name is required. Valid values: `bash`, `zsh`, `fish`, `powershell`.

### Installation

#### Bash

Linux:

```bash
shipyard completion bash > /etc/bash_completion.d/shipyard
```

macOS:

```bash
shipyard completion bash > /usr/local/etc/bash_completion.d/shipyard
```

Or add to `~/.bashrc`:

```bash
echo 'source <(shipyard completion bash)' >> ~/.bashrc
```

#### Zsh

```bash
shipyard completion zsh > "${fpath[1]}/_shipyard"
```

Or add to `~/.zshrc`:

```bash
echo 'source <(shipyard completion zsh)' >> ~/.zshrc
echo 'compdef _shipyard shipyard' >> ~/.zshrc
```

#### Fish

```bash
shipyard completion fish > ~/.config/fish/completions/shipyard.fish
```

#### PowerShell

Add to your profile:

```powershell
shipyard completion powershell | Out-String | Invoke-Expression
```

Or save to profile:

```powershell
shipyard completion powershell >> $PROFILE
```

After installing, restart your shell or source the completion file.

### Features

Completions include:

- Command names (`init`, `add`, `version`, etc.)
- Flag names and values
- Package names from `shipyard.yaml`
- Change types (`patch`, `minor`, `major`)

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success - completion script generated |
| 1 | Error - invalid shell name |

### Related Commands

- [`upgrade`](./upgrade.md) - Upgrade shipyard

### See Also

- [Getting Started](../getting-started.md) - Initial setup

---

## config show - Read the ship's charter

### Synopsis

```bash
shipyard config show [OPTIONS]
shipyard config view [OPTIONS]
shipyard cfg show [OPTIONS]
shipyard cfg view [OPTIONS]
```

**Aliases:** `view` (for `show`), `cfg` (for `config`)

### Description

The `config show` command displays the current shipyard configuration with all defaults applied. It:

1. Loads the configuration from `.shipyard/shipyard.yaml`
2. Resolves any remote/extended configurations
3. Applies default values for unset fields
4. Outputs the full resolved configuration

Outputs as YAML by default, or JSON with the `--json` flag.

**Maritime Metaphor**: Read the ship's charter—see the full orders including all standing instructions.

### Global Options

These options are available for all shipyard commands:

| Option | Short | Description |
|--------|-------|-------------|
| `--config <path>` | `-c` | Config file (default: `.shipyard/shipyard.yaml`) |
| `--json` | `-j` | Output in JSON format |
| `--quiet` | `-q` | Suppress non-error output |
| `--verbose` | `-v` | Verbose output |

### Examples

#### Default YAML Output

```bash
shipyard config show
```

```yaml
packages:
  - name: core
    path: ./core
    ecosystem: go
    dependencies: []
templates:
  changelog:
    source: builtin:default
  tagname:
    source: builtin:default
  releasenotes:
    source: builtin:default
consignments:
  path: .shipyard/consignments
history:
  path: .shipyard/history.json
github:
  owner: ""
  repo: ""
prerelease:
  stages: []
```

#### JSON Output

```bash
shipyard config show --json
```

```json
{
  "Packages": [
    {
      "Name": "core",
      "Path": "./core",
      "Ecosystem": "go"
    }
  ],
  "Templates": {
    "Changelog": { "Source": "builtin:default" },
    "TagName": { "Source": "builtin:default" }
  }
}
```

#### Multi-Package Repository

```bash
shipyard config show
```

Shows all packages, their ecosystems, dependency relationships, and template overrides.

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success - configuration displayed |
| 1 | Error - failed to load or marshal configuration |

### Behavior Details

#### Resolved Configuration

The output shows the **resolved** configuration after applying:
- Default values for unset fields
- Merged remote/extended configurations
- Default template sources (`builtin:default`)
- Default paths for consignments and history

#### YAML vs JSON

- **YAML** (default): Uses lowercase keys matching the config file format
- **JSON** (`--json`): Uses PascalCase keys matching Go struct field names

#### Not Initialized

If no configuration file exists, returns an error.

### Related Commands

- [`init`](./init.md) - Initialize shipyard configuration
- [`validate`](./validate.md) - Validate configuration for errors

### See Also

- [Configuration Reference](../configuration.md) - Full configuration file format

---

## init - Set sail - prepare your repository

### Synopsis

```bash
shipyard init [OPTIONS]
shipyard setup [OPTIONS]
```

**Aliases:** `setup`

### Description

The `init` command prepares a repository for versioning with Shipyard. It:

1. Verifies the current directory is a git repository
2. Creates the `.shipyard/` directory structure
3. Detects packages in the repository
4. Generates `shipyard.yaml` configuration
5. Initializes an empty `history.json`

Supports interactive mode (prompts for configuration) and non-interactive mode (`--yes`).

**Maritime Metaphor**: Prepare your repository for the versioning voyage ahead—set up cargo manifests, navigation charts, and the captain's log.

### Global Options

These options are available for all shipyard commands:

| Option | Short | Description |
|--------|-------|-------------|
| `--config <path>` | `-c` | Config file (default: `.shipyard/shipyard.yaml`) |
| `--json` | `-j` | Output in JSON format |
| `--quiet` | `-q` | Suppress non-error output |
| `--verbose` | `-v` | Verbose output |

### Options

#### `--force`, `-f`

Force re-initialization if already initialized.

```bash
shipyard init --force
```

#### `--remote <url>`, `-r`

Extend from a remote configuration URL.

```bash
shipyard init --remote https://example.com/shipyard-config.yaml
```

#### `--yes`, `-y`

Skip all prompts and accept defaults. Uses auto-detected packages.

```bash
shipyard init --yes
```

### Examples

#### Interactive Mode (Default)

```bash
shipyard init
```

Prompts for:
- Repository type (single package or monorepo)
- Package selection/configuration
- Package details (name, path, ecosystem)

#### Non-Interactive Mode

```bash
shipyard init --yes
```

Uses auto-detected packages or creates a default package if none found.

#### Re-initialize

```bash
shipyard init --force
```

#### With Remote Config

```bash
shipyard init --remote https://example.com/shared-config.yaml
```

### Output

On success:

```
✓ Shipyard initialized successfully

Configuration:          .shipyard/shipyard.yaml
Consignments directory: .shipyard/consignments
History file:           .shipyard/history.json
```

### Created Files

| Path | Description |
|------|-------------|
| `.shipyard/shipyard.yaml` | Main configuration file |
| `.shipyard/consignments/` | Directory for pending consignments |
| `.shipyard/history.json` | Version history (empty array initially) |

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success - repository initialized |
| 1 | Error - not a git repo, already initialized, or file operation failed |

### Behavior Details

#### Package Detection

Automatically detects packages by looking for:
- `package.json` (npm)
- `go.mod` (Go)
- `Cargo.toml` (Cargo)
- `Chart.yaml` (Helm)
- `setup.py` / `pyproject.toml` (Python)
- `deno.json` (Deno)

#### Already Initialized

Without `--force`, returns an error if `.shipyard/shipyard.yaml` exists.

#### Git Requirement

Must be run inside a git repository.

#### Default Package

If no packages are detected in `--yes` mode, creates a default package:

```yaml
packages:
  - name: default
    path: ./
    ecosystem: go
```

### Related Commands

- [`add`](./add.md) - Create consignments after initialization
- [`status`](./status.md) - View pending consignments

### See Also

- [Configuration Reference](../configuration.md) - Full shipyard.yaml format
- [Getting Started](../getting-started.md) - First-time setup guide

---

## prerelease - Create or increment a pre-release version at the current stage

### Synopsis

```bash
shipyard version prerelease [OPTIONS]
shipyard version pre [OPTIONS]
shipyard version rc [OPTIONS]
```

**Aliases:** `pre`, `rc`

Creates or increments a pre-release version at the current stage based on pending consignments.

### Description

The `prerelease` command creates pre-release versions for testing changes before creating a stable release. Like charting test waters before the main voyage, pre-releases let you validate changes without committing to a final version.

**Key behaviors:**

- **Stage-based**: Pre-release stage comes from `.shipyard/prerelease.yml` state file, not a command argument
- **First pre-release**: Automatically starts at the stage with the lowest `order` value
- **Incremental**: Subsequent runs increment the counter at the current stage (e.g., `alpha.1` → `alpha.2`)
- **Target version**: Calculated from pending consignments, just like the `version` command
- **Consignments preserved**: Consignments remain in `.shipyard/consignments/` for the eventual stable release
- **Changelog deferred**: Changelog updates are deferred until stable release
- **Git tags**: Creates tags using stage-specific templates

To advance to the next stage, use [`shipyard version promote`](./promote.md). To create timestamp-based snapshot builds, use [`shipyard version snapshot`](./snapshot.md).

### Global Options

| Option | Description |
|--------|-------------|
| `--config`, `-c` | Path to configuration file (default: `.shipyard/shipyard.yaml`) |
| `--json`, `-j` | Output in JSON format |
| `--quiet`, `-q` | Suppress all output except errors |
| `--verbose`, `-v` | Enable verbose logging |

### Options

#### `--preview`

Show what pre-release version would be created without making any changes.

**Example:**
```bash
$ shipyard version prerelease --preview
📦 Preview: Pre-release version changes
  - my-api: 1.2.0-beta.1 → 1.2.0-beta.2 (beta)
    Target version: 1.2.0
    Based on consignments:
      - 20240130-120000-abc123: Add new API endpoint

ℹ Preview mode: no changes made
```

#### `--no-commit`

Apply version changes to files but skip creating a git commit. Useful for reviewing changes before committing.

**Example:**
```bash
$ shipyard version prerelease --no-commit
📦 Creating pre-release versions...
  - my-api: 1.2.0-alpha.1 → 1.2.0-alpha.2 (alpha)
✓ Updated version files
⊘ Skipped git commit (--no-commit)
```

#### `--no-tag`

Create the git commit but skip creating git tags. Useful for testing version file updates.

**Example:**
```bash
$ shipyard version prerelease --no-tag
📦 Creating pre-release versions...
  - my-api: 1.2.0-alpha.1 → 1.2.0-alpha.2 (alpha)
✓ Created commit: "chore: pre-release my-api v1.2.0-alpha.2"
⊘ Skipped git tags (--no-tag)
```

#### `--package <name>`

Create pre-release only for specific packages in a multi-package repository.

**Example:**
```bash
$ shipyard version prerelease --package cli
📦 Creating pre-release versions...
  - cli: 2.0.0 → 2.1.0-alpha.1 (alpha)
  - api: skipped (not in --package filter)
✓ Created commit: "chore: pre-release cli v2.1.0-alpha.1"
✓ Created tag: cli-v2.1.0-alpha.1
```

### Stage System

#### Stage Determination

The current stage is read from `.shipyard/prerelease.yml`:

- **First pre-release**: Starts at the stage with the lowest `order` value
- **Subsequent pre-releases**: Increments the counter at the current stage
- **Stage advancement**: Use `shipyard version promote` to move to the next stage

#### Stage Configuration

Stages are defined in `.shipyard/shipyard.yaml`:

```yaml
prerelease:
  stages:
    - name: alpha
      order: 1
      tagTemplate: "v{{.Version}}-alpha.{{.Counter}}"
    - name: beta
      order: 2
      tagTemplate: "v{{.Version}}-beta.{{.Counter}}"
    - name: rc
      order: 3
      tagTemplate: "v{{.Version}}-rc.{{.Counter}}"
```

**Stage fields:**

- `name`: Stage identifier (used in state file)
- `order`: Numeric order (lower = earlier stage)
- `tagTemplate`: Go template for rendering tags

#### Custom Stages

Teams can define their own stages to match their workflow:

```yaml
prerelease:
  stages:
    - name: dev
      order: 1
      tagTemplate: "v{{.Version}}-dev.{{.Counter}}"
    - name: staging
      order: 2
      tagTemplate: "v{{.Version}}-staging.{{.Counter}}"
    - name: prod-ready
      order: 3
      tagTemplate: "v{{.Version}}-rc.{{.Counter}}"
```

Not all teams use alpha/beta/rc—customize stages to fit your development process!

### Workflow

The `prerelease` command executes the following steps:

1. **Load Configuration** - Read `.shipyard/shipyard.yaml` for stage definitions
2. **Load State** - Read `.shipyard/prerelease.yml` for current stage and counter
3. **Determine Stage** - If first pre-release, use stage with lowest order
4. **Read Consignments** - Load pending consignments from `.shipyard/consignments/`
5. **Calculate Target Version** - Calculate next stable version from consignments
6. **Check Target Version** - If target changed, warn user and reset counter
7. **Increment Counter** - Increment counter for current stage
8. **Render Tag** - Apply stage's tag template with version and counter
9. **Update Version Files** - Write pre-release version to ecosystem files
10. **Git Operations** - Create commit and tags (unless `--no-commit`)
11. **Update State** - Save stage, counter, and target to `.shipyard/prerelease.yml`
12. **Skip**: Consignment archival and changelog generation (deferred to stable release)

### Important Behaviors

**Target version changes**: If consignments are modified between pre-releases, the target version may change. When this happens:
- The counter resets to 1
- A warning is displayed
- The new target version is used

**State file creation**: The `.shipyard/prerelease.yml` file is created automatically on the first pre-release if it doesn't exist.

**Pre-release versions in history**: Pre-release versions are NOT recorded in `.shipyard/history.json`. Only stable releases are tracked in history.

**Changelog updates**: Changelogs are NOT regenerated during pre-releases. Changelog updates happen only on stable release.

### Configuration

#### Main Config (`.shipyard/shipyard.yaml`)

Define stages with order and tag templates:

```yaml
prerelease:
  stages:
    - name: alpha
      order: 1
      tagTemplate: "v{{.Version}}-alpha.{{.Counter}}"
    - name: beta
      order: 2
      tagTemplate: "v{{.Version}}-beta.{{.Counter}}"
    - name: rc
      order: 3
      tagTemplate: "v{{.Version}}-rc.{{.Counter}}"

  # Snapshot template (used by 'shipyard version snapshot')
  snapshotTagTemplate: "v{{.Version}}-snapshot.{{.Timestamp}}"
```

#### Template Variables

Tag templates support the following variables:

- `{{.Version}}`: Target stable version (e.g., "1.2.0")
- `{{.Counter}}`: Current pre-release counter (e.g., 1, 2, 3)
- `{{.Package}}`: Package name (for multi-package projects)
- `{{.Timestamp}}`: Timestamp in format YYYYMMDD-HHMMSS (snapshots only)

**Examples:**

```yaml
# Standard format
tagTemplate: "v{{.Version}}-alpha.{{.Counter}}"  # v1.2.0-alpha.1

# With package prefix (multi-package repos)
tagTemplate: "{{.Package}}-v{{.Version}}-beta.{{.Counter}}"  # api-v1.2.0-beta.1

# Custom identifier
tagTemplate: "v{{.Version}}-preview.{{.Counter}}"  # v1.2.0-preview.1
```

#### State File (`.shipyard/prerelease.yml`)

Tracks dynamic pre-release state:

```yaml
packages:
  my-api:
    stage: beta
    counter: 2
    targetVersion: 1.2.0
  shared-lib:
    stage: alpha
    counter: 5
    targetVersion: 0.5.0
```

**State fields:**

- `stage`: Current stage name (matches main config stage)
- `counter`: Current pre-release counter for this stage
- `targetVersion`: Target stable version calculated from consignments

#### Why Separate Files?

- **Main config** (`.shipyard/shipyard.yaml`): Static stage definitions, committed to git
- **State file** (`.shipyard/prerelease.yml`): Dynamic tracking data, can be gitignored or committed
- Keeps main configuration clean and focused on project structure

### Examples

#### First Pre-Release (No State File Exists)

```bash
$ shipyard version prerelease
📦 Creating pre-release versions...
  - my-api: 1.1.5 → 1.2.0-alpha.1 (alpha, first pre-release)
✓ Created .shipyard/prerelease.yml
✓ Created commit: "chore: pre-release my-api v1.2.0-alpha.1"
✓ Created tag: v1.2.0-alpha.1
```

#### Increment Pre-Release (Same Stage)

```bash
$ shipyard version prerelease
📦 Creating pre-release versions...
  - my-api: 1.2.0-alpha.1 → 1.2.0-alpha.2 (alpha)
✓ Created commit: "chore: pre-release my-api v1.2.0-alpha.2"
✓ Created tag: v1.2.0-alpha.2
```

#### Promote to Next Stage

```bash
$ shipyard version promote
📦 Promoting to next stage...
  - my-api: 1.2.0-alpha.2 → 1.2.0-beta.1 (alpha → beta)
✓ Updated .shipyard/prerelease.yml
✓ Created commit: "chore: pre-release my-api v1.2.0-beta.1"
✓ Created tag: v1.2.0-beta.1
```

#### Preview Changes

```bash
$ shipyard version prerelease --preview
📦 Preview: Pre-release version changes
  - my-api: 1.2.0-beta.1 → 1.2.0-beta.2 (beta)
    Target version: 1.2.0
    Based on consignments:
      - 20240130-120000-abc123: Add new API endpoint

ℹ Preview mode: no changes made
```

#### Specific Package Only

```bash
$ shipyard version prerelease --package cli
📦 Creating pre-release versions...
  - cli: 2.0.0 → 2.1.0-alpha.1 (alpha)
  - api: skipped (not in --package filter)
✓ Created commit: "chore: pre-release cli v2.1.0-alpha.1"
✓ Created tag: cli-v2.1.0-alpha.1
```

#### Target Version Changed (Consignments Modified)

```bash
# Was on 1.2.0-beta.2, then added major change
$ shipyard version prerelease
⚠ Warning: Target version changed from 1.2.0 to 2.0.0 (consignments modified)
📦 Creating pre-release versions...
  - my-api: 1.2.0-beta.2 → 2.0.0-beta.1 (beta, counter reset)
✓ Updated .shipyard/prerelease.yml
✓ Created commit: "chore: pre-release my-api v2.0.0-beta.1"
✓ Created tag: v2.0.0-beta.1
```

#### Promote to Stable Release

```bash
$ shipyard version
📦 Promoting pre-release to stable...
  - my-api: 1.2.0-rc.1 → 1.2.0 (minor)
    - 20240130-120000-abc123: Add new API endpoint
✓ Archived consignments to history
✓ Updated CHANGELOG.md
✓ Deleted .shipyard/prerelease.yml
✓ Created commit: "chore: release my-api v1.2.0"
✓ Created tag: v1.2.0
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error - validation, file, or git operation failed |
| 2 | No consignments to base pre-release on |

### Behavior Details

#### Consignment Changes During Pre-Release

If consignments are added, removed, or modified between pre-release versions:
- Target version recalculates
- Counter resets to 1
- User is warned about the target version change

**Example:**

```bash
# Start with 1.2.0-alpha.1 based on minor change
$ shipyard version prerelease

# Add major change consignment
$ shipyard consign major "Breaking API change"

# Create new pre-release
$ shipyard version prerelease
⚠ Warning: Target version changed from 1.2.0 to 2.0.0
📦 Creating pre-release versions...
  - my-api: 1.2.0-alpha.1 → 2.0.0-alpha.1 (counter reset)
```

#### Stage Progression

Typical workflow through stages:

1. **Development** → `prerelease` repeatedly (alpha.1, alpha.2, ...)
2. **Promotion** → `promote` (alpha → beta)
3. **Testing** → `prerelease` repeatedly (beta.1, beta.2, ...)
4. **Stabilization** → `promote` (beta → rc)
5. **Final testing** → `prerelease` repeatedly (rc.1, rc.2, ...)
6. **Release** → `version` (promotes to stable 1.2.0, deletes state file)

Stages follow the order defined in configuration. Use `promote` to advance between stages.

#### Snapshot Behavior

For timestamp-based builds independent of the stage system, use [`shipyard version snapshot`](./snapshot.md):

- Snapshots use timestamp: `YYYYMMDD-HHMMSS`
- Don't affect pre-release stage/counter tracking
- Useful for PR builds, CI pipelines, ad-hoc testing
- Independent of stage-based workflow

#### Multi-Package Projects

Each package tracks its own pre-release state independently:

- Can have `api@1.2.0-alpha.1` and `cli@3.0.0-beta.2` simultaneously
- State file contains separate entries for each package
- Dependencies: Linked dependencies inherit pre-release stage

**Example state file:**

```yaml
packages:
  api:
    stage: alpha
    counter: 1
    targetVersion: 1.2.0
  cli:
    stage: beta
    counter: 2
    targetVersion: 3.0.0
```

#### State File Management

The `.shipyard/prerelease.yml` state file:

- Created automatically on first pre-release
- Deleted automatically when promoting to stable release with `shipyard version`
- **Should be committed to git** (not `.gitignore`d) - this allows your team to track pre-release state across branches
- `shipyard version` will remove the file entirely for single-package repos, or remove the entry for packages promoted to stable in multi-package repos
- If deleted manually, next prerelease starts at lowest order stage

#### Git Requirements

Same requirements as the `version` command:

- Repository must be a git repository
- Working directory must be clean (or use `--no-commit`)
- Git `user.name` and `user.email` must be configured

### Integration with CI/CD

#### Automatic Pre-Release on Push

Example GitHub Actions workflow for automatic pre-releases:

```yaml
name: Pre-Release
on:
  push:
    branches: [develop]

jobs:
  prerelease:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0  # Fetch all history for tags

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Install Shipyard
        run: |
          curl -fsSL https://shipyard.build/install.sh | sh

      - name: Create pre-release
        run: shipyard version prerelease

      - name: Push changes
        run: |
          git push origin develop
          git push --tags
```

#### Stage Promotion on Branch

Example workflow for promoting to beta when merged to beta branch:

```yaml
name: Promote to Beta
on:
  push:
    branches: [beta]

jobs:
  promote:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0

      - name: Install Shipyard
        run: curl -fsSL https://shipyard.build/install.sh | sh

      - name: Promote to beta
        run: shipyard version promote

      - name: Push changes
        run: |
          git push origin beta
          git push --tags
```

### Related Commands

- [`version`](./version.md) - Create stable release, clears pre-release state
- [`version promote`](./promote.md) - Promote to next stage or stable
- [`version snapshot`](./snapshot.md) - Create timestamped snapshot build
- [`release-notes`](./release-notes.md) - Generate notes (includes pre-release tags)
- [`release`](./release.md) - Publish release (supports `--prerelease` flag)

### See Also

- [Configuration Reference](../configuration.md)
- [Tag Templates](../tag-templates.md)
- [Consignment Format](../consignment-format.md)
- [CI/CD Integration Guide](../ci-cd.md)

---

## promote - Advance through the harbor channel

Promote a pre-release to the next stage or stable release.

### Synopsis

```bash
shipyard version promote [OPTIONS]
shipyard version advance [OPTIONS]
```

**Aliases:** `advance`

Advances a pre-release to the next stage in order, or promotes to stable release.

### Description

The `promote` command advances pre-releases through configured stages. Like navigating from calm testing waters to the open ocean, promotion moves your changes closer to a stable release.

**Key behaviors:**

- **Stage advancement**: Moves to the next stage based on `order` values in configuration
- **Counter reset**: Resets counter to 1 when advancing stages
- **State tracking**: Updates `.shipyard/prerelease.yml` with new stage
- **Target version**: Maintains the same target version (unless consignments changed)
- **Explicit promotion**: At the highest stage, returns error— use `shipyard version` to explicitly promote to stable

### Global Options

| Option | Description |
|--------|-------------|
| `--config`, `-c` | Path to configuration file (default: `.shipyard/shipyard.yaml`) |
| `--json`, `-j` | Output in JSON format |
| `--quiet`, `-q` | Suppress all output except errors |
| `--verbose`, `-v` | Enable verbose logging |

### Options

#### `--preview`

Show what promotion would do without making any changes.

**Example:**
```bash
$ shipyard version promote --preview
📦 Preview: Promote to next stage
  - my-api: 1.2.0-alpha.5 → 1.2.0-beta.1 (alpha → beta)
    Target version: 1.2.0

ℹ Preview mode: no changes made
```

#### `--no-commit`

Apply version changes to files but skip creating a git commit.

**Example:**
```bash
$ shipyard version promote --no-commit
📦 Promoting to next stage...
  - my-api: 1.2.0-alpha.5 → 1.2.0-beta.1 (alpha → beta)
✓ Updated version files
⊘ Skipped git commit (--no-commit)
```

#### `--no-tag`

Create the git commit but skip creating git tags.

**Example:**
```bash
$ shipyard version promote --no-tag
📦 Promoting to next stage...
  - my-api: 1.2.0-alpha.5 → 1.2.0-beta.1 (alpha → beta)
✓ Created commit: "chore: pre-release my-api v1.2.0-beta.1"
⊘ Skipped git tags (--no-tag)
```

#### `--package <name>`

Promote only specific packages in a multi-package repository.

**Example:**
```bash
$ shipyard version promote --package cli
📦 Promoting to next stage...
  - cli: 2.0.0-alpha.3 → 2.0.0-beta.1 (alpha → beta)
  - api: skipped (not in --package filter)
✓ Created commit: "chore: pre-release cli v2.0.0-beta.1"
✓ Created tag: cli-v2.0.0-beta.1
```

### Stage Progression

#### Stage Order

Stages are defined with numeric `order` values in `.shipyard/shipyard.yaml`:

```yaml
prerelease:
  stages:
    - name: alpha
      order: 1
      tagTemplate: "v{{.Version}}-alpha.{{.Counter}}"
    - name: beta
      order: 2
      tagTemplate: "v{{.Version}}-beta.{{.Counter}}"
    - name: rc
      order: 3
      tagTemplate: "v{{.Version}}-rc.{{.Counter}}"
```

The `promote` command advances to the next stage by `order`:
- `alpha` (order: 1) → `beta` (order: 2)
- `beta` (order: 2) → `rc` (order: 3)
- `rc` (order: 3) → Error (use `shipyard version` for stable)

#### Counter Reset

When promoting to a new stage, the counter always resets to 1:

```bash
# Currently at alpha.5
$ shipyard version promote
# Output: beta.1 (not beta.5)
```

This makes it clear that you're starting fresh at the new stage.

### Workflow

The `promote` command executes the following steps:

1. **Load Configuration** - Read `.shipyard/shipyard.yaml` for stage definitions
2. **Load State** - Read `.shipyard/prerelease.yml` for current stage
3. **Find Next Stage** - Identify the stage with the next-highest `order` value
4. **Check Highest Stage** - If already at highest stage, return error
5. **Reset Counter** - Set counter to 1 for new stage
6. **Read Consignments** - Load pending consignments (for target version calculation)
7. **Calculate Target Version** - Verify target version hasn't changed
8. **Render Tag** - Apply new stage's tag template
9. **Update Version Files** - Write new pre-release version to ecosystem files
10. **Git Operations** - Create commit and tags (unless `--no-commit`)
11. **Update State** - Save new stage, counter=1, and target to `.shipyard/prerelease.yml`
12. **Skip**: Consignment archival, changelog generation (deferred to stable release)

### Important Behaviors

**Highest stage error**: If you're at the highest stage (by `order`), `promote` returns an error. To promote to stable release, use `shipyard version` instead—this makes the intent explicit.

**Target version changes**: If consignments changed and the target version is different, a warning is displayed but promotion continues.

**Stage order gaps**: Stages don't need consecutive order numbers. `promote` finds the next stage by the next-highest `order` value.

### Examples

#### Promote from Alpha to Beta

```bash
$ shipyard version promote
📦 Promoting to next stage...
  - my-api: 1.2.0-alpha.5 → 1.2.0-beta.1 (alpha → beta)
✓ Updated .shipyard/prerelease.yml
✓ Created commit: "chore: pre-release my-api v1.2.0-beta.1"
✓ Created tag: v1.2.0-beta.1
```

#### Promote from Beta to RC

```bash
$ shipyard version promote
📦 Promoting to next stage...
  - my-api: 1.2.0-beta.3 → 1.2.0-rc.1 (beta → rc)
✓ Updated .shipyard/prerelease.yml
✓ Created commit: "chore: pre-release my-api v1.2.0-rc.1"
✓ Created tag: v1.2.0-rc.1
```

#### Already at Highest Stage

```bash
# Currently at rc.2 (highest stage)
$ shipyard version promote
❌ Error: Already at highest pre-release stage 'rc'
   Use 'shipyard version' to promote to stable release
```

#### Preview Promotion

```bash
$ shipyard version promote --preview
📦 Preview: Promote to next stage
  - my-api: 1.2.0-alpha.5 → 1.2.0-beta.1 (alpha → beta)
    Target version: 1.2.0
    Based on consignments:
      - 20240130-120000-abc123: Add new API endpoint

ℹ Preview mode: no changes made
```

#### Multi-Package Promotion

```bash
$ shipyard version promote
📦 Promoting to next stage...
  - api: 1.2.0-alpha.2 → 1.2.0-beta.1 (alpha → beta)
  - cli: 2.1.0-alpha.5 → 2.1.0-beta.1 (alpha → beta)
✓ Created commits for 2 packages
✓ Created tags: v1.2.0-beta.1, cli-v2.1.0-beta.1
```

#### Target Version Changed During Promotion

```bash
# Currently at alpha.5, then added major change consignment
$ shipyard version promote
⚠ Warning: Target version changed from 1.2.0 to 2.0.0 (consignments modified)
📦 Promoting to next stage...
  - my-api: 1.2.0-alpha.5 → 2.0.0-beta.1 (alpha → beta)
✓ Updated .shipyard/prerelease.yml
✓ Created commit: "chore: pre-release my-api v2.0.0-beta.1"
✓ Created tag: v2.0.0-beta.1
```

#### Promote to Stable Release

To promote from the highest pre-release stage to a stable release, use the `version` command:

```bash
$ shipyard version
📦 Promoting pre-release to stable...
  - my-api: 1.2.0-rc.1 → 1.2.0 (minor)
    - 20240130-120000-abc123: Add new API endpoint
✓ Archived consignments to history
✓ Updated CHANGELOG.md
✓ Deleted .shipyard/prerelease.yml
✓ Created commit: "chore: release my-api v1.2.0"
✓ Created tag: v1.2.0
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error - validation, file, or git operation failed |
| 2 | Already at highest stage (use `shipyard version` for stable) |
| 3 | No pre-release state exists (use `shipyard version prerelease` first) |

### Configuration

#### Stage Definitions

Stages must be defined in `.shipyard/shipyard.yaml` with `order` values:

```yaml
prerelease:
  stages:
    - name: alpha
      order: 1
      tagTemplate: "v{{.Version}}-alpha.{{.Counter}}"
    - name: beta
      order: 2
      tagTemplate: "v{{.Version}}-beta.{{.Counter}}"
    - name: rc
      order: 3
      tagTemplate: "v{{.Version}}-rc.{{.Counter}}"
```

#### Custom Stage Order

Use any numeric order values—they don't need to be consecutive:

```yaml
prerelease:
  stages:
    - name: dev
      order: 10
      tagTemplate: "v{{.Version}}-dev.{{.Counter}}"
    - name: staging
      order: 20
      tagTemplate: "v{{.Version}}-staging.{{.Counter}}"
    - name: prod-candidate
      order: 30
      tagTemplate: "v{{.Version}}-rc.{{.Counter}}"
```

The `promote` command will advance: `dev` → `staging` → `prod-candidate`.

#### State File

The current stage is tracked in `.shipyard/prerelease.yml`:

```yaml
packages:
  my-api:
    stage: beta
    counter: 1
    targetVersion: 1.2.0
```

After promotion, the state file is updated with the new stage and counter reset to 1.

### Behavior Details

#### Stage Advancement Logic

Promotion finds the next stage by:
1. Reading current stage from `.shipyard/prerelease.yml`
2. Finding the stage definition in config by name
3. Selecting the stage with the next-highest `order` value
4. Returning error if no higher-order stage exists

**Example:**

```yaml
# Current state
packages:
  my-api:
    stage: alpha  # order: 1

# After promote
packages:
  my-api:
    stage: beta   # order: 2
    counter: 1    # Reset to 1
```

#### Why Require Explicit `version` for Stable?

At the highest pre-release stage, `promote` returns an error instead of automatically promoting to stable. This is intentional:

- **Explicit intent**: Stable releases are significant—should be explicit
- **Different behavior**: Stable release archives consignments, updates changelog
- **Clear semantics**: `promote` = stage advancement, `version` = stable release

This prevents accidental stable releases and makes the intent clear in scripts and CI/CD.

#### Multi-Package Stages

In multi-package projects, each package can be at a different stage:

```yaml
packages:
  api:
    stage: beta
    counter: 2
    targetVersion: 1.2.0
  cli:
    stage: alpha
    counter: 5
    targetVersion: 2.1.0
```

When promoting:
- `--package api` promotes only `api` from `beta` to `rc`
- No `--package` flag promotes all packages to their next stages

#### Target Version Tracking

The target version is recalculated on each promotion to detect consignment changes:

- **Same target**: Promotion proceeds normally
- **Changed target**: Warning displayed, promotion continues with new target

This helps catch cases where significant changes were added between promotions.

#### Git Requirements

Same requirements as other version commands:

- Repository must be a git repository
- Working directory must be clean (or use `--no-commit`)
- Git `user.name` and `user.email` must be configured

### Integration with CI/CD

#### Branch-Based Stage Promotion

Example GitHub Actions workflow that promotes when merging to specific branches:

```yaml
name: Auto-Promote Pre-Release
on:
  push:
    branches: [beta, rc]

jobs:
  promote:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0

      - name: Install Shipyard
        run: curl -fsSL https://shipyard.build/install.sh | sh

      - name: Promote to next stage
        run: shipyard version promote

      - name: Push changes
        run: |
          git push origin ${{ github.ref_name }}
          git push --tags
```

#### Manual Promotion Workflow

Example workflow dispatch for manual promotions:

```yaml
name: Promote Pre-Release Stage
on:
  workflow_dispatch:
    inputs:
      package:
        description: 'Package to promote (leave empty for all)'
        required: false

jobs:
  promote:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0

      - name: Install Shipyard
        run: curl -fsSL https://shipyard.build/install.sh | sh

      - name: Promote stage
        run: |
          if [ -n "${{ github.event.inputs.package }}" ]; then
            shipyard version promote --package "${{ github.event.inputs.package }}"
          else
            shipyard version promote
          fi

      - name: Push changes
        run: |
          git push
          git push --tags
```

#### Stage Gate with Tests

Example workflow that runs tests before promoting:

```yaml
name: Promote with Tests
on:
  workflow_dispatch:

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run tests
        run: make test

  promote:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0

      - name: Install Shipyard
        run: curl -fsSL https://shipyard.build/install.sh | sh

      - name: Promote to next stage
        run: shipyard version promote

      - name: Push changes
        run: |
          git push
          git push --tags
```

### Related Commands

- [`version prerelease`](./prerelease.md) - Create or increment pre-release at current stage
- [`version snapshot`](./snapshot.md) - Create timestamped snapshot build
- [`version`](./version.md) - Promote to stable release (from highest stage)
- [`consign`](./consign.md) - Record changes that will be promoted
- [`status`](./status.md) - View current pre-release stage

### See Also

- [Configuration Reference](../configuration.md)
- [Pre-Release Reference](./prerelease.md)
- [Tag Templates](../tag-templates.md)
- [CI/CD Integration Guide](../ci-cd.md)

---

## release - Signal arrival at port

### Synopsis

```bash
shipyard release [OPTIONS]
shipyard publish [OPTIONS]
```

**Aliases:** `publish`

### Description

The `release` command publishes a version release to GitHub. It:

1. Reads version history from `.shipyard/history.json`
2. Finds the entry for the specified package/tag
3. Generates release notes from the history entry
4. Creates a GitHub release using the existing git tag

**Prerequisites**: Run `shipyard version` first to create version tags, then push them with `git push --tags`.

**Maritime Metaphor**: Announce your arrival at port—declare the cargo delivered and the voyage complete.

### Global Options

These options are available for all shipyard commands:

| Option | Short | Description |
|--------|-------|-------------|
| `--config <path>` | `-c` | Config file (default: `.shipyard/shipyard.yaml`) |
| `--json` | `-j` | Output in JSON format |
| `--quiet` | `-q` | Suppress non-error output |
| `--verbose` | `-v` | Verbose output |

### Options

#### `--package <name>`, `-p`

Package to release. Required for multi-package repositories.

```bash
shipyard release --package my-api
```

#### `--draft`

Create as a draft release (not published publicly).

```bash
shipyard release --draft
```

#### `--prerelease`

Mark the release as a prerelease.

```bash
shipyard release --prerelease
```

#### `--tag <tag>`

Use a specific tag instead of the latest for the package.

```bash
shipyard release --tag my-api/v1.2.0
```

### Configuration

GitHub settings must be configured in `.shipyard/shipyard.yaml`:

```yaml
github:
  owner: myorg
  repo: myrepo
```

The `GITHUB_TOKEN` environment variable must be set with appropriate permissions.

### Examples

#### Basic Usage

```bash
shipyard release --package my-api
```

```
✓ Release published successfully
  Package: my-api
  Version: 1.3.0
  Tag: my-api/v1.3.0
  URL: https://github.com/myorg/myrepo/releases/tag/my-api/v1.3.0
```

#### Draft Release

```bash
shipyard release --package my-api --draft
```

#### Specific Version

```bash
shipyard release --package my-api --tag my-api/v1.2.0
```

#### Single-Package Repository

For repos with one package, `--package` is auto-detected:

```bash
shipyard release
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success - release published |
| 1 | Error - missing config, missing token, tag not found, or API failure |

### Behavior Details

#### Package Requirement

For multi-package repositories, `--package` is required. Single-package repos auto-detect.

#### Tag Selection

Without `--tag`, uses the most recent release for the package from history.

#### Release Notes

Generated automatically from the history entry using the `releaseNotes` template.

#### Release Title

Extracted from the first line of the release notes. Falls back to `{package} v{version}` if:
- Release notes are empty
- First line is a markdown heading (`#`)

#### GitHub Token

Requires `GITHUB_TOKEN` environment variable with `repo` scope permissions.

#### Tag Must Exist

The git tag must already exist locally and be pushed to the remote. Run:

```bash
shipyard version
git push --tags
shipyard release --package my-api
```

### Related Commands

- [`version`](./version.md) - Create version tags
- [`release-notes`](./release-notes.md) - Generate release notes manually

### See Also

- [Configuration Reference](../configuration.md) - GitHub settings

---

## release-notes - Tell the tale of your voyage

### Synopsis

```bash
shipyard release-notes [OPTIONS]
shipyard notes [OPTIONS]
shipyard changelog [OPTIONS]
```

**Aliases:** `notes`, `changelog`

### Description

The `release-notes` command generates release notes from version history. It:

1. Reads entries from `.shipyard/history.json`
2. Filters by package, version, or metadata
3. Renders output using a template
4. Writes to stdout or a file

**Maritime Metaphor**: Recount the journey from the captain's log—tales of ports visited and cargo delivered.

### Global Options

These options are available for all shipyard commands:

| Option | Short | Description |
|--------|-------|-------------|
| `--config <path>` | `-c` | Config file (default: `.shipyard/shipyard.yaml`) |
| `--json` | `-j` | Output in JSON format |
| `--quiet` | `-q` | Suppress non-error output |
| `--verbose` | `-v` | Verbose output |

### Options

#### `--package <name>`, `-p`

Filter release notes by package name. Required for multi-package repositories.

```bash
shipyard release-notes --package my-api
```

#### `--output <path>`, `-o`

Write output to a file instead of stdout.

```bash
shipyard release-notes --output RELEASE_NOTES.md
```

#### `--version <version>`

Generate notes for a specific version only.

```bash
shipyard release-notes --version 1.2.0
```

#### `--all-versions`

Show complete history instead of just the latest version. Automatically uses the changelog template.

```bash
shipyard release-notes --all-versions
```

#### `--filter <key=value>`

Filter by custom metadata. Can be repeated for multiple filters.

```bash
shipyard release-notes --filter team=platform
shipyard release-notes --filter team=platform --filter scope=api
```

#### `--template <name>`

Specify which template to use. Can be a builtin name or path to a `.tmpl` file.

```bash
shipyard release-notes --template builtin:grouped
shipyard release-notes --template .shipyard/templates/custom-notes.tmpl
```

### Examples

#### Latest Version (Default)

```bash
shipyard release-notes --package my-api
```

```
# my-api v1.3.0

Released: 2024-01-30

## Changes

- **feat**: Add new API endpoint
- **fix**: Fix memory leak in handler
```

#### Specific Version

```bash
shipyard release-notes --package my-api --version 1.2.0
```

#### Complete History

```bash
shipyard release-notes --package my-api --all-versions
```

#### Write to File

```bash
shipyard release-notes --package my-api --output docs/RELEASE_NOTES.md
```

#### Filter by Metadata

```bash
shipyard release-notes --package my-api --filter team=backend --all-versions
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error - missing required flags, invalid filters, or file operation failed |

### Behavior Details

#### Package Requirement

For multi-package repositories, `--package` is required. For single-package repos, the package is auto-detected.

#### Default Behavior

Without `--version` or `--all-versions`, only the latest version is shown.

#### Template Selection

- Default (single version): `release-notes` template
- With `--all-versions`: `changelog` template
- With `--template`: Uses specified template

#### Metadata Validation

Filter keys and values are validated against metadata fields defined in `shipyard.yaml`. Invalid keys or values return an error.

#### Empty History

If no releases are found in history:

```
No releases found in history
```

Exit code: 0 (success)

### Related Commands

- [`version`](./version.md) - Create new versions (populates history)
- [`release`](./release.md) - Publish releases to GitHub

### See Also

- [Tag Generation Guide](../tag-generation.md) - Template customization
- [Configuration Reference](../configuration.md) - Metadata field definitions

---

## remove - Jettison cargo from the manifest

### Synopsis

```bash
shipyard remove [OPTIONS]
shipyard rm [OPTIONS]
shipyard delete [OPTIONS]
```

**Aliases:** `rm`, `delete`

### Description

The `remove` command removes one or more pending consignments from the manifest. It:

1. Validates that `--id` or `--all` is specified
2. Locates consignment files in `.shipyard/consignments/`
3. Deletes the specified files

Use `--id` to remove specific consignments by ID, or `--all` to clear all pending consignments.

**Maritime Metaphor**: Unload cargo from the manifest before setting sail—discard changes that are no longer needed.

### Global Options

These options are available for all shipyard commands:

| Option | Short | Description |
|--------|-------|-------------|
| `--config <path>` | `-c` | Config file (default: `.shipyard/shipyard.yaml`) |
| `--json` | `-j` | Output in JSON format |
| `--quiet` | `-q` | Suppress non-error output |
| `--verbose` | `-v` | Verbose output |

### Options

#### `--id <consignment-id>`

Consignment ID(s) to remove. Can be repeated.

```bash
shipyard remove --id 20240101-120000-abc123
shipyard remove --id c1 --id c2
```

#### `--all`

Remove all pending consignments.

```bash
shipyard remove --all
```

### Examples

#### Remove Specific Consignment

```bash
shipyard remove --id 20240130-120000-abc123
```

```
✓ Removed 1 consignment(s)
  - 20240130-120000-abc123
```

#### Remove Multiple by ID

```bash
shipyard remove --id 20240130-120000-abc123 --id 20240131-090000-def456
```

#### Remove All Pending

```bash
shipyard remove --all
```

```
✓ Removed 3 consignment(s)
  - 20240130-120000-abc123
  - 20240131-090000-def456
  - 20240201-140000-ghi789
```

#### JSON Output

```bash
shipyard remove --all --json
```

```json
{
  "removed": ["20240130-120000-abc123", "20240131-090000-def456"],
  "count": 2
}
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success - consignment(s) removed (or none to remove with `--all`) |
| 1 | Error - missing flags, consignment not found, or failed to load config |

### Behavior Details

#### Flag Requirement

Either `--id` or `--all` must be specified. Running `remove` without flags returns an error.

#### Not Found

If a consignment ID specified with `--id` does not exist, the command returns an error.

#### Empty Directory

Running `remove --all` with no pending consignments exits successfully with a message:

```
No pending consignments to remove
```

#### JSON Output

With `--json`, outputs a JSON object with `removed` (array of IDs) and `count` (number removed).

### Related Commands

- [`add`](./add.md) - Create new consignments
- [`status`](./status.md) - View pending consignments
- [`version`](./version.md) - Process consignments into versions

### See Also

- [Consignment Format](../consignment-format.md) - Structure of consignment files

---

## snapshot - Create a timestamped snapshot pre-release version

### Synopsis

```bash
shipyard version snapshot [OPTIONS]
shipyard version snap [OPTIONS]
```

**Aliases:** `snap`

Creates a snapshot pre-release version with a timestamp identifier, independent of the stage-based pre-release system.

### Description

The `snapshot` command creates timestamped pre-release versions for ad-hoc testing and builds. Unlike stage-based pre-releases (`alpha`, `beta`, etc.), snapshots:

- **Use timestamps**: Version identifier includes `YYYYMMDD-HHMMSS` format
- **Don't affect state**: Don't modify `.shipyard/prerelease.yml` or stage tracking
- **Independent**: Can be created alongside stage-based pre-releases or stable releases
- **Template-driven**: Use `snapshotTagTemplate` from configuration

Snapshots are ideal for:
- Pull request builds
- CI/CD pipeline artifacts
- Ad-hoc testing builds
- Development snapshots

Like taking a quick navigational reading, snapshots capture the current state without committing to a formal pre-release stage.

### Global Options

| Option | Description |
|--------|-------------|
| `--config`, `-c` | Path to configuration file (default: `.shipyard/shipyard.yaml`) |
| `--json`, `-j` | Output in JSON format |
| `--quiet`, `-q` | Suppress all output except errors |
| `--verbose`, `-v` | Enable verbose logging |

### Options

#### `--preview`

Show what snapshot version would be created without making any changes.

**Example:**
```bash
$ shipyard version snapshot --preview
📦 Preview: Snapshot version
  - my-api: 1.2.0 → 1.2.0-snapshot.20260204-153045
    Target version: 1.2.0
    Based on consignments:
      - 20240130-120000-abc123: Add new API endpoint

ℹ Preview mode: no changes made
```

#### `--no-commit`

Apply version changes to files but skip creating a git commit.

**Example:**
```bash
$ shipyard version snapshot --no-commit
📦 Creating snapshot version...
  - my-api: 1.2.0 → 1.2.0-snapshot.20260204-153045
✓ Updated version files
⊘ Skipped git commit (--no-commit)
```

#### `--no-tag`

Create the git commit but skip creating git tags.

**Example:**
```bash
$ shipyard version snapshot --no-tag
📦 Creating snapshot version...
  - my-api: 1.2.0 → 1.2.0-snapshot.20260204-153045
✓ Created commit: "chore: snapshot my-api v1.2.0-snapshot.20260204-153045"
⊘ Skipped git tags (--no-tag)
```

#### `--package <name>`

Create snapshot only for specific packages in a multi-package repository.

**Example:**
```bash
$ shipyard version snapshot --package cli
📦 Creating snapshot version...
  - cli: 2.0.0 → 2.0.0-snapshot.20260204-153045
  - api: skipped (not in --package filter)
✓ Created commit: "chore: snapshot cli v2.0.0-snapshot.20260204-153045"
✓ Created tag: cli-v2.0.0-snapshot.20260204-153045
```

### Workflow

The `snapshot` command executes the following steps:

1. **Load Configuration** - Read `.shipyard/shipyard.yaml` for snapshot template
2. **Read Consignments** - Load pending consignments from `.shipyard/consignments/`
3. **Calculate Target Version** - Calculate next stable version from consignments
4. **Generate Timestamp** - Create timestamp in `YYYYMMDD-HHMMSS` format
5. **Render Tag** - Apply snapshot tag template with version and timestamp
6. **Update Version Files** - Write snapshot version to ecosystem files
7. **Git Operations** - Create commit and tags (unless `--no-commit`)
8. **Skip**: State file updates, consignment archival, changelog generation

### Important Behaviors

**No state changes**: The `.shipyard/prerelease.yml` file is not created or modified. Snapshots are independent of stage-based pre-releases.

**Target version**: Based on pending consignments, just like the `version` and `prerelease` commands.

**Timestamp format**: Always uses UTC time in `YYYYMMDD-HHMMSS` format for consistency.

### Configuration

#### Snapshot Template

Configure the snapshot tag template in `.shipyard/shipyard.yaml`:

```yaml
prerelease:
  snapshotTagTemplate: "v{{.Version}}-snapshot.{{.Timestamp}}"
```

#### Template Variables

Snapshot templates support:

- `{{.Version}}`: Target stable version (e.g., "1.2.0")
- `{{.Timestamp}}`: UTC timestamp in format `YYYYMMDD-HHMMSS`
- `{{.Package}}`: Package name

**Examples:**

```yaml
# Standard snapshot format
snapshotTagTemplate: "v{{.Version}}-snapshot.{{.Timestamp}}"
# Output: v1.2.0-snapshot.20260204-153045

# With package prefix
snapshotTagTemplate: "{{.Package}}-v{{.Version}}-snapshot.{{.Timestamp}}"
# Output: api-v1.2.0-snapshot.20260204-153045

# Short timestamp format
snapshotTagTemplate: "v{{.Version}}-{{.Timestamp}}"
# Output: v1.2.0-20260204-153045

# Build-style identifier
snapshotTagTemplate: "v{{.Version}}+build.{{.Timestamp}}"
# Output: v1.2.0+build.20260204-153045
```

### Examples

#### Create Snapshot

```bash
$ shipyard version snapshot
📦 Creating snapshot version...
  - my-api: 1.2.0 → 1.2.0-snapshot.20260204-153045
✓ Created commit: "chore: snapshot my-api v1.2.0-snapshot.20260204-153045"
✓ Created tag: v1.2.0-snapshot.20260204-153045
```

#### Preview Snapshot

```bash
$ shipyard version snapshot --preview
📦 Preview: Snapshot version
  - my-api: 1.2.0 → 1.2.0-snapshot.20260204-153045
    Target version: 1.2.0
    Based on consignments:
      - 20240130-120000-abc123: Add new API endpoint

ℹ Preview mode: no changes made
```

#### Snapshot Without Git Operations

Useful for CI builds where you want the version file updated but handle git operations separately:

```bash
$ shipyard version snapshot --no-commit --no-tag
📦 Creating snapshot version...
  - my-api: 1.2.0 → 1.2.0-snapshot.20260204-153045
✓ Updated version files
⊘ Skipped git commit (--no-commit)
⊘ Skipped git tags (--no-tag)
```

#### Multi-Package Snapshot

```bash
$ shipyard version snapshot
📦 Creating snapshot version...
  - api: 1.2.0 → 1.2.0-snapshot.20260204-153045
  - cli: 2.1.0 → 2.1.0-snapshot.20260204-153045
✓ Created commits for 2 packages
✓ Created tags: v1.2.0-snapshot.20260204-153045, cli-v2.1.0-snapshot.20260204-153045
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error - validation, file, or git operation failed |
| 2 | No consignments to base snapshot on |

### Behavior Details

#### Snapshot Independence

Snapshots are completely independent of stage-based pre-releases:

- Creating a snapshot doesn't affect `.shipyard/prerelease.yml`
- Can create snapshots while in a pre-release stage
- Snapshots don't increment or reset pre-release counters

**Example workflow:**

```bash
# Start alpha pre-release
$ shipyard version prerelease
# Output: 1.2.0-alpha.1

# Create PR snapshot
$ shipyard version snapshot
# Output: 1.2.0-snapshot.20260204-120000

# Continue alpha pre-releases
$ shipyard version prerelease
# Output: 1.2.0-alpha.2 (counter incremented correctly)
```

#### Snapshot Ordering

Snapshots use timestamps, which naturally order chronologically:

```
v1.2.0-snapshot.20260204-120000
v1.2.0-snapshot.20260204-121500
v1.2.0-snapshot.20260204-130245
```

This makes it easy to identify and sort snapshot builds in CI/CD systems.

#### Consignments

Like stage-based pre-releases, snapshots:
- Calculate target version from pending consignments
- Don't archive consignments
- Don't update changelogs
- Preserve consignments for eventual stable release

#### Git Requirements

Same requirements as other version commands:

- Repository must be a git repository
- Working directory must be clean (or use `--no-commit`)
- Git `user.name` and `user.email` must be configured

### Integration with CI/CD

#### Pull Request Snapshots

Example GitHub Actions workflow for PR snapshots:

```yaml
name: PR Snapshot Build
on:
  pull_request:

jobs:
  snapshot:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0

      - name: Install Shipyard
        run: curl -fsSL https://shipyard.build/install.sh | sh

      - name: Create snapshot (no git operations)
        run: shipyard version snapshot --no-commit --no-tag

      - name: Build artifacts
        run: make build

      - name: Upload artifacts
        uses: actions/upload-artifact@v3
        with:
          name: snapshot-${{ github.event.pull_request.number }}
          path: dist/

      - name: Comment PR with snapshot info
        uses: actions/github-script@v6
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: '✅ Snapshot build created. Download artifacts to test.'
            })
```

#### Scheduled Nightly Snapshots

```yaml
name: Nightly Snapshot
on:
  schedule:
    - cron: '0 2 * * *'  # 2 AM UTC daily

jobs:
  snapshot:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0

      - name: Install Shipyard
        run: curl -fsSL https://shipyard.build/install.sh | sh

      - name: Create snapshot
        run: shipyard version snapshot

      - name: Push snapshot tag
        run: git push --tags

      - name: Build and publish
        run: |
          make build
          make publish-snapshot
```

#### Docker Image Tagging

Use snapshot versions for Docker image tags:

```yaml
- name: Create snapshot
  run: shipyard version snapshot --no-commit --no-tag

- name: Extract version
  id: version
  run: echo "version=$(cat VERSION)" >> $GITHUB_OUTPUT

- name: Build Docker image
  run: |
    docker build -t myapp:${{ steps.version.outputs.version }} .
    docker push myapp:${{ steps.version.outputs.version }}
```

### Related Commands

- [`version prerelease`](./prerelease.md) - Create stage-based pre-release (alpha, beta, rc)
- [`version promote`](./promote.md) - Promote to next stage or stable
- [`version`](./version.md) - Create stable release
- [`consign`](./consign.md) - Record changes that will be in snapshot
- [`release`](./release.md) - Publish release (snapshots can use `--prerelease` flag)

### See Also

- [Configuration Reference](../configuration.md)
- [Pre-Release Reference](./prerelease.md)
- [Tag Templates](../tag-templates.md)
- [CI/CD Integration Guide](../ci-cd.md)

---

## status - Check cargo and chart your course

### Synopsis

```bash
shipyard status [OPTIONS]
shipyard ls [OPTIONS]
shipyard list [OPTIONS]
```

**Aliases:** `ls`, `list`

### Description

The `status` command shows pending consignments and their calculated version bumps. It:

1. Reads pending consignments from `.shipyard/consignments/`
2. Groups them by package
3. Calculates version bumps (including dependency propagation)
4. Displays results in table or JSON format

**Maritime Metaphor**: Review pending cargo and see which ports of call (versions) await.

### Global Options

These options are available for all shipyard commands:

| Option | Short | Description |
|--------|-------|-------------|
| `--config <path>` | `-c` | Config file (default: `.shipyard/shipyard.yaml`) |
| `--json` | `-j` | Output in JSON format |
| `--quiet` | `-q` | Suppress non-error output |
| `--verbose` | `-v` | Verbose output |

### Options

#### `--package <name>`, `-p`

Filter by package name(s). Can be repeated.

```bash
shipyard status --package core
shipyard status --package core --package api
```

#### `--quiet`, `-q`

Minimal output.

```bash
shipyard status --quiet
```

#### `--verbose`, `-v`

Verbose output with timestamps and metadata.

```bash
shipyard status --verbose
```

### Examples

#### Basic Usage

```bash
shipyard status
```

```
📦 Pending consignments

core (1.2.3 → 1.3.0)
  - 20240130-120000-abc123: Add new feature (minor)
  - 20240130-110000-def456: Fix null pointer (patch)

api (2.0.0 → 2.0.1)
  - 20240130-120000-abc123: Add new feature (minor)
```

#### Filter by Package

```bash
shipyard status --package core
```

#### JSON Output

```bash
shipyard status --json
```

#### Verbose Mode

```bash
shipyard status --verbose
```

Shows additional details like timestamps, metadata, and propagation info.

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success (even with no consignments) |
| 1 | Error - not initialized or failed to read consignments |

### Behavior Details

#### No Consignments

```
No pending consignments
```

Exit code: 0 (success)

#### Version Calculation

Shows what version each package would become if `shipyard version` were run now. Includes:
- Direct bumps from consignments affecting the package
- Propagated bumps from dependencies

#### Package Filtering

With `--package`, only shows consignments affecting those packages.

### Related Commands

- [`add`](./add.md) - Create new consignments
- [`version`](./version.md) - Process consignments into versions

### See Also

- [Consignment Format](../consignment-format.md) - Structure of consignment files

---

## upgrade - Refit the shipyard with latest provisions

### Synopsis

```bash
shipyard upgrade [OPTIONS]
shipyard update [OPTIONS]
shipyard self-update [OPTIONS]
```

**Aliases:** `update`, `self-update`

### Description

The `upgrade` command upgrades shipyard to the latest (or specified) version. It:

1. Detects how shipyard was installed
2. Fetches the latest release from GitHub
3. Compares versions
4. Upgrades using the appropriate method

Supports Homebrew, npm, Go install, and script installations. Docker installations must be upgraded manually.

### Global Options

These options are available for all shipyard commands:

| Option | Short | Description |
|--------|-------|-------------|
| `--config <path>` | `-c` | Config file (default: `.shipyard/shipyard.yaml`) |
| `--json` | `-j` | Output in JSON format |
| `--quiet` | `-q` | Suppress non-error output |
| `--verbose` | `-v` | Verbose output |

### Options

#### `--yes`, `-y`

Skip confirmation prompt.

```bash
shipyard upgrade --yes
```

#### `--version <version>`

Upgrade to a specific version instead of latest.

```bash
shipyard upgrade --version v1.2.0
```

#### `--force`

Force upgrade even if already on the latest version.

```bash
shipyard upgrade --force
```

#### `--dry-run`

Show upgrade plan without executing.

```bash
shipyard upgrade --dry-run
```

### Examples

#### Basic Usage

```bash
shipyard upgrade
```

```
Checking installation... ✓
Checking for updates... ✓

Current version: 1.2.0
Latest version:  1.3.0

Upgrade shipyard from 1.2.0 to 1.3.0? [Y/n]
```

#### Skip Confirmation

```bash
shipyard upgrade --yes
```

#### Preview Upgrade

```bash
shipyard upgrade --dry-run
```

#### Force Reinstall

```bash
shipyard upgrade --force
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success - upgraded or already on latest |
| 1 | Error - detection failed, network error, or upgrade failed |

### Behavior Details

#### Installation Detection

Automatically detects:
- **Homebrew** - Uses `brew upgrade`
- **npm** - Uses `npm update -g`
- **Go install** - Uses `go install`
- **Script install** - Downloads binary from GitHub releases

#### Docker

Docker installations cannot be auto-upgraded:

```
Cannot upgrade: Docker installations must be upgraded manually

To upgrade Docker installations:
  docker pull natonathan/shipyard:latest
```

#### Already on Latest

```
✓ Already on latest version (1.3.0)
```

Use `--force` to reinstall anyway.

#### Network Requirements

Requires internet access to fetch release information from GitHub.

### Related Commands

- [`completion`](./completion.md) - Shell completions

### See Also

- [GitHub Releases](https://github.com/NatoNathan/shipyard/releases) - All versions

---

## validate - Inspect the hull before departure

### Synopsis

```bash
shipyard validate [OPTIONS]
shipyard check [OPTIONS]
shipyard lint [OPTIONS]
```

**Aliases:** `check`, `lint`

### Description

The `validate` command checks the health of your shipyard setup. It:

1. Loads and validates the configuration file
2. Validates dependency references between packages
3. Parses all pending consignment files for errors
4. Builds the dependency graph and checks for cycles

Reports errors and warnings found during validation.

**Maritime Metaphor**: Inspect the hull and rigging before departure—ensure everything is seaworthy.

### Global Options

These options are available for all shipyard commands:

| Option | Short | Description |
|--------|-------|-------------|
| `--config <path>` | `-c` | Config file (default: `.shipyard/shipyard.yaml`) |
| `--json` | `-j` | Output in JSON format |
| `--quiet` | `-q` | Suppress non-error output |
| `--verbose` | `-v` | Verbose output |

### Examples

#### Basic Usage

```bash
shipyard validate
```

```
✓ Validation passed
```

#### JSON Output

```bash
shipyard validate --json
```

```json
{
  "valid": true,
  "errors": [],
  "warnings": []
}
```

#### Quiet Mode

```bash
shipyard validate --quiet
```

Exits silently with code 0 on success, or code 1 with "validation failed" on failure.

#### With Validation Errors

```bash
shipyard validate
```

```
Errors:
  - config validation: package "core" references unknown dependency "missing-lib"
  - consignment 20240130-120000-abc123.md: unknown package "nonexistent"

Validation failed
```

#### With Warnings

```bash
shipyard validate
```

```
Warnings:
  - dependency cycle detected: core -> api -> core

✓ Validation passed
```

Cycles are reported as warnings, not errors.

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Validation passed (warnings may be present) |
| 1 | Validation failed - errors found in config, consignments, or dependencies |

### Behavior Details

#### What Is Validated

| Check | Category | Severity |
|-------|----------|----------|
| Config file loads successfully | Config | Error |
| Config passes schema validation | Config | Error |
| Package dependency references exist | Dependencies | Error |
| Consignment files parse correctly | Consignments | Error |
| Dependency graph has no cycles | Graph | Warning |

#### Quiet Mode

With `--quiet`, produces no output on success. On failure, outputs "validation failed" and exits with code 1.

#### JSON Output

With `--json`, outputs a JSON object:

```json
{
  "valid": false,
  "errors": ["config validation: ..."],
  "warnings": ["dependency cycle detected: ..."]
}
```

#### Warnings vs Errors

- **Errors** cause validation to fail (exit code 1)
- **Warnings** are informational only (validation still passes)

Currently, dependency cycles are the only condition that produces warnings.

### Related Commands

- [`config show`](./config-show.md) - Display resolved configuration
- [`status`](./status.md) - View pending consignments and version bumps

### See Also

- [Configuration Reference](../configuration.md) - Configuration file format
- [Consignment Format](../consignment-format.md) - Structure of consignment files

---

## version - Set sail to the next port

### Synopsis

```bash
shipyard version [OPTIONS]
shipyard bump [OPTIONS]
shipyard sail [OPTIONS]
```

**Aliases:** `bump`, `sail`

### Description

The `version` command processes pending consignments and creates new package versions. It:

1. Calculates new version numbers based on change types
2. Updates ecosystem-specific version files
3. Archives consignments to history
4. Generates changelogs from complete history
5. Creates a git commit and tags

**Maritime Metaphor**: The ship leaves port with its cargo (consignments), and each package reaches its next destination (version).

### Global Options

These options are available for all shipyard commands:

| Option | Short | Description |
|--------|-------|-------------|
| `--config <path>` | `-c` | Config file (default: `.shipyard/shipyard.yaml`) |
| `--json` | `-j` | Output in JSON format |
| `--quiet` | `-q` | Suppress non-error output |
| `--verbose` | `-v` | Verbose output |

### Options

#### `--preview`

Show what changes would be made without applying them.

```bash
shipyard version --preview
```

#### `--no-commit`

Apply version changes but skip creating a git commit. Tags are also skipped.

```bash
shipyard version --no-commit
```

#### `--no-tag`

Create the commit but skip creating git tags.

```bash
shipyard version --no-tag
```

#### `--package <name>`

Process consignments only for specified package(s). Can be repeated.

```bash
shipyard version --package my-api
shipyard version --package cli --package sdk
```

### Workflow

The command executes these phases:

1. **Validation** - Read and validate pending consignments
2. **Dependency Graph** - Build package dependency map
3. **Version Calculation** - Determine new versions based on change types
4. **Preview** (if `--preview`) - Display changes and exit
5. **Update Version Files** - Write new versions to ecosystem files
6. **Generate Tags** - Render tag names and messages from templates
7. **Archive Consignments** - Append to `history.json` with version context
8. **Generate Changelogs** - Regenerate from complete history (including new version)
9. **Delete Consignments** - Remove processed `.md` files
10. **Git Operations** - Create commit and tags (unless `--no-commit`)

**Note**: Changelogs are generated *after* archiving so the new version appears in the output.

### Configuration

Configuration is read from `.shipyard/shipyard.yaml`:

```yaml
packages:
  - name: my-api
    path: packages/api
    ecosystem: npm
    templates:                           # optional, overrides global
      changelog:
        source: builtin:grouped
      tagName:
        source: builtin:npm
    dependencies:                        # optional
      - name: shared-types
        type: linked  # or: fixed

templates:
  changelog:
    source: builtin:default              # or path to .tmpl file
  tagName:
    source: builtin:go                   # or path to .tmpl file
  releaseNotes:
    source: builtin:default              # or inline (see below)
    inline: |                            # alternative to source
      # {{.Package}} {{.Version}}
      {{range .Consignments}}
      - {{.Summary}}
      {{end}}

consignments:
  path: .shipyard/consignments

history:
  path: .shipyard/history.json
```

### Supported Ecosystems

- **go** - `VERSION` file (or tag-only)
- **npm** - `package.json`
- **python** - `__version__` in `__init__.py` or `setup.py`
- **helm** - `Chart.yaml`
- **cargo** - `Cargo.toml`
- **deno** - `deno.json`

### Template Options

Templates are configured globally under `templates:` with `source:` (file path or builtin) or `inline:` (embedded template).

**Builtins**:
- `builtin:default` - Available for changelog, tagName, and releaseNotes
- `builtin:go` - Go module style tags (v-prefixed)
- `builtin:npm` - NPM style tags
- `builtin:grouped` - Changelog grouped by change type

See [Tag Generation Guide](../tag-generation.md) for template details.

### Examples

#### Basic Usage

```bash
shipyard version
```

```
📦 Versioning packages...
  - my-api: 1.2.3 → 1.3.0 (minor)
  - shared-types: 0.5.0 → 0.5.1 (patch)
✓ Created commit: "chore: release my-api v1.3.0, shared-types v0.5.1"
✓ Created tags: my-api/v1.3.0, shared-types/v0.5.1
```

#### Preview Changes

```bash
shipyard version --preview
```

```
📦 Preview: Version changes
  - my-api: 1.2.3 → 1.3.0 (minor)
    - 20240130-120000-abc123: Add new API endpoint

ℹ Preview mode: no changes made
```

#### Manual Review Before Commit

```bash
shipyard version --no-commit
git diff
git add -A && git commit -m "chore: release my-api v1.3.0"
git tag -a my-api/v1.3.0 -m "Release my-api v1.3.0"
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success (or no consignments to process) |
| 1 | Error - validation, file, or git operation failed |

### Behavior Details

#### No Consignments

Exits successfully with no-op message.

#### Package Filtering

With `--package`, unmatched consignments remain in `.shipyard/consignments/`.

#### Version Propagation

When a dependency is versioned, dependents are also bumped:
- **linked**: Same change type as the dependency
- **fixed**: Patch bump

#### Tag Format

Tags follow git commit message format:
- Single line → lightweight tag
- Multiple lines (blank line separator) → annotated tag with message body

#### Git Requirements

- Repository must be initialized
- Working directory must be clean
- `user.name` and `user.email` must be configured

### Related Commands

- [`consign`](./consign.md) - Record a new change
- [`releasenotes`](./releasenotes.md) - Generate release notes from history
- [`changelog`](./changelog.md) - Generate changelog from history

### See Also

- [Tag Generation Guide](../tag-generation.md)
- [Configuration Reference](../configuration.md)
- [Consignment Format](../consignment-format.md)
