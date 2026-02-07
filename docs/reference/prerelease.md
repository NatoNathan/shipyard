# prerelease

Create or increment a pre-release version at the current stage.

## Synopsis

```bash
shipyard version prerelease [OPTIONS]
shipyard version pre [OPTIONS]
shipyard version rc [OPTIONS]
```

**Aliases:** `pre`, `rc`

Creates or increments a pre-release version at the current stage based on pending consignments.

## Description

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

## Global Options

| Option | Description |
|--------|-------------|
| `--config`, `-c` | Path to configuration file (default: `.shipyard/shipyard.yaml`) |
| `--json`, `-j` | Output in JSON format |
| `--quiet`, `-q` | Suppress all output except errors |
| `--verbose`, `-v` | Enable verbose logging |

## Options

### `--preview`

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

### `--no-commit`

Apply version changes to files but skip creating a git commit. Useful for reviewing changes before committing.

**Example:**
```bash
$ shipyard version prerelease --no-commit
📦 Creating pre-release versions...
  - my-api: 1.2.0-alpha.1 → 1.2.0-alpha.2 (alpha)
✓ Updated version files
⊘ Skipped git commit (--no-commit)
```

### `--no-tag`

Create the git commit but skip creating git tags. Useful for testing version file updates.

**Example:**
```bash
$ shipyard version prerelease --no-tag
📦 Creating pre-release versions...
  - my-api: 1.2.0-alpha.1 → 1.2.0-alpha.2 (alpha)
✓ Created commit: "chore: pre-release my-api v1.2.0-alpha.2"
⊘ Skipped git tags (--no-tag)
```

### `--package <name>`

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

## Stage System

### Stage Determination

The current stage is read from `.shipyard/prerelease.yml`:

- **First pre-release**: Starts at the stage with the lowest `order` value
- **Subsequent pre-releases**: Increments the counter at the current stage
- **Stage advancement**: Use `shipyard version promote` to move to the next stage

### Stage Configuration

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

### Custom Stages

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

## Workflow

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

## Configuration

### Main Config (`.shipyard/shipyard.yaml`)

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

### Template Variables

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

### State File (`.shipyard/prerelease.yml`)

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

### Why Separate Files?

- **Main config** (`.shipyard/shipyard.yaml`): Static stage definitions, committed to git
- **State file** (`.shipyard/prerelease.yml`): Dynamic tracking data, can be gitignored or committed
- Keeps main configuration clean and focused on project structure

## Examples

### First Pre-Release (No State File Exists)

```bash
$ shipyard version prerelease
📦 Creating pre-release versions...
  - my-api: 1.1.5 → 1.2.0-alpha.1 (alpha, first pre-release)
✓ Created .shipyard/prerelease.yml
✓ Created commit: "chore: pre-release my-api v1.2.0-alpha.1"
✓ Created tag: v1.2.0-alpha.1
```

### Increment Pre-Release (Same Stage)

```bash
$ shipyard version prerelease
📦 Creating pre-release versions...
  - my-api: 1.2.0-alpha.1 → 1.2.0-alpha.2 (alpha)
✓ Created commit: "chore: pre-release my-api v1.2.0-alpha.2"
✓ Created tag: v1.2.0-alpha.2
```

### Promote to Next Stage

```bash
$ shipyard version promote
📦 Promoting to next stage...
  - my-api: 1.2.0-alpha.2 → 1.2.0-beta.1 (alpha → beta)
✓ Updated .shipyard/prerelease.yml
✓ Created commit: "chore: pre-release my-api v1.2.0-beta.1"
✓ Created tag: v1.2.0-beta.1
```

### Preview Changes

```bash
$ shipyard version prerelease --preview
📦 Preview: Pre-release version changes
  - my-api: 1.2.0-beta.1 → 1.2.0-beta.2 (beta)
    Target version: 1.2.0
    Based on consignments:
      - 20240130-120000-abc123: Add new API endpoint

ℹ Preview mode: no changes made
```

### Specific Package Only

```bash
$ shipyard version prerelease --package cli
📦 Creating pre-release versions...
  - cli: 2.0.0 → 2.1.0-alpha.1 (alpha)
  - api: skipped (not in --package filter)
✓ Created commit: "chore: pre-release cli v2.1.0-alpha.1"
✓ Created tag: cli-v2.1.0-alpha.1
```

### Target Version Changed (Consignments Modified)

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

### Promote to Stable Release

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

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error - validation, file, or git operation failed |
| 2 | No consignments to base pre-release on |

## Behavior Details

### Consignment Changes During Pre-Release

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

### Stage Progression

Typical workflow through stages:

1. **Development** → `prerelease` repeatedly (alpha.1, alpha.2, ...)
2. **Promotion** → `promote` (alpha → beta)
3. **Testing** → `prerelease` repeatedly (beta.1, beta.2, ...)
4. **Stabilization** → `promote` (beta → rc)
5. **Final testing** → `prerelease` repeatedly (rc.1, rc.2, ...)
6. **Release** → `version` (promotes to stable 1.2.0, deletes state file)

Stages follow the order defined in configuration. Use `promote` to advance between stages.

### Snapshot Behavior

For timestamp-based builds independent of the stage system, use [`shipyard version snapshot`](./snapshot.md):

- Snapshots use timestamp: `YYYYMMDD-HHMMSS`
- Don't affect pre-release stage/counter tracking
- Useful for PR builds, CI pipelines, ad-hoc testing
- Independent of stage-based workflow

### Multi-Package Projects

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

### State File Management

The `.shipyard/prerelease.yml` state file:

- Created automatically on first pre-release
- Deleted automatically when promoting to stable release with `shipyard version`
- **Should be committed to git** (not `.gitignore`d) - this allows your team to track pre-release state across branches
- `shipyard version` will remove the file entirely for single-package repos, or remove the entry for packages promoted to stable in multi-package repos
- If deleted manually, next prerelease starts at lowest order stage

### Git Requirements

Same requirements as the `version` command:

- Repository must be a git repository
- Working directory must be clean (or use `--no-commit`)
- Git `user.name` and `user.email` must be configured

## Integration with CI/CD

### Automatic Pre-Release on Push

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

### Stage Promotion on Branch

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

## Related Commands

- [`version`](./version.md) - Create stable release, clears pre-release state
- [`version promote`](./promote.md) - Promote to next stage or stable
- [`version snapshot`](./snapshot.md) - Create timestamped snapshot build
- [`release-notes`](./release-notes.md) - Generate notes (includes pre-release tags)
- [`release`](./release.md) - Publish release (supports `--prerelease` flag)

## See Also

- [Configuration Reference](../configuration.md)
- [Tag Templates](../tag-templates.md)
- [Consignment Format](../consignment-format.md)
- [CI/CD Integration Guide](../ci-cd.md)
