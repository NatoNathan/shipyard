# snapshot

Create a timestamped snapshot pre-release version.

## Synopsis

```bash
shipyard version snapshot [OPTIONS]
```

Creates a snapshot pre-release version with a timestamp identifier, independent of the stage-based pre-release system.

## Description

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

## Global Options

| Option | Description |
|--------|-------------|
| `--config`, `-c` | Path to configuration file (default: `.shipyard/shipyard.yaml`) |
| `--json`, `-j` | Output in JSON format |
| `--quiet`, `-q` | Suppress all output except errors |
| `--verbose`, `-v` | Enable verbose logging |

## Options

### `--preview`

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

### `--no-commit`

Apply version changes to files but skip creating a git commit.

**Example:**
```bash
$ shipyard version snapshot --no-commit
📦 Creating snapshot version...
  - my-api: 1.2.0 → 1.2.0-snapshot.20260204-153045
✓ Updated version files
⊘ Skipped git commit (--no-commit)
```

### `--no-tag`

Create the git commit but skip creating git tags.

**Example:**
```bash
$ shipyard version snapshot --no-tag
📦 Creating snapshot version...
  - my-api: 1.2.0 → 1.2.0-snapshot.20260204-153045
✓ Created commit: "chore: snapshot my-api v1.2.0-snapshot.20260204-153045"
⊘ Skipped git tags (--no-tag)
```

### `--package <name>`

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

## Workflow

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

## Configuration

### Snapshot Template

Configure the snapshot tag template in `.shipyard/shipyard.yaml`:

```yaml
prerelease:
  snapshotTagTemplate: "v{{.Version}}-snapshot.{{.Timestamp}}"
```

### Template Variables

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

## Examples

### Create Snapshot

```bash
$ shipyard version snapshot
📦 Creating snapshot version...
  - my-api: 1.2.0 → 1.2.0-snapshot.20260204-153045
✓ Created commit: "chore: snapshot my-api v1.2.0-snapshot.20260204-153045"
✓ Created tag: v1.2.0-snapshot.20260204-153045
```

### Preview Snapshot

```bash
$ shipyard version snapshot --preview
📦 Preview: Snapshot version
  - my-api: 1.2.0 → 1.2.0-snapshot.20260204-153045
    Target version: 1.2.0
    Based on consignments:
      - 20240130-120000-abc123: Add new API endpoint

ℹ Preview mode: no changes made
```

### Snapshot Without Git Operations

Useful for CI builds where you want the version file updated but handle git operations separately:

```bash
$ shipyard version snapshot --no-commit --no-tag
📦 Creating snapshot version...
  - my-api: 1.2.0 → 1.2.0-snapshot.20260204-153045
✓ Updated version files
⊘ Skipped git commit (--no-commit)
⊘ Skipped git tags (--no-tag)
```

### Multi-Package Snapshot

```bash
$ shipyard version snapshot
📦 Creating snapshot version...
  - api: 1.2.0 → 1.2.0-snapshot.20260204-153045
  - cli: 2.1.0 → 2.1.0-snapshot.20260204-153045
✓ Created commits for 2 packages
✓ Created tags: v1.2.0-snapshot.20260204-153045, cli-v2.1.0-snapshot.20260204-153045
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error - validation, file, or git operation failed |
| 2 | No consignments to base snapshot on |

## Behavior Details

### Snapshot Independence

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

### Snapshot Ordering

Snapshots use timestamps, which naturally order chronologically:

```
v1.2.0-snapshot.20260204-120000
v1.2.0-snapshot.20260204-121500
v1.2.0-snapshot.20260204-130245
```

This makes it easy to identify and sort snapshot builds in CI/CD systems.

### Consignments

Like stage-based pre-releases, snapshots:
- Calculate target version from pending consignments
- Don't archive consignments
- Don't update changelogs
- Preserve consignments for eventual stable release

### Git Requirements

Same requirements as other version commands:

- Repository must be a git repository
- Working directory must be clean (or use `--no-commit`)
- Git `user.name` and `user.email` must be configured

## Integration with CI/CD

### Pull Request Snapshots

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

### Scheduled Nightly Snapshots

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

### Docker Image Tagging

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

## Related Commands

- [`version prerelease`](./prerelease.md) - Create stage-based pre-release (alpha, beta, rc)
- [`version promote`](./promote.md) - Promote to next stage or stable
- [`version`](./version.md) - Create stable release
- [`consign`](./consign.md) - Record changes that will be in snapshot
- [`release`](./release.md) - Publish release (snapshots can use `--prerelease` flag)

## See Also

- [Configuration Reference](../configuration.md)
- [Pre-Release Reference](./prerelease.md)
- [Tag Templates](../tag-templates.md)
- [CI/CD Integration Guide](../ci-cd.md)
