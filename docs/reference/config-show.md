# config show - Read the ship's charter

## Synopsis

```bash
shipyard config show [OPTIONS]
shipyard config view [OPTIONS]
shipyard cfg show [OPTIONS]
shipyard cfg view [OPTIONS]
```

**Aliases:** `view` (for `show`), `cfg` (for `config`)

## Description

The `config show` command displays the current shipyard configuration with all defaults applied. It:

1. Loads the configuration from `.shipyard/shipyard.yaml`
2. Resolves any remote/extended configurations
3. Applies default values for unset fields
4. Outputs the full resolved configuration

Outputs as YAML by default, or JSON with the `--json` flag.

**Maritime Metaphor**: Read the ship's charter—see the full orders including all standing instructions.

## Global Options

These options are available for all shipyard commands:

| Option | Short | Description |
|--------|-------|-------------|
| `--config <path>` | `-c` | Config file (default: `.shipyard/shipyard.yaml`) |
| `--json` | `-j` | Output in JSON format |
| `--quiet` | `-q` | Suppress non-error output |
| `--verbose` | `-v` | Verbose output |

## Examples

### Default YAML Output

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

### JSON Output

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

### Multi-Package Repository

```bash
shipyard config show
```

Shows all packages, their ecosystems, dependency relationships, and template overrides.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success - configuration displayed |
| 1 | Error - failed to load or marshal configuration |

## Behavior Details

### Resolved Configuration

The output shows the **resolved** configuration after applying:
- Default values for unset fields
- Merged remote/extended configurations
- Default template sources (`builtin:default`)
- Default paths for consignments and history

### YAML vs JSON

- **YAML** (default): Uses lowercase keys matching the config file format
- **JSON** (`--json`): Uses PascalCase keys matching Go struct field names

### Not Initialized

If no configuration file exists, returns an error.

## Related Commands

- [`init`](./init.md) - Initialize shipyard configuration
- [`validate`](./validate.md) - Validate configuration for errors

## See Also

- [Configuration Reference](../configuration.md) - Full configuration file format
