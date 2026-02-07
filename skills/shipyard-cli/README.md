# Shipyard CLI Skill

This skill teaches AI agents how to use Shipyard for semantic versioning and release management.

## Structure

```
shipyard-cli/
├── SKILL.md                          # Core skill content (loaded when triggered)
└── references/                       # Detailed documentation (loaded as needed)
    ├── commands.md                   # Complete command reference (all 14 commands)
    ├── configuration.md              # shipyard.yaml deep dive
    ├── workflows.md                  # Common workflow patterns and CI/CD
    ├── templates.md                  # Template system documentation
    └── history-conversion.md         # Converting existing changelogs
```

## Philosophy

This skill provides **context and patterns**, not scripts. AI agents use this knowledge to:
- Write customized workflows for the user's specific needs
- Understand Shipyard commands and when to use them
- Adapt patterns to different project structures
- Generate appropriate shell scripts when needed

Scripts are not included because:
- Shipyard already provides CLI commands for core operations
- Workflows vary significantly by project and team
- Teaching patterns is more valuable than providing exact code
- Agents can customize solutions based on context

## Installation

This skill can be installed using the vercel-labs/skills protocol:

```bash
# Install to Claude Code
npx skills add NatoNathan/shipyard --agent claude-code

# Install to multiple agents
npx skills add NatoNathan/shipyard --agent claude-code --agent cursor

# Install globally
npx skills add NatoNathan/shipyard --global
```

Or use Shipyard's built-in skill command (once implemented):

```bash
# Install to Claude Code
shipyard skill add NatoNathan/shipyard --agent claude-code

# Install to multiple agents globally
shipyard skill add NatoNathan/shipyard -g -a claude-code -a cursor
```

## Usage

Once installed, AI agents will automatically use this skill when users ask about:
- Creating consignments
- Bumping versions
- Creating releases
- Managing changelogs
- Converting existing changelog files
- Monorepo versioning
- Shipyard configuration

## Skill Triggers

The skill is triggered when users ask to:
- "create a consignment"
- "add a shipyard consignment"
- "bump version"
- "shipyard version"
- "create release"
- "publish release"
- "generate changelog"
- "check shipyard status"
- "convert changelog to history"
- "initialize shipyard"

Or mention:
- Shipyard semantic versioning
- Monorepo versioning
- Consignment-based release management
- Changelog conversion

## What's Included

### SKILL.md (~1,800 words)
Core concepts and workflow:
- What is Shipyard and consignments
- Core workflow: add → version → release
- Command quick reference
- When to use Shipyard
- Configuration basics
- Common workflows
- Best practices

### references/commands.md
Complete reference for all 14 commands:
- `init`, `add`, `status`, `version`, `release`
- `release-notes`, `validate`, `remove`, `snapshot`
- `promote`, `prerelease`, `config-show`, `completion`, `upgrade`
- All options, flags, examples, exit codes

### references/configuration.md
Deep dive into shipyard.yaml:
- Package definitions
- Template system
- Dependency management
- GitHub integration
- Metadata fields
- Multi-ecosystem support

### references/workflows.md
Common patterns and CI/CD integration:
- Single package workflows
- Monorepo workflows
- Release workflows (standard, pre-release, draft)
- CI/CD examples (GitHub Actions, GitLab CI)
- Team workflows
- Best practices

### references/templates.md
Template system documentation:
- Built-in templates
- Custom templates
- Template data structures
- Template functions
- Examples

### references/history-conversion.md
Converting existing changelogs:
- History format structure
- Parsing Keep a Changelog format
- Parsing conventional commits
- Parsing GitHub releases
- Change type mapping
- Validation strategies

## Progressive Disclosure

The skill uses progressive disclosure to manage context efficiently:

1. **SKILL.md** (~1,800 words) - Always loaded when skill triggers
2. **references/** - Loaded by agent as needed for specific tasks

This keeps the initial context load small while providing comprehensive documentation when required.

## Validation

Shipyard includes built-in validation:

```bash
# Validate configuration and consignments
shipyard validate

# Preview changes before applying
shipyard version --preview

# Check current status
shipyard status
```

## Example Usage

Agents will generate customized scripts based on patterns from references/workflows.md:

**User**: "Set up Shipyard for my monorepo with api and sdk packages"

**Agent**: Reads SKILL.md + references/workflows.md, then generates:
```bash
shipyard init
# ... customized initialization steps ...
```

**User**: "Convert my CHANGELOG.md to Shipyard format"

**Agent**: Reads references/history-conversion.md, then generates appropriate parsing code based on the changelog format.

## Contributing

To improve this skill:

1. Keep SKILL.md lean (< 2,000 words)
2. Move detailed content to references/
3. Focus on patterns and concepts, not exact scripts
4. Use imperative/infinitive form in documentation
5. Include specific trigger phrases in description
6. Document the "why" and "when", not just the "how"

## Version

Version: 1.0.0

## License

Same as Shipyard CLI
