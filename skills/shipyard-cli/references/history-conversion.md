# Converting Changelogs to Shipyard History

This guide explains how to convert existing CHANGELOG.md files or git history to Shipyard's history.json format.

## Why Convert?

Converting existing changelog data to Shipyard format provides:
- Backward compatibility with existing release history
- Complete version history for changelog regeneration
- Consistent format across old and new releases
- Ability to use Shipyard templates on historical data

## History Format

Shipyard stores release history in `.shipyard/history.json`:

```json
[
  {
    "version": "1.2.3",
    "package": "my-package",
    "tag": "v1.2.3",
    "timestamp": "2024-01-15T10:30:00Z",
    "consignments": [
      {
        "id": "20240115-103000-abc123",
        "summary": "Add new feature\n\nDetailed description here...",
        "changeType": "minor",
        "metadata": {
          "author": "dev@example.com",
          "issue": "JIRA-123"
        }
      },
      {
        "id": "20240115-103001-def456",
        "summary": "Fix bug in authentication",
        "changeType": "patch"
      }
    ]
  },
  {
    "version": "1.2.2",
    "package": "my-package",
    "tag": "v1.2.2",
    "timestamp": "2024-01-10T14:20:00Z",
    "consignments": [
      {
        "id": "20240110-142000-ghi789",
        "summary": "Security fix for CVE-2024-1234",
        "changeType": "patch"
      }
    ]
  }
]
```

## Conversion Methods

### Method 1: Keep a Changelog Format

