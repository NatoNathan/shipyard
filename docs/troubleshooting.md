# Troubleshooting Guide

This guide covers common errors, their causes, and solutions when using Shipyard.

## Table of Contents

- [Initialization Errors](#initialization-errors)
- [Configuration Errors](#configuration-errors)
- [Consignment Errors](#consignment-errors)
- [Version Errors](#version-errors)
- [Dependency Errors](#dependency-errors)
- [Git Errors](#git-errors)
- [Template Errors](#template-errors)
- [General Troubleshooting Tips](#general-troubleshooting-tips)
- [Need More Help?](#need-more-help)

---

## Initialization Errors

### Error: "Not a git repository"

**Full Error**:
```
Error: Not a git repository
Location: /path/to/project

Shipyard requires a git repository to function.
Initialize git first:
  git init
```

**Cause**: You're trying to initialize Shipyard in a directory that is not a git repository.

**Solution**:
```bash
# Initialize git first
git init

# Then initialize Shipyard
shipyard init
```

---

### Error: "Shipyard already initialized"

**Full Error**:
```
Error: Shipyard already initialized
Location: .shipyard/shipyard.yaml

Use --force to overwrite existing configuration:
  shipyard init --force
```

**Cause**: `.shipyard/shipyard.yaml` already exists in the repository.

**Solutions**:

1. **If you want to keep the existing configuration**, no action needed.

2. **If you want to reinitialize with defaults**:
   ```bash
   shipyard init --force
   ```
   **Warning**: This will overwrite your existing configuration!

3. **If you want to keep some settings**, back up first:
   ```bash
   # Backup existing config
   cp .shipyard/shipyard.yaml .shipyard/shipyard.yaml.backup

   # Reinitialize
   shipyard init --force

   # Manually merge settings from backup
   ```

---

## Configuration Errors

### Error: "Failed to load configuration"

**Full Error**:
```
Error: config error: failed to load configuration: failed to read config file: open .shipyard/shipyard.yaml: no such file or directory
```

**Cause**: Shipyard configuration file doesn't exist or is in the wrong location.

**Solutions**:

1. **Initialize Shipyard**:
   ```bash
   shipyard init
   ```

2. **Check config file location**:
   ```bash
   # Default location
   ls .shipyard/shipyard.yaml

   # Or specify custom path
   shipyard status --config /path/to/config.yaml
   ```

3. **Verify file permissions**:
   ```bash
   # Check if file is readable
   cat .shipyard/shipyard.yaml
   ```

---

### Error: "Invalid package reference"

**Full Error**:
```
Error: config error: invalid package reference: package "api-client" depends on non-existent package "core"
```

**Cause**: A package lists a dependency that doesn't exist in the configuration.

**Solution**:

Edit `.shipyard/shipyard.yaml` to either:

1. **Add the missing package**:
   ```yaml
   packages:
     # Add the missing package
     - name: "core"
       path: "./packages/core"
       ecosystem: "go"

     # Now api-client can reference it
     - name: "api-client"
       path: "./clients/api"
       ecosystem: "npm"
       dependencies:
         - package: "core"
           strategy: "linked"
   ```

2. **Remove the invalid dependency**:
   ```yaml
   packages:
     - name: "api-client"
       path: "./clients/api"
       ecosystem: "npm"
       dependencies: []  # Remove invalid dependency
   ```

**Validation**:
```bash
# Verify configuration is valid
shipyard status
```

---

### Error: "Failed to parse package.json / pyproject.toml"

**Full Error**:
```
Error: failed to parse package.json: unexpected end of JSON input
```

**Cause**: Version file has invalid syntax (malformed JSON, TOML, etc.).

**Solutions**:

1. **Check file syntax**:
   ```bash
   # For package.json (NPM)
   cat package.json | jq .

   # For pyproject.toml (Python)
   cat pyproject.toml | toml-test

   # For Cargo.toml (Rust)
   cargo check
   ```

2. **Common JSON errors**:
   - Trailing commas
   - Unquoted keys
   - Missing closing brackets

3. **Fix or regenerate the file**:
   ```bash
   # For NPM
   npm init --yes

   # For Python
   poetry init

   # For Rust
   cargo init
   ```

---

### Error: "No version field found"

**Full Error**:
```
Error: no version field found in package.json
```

**Cause**: Version file exists but doesn't contain a `version` field.

**Solution**:

Add the version field to your file:

**For package.json (NPM)**:
```json
{
  "name": "my-package",
  "version": "1.0.0"
}
```

**For pyproject.toml (Python)**:
```toml
[project]
name = "my-package"
version = "1.0.0"
```

**For Cargo.toml (Rust)**:
```toml
[package]
name = "my-package"
version = "1.0.0"
```

**For version.go (Go)**:
```go
package main

const Version = "1.0.0"
```

---

### Error: "Circular dependency detected"

**Full Error**:
```
Error: dependency error: circular dependency detected: core -> api-client -> core
```

**Cause**: Two or more packages depend on each other, creating a cycle.

**Understanding**: Circular dependencies are **valid** in Shipyard and treated as a single version bump unit. This error typically appears during validation, not during version application.

**If you see this error during `shipyard status`**, it's informational. Both packages in the cycle will receive the same version bump.

**Example**:
```yaml
packages:
  - name: "service-a"
    dependencies:
      - package: "service-b"
        strategy: "linked"

  - name: "service-b"
    dependencies:
      - package: "service-a"
        strategy: "linked"
```

**Result**: When either service gets a version bump, both services will be bumped together.

---

### Error: "Remote config fetch failed"

**Full Error**:
```
Error: config error: failed to fetch remote config: Get "https://example.com/config.yaml": dial tcp: lookup example.com: no such host
```

**Cause**: Cannot fetch remote configuration from URL or Git repository.

**Solutions**:

1. **Check URL accessibility**:
   ```bash
   # Test HTTP URL
   curl -I https://example.com/config.yaml

   # Test Git repository
   git ls-remote git@github.com:org/configs.git
   ```

2. **Verify authentication** (for private resources):
   ```yaml
   extends:
     - url: "https://api.internal.com/config.yaml"
       auth: "env:CONFIG_TOKEN"  # Ensure env var is set
   ```

   ```bash
   # Check environment variable
   echo $CONFIG_TOKEN
   ```

3. **Check network connectivity**:
   ```bash
   # Test DNS resolution
   nslookup example.com

   # Check VPN connection if using internal resources
   ```

4. **Use explicit format for clarity**:
   ```yaml
   extends:
     - url: "https://example.com/config.yaml"
       auth: "env:GITHUB_TOKEN"
   ```

**Example**: [Remote config examples](../examples/remote-config/)

---

## Consignment Errors

### Error: "No editor available"

**Full Error**:
```
Error: no editor available: $EDITOR not set and no default editor found
```

**Cause**: Shipyard cannot determine which editor to use for creating consignment summaries.

**Solutions**:

1. **Set EDITOR environment variable**:
   ```bash
   # Temporary (current session)
   export EDITOR=vim
   shipyard add

   # Permanent (add to ~/.bashrc or ~/.zshrc)
   echo 'export EDITOR=vim' >> ~/.bashrc
   source ~/.bashrc
   ```

2. **Use --summary flag** (non-interactive mode):
   ```bash
   shipyard add --summary "Your change description" --bump patch
   ```

3. **Specify editor via flag**:
   ```bash
   shipyard add --editor=nano
   ```

---

### Error: "Required metadata field missing"

**Full Error**:
```
Error: validation error: author: required field 'author' is missing
```

**Cause**: Configuration defines required metadata fields, but they weren't provided.

**Solution**:

Provide the required metadata:

**Interactive mode** (prompts for required fields):
```bash
shipyard add
```

**Non-interactive mode**:
```bash
shipyard add \
  --summary "Fix authentication bug" \
  --bump patch \
  --metadata author="dev@example.com" \
  --metadata issue="PROJ-123"
```

**Check required fields** in configuration:
```yaml
# .shipyard/shipyard.yaml
metadata:
  fields:
    - name: "author"
      required: true  # This field is required
      type: "string"
```

**Example**: [NPM package with metadata](../examples/single-repo/npm-package/)

---

### Error: "Summary cannot be empty"

**Full Error**:
```
Error: validation error: summary: summary cannot be empty
```

**Cause**: Consignment summary is empty or only whitespace.

**Solution**:

Provide a meaningful summary:

```bash
# Good examples
shipyard add --summary "Fix null pointer error in authentication module"
shipyard add --summary "Add caching layer to API endpoints"
shipyard add --summary "Update dependencies to latest versions"

# Bad examples (too vague)
shipyard add --summary "Fix bug"
shipyard add --summary "Update"
```

---

## Version Errors

### Error: "Uncommitted changes in repository"

**Full Error**:
```
Error: git error: repository has uncommitted changes

Commit or stash your changes before running 'shipyard version':
  git add .
  git commit -m "Your changes"
```

**Cause**: Git working directory has uncommitted changes. Shipyard requires a clean working tree to prevent mixing version bump changes with other work.

**Solutions**:

1. **Commit pending changes**:
   ```bash
   git add .
   git commit -m "Your changes"
   shipyard version
   ```

2. **Stash changes temporarily**:
   ```bash
   git stash
   shipyard version
   git stash pop  # Restore stashed changes
   ```

3. **Skip commit creation** (not recommended):
   ```bash
   shipyard version --no-commit
   ```

---

### Error: "Version file not found"

**Full Error**:
```
Error: no version file found in Python project at ./packages/core
```

**Cause**: Package directory doesn't contain the expected version file for its ecosystem.

**Solution**:

Create the appropriate version file:

**For Go ecosystem**:
```bash
# Create version.go
cat > ./packages/core/version.go <<EOF
package core

const Version = "1.0.0"
EOF
```

**For NPM ecosystem**:
```bash
# Create or fix package.json
npm init --yes
# Edit package.json to set version: "1.0.0"
```

**For Python ecosystem**:
```bash
# Create pyproject.toml
cat > ./packages/core/pyproject.toml <<EOF
[project]
name = "core"
version = "1.0.0"
EOF
```

**For Docker ecosystem**:
```bash
# Add version label to Dockerfile
echo 'LABEL version="1.0.0"' >> Dockerfile
```

**Example**: [Ecosystem examples](../examples/single-repo/)

---

### Error: "No pending consignments"

**Full Error**:
```
No pending consignments
```

**Cause**: There are no consignments to process in `.shipyard/consignments/`.

**Solution**:

This is not an error - just means there are no pending changes. Create a consignment first:

```bash
shipyard add --summary "Your change" --bump patch
shipyard version
```

---

## Dependency Errors

### Error: "Dependency package not found"

**Full Error**:
```
Error: dependency error: package "web-app" depends on non-existent package "api-client"
```

**Cause**: Package configuration references a dependency that doesn't exist.

**Solution**: See [Configuration Errors - Invalid package reference](#error-invalid-package-reference)

---

## Git Errors

### Error: "Tag already exists"

**Full Error**:
```
Error: git error: tag v1.0.0 already exists

Delete the existing tag first:
  git tag -d v1.0.0
  git push origin :refs/tags/v1.0.0  # Also delete from remote
```

**Cause**: Trying to create a git tag that already exists.

**Solutions**:

1. **Delete local and remote tag**:
   ```bash
   # Delete local tag
   git tag -d v1.0.0

   # Delete remote tag
   git push origin :refs/tags/v1.0.0

   # Re-run version command
   shipyard version
   ```

2. **Skip tag creation** (not recommended):
   ```bash
   shipyard version --no-tag
   ```

3. **Use different version**:
   ```bash
   # Manually edit version file to use different version
   # Then run shipyard version
   ```

---

### Error: "Not a git repository"

**Full Error**:
```
Error: git error: not a git repository
```

**Cause**: Current directory is not a git repository or `.git` directory is missing.

**Solution**:

```bash
# Check if .git exists
ls -la .git

# If not, initialize git
git init

# Then initialize Shipyard
shipyard init
```

---

### Error: "Git remote not configured"

**Full Error**:
```
Error: git error: no remote repository configured
```

**Cause**: Trying to use `--publish` flag but git remote is not configured.

**Solution**:

```bash
# Add git remote
git remote add origin git@github.com:your-org/your-repo.git

# Or use HTTPS
git remote add origin https://github.com/your-org/your-repo.git

# Verify remote
git remote -v
```

---

## Template Errors

### Error: "Template rendering failed"

**Full Error**:
```
Error: template error: failed to render changelog template: template: changelog:10:15: executing "changelog" at <.InvalidField>: can't evaluate field InvalidField
```

**Cause**: Template references a field that doesn't exist or has syntax errors.

**Solution**:

1. **Check template syntax** in `.shipyard/shipyard.yaml`:
   ```yaml
   templates:
     changelog: |
       # Changelog

       ## Version {{.Version}}  # Correct
       # NOT: {{.InvalidField}}  # Incorrect - field doesn't exist

       {{range .Consignments}}
       - {{.Summary}}
       {{end}}
   ```

2. **Available template variables**:
   - `.Version` - Current version
   - `.Package` - Package name
   - `.Date` - Release date
   - `.Consignments` - Array of consignments
   - `.Consignments[].Summary` - Change description
   - `.Consignments[].ChangeType` - "major", "minor", or "patch"
   - `.Consignments[].Metadata` - Custom metadata fields

3. **Test template locally**:
   ```bash
   # Dry run to see template output
   shipyard version --preview
   ```

**Example**: [Template examples](../examples/templates/)

---

### Error: "Remote config fetch failed"

See [Configuration Errors - Remote config fetch failed](#error-remote-config-fetch-failed)

---

## General Troubleshooting Tips

### Enable Verbose Logging

Get detailed information about what Shipyard is doing:

```bash
# Add --verbose flag to any command
shipyard status --verbose
shipyard version --verbose

# Or use global flag
shipyard --verbose status
```

### Check Configuration

Verify your configuration is valid:

```bash
# Display current configuration
cat .shipyard/shipyard.yaml

# Validate by running status
shipyard status
```

### Verify Git State

Ensure git repository is in expected state:

```bash
# Check git status
git status

# Check current branch
git branch

# Check remote
git remote -v

# Check tags
git tag -l
```

### Reset to Clean State

If things are really broken, reset to a clean state:

```bash
# Back up consignments
cp -r .shipyard/consignments .shipyard/consignments.backup

# Remove Shipyard configuration
rm -rf .shipyard/

# Reinitialize
shipyard init

# Restore consignments
cp -r .shipyard/consignments.backup/* .shipyard/consignments/
```

### Check File Permissions

Ensure Shipyard can read/write necessary files:

```bash
# Check directory permissions
ls -la .shipyard/

# Fix permissions if needed
chmod 755 .shipyard/
chmod 644 .shipyard/shipyard.yaml
chmod 755 .shipyard/consignments/
```

### Validate Version Files

Ensure version files are in expected format:

```bash
# For Go
cat version.go

# For NPM
cat package.json | jq .version

# For Python
cat pyproject.toml | grep version
```

### Clear Stuck State

If Shipyard seems stuck or corrupted:

```bash
# Remove history file (will lose version history)
rm .shipyard/history.json

# Or back it up first
mv .shipyard/history.json .shipyard/history.json.backup
```

### Use Preview Mode

Test changes without applying them:

```bash
# Preview what would happen
shipyard version --preview

# Shows:
# - Version bumps
# - Changelog changes
# - Tags that would be created
```

---

## Need More Help?

### Check Documentation

- [Configuration Schema](https://shipyard.tamez.dev/docs/config) - Full config reference
- [CLI Interface](https://shipyard.tamez.dev/docs/cli) - Command reference
- [Examples](../examples/) - Real-world configurations

### Search Existing Issues

Check if someone else has encountered the same problem:

1. Visit https://github.com/natonathan/shipyard/issues
2. Search for your error message or problem
3. Read existing solutions

### Open a New Issue

If you can't find a solution:

1. Go to https://github.com/natonathan/shipyard/issues/new
2. Include:
   - Full error message
   - Shipyard version (`shipyard --version`)
   - Configuration file (redact sensitive data)
   - Steps to reproduce
   - Expected vs actual behavior

### Enable Debug Mode

Provide debug output when reporting issues:

```bash
# Run with verbose flag
shipyard --verbose <command> 2>&1 | tee debug.log

# Attach debug.log to your issue
```

---

**Quick Reference Card**:

| Problem | Quick Fix |
|---------|-----------|
| Not initialized | `shipyard init` |
| No editor | `export EDITOR=vim` |
| Git unclean | `git add . && git commit` |
| Tag exists | `git tag -d <tag>` |
| Config invalid | Check `.shipyard/shipyard.yaml` syntax |
| No version file | Create appropriate file for ecosystem |
| Template error | Check template syntax with `--preview` |
| Remote fetch fail | Check network/auth/URL |

---

**Still stuck?** Open an issue at https://github.com/natonathan/shipyard/issues with:
- Error message
- Steps to reproduce
- Debug output (`--verbose` flag)
