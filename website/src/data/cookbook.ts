export interface Recipe {
  id: string
  title: string
  description: string
  language: 'yaml' | 'bash' | 'markdown'
  code: string
}

export interface CookbookCategory {
  id: string
  title: string
  description: string
  recipes: Recipe[]
}

export const cookbookCategories: CookbookCategory[] = [
  {
    id: 'single-repo',
    title: 'Single Repository',
    description: 'Minimal configs for single-package repositories, one per ecosystem.',
    recipes: [
      {
        id: 'single-go',
        title: 'Go module',
        description: 'Single Go module using version.go or go.mod version comment.',
        language: 'yaml',
        code: `# .shipyard/shipyard.yaml
packages:
  - name: "app"
    path: "./"
    ecosystem: "go"
    dependencies: []

templates:
  changelog: "builtin:default"
  tagName: "builtin:default"   # produces v1.2.3
  releaseNotes: "builtin:default"

github:
  enabled: true
  owner: "your-org"
  repo: "your-repo"`,
      },
      {
        id: 'single-npm',
        title: 'NPM package',
        description: 'NPM package with custom metadata fields for author and issue tracking.',
        language: 'yaml',
        code: `# .shipyard/shipyard.yaml
metadata:
  fields:
    - name: "author"
      required: true
      type: "string"
    - name: "issue"
      required: false
      type: "string"

packages:
  - name: "my-package"
    path: "./"
    ecosystem: "npm"
    dependencies: []

templates:
  changelog: |
    # Changelog
    ## {{.Version}} — {{.Date | date "2006-01-02"}}
    {{range .Consignments}}
    ### {{.ChangeType | upper}}
    {{.Summary}}
    *By {{.Metadata.author}}{{if .Metadata.issue}} — {{.Metadata.issue}}{{end}}*
    {{end}}
  tagName: "v{{.Version}}"`,
      },
      {
        id: 'single-python',
        title: 'Python package',
        description: 'Python package with Keep a Changelog format and pip install snippet in release notes.',
        language: 'yaml',
        code: `# .shipyard/shipyard.yaml
packages:
  - name: "mylib"
    path: "./"
    ecosystem: "python"
    dependencies: []

templates:
  changelog: |
    # Changelog
    All notable changes to this project are documented here.
    The format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

    ## [{{.Version}}] - {{.Date | date "2006-01-02"}}
    {{range .Consignments}}
    ### {{.ChangeType | title}}
    - {{.Summary}}
    {{end}}
    [{{.Version}}]: https://github.com/your-org/mylib/compare/v{{.PreviousVersion}}...v{{.Version}}

  tagName: "v{{.Version}}"

  releaseNotes: |
    # Release {{.Version}}
    ## Installation
    \`\`\`bash
    pip install mylib=={{.Version}}
    \`\`\`
    ## Changes
    {{range .Consignments}}
    ### {{.ChangeType | upper}}
    {{.Summary}}
    {{end}}`,
      },
      {
        id: 'single-cargo',
        title: 'Cargo (Rust)',
        description: 'Rust crate with standard semver tagging.',
        language: 'yaml',
        code: `# .shipyard/shipyard.yaml
packages:
  - name: "my-crate"
    path: "./"
    ecosystem: "cargo"
    dependencies: []

templates:
  changelog: "builtin:default"
  tagName: "v{{.Version}}"
  releaseNotes: "builtin:default"

github:
  enabled: true
  owner: "your-org"
  repo: "your-crate"`,
      },
      {
        id: 'single-deno',
        title: 'Deno module',
        description: 'Deno module with version tracked in deno.json.',
        language: 'yaml',
        code: `# .shipyard/shipyard.yaml
packages:
  - name: "my-module"
    path: "./"
    ecosystem: "deno"
    dependencies: []

templates:
  changelog: "builtin:default"
  tagName: "v{{.Version}}"
  releaseNotes: |
    # {{.Package}} v{{.Version}}
    \`\`\`ts
    import { something } from "jsr:@your-scope/my-module@{{.Version}}";
    \`\`\`
    {{range .Consignments}}
    ## {{.ChangeType | upper}}
    {{.Summary}}
    {{end}}`,
      },
    ],
  },
  {
    id: 'monorepo',
    title: 'Monorepo',
    description: 'Multi-package setups with independent versioning, linked dependencies, and mixed ecosystems.',
    recipes: [
      {
        id: 'monorepo-independent',
        title: 'Independent packages',
        description: 'Multiple packages versioned independently with no dependency propagation.',
        language: 'yaml',
        code: `# .shipyard/shipyard.yaml
packages:
  - name: "web-app"
    path: "./apps/web"
    ecosystem: "npm"
    dependencies: []

  - name: "api-server"
    path: "./services/api"
    ecosystem: "go"
    dependencies: []

  - name: "worker"
    path: "./services/worker"
    ecosystem: "go"
    dependencies: []

  - name: "utils"
    path: "./packages/utils"
    ecosystem: "npm"
    dependencies: []

templates:
  changelog: "builtin:default"
  tagName: "{{.Package}}/v{{.Version}}"
  releaseNotes: |
    # {{.Package}} v{{.Version}}
    {{range .Consignments}}
    ## {{.ChangeType | upper}}
    {{.Summary}}
    {{end}}`,
      },
      {
        id: 'monorepo-linked',
        title: 'Linked dependencies',
        description: 'Version propagation: core minor bump cascades to dependents with configurable bump mapping.',
        language: 'yaml',
        code: `# .shipyard/shipyard.yaml
# Dependency graph: core → auth-lib → api-client → web-app
#                                                  → mobile-app
packages:
  - name: "core"
    path: "./packages/core"
    ecosystem: "go"
    dependencies: []

  - name: "auth-lib"
    path: "./packages/auth"
    ecosystem: "go"
    dependencies:
      - package: "core"
        strategy: "linked"     # identical bump mapping (default)

  - name: "api-client"
    path: "./clients/api"
    ecosystem: "npm"
    dependencies:
      - package: "core"
        strategy: "linked"
      - package: "auth-lib"
        strategy: "linked"

  - name: "web-app"
    path: "./apps/web"
    ecosystem: "npm"
    dependencies:
      - package: "api-client"
        strategy: "linked"
        bumpMapping:
          major: "patch"   # breaking upstream → patch here
          minor: "patch"
          patch: "patch"

  - name: "backend"
    path: "./services/backend"
    ecosystem: "go"
    dependencies:
      - package: "core"
        strategy: "fixed"   # versions independently

templates:
  changelog: "builtin:default"
  tagName: "{{.Package}}/v{{.Version}}"`,
      },
      {
        id: 'monorepo-mixed',
        title: 'Mixed ecosystems',
        description: 'Go core, Python and JS SDKs, Docker images — all linked across ecosystem boundaries.',
        language: 'yaml',
        code: `# .shipyard/shipyard.yaml
packages:
  - name: "core"
    path: "./packages/core"
    ecosystem: "go"
    dependencies: []

  - name: "python-sdk"
    path: "./sdks/python"
    ecosystem: "python"
    dependencies:
      - package: "core"
        strategy: "linked"

  - name: "js-sdk"
    path: "./sdks/javascript"
    ecosystem: "npm"
    dependencies:
      - package: "core"
        strategy: "linked"

  - name: "api-server"
    path: "./services/api"
    ecosystem: "go"
    dependencies:
      - package: "core"
        strategy: "linked"

  - name: "api-server-image"
    path: "./docker/api-server"
    ecosystem: "docker"
    dependencies:
      - package: "api-server"
        strategy: "linked"
    templates:
      tagName: "api-server-{{.Version}}"

templates:
  changelog: "builtin:default"
  tagName: "{{.Package}}/v{{.Version}}"`,
      },
    ],
  },
  {
    id: 'cicd',
    title: 'CI / CD',
    description: 'GitHub Actions and workflow patterns for automated releases and PR enforcement.',
    recipes: [
      {
        id: 'ci-github-release',
        title: 'GitHub Actions — release on tag',
        description: 'Trigger the Dagger release pipeline on v* tag push.',
        language: 'yaml',
        code: `# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write
  packages: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: dagger/dagger-for-github@v7
        with:
          version: "latest"

      - name: Run Dagger Release Pipeline
        env:
          GITHUB_TOKEN: \${{ secrets.GITHUB_TOKEN }}
          NPM_TOKEN: \${{ secrets.NPM_TOKEN }}
        run: |
          dagger call -m ./dagger release \\
            --source=. \\
            --version="\${{ github.ref_name }}" \\
            --github-token=env:GITHUB_TOKEN \\
            --npm-token=env:NPM_TOKEN \\
            --docker-registry=ghcr.io/your-org/your-repo \\
            --docker-token=env:GITHUB_TOKEN`,
      },
      {
        id: 'ci-require-consignment',
        title: 'GitHub Actions — require consignment in PR',
        description: 'Block merge if code changes land without a consignment file. Supports skip-consignment label.',
        language: 'yaml',
        code: `# .github/workflows/require-consignment.yml
name: Require Consignment

on:
  pull_request:
    branches: [main]
    types: [opened, synchronize, reopened, labeled, unlabeled]

jobs:
  check-consignment:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: "1.21"

      - name: Install Shipyard
        run: go install github.com/NatoNathan/shipyard/cmd/shipyard@latest

      - name: Check for skip label
        id: skip
        run: |
          SKIP=\${{ contains(github.event.pull_request.labels.*.name, 'skip-consignment') }}
          echo "skip=\$SKIP" >> "\$GITHUB_OUTPUT"

      - name: Check for consignments
        if: steps.skip.outputs.skip == 'false'
        run: |
          git fetch origin \${{ github.base_ref }}
          CHANGED=$(git diff --name-only origin/\${{ github.base_ref }}...HEAD)

          # Only check if non-doc files changed
          if ! echo "\$CHANGED" | grep -qvE '^(README|CONTRIBUTING|docs/|examples/|LICENSE|\.github/)'; then
            echo "✅ Only docs changed — no consignment required"
            exit 0
          fi

          CONSIGNMENTS=$(git diff --name-only --diff-filter=A origin/\${{ github.base_ref }}...HEAD \\
            | grep '^\.shipyard/consignments/.*\\.md\$' || true)

          if [ -z "\$CONSIGNMENTS" ]; then
            echo "❌ No consignment found. Run: shipyard add --summary '...' --bump patch"
            exit 1
          fi

          echo "✅ Found consignments:"
          echo "\$CONSIGNMENTS"`,
      },
      {
        id: 'ci-auto-version',
        title: 'GitHub Actions — auto version on merge',
        description: 'Run shipyard version automatically when consignment files are merged to main.',
        language: 'yaml',
        code: `# .github/workflows/auto-version.yml
name: Auto Version

on:
  push:
    branches: [main]
    paths:
      - ".shipyard/consignments/**"

permissions:
  contents: write

jobs:
  version:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/create-github-app-token@v1
        id: token
        with:
          app-id: \${{ vars.APP_ID }}
          private-key: \${{ secrets.APP_PRIVATE_KEY }}

      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
          token: \${{ steps.token.outputs.token }}

      - uses: actions/setup-go@v5
        with:
          go-version: "1.21"

      - name: Install Shipyard
        run: go install github.com/NatoNathan/shipyard/cmd/shipyard@latest

      - name: Bump versions
        run: |
          git config user.name "Release Bot"
          git config user.email "bot@example.com"
          shipyard version

      - name: Push tags
        run: git push --follow-tags`,
      },
    ],
  },
  {
    id: 'templates',
    title: 'Custom Templates',
    description: 'Go template snippets for changelogs, tag names, and release notes.',
    recipes: [
      {
        id: 'template-keepachangelog',
        title: 'Keep a Changelog format',
        description: 'Changelog output that follows the keepachangelog.com standard with comparison links.',
        language: 'yaml',
        code: `# .shipyard/shipyard.yaml (templates section)
templates:
  changelog: |
    # Changelog
    All notable changes to this project are documented here.
    The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

    ## [{{.Version}}] - {{.Date | date "2006-01-02"}}
    {{range .Consignments}}
    ### {{.ChangeType | title}}
    - {{.Summary}}
    {{end}}

    [{{.Version}}]: https://github.com/your-org/your-repo/compare/v{{.PreviousVersion}}...v{{.Version}}`,
      },
      {
        id: 'template-monorepo-tags',
        title: 'Monorepo tag strategies',
        description: 'Package-prefixed tags for monorepos, with per-package overrides.',
        language: 'yaml',
        code: `# .shipyard/shipyard.yaml
# Global default: package/vX.Y.Z
templates:
  tagName: "{{.Package}}/v{{.Version}}"

packages:
  - name: "core"
    path: "./packages/core"
    ecosystem: "go"
    # Uses global: core/v1.2.3

  - name: "web-app"
    path: "./apps/web"
    ecosystem: "npm"
    templates:
      tagName: "web/{{.Version}}"   # override: web/1.2.3

  - name: "api-client"
    path: "./clients/api"
    ecosystem: "npm"
    # Uses global: api-client/v1.2.3

# Available template variables:
#   .Package      — package name
#   .Version      — new version (e.g. "1.2.3")
#   .PrevVersion  — previous version
#   .Ecosystem    — go, npm, python, etc.`,
      },
      {
        id: 'template-detailed-release',
        title: 'Detailed release notes with metadata',
        description: 'Release notes that surface custom metadata fields like author, team, and issue links.',
        language: 'yaml',
        code: `# .shipyard/shipyard.yaml
metadata:
  fields:
    - name: "author"
      required: true
      type: "string"
    - name: "issue"
      required: false
      type: "string"
    - name: "breaking"
      required: false
      type: "boolean"

templates:
  releaseNotes: |
    # {{.Package}} v{{.Version}}
    Released: {{.Date | date "January 2, 2006"}}

    {{range .Consignments}}
    ## {{.ChangeType | upper}}{{if .Metadata.breaking}} ⚠️ BREAKING{{end}}

    {{.Summary}}

    *By {{.Metadata.author}}{{if .Metadata.issue}} — [{{.Metadata.issue}}](https://github.com/your-org/repo/issues/{{trimPrefix "#" .Metadata.issue}}){{end}}*
    {{end}}

    ---
    **Full changelog**: https://github.com/your-org/repo/blob/main/CHANGELOG.md`,
      },
    ],
  },
]
