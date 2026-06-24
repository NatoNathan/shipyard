# promote - Advance through the harbor channel

Promote a pre-release to the next stage or stable release.

## Synopsis

```bash
shipyard version promote [OPTIONS]
shipyard version advance [OPTIONS]
```

**Aliases:** `advance`

Advances a pre-release to the next stage in order, or promotes to stable release.

## Description

The `promote` command advances pre-releases through configured stages. Like navigating from calm testing waters to the open ocean, promotion moves your changes closer to a stable release.

**Key behaviors:**

- **Stage advancement**: Moves to the next stage based on `order` values in configuration
- **Counter reset**: Resets counter to 1 when advancing stages
- **State tracking**: Updates `.shipyard/prerelease.yml` with new stage
- **Target version**: Maintains the same target version (unless consignments changed)
- **Explicit promotion**: At the highest stage, returns error— use `shipyard version` to explicitly promote to stable

## Global Options

| Option | Description |
|--------|-------------|
| `--json`, `-j` | Output in JSON format |
| `--quiet`, `-q` | Suppress all output except errors |
| `--verbose`, `-v` | Enable verbose logging |

## Options

### `--preview`

Show what promotion would do without making any changes.

**Example:**
```bash
$ shipyard version promote --preview
📦 Preview: Promote to next stage
  - my-api: 1.2.0-alpha.5 → 1.2.0-beta.1 (alpha → beta)
    Target version: 1.2.0

ℹ Preview mode: no changes made
```

### `--no-commit`

Apply version changes to files but skip creating a git commit.

**Example:**
```bash
$ shipyard version promote --no-commit
📦 Promoting to next stage...
  - my-api: 1.2.0-alpha.5 → 1.2.0-beta.1 (alpha → beta)
✓ Updated version files
⊘ Skipped git commit (--no-commit)
```

### `--no-tag`

Create the git commit but skip creating git tags.

**Example:**
```bash
$ shipyard version promote --no-tag
📦 Promoting to next stage...
  - my-api: 1.2.0-alpha.5 → 1.2.0-beta.1 (alpha → beta)
✓ Created commit: "chore: pre-release my-api v1.2.0-beta.1"
⊘ Skipped git tags (--no-tag)
```

### `--package <name>`

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

## Stage Progression

### Stage Order

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

### Counter Reset

When promoting to a new stage, the counter always resets to 1:

```bash
# Currently at alpha.5
$ shipyard version promote
# Output: beta.1 (not beta.5)
```

This makes it clear that you're starting fresh at the new stage.

## Workflow

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

## Examples

### Promote from Alpha to Beta

```bash
$ shipyard version promote
📦 Promoting to next stage...
  - my-api: 1.2.0-alpha.5 → 1.2.0-beta.1 (alpha → beta)
✓ Updated .shipyard/prerelease.yml
✓ Created commit: "chore: pre-release my-api v1.2.0-beta.1"
✓ Created tag: v1.2.0-beta.1
```

### Promote from Beta to RC

```bash
$ shipyard version promote
📦 Promoting to next stage...
  - my-api: 1.2.0-beta.3 → 1.2.0-rc.1 (beta → rc)
✓ Updated .shipyard/prerelease.yml
✓ Created commit: "chore: pre-release my-api v1.2.0-rc.1"
✓ Created tag: v1.2.0-rc.1
```

### Already at Highest Stage

```bash
# Currently at rc.2 (highest stage)
$ shipyard version promote
❌ Error: Already at highest pre-release stage 'rc'
   Use 'shipyard version' to promote to stable release
```

### Preview Promotion

```bash
$ shipyard version promote --preview
📦 Preview: Promote to next stage
  - my-api: 1.2.0-alpha.5 → 1.2.0-beta.1 (alpha → beta)
    Target version: 1.2.0
    Based on consignments:
      - 20240130-120000-abc123: Add new API endpoint

ℹ Preview mode: no changes made
```

### Multi-Package Promotion

```bash
$ shipyard version promote
📦 Promoting to next stage...
  - api: 1.2.0-alpha.2 → 1.2.0-beta.1 (alpha → beta)
  - cli: 2.1.0-alpha.5 → 2.1.0-beta.1 (alpha → beta)
✓ Created commits for 2 packages
✓ Created tags: v1.2.0-beta.1, cli-v2.1.0-beta.1
```

### Target Version Changed During Promotion

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

### Promote to Stable Release

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

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error - validation, file, or git operation failed |
| 2 | Already at highest stage (use `shipyard version` for stable) |
| 3 | No pre-release state exists (use `shipyard version prerelease` first) |

## Configuration

### Stage Definitions

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

### Custom Stage Order

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

### State File

The current stage is tracked in `.shipyard/prerelease.yml`:

```yaml
packages:
  my-api:
    stage: beta
    counter: 1
    targetVersion: 1.2.0
```

After promotion, the state file is updated with the new stage and counter reset to 1.

## Behavior Details

### Stage Advancement Logic

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

### Why Require Explicit `version` for Stable?

At the highest pre-release stage, `promote` returns an error instead of automatically promoting to stable. This is intentional:

- **Explicit intent**: Stable releases are significant—should be explicit
- **Different behavior**: Stable release archives consignments, updates changelog
- **Clear semantics**: `promote` = stage advancement, `version` = stable release

This prevents accidental stable releases and makes the intent clear in scripts and CI/CD.

### Multi-Package Stages

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

### Target Version Tracking

The target version is recalculated on each promotion to detect consignment changes:

- **Same target**: Promotion proceeds normally
- **Changed target**: Warning displayed, promotion continues with new target

This helps catch cases where significant changes were added between promotions.

### Git Requirements

Same requirements as other version commands:

- Repository must be a git repository
- Working directory must be clean (or use `--no-commit`)
- Git `user.name` and `user.email` must be configured

## Integration with CI/CD

### Branch-Based Stage Promotion

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

### Manual Promotion Workflow

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

### Stage Gate with Tests

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

## Related Commands

- [`version prerelease`](./prerelease.md) - Create or increment pre-release at current stage
- [`version snapshot`](./snapshot.md) - Create timestamped snapshot build
- [`version`](./version.md) - Promote to stable release (from highest stage)
- [`consign`](./add.md) - Record changes that will be promoted
- [`status`](./status.md) - View current pre-release stage

## See Also

- [Configuration Reference](../configuration.md)
- [Pre-Release Reference](./prerelease.md)
- [Tag Templates](../tag-generation.md)
- [CI/CD Integration Guide](../../README.md#development)