For changelogs following the [Keep a Changelog](https://keepachangelog.com/) format:

```markdown
# Changelog

## [1.2.3] - 2024-01-15

### Added
- New API endpoint for user profiles
- Support for OAuth authentication

### Fixed
- Memory leak in background service
- Race condition in cache

## [1.2.2] - 2024-01-10

### Security
- Fixed CVE-2024-1234 vulnerability

### Changed
- Updated dependencies
```

**Mapping Rules:**

| Changelog Section | Change Type |
|-------------------|-------------|
| Added | minor |
| Changed | minor |
| Fixed | patch |
| Security | patch |
| Removed | major |
| Deprecated | minor |
| Breaking | major |

**Conversion Script:**

```bash
#!/bin/bash
# Parse Keep a Changelog format

parse_changelog() {
  local file=$1
  local package=$2
  local entries='[]'

  # Parse version blocks
  while IFS= read -r line; do
    if [[ $line =~ ^\#\#\ \[([0-9]+\.[0-9]+\.[0-9]+)\]\ -\ ([0-9]{4}-[0-9]{2}-[0-9]{2}) ]]; then
      version="${BASH_REMATCH[1]}"
      date="${BASH_REMATCH[2]}"

      # Read changes until next version or EOF
      changes=()
      while IFS= read -r change_line; do
        [[ $change_line =~ ^\#\# ]] && break

        # Parse sections
        if [[ $change_line =~ ^\#\#\#\ (Added|Changed|Fixed|Security|Removed) ]]; then
          section="${BASH_REMATCH[1]}"
          change_type=$(get_change_type "$section")
        elif [[ $change_line =~ ^-\ (.+) ]]; then
          summary="${BASH_REMATCH[1]}"
          consignment=$(create_consignment "$summary" "$change_type" "$date")
          changes+=("$consignment")
        fi
      done

      # Create history entry
      entry=$(jq -n \
        --arg version "$version" \
        --arg package "$package" \
        --arg tag "v$version" \
        --arg timestamp "${date}T00:00:00Z" \
        --argjson consignments "$(printf '%s\n' "${changes[@]}" | jq -s .)" \
        '{version: $version, package: $package, tag: $tag, timestamp: $timestamp, consignments: $consignments}')

      entries=$(echo "$entries" | jq --argjson entry "$entry" '. + [$entry]')
    fi
  done < "$file"

  echo "$entries"
}

get_change_type() {
  case $1 in
    Added|Changed|Deprecated) echo "minor" ;;
    Fixed|Security) echo "patch" ;;
    Removed|Breaking) echo "major" ;;
    *) echo "patch" ;;
  esac
}

create_consignment() {
  local summary=$1
  local change_type=$2
  local date=$3
  local id=$(date -d "$date" +%Y%m%d-000000)-$(openssl rand -hex 3)

  jq -n \
    --arg id "$id" \
    --arg summary "$summary" \
    --arg changeType "$change_type" \
    '{id: $id, summary: $summary, changeType: $changeType}'
}

# Usage
parse_changelog CHANGELOG.md my-package > .shipyard/history.json
```

### Method 2: Conventional Commits

Extract history from git commits using conventional commit format:

```bash
#!/bin/bash
# Parse conventional commits from git history

parse_git_history() {
  local from_ref=$1
  local to_ref=${2:-HEAD}
  local package=$3
  local entries='[]'

  # Get all tags between refs
  tags=$(git tag --merged "$to_ref" --no-merged "$from_ref" --sort=version:refname)

  for tag in $tags; do
    version=$(echo "$tag" | sed 's/^v//')
    timestamp=$(git log -1 --format=%aI "$tag")

    # Get commits since previous tag
    prev_tag=$(git describe --tags --abbrev=0 "$tag^" 2>/dev/null || echo "$from_ref")
    commits=$(git log --format="%H|%aI|%s%n%b%n---" "$prev_tag..$tag")

    # Parse commits
    consignments='[]'
    while IFS='|' read -r hash date subject; do
      # Parse conventional commit
      if [[ $subject =~ ^(feat|fix|docs|style|refactor|perf|test|chore)(\(.+\))?!?:\ (.+) ]]; then
        type="${BASH_REMATCH[1]}"
        scope="${BASH_REMATCH[2]}"
        summary="${BASH_REMATCH[3]}"
        breaking="${BASH_REMATCH[4]}"

        # Read body until ---
        body=""
        while IFS= read -r line; do
          [[ $line == "---" ]] && break
          body="$body\n$line"
        done

        # Determine change type
        change_type="patch"
        [[ $type == "feat" ]] && change_type="minor"
        [[ $subject =~ ! || $body =~ BREAKING\ CHANGE ]] && change_type="major"

        # Create consignment
        id=$(date -d "$date" +%Y%m%d-%H%M%S)-$(echo "$hash" | cut -c1-6)
        full_summary="$summary"
        [[ -n $body ]] && full_summary="$summary\n$body"

        consignment=$(jq -n \
          --arg id "$id" \
          --arg summary "$full_summary" \
          --arg changeType "$change_type" \
          '{id: $id, summary: $summary, changeType: $changeType}')

        consignments=$(echo "$consignments" | jq --argjson c "$consignment" '. + [$c]')
      fi
    done <<< "$commits"

    # Create history entry
    entry=$(jq -n \
      --arg version "$version" \
      --arg package "$package" \
      --arg tag "$tag" \
      --arg timestamp "$timestamp" \
      --argjson consignments "$consignments" \
      '{version: $version, package: $package, tag: $tag, timestamp: $timestamp, consignments: $consignments}')

    entries=$(echo "$entries" | jq --argjson entry "$entry" '. + [$entry]')
  done

  echo "$entries"
}

# Usage
parse_git_history v1.0.0 HEAD my-package > .shipyard/history.json
```

### Method 3: GitHub Releases

Extract history from GitHub releases API:

```bash
#!/bin/bash
# Fetch and parse GitHub releases

parse_github_releases() {
  local owner=$1
  local repo=$2
  local package=$3
  local entries='[]'

  # Fetch releases
  releases=$(gh api "repos/$owner/$repo/releases" --paginate)

  # Parse each release
  echo "$releases" | jq -c '.[]' | while read -r release; do
    tag=$(echo "$release" | jq -r '.tag_name')
    version=$(echo "$tag" | sed 's/^v//')
    timestamp=$(echo "$release" | jq -r '.published_at')
    body=$(echo "$release" | jq -r '.body')

    # Parse release notes to consignments
    consignments='[]'

    # Extract bullet points as consignments
    echo "$body" | grep -E '^[-*]\s+' | while read -r line; do
      summary=$(echo "$line" | sed 's/^[-*]\s*//')

      # Infer change type from content
      change_type="patch"
      [[ $summary =~ ^(feat|add|new|feature): ]] && change_type="minor"
      [[ $summary =~ ^(breaking|remove|drop): ]] && change_type="major"

      # Generate ID
      id=$(date -d "$timestamp" +%Y%m%d-%H%M%S)-$(openssl rand -hex 3)

      consignment=$(jq -n \
        --arg id "$id" \
        --arg summary "$summary" \
        --arg changeType "$change_type" \
        '{id: $id, summary: $summary, changeType: $changeType}')

      consignments=$(echo "$consignments" | jq --argjson c "$consignment" '. + [$c]')
    done

    # Create history entry
    entry=$(jq -n \
      --arg version "$version" \
      --arg package "$package" \
      --arg tag "$tag" \
      --arg timestamp "$timestamp" \
      --argjson consignments "$consignments" \
      '{version: $version, package: $package, tag: $tag, timestamp: $timestamp, consignments: $consignments}')

    entries=$(echo "$entries" | jq --argjson entry "$entry" '. + [$entry]')
  done

  echo "$entries"
}

# Usage
parse_github_releases myorg myrepo my-package > .shipyard/history.json
```

### Method 4: Manual Conversion

For complex changelogs or non-standard formats, manual conversion may be necessary:

```json
[
  {
    "version": "1.2.3",
    "package": "my-package",
    "tag": "v1.2.3",
    "timestamp": "2024-01-15T10:30:00Z",
    "consignments": [
      {
        "id": "20240115-103000-abc123",
        "summary": "Add new feature",
        "changeType": "minor"
      }
    ]
  }
]
```

**Guidelines:**
1. Create one entry per version
2. Maintain reverse chronological order (newest first)
3. Generate unique IDs for consignments
4. Map changes to appropriate change types
5. Include detailed summaries if available

## Consignment ID Generation

Generate IDs in format `YYYYMMDD-HHMMSS-{random6}`:

```bash
# Using date from changelog
date -d "2024-01-15" +%Y%m%d-000000

# Append random suffix
echo "$(date -d "2024-01-15" +%Y%m%d-000000)-$(openssl rand -hex 3)"
# Output: 20240115-000000-a1b2c3
```

## Change Type Mapping

**Conventional Commits:**
| Commit Type | Change Type |
|-------------|-------------|
| feat | minor |
| fix | patch |
| perf | patch |
| refactor | patch |
| docs | patch |
| style | patch |
| test | patch |
| chore | patch |
| BREAKING CHANGE | major |
| ! suffix | major |

**Keep a Changelog:**
| Section | Change Type |
|---------|-------------|
| Added | minor |
| Changed | minor |
| Deprecated | minor |
| Fixed | patch |
| Security | patch |
| Removed | major |

**Semantic Versioning:**
| Version Change | Change Type |
|----------------|-------------|
| X.0.0 | major |
| 0.X.0 | minor |
| 0.0.X | patch |

## Validation

After conversion, validate the history file:

```bash
# Check JSON syntax
jq . .shipyard/history.json

# Validate with Shipyard
shipyard validate

# Test changelog generation
shipyard release-notes --package my-package
```

## Complete Example

Converting a Keep a Changelog file:

```markdown
# Changelog

## [1.2.0] - 2024-01-15

### Added
- User authentication system
- OAuth2 provider support

### Fixed
- Memory leak in background worker

## [1.1.0] - 2024-01-01

### Added
- New API endpoints
- Rate limiting

### Changed
- Updated dependencies
```

**Result:**

```json
[
  {
    "version": "1.2.0",
    "package": "my-api",
    "tag": "v1.2.0",
    "timestamp": "2024-01-15T00:00:00Z",
    "consignments": [
      {
        "id": "20240115-000000-abc123",
        "summary": "User authentication system",
        "changeType": "minor"
      },
      {
        "id": "20240115-000001-def456",
        "summary": "OAuth2 provider support",
        "changeType": "minor"
      },
      {
        "id": "20240115-000002-ghi789",
        "summary": "Memory leak in background worker",
        "changeType": "patch"
      }
    ]
  },
  {
    "version": "1.1.0",
    "package": "my-api",
    "tag": "v1.1.0",
    "timestamp": "2024-01-01T00:00:00Z",
    "consignments": [
      {
        "id": "20240101-000000-jkl012",
        "summary": "New API endpoints",
        "changeType": "minor"
      },
      {
        "id": "20240101-000001-mno345",
        "summary": "Rate limiting",
        "changeType": "minor"
      },
      {
        "id": "20240101-000002-pqr678",
        "summary": "Updated dependencies",
        "changeType": "minor"
      }
    ]
  }
]
```

## Using Converted History

Once converted, Shipyard uses the history for:

### Changelog Generation

```bash
# Regenerate changelog from history
shipyard release-notes --package my-api > CHANGELOG.md
```

### Release Notes

```bash
# Generate release notes for specific version
shipyard release-notes --package my-api --version 1.2.0
```

### Status Checking

```bash
# Shows both pending consignments and history
shipyard status
```

## Best Practices

### During Conversion

- Start with most recent versions (easier to parse)
- Verify timestamps match actual release dates
- Include all available detail in summaries
- Map change types conservatively (prefer patch over minor)
- Validate after each major section

### After Conversion

- Commit history.json to repository
- Delete or archive old CHANGELOG.md
- Use Shipyard for all new changes
- Regenerate CHANGELOG.md from history
- Update CI/CD to use Shipyard

### Maintaining History

- Never manually edit history.json (Shipyard manages it)
- Keep history.json in version control
- Back up history file before major changes
- Validate history after merges

## Troubleshooting

### Invalid JSON

```bash
# Check syntax
jq . .shipyard/history.json
# If error, use jq to find line number
```

### Missing Timestamps

```bash
# Add default timestamps
jq 'map(. + {timestamp: (.timestamp // "2024-01-01T00:00:00Z")})' \
  .shipyard/history.json > .shipyard/history.json.tmp
mv .shipyard/history.json.tmp .shipyard/history.json
```

### Duplicate IDs

```bash
# Check for duplicates
jq -r '.[].consignments[].id' .shipyard/history.json | sort | uniq -d

# Regenerate IDs if duplicates found
# (manual process - each entry needs unique ID)
```

### Invalid Change Types

```bash
# Check all change types are valid
jq -r '.[].consignments[].changeType' .shipyard/history.json | sort -u

# Should only output: major, minor, patch
```

## Migration Checklist

- [ ] Backup existing CHANGELOG.md
- [ ] Choose conversion method
- [ ] Run conversion script
- [ ] Validate JSON syntax
- [ ] Validate with `shipyard validate`
- [ ] Test changelog generation
- [ ] Compare with original CHANGELOG.md
- [ ] Commit history.json
- [ ] Update documentation
- [ ] Update CI/CD configuration
- [ ] Train team on new workflow
