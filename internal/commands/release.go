package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/github"
	"github.com/NatoNathan/shipyard/internal/history"
	"github.com/NatoNathan/shipyard/internal/template"
	"github.com/NatoNathan/shipyard/internal/ui"
	"github.com/spf13/cobra"
)

// ReleaseOptions holds options for the release command.
type ReleaseOptions struct {
	Version string // --version: The version/tag to release (e.g. "v1.2.3")
	Package string // --package: Filter release notes to a package
	Draft   bool   // --draft: Create as a draft release
	Notes   string // --notes: Custom release notes body (overrides generated)
}

// NewReleaseCommand creates the release command.
func NewReleaseCommand() *cobra.Command {
	opts := &ReleaseOptions{}

	cmd := &cobra.Command{
		Use:   "release",
		Short: "Launch your vessel into harbor",
		Long: `Create a GitHub release for a tagged version. Generates release notes from
the captain's log (history) and publishes them as a GitHub release.

Requires GitHub configuration in shipyard.yaml:
  github:
    owner: your-org
    repo: your-repo
    token: "env:GITHUB_TOKEN"

Examples:
  # Release the latest tagged version
  shipyard release --version v1.2.3

  # Create a draft release for review
  shipyard release --version v1.2.3 --draft

  # Release with custom notes
  shipyard release --version v1.2.3 --notes "Hotfix for auth bug"

  # Release notes scoped to a specific package
  shipyard release --version v1.2.3 --package core
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRelease(opts)
		},
	}

	cmd.Flags().StringVar(&opts.Version, "version", "", "Version tag to release (required, e.g. v1.2.3)")
	cmd.Flags().StringVarP(&opts.Package, "package", "p", "", "Filter release notes to a specific package")
	cmd.Flags().BoolVar(&opts.Draft, "draft", false, "Create as a draft release")
	cmd.Flags().StringVar(&opts.Notes, "notes", "", "Custom release notes (overrides generated notes)")

	_ = cmd.MarkFlagRequired("version")

	RegisterPackageCompletions(cmd, "package")

	return cmd
}

// runRelease executes the release command.
func runRelease(opts *ReleaseOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	return runReleaseWithDir(cwd, opts)
}

// runReleaseWithDir executes the release command in a specific directory.
// Exported logic so the version command can call it after tagging.
func runReleaseWithDir(projectPath string, opts *ReleaseOptions) error {
	// Load configuration
	cfg, err := config.LoadFromDir(projectPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate GitHub config
	if cfg.GitHub.Owner == "" || cfg.GitHub.Repo == "" {
		return fmt.Errorf("GitHub owner and repo must be configured in shipyard.yaml under the github section")
	}

	// Create GitHub client
	client, err := github.NewClient(cfg.GitHub.Owner, cfg.GitHub.Repo, cfg.GitHub.Token)
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %w", err)
	}

	// Generate release notes
	notes := opts.Notes
	if notes == "" {
		generated, err := generateNotesForVersion(projectPath, opts.Version, opts.Package)
		if err != nil {
			fmt.Println(ui.WarningMessage(fmt.Sprintf("Could not generate release notes: %v", err)))
			notes = fmt.Sprintf("Release %s", opts.Version)
		} else {
			notes = generated
		}
	}

	// Build release title
	title := opts.Version
	if opts.Package != "" {
		title = fmt.Sprintf("%s - %s", opts.Package, opts.Version)
	}

	// Check for existing release and clean up if present
	existing, err := client.GetReleaseByTag(opts.Version)
	if err != nil {
		return fmt.Errorf("failed to check for existing release: %w", err)
	}
	if existing != nil {
		fmt.Println(ui.InfoMessage(fmt.Sprintf("Replacing existing release for %s", opts.Version)))
		if err := client.DeleteRelease(existing.ID); err != nil {
			return fmt.Errorf("failed to delete existing release: %w", err)
		}
	}

	// Create the release
	status := "release"
	if opts.Draft {
		status = "draft release"
	}
	fmt.Println(ui.InfoMessage(fmt.Sprintf("Creating %s for %s...", status, opts.Version)))

	release, err := client.CreateRelease(opts.Version, title, notes, opts.Draft)
	if err != nil {
		return fmt.Errorf("failed to create GitHub release: %w", err)
	}

	fmt.Println()
	fmt.Println(ui.SuccessMessage(fmt.Sprintf("GitHub %s created: %s", status, opts.Version)))
	fmt.Println(ui.KeyValue("URL", release.HTMLURL))
	if opts.Draft {
		fmt.Println(ui.KeyValue("Status", "draft (publish from GitHub when ready)"))
	}
	fmt.Println()

	return nil
}

// generateNotesForVersion generates release notes from history for a given version.
func generateNotesForVersion(projectPath, version, packageFilter string) (string, error) {
	historyPath := filepath.Join(projectPath, ".shipyard", "history.json")
	entries, err := history.ReadHistory(historyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read history: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Sprintf("Release %s", version), nil
	}

	// Strip the "v" prefix for history matching since history stores bare versions
	versionBare := strings.TrimPrefix(version, "v")

	// Filter by version (try both with and without v prefix)
	filtered := history.FilterByVersion(entries, version)
	if len(filtered) == 0 {
		filtered = history.FilterByVersion(entries, versionBare)
	}

	// Filter by package if specified
	if packageFilter != "" {
		filtered = history.FilterByPackage(filtered, packageFilter)
	}

	// If no entries match this specific version, use all entries as context
	if len(filtered) == 0 {
		return fmt.Sprintf("Release %s", version), nil
	}

	notes, err := template.RenderReleaseNotes(filtered)
	if err != nil {
		return "", fmt.Errorf("failed to render release notes: %w", err)
	}

	return notes, nil
}
