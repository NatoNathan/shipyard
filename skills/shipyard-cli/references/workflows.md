# Shipyard Workflows

This document describes common workflows and patterns for using Shipyard in different scenarios.

## Table of Contents

1. [Single Package Workflows](#single-package-workflows)
2. [Monorepo Workflows](#monorepo-workflows)
3. [Release Workflows](#release-workflows)
4. [CI/CD Integration](#cicd-integration)
5. [Migration Workflows](#migration-workflows)
6. [Team Workflows](#team-workflows)

## Single Package Workflows

### Basic Single Package Workflow

```bash
# 1. Initialize Shipyard
shipyard init --yes

# 2. Make code changes and add consignment
shipyard add --type minor --summary "Add user authentication"

# 3. Continue development, add more consignments
shipyard add --type patch --summary "Fix login validation bug"

# 4. Check pending changes
shipyard status

# 5. Preview version bump
shipyard version --preview

# 6. Apply version bump
shipyard version

# 7. Push to remote
git push origin main --tags

# 8. Create GitHub release
shipyard release
```

### Feature Branch Workflow

```bash
# 1. Create feature branch
git checkout -b feature/user-auth

# 2. Make changes and create consignments
shipyard add --type minor --summary "Add OAuth2 support"
shipyard add --type patch --summary "Update dependencies"

# 3. Commit consignments
git add .shipyard/consignments/
git commit -m "feat: add OAuth2 support"

# 4. Push feature branch
git push origin feature/user-auth

# 5. Create PR (consignments are reviewed)

# 6. After merge to main, apply version
git checkout main
git pull
shipyard version
git push --tags
```

### Hotfix Workflow

```bash
# 1. Create hotfix branch from production tag
git checkout -b hotfix/security-fix v1.2.3

# 2. Make critical fix
# ... code changes ...

# 3. Create patch consignment
shipyard add --type patch --summary "Fix security vulnerability CVE-2024-1234"

# 4. Apply version (creates v1.2.4)
shipyard version

# 5. Push hotfix
git push origin hotfix/security-fix --tags

# 6. Create release
shipyard release

# 7. Merge back to main
git checkout main
git merge hotfix/security-fix
git push
```

## Monorepo Workflows

### Basic Monorepo Workflow

```bash
# 1. Initialize monorepo
shipyard init
# Select multiple packages during setup

# 2. Add changes affecting single package
shipyard add --package api --type minor --summary "Add new endpoint"

# 3. Add changes affecting multiple packages
shipyard add --package api --package sdk --type major --summary "Breaking API change"

# 4. Check status shows all packages
shipyard status

# 5. Preview version propagation
shipyard version --preview

# 6. Apply versions (handles dependencies)
shipyard version

# 7. Push all tags
git push --tags

# 8. Release packages individually
shipyard release --package api
shipyard release --package sdk
```

### Linked Dependencies Workflow

```yaml
# Configuration: packages/api and packages/sdk are linked
packages:
  - name: api
    path: packages/api
    ecosystem: npm

  - name: sdk
    path: packages/sdk
    ecosystem: npm
    dependencies:
      - name: api
        type: linked
```

```bash
# 1. Change API (minor bump)
shipyard add --package api --type minor --summary "Add new endpoint"

# 2. Preview shows SDK also gets minor bump
shipyard version --preview
# Output:
#   - api: 1.0.0 → 1.1.0 (minor)
#   - sdk: 2.0.0 → 2.1.0 (minor) [propagated]

# 3. Apply versions
shipyard version

# 4. Both packages versioned together
git push --tags
```

### Independent Package Workflow

```bash
# 1. Version only specific package
shipyard add --package web-app --type patch --summary "Fix CSS bug"

# 2. Version only web-app (other packages unchanged)
shipyard version --package web-app

# 3. Other consignments remain for later
shipyard status
# Shows pending consignments for other packages

# 4. Version another package later
shipyard add --package api --type minor --summary "Add endpoint"
shipyard version --package api
```

### Multi-Package Release Workflow

```bash
# 1. Add changes to multiple packages over time
shipyard add --package api --type minor --summary "Add feature A"
shipyard add --package sdk --type minor --summary "Add feature B"
shipyard add --package web --type patch --summary "Fix bug C"

# 2. Check all pending changes
shipyard status

# 3. Version all packages at once
shipyard version

# 4. Selective releases based on priority
git push --tags

# Critical packages first
shipyard release --package api
shipyard release --package sdk

# Less critical later
shipyard release --package web
```

## Release Workflows

### Standard Release Workflow

```bash
# 1. Accumulate changes during development
shipyard add --type minor --summary "Feature 1"
shipyard add --type minor --summary "Feature 2"
shipyard add --type patch --summary "Bug fix"

# 2. When ready to release, check status
shipyard status

# 3. Preview version (highest change type wins)
shipyard version --preview
# Output: 1.0.0 → 1.1.0 (minor)

# 4. Apply version
shipyard version

# 5. Push to remote
git push origin main --tags

# 6. Create GitHub release
shipyard release

# 7. Verify release
# Check GitHub releases page
```

### Pre-release Workflow

```bash
# 1. Create snapshot for testing
shipyard snapshot --package my-api
# Creates: 1.2.3-20240315120000

# 2. Test snapshot version

# 3. If testing passes, promote to stable
shipyard promote --package my-api
# Creates: 1.2.3

# Alternative: Named pre-releases
shipyard prerelease --package my-api --identifier beta
# Creates: 1.2.3-beta.1

# Increment beta
shipyard prerelease --package my-api --identifier beta
# Creates: 1.2.3-beta.2

# Promote when ready
shipyard promote --package my-api
# Creates: 1.2.3
```

### Release Candidate Workflow

```bash
# 1. Add all changes for release
shipyard add --type minor --summary "Feature complete"

# 2. Create RC1
shipyard prerelease --identifier rc
# Creates: 1.3.0-rc.1

# 3. Deploy and test RC1

# 4. If issues found, fix and create RC2
shipyard add --type patch --summary "Fix RC1 issues"
shipyard prerelease --identifier rc
# Creates: 1.3.0-rc.2

# 5. When RC passes testing, promote
shipyard promote
# Creates: 1.3.0

# 6. Release to production
git push --tags
shipyard release
```

### Draft Release Workflow

```bash
# 1. Version and create draft release
shipyard version
git push --tags

shipyard release --draft

# 2. Manually edit release notes on GitHub
# Add screenshots, videos, etc.

# 3. Publish when ready via GitHub UI
```

## CI/CD Integration

### GitHub Actions - Basic

```yaml
name: Release

on:
  push:
    branches: [main]

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0

      - name: Check for pending consignments
        id: check
        run: |
          if [ -n "$(ls -A .shipyard/consignments/)" ]; then
            echo "has_consignments=true" >> $GITHUB_OUTPUT
          fi

      - name: Apply versions
        if: steps.check.outputs.has_consignments == 'true'
        run: |
          shipyard version

      - name: Push tags
        if: steps.check.outputs.has_consignments == 'true'
        run: |
          git push --tags

      - name: Create releases
        if: steps.check.outputs.has_consignments == 'true'
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          for pkg in $(shipyard status --json | jq -r '.packages[].name'); do
            shipyard release --package $pkg
          done
```

### GitHub Actions - With Validation

```yaml
name: CI

on:
  pull_request:
    branches: [main]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Validate Shipyard configuration
        run: shipyard validate

      - name: Check consignment required
        run: |
          # Check if code changed (not just docs)
          if git diff --name-only origin/main | grep -v "^docs/"; then
            # Ensure at least one consignment exists
            if [ -z "$(ls -A .shipyard/consignments/)" ]; then
              echo "Error: Code changes require a consignment"
              exit 1
            fi
          fi

      - name: Preview version changes
        run: shipyard version --preview
```

### GitHub Actions - Monorepo

```yaml
name: Release Monorepo

on:
  workflow_dispatch:
    inputs:
      packages:
        description: 'Packages to release (comma-separated, or "all")'
        required: true
        default: 'all'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0

      - name: Apply versions
        run: |
          if [ "${{ inputs.packages }}" = "all" ]; then
            shipyard version
          else
            for pkg in $(echo "${{ inputs.packages }}" | tr ',' ' '); do
              shipyard version --package $pkg
            done
          fi

      - name: Push tags
        run: git push --tags

      - name: Create releases
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          if [ "${{ inputs.packages }}" = "all" ]; then
            for pkg in $(shipyard status --json | jq -r '.packages[].name'); do
              shipyard release --package $pkg || echo "Package $pkg already released"
            done
          else
            for pkg in $(echo "${{ inputs.packages }}" | tr ',' ' '); do
              shipyard release --package $pkg
            done
          fi
```

### GitLab CI

```yaml
stages:
  - validate
  - release

validate:
  stage: validate
  script:
    - shipyard validate
    - shipyard version --preview
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"

release:
  stage: release
  script:
    - git config user.name "GitLab CI"
    - git config user.email "ci@gitlab.com"
    - shipyard version
    - git push --tags
    - |
      for pkg in $(shipyard status --json | jq -r '.packages[].name'); do
        shipyard release --package $pkg
      done
  rules:
    - if: $CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH
  only:
    - main
```

## Migration Workflows

### From Manual Versioning

```bash
# 1. Initialize Shipyard
shipyard init

# 2. Convert existing CHANGELOG.md (see history-conversion.md)
./examples/convert-changelog.sh CHANGELOG.md > .shipyard/history.json

# 3. Verify conversion
shipyard validate

# 4. Start using Shipyard for new changes
shipyard add --type minor --summary "First Shipyard-managed change"

# 5. Continue with normal workflow
shipyard version
```

### From Conventional Commits

```bash
# 1. Initialize Shipyard
shipyard init

# 2. Extract history from git commits
./examples/convert-changelog.sh --from-git v1.0.0..HEAD > .shipyard/history.json

# 3. Validate
shipyard validate

# 4. Continue development with consignments
shipyard add --type minor --summary "New feature"
```

### From Lerna/Changesets

```bash
# 1. Initialize Shipyard
shipyard init

# 2. Convert existing changesets to consignments
# For each changeset file:
for file in .changeset/*.md; do
  # Parse changeset format
  # Extract package, type, summary
  # Create equivalent consignment
  shipyard add --package $pkg --type $type --summary "$summary"
done

# 3. Remove old changeset files
rm -rf .changeset/

# 4. Update CI/CD to use Shipyard
# Replace lerna/changesets commands with shipyard
```

## Team Workflows

### PR Review Workflow

```bash
# Developer creates feature
git checkout -b feature/new-api

# Make changes
# ...

# Create consignment
shipyard add --package api --type minor --summary "Add new API endpoint"

# Commit everything
git add .
git commit -m "feat: add new API endpoint"
git push origin feature/new-api

# Create PR
# Reviewers can see:
# 1. Code changes
# 2. Consignment (documents the change)
# 3. Version impact (via shipyard version --preview)

# After approval and merge
git checkout main
git pull
shipyard version
git push --tags
```

### Release Manager Workflow

```bash
# Weekly release process

# 1. Review accumulated consignments
shipyard status

# 2. Check for any issues
shipyard validate

# 3. Preview version changes
shipyard version --preview

# 4. If OK, apply versions
shipyard version

# 5. Review changes before pushing
git log -1 --stat
git tag

# 6. Push to remote
git push origin main --tags

# 7. Create releases
for pkg in api sdk web; do
  shipyard release --package $pkg
done

# 8. Announce releases
# Send team notification with release notes
```

### Distributed Team Workflow

```bash
# Team members in different timezones

# Team member 1 (morning)
shipyard add --package api --type minor --summary "Add feature A"
git add .shipyard/consignments/
git commit -m "feat: add feature A"
git push

# Team member 2 (afternoon)
git pull
shipyard add --package sdk --type minor --summary "Add feature B"
git add .shipyard/consignments/
git commit -m "feat: add feature B"
git push

# Team member 3 (evening)
git pull
shipyard add --package web --type patch --summary "Fix bug C"
git add .shipyard/consignments/
git commit -m "fix: fix bug C"
git push

# Release manager (next day)
git pull
shipyard status  # See all accumulated changes
shipyard version # Version all packages
git push --tags
```

## Best Practices

### Consignment Creation

- Create consignments as changes are made (not at release time)
- One consignment per logical change
- Use descriptive summaries
- Include metadata (author, issue links)
- Commit consignments with code changes

### Version Application

- Preview before applying (`--preview`)
- Validate configuration first
- Check status regularly
- Apply versions when ready to release
- Push tags immediately after versioning

### Release Process

- Version all packages together (monorepo)
- Release critical packages first
- Use draft releases for manual editing
- Test pre-releases before promoting
- Automate with CI/CD where possible

### Team Coordination

- Review consignments in PRs
- Discuss change types in code review
- Schedule regular release cycles
- Document release process
- Use metadata for traceability

### CI/CD Integration

- Validate on every PR
- Require consignments for code changes
- Automate version application
- Create releases automatically
- Notify team of releases
