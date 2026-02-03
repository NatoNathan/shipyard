package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/github"
	"github.com/NatoNathan/shipyard/internal/history"
	"github.com/NatoNathan/shipyard/internal/template"
	"github.com/NatoNathan/shipyard/pkg/semver"
	"github.com/spf13/cobra"
)

// ReleaseOptions holds options for the release command
type ReleaseOptions struct {
	Package     string
	Draft       bool
	Prerelease  bool
	Tag         string
	JSON        bool   // Output in JSON format
	Quiet       bool   // Suppress output
}

// NewReleaseCommand creates the release command
func NewReleaseCommand() *cobra.Command {
	opts := &ReleaseOptions{}

	cmd := &cobra.Command{
		Use:   "release",
		Short: "Publish release to GitHub",
		Long: `Publish a version release to GitHub. Creates a GitHub release using an existing
git tag. The tag must already exist locally and be pushed to the remote.

Run 'shipyard version' first to create version tags, then push them with 'git push --tags'.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Extract global flags
			globalFlags := GetGlobalFlags(cmd)
			opts.JSON = globalFlags.JSON
			opts.Quiet = globalFlags.Quiet
			return runRelease(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Package, "package", "p", "", "Package to release (required for multi-package repos)")
	cmd.Flags().BoolVar(&opts.Draft, "draft", false, "Create as draft release")
	cmd.Flags().BoolVar(&opts.Prerelease, "prerelease", false, "Mark as prerelease")
	cmd.Flags().StringVar(&opts.Tag, "tag", "", "Use specific tag instead of latest for package")

	// Register package name completion
	RegisterPackageCompletions(cmd, "package")

	return cmd
}

// runRelease executes the release command
func runRelease(opts *ReleaseOptions) error {
	// Get current directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Load configuration
	cfg, err := config.LoadFromDir(cwd)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Verify GitHub configuration
	if cfg.GitHub.Owner == "" || cfg.GitHub.Repo == "" {
		return fmt.Errorf("GitHub not configured in .shipyard.yaml (set github.owner and github.repo)")
	}

	// Verify GITHUB_TOKEN
	if os.Getenv("GITHUB_TOKEN") == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable not set")
	}

	// Determine package to release
	if len(cfg.Packages) > 1 && opts.Package == "" {
		return fmt.Errorf("--package is required for multi-package repositories")
	}

	// Auto-detect package for single-package repos
	if len(cfg.Packages) == 1 && opts.Package == "" {
		opts.Package = cfg.Packages[0].Name
	}

	// Read history to find latest entry for package
	historyPath := filepath.Join(cwd, ".shipyard", "history.json")
	entries, err := history.ReadHistory(historyPath)
	if err != nil {
		return fmt.Errorf("failed to read history: %w", err)
	}

	// Filter by package
	entries = history.FilterByPackage(entries, opts.Package)
	if len(entries) == 0 {
		return fmt.Errorf("no releases found for package %s", opts.Package)
	}

	// Use specific tag if provided, otherwise use latest
	var selectedEntry history.Entry
	if opts.Tag != "" {
		// Find entry by tag
		found := false
		for _, entry := range entries {
			if entry.Tag == opts.Tag {
				selectedEntry = entry
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("tag %s not found in history", opts.Tag)
		}
	} else {
		// Sort by timestamp and use latest
		entries = history.SortByTimestamp(entries, true) // newest first
		selectedEntry = entries[0]
	}

	// Parse version
	version, err := semver.Parse(selectedEntry.Version)
	if err != nil {
		return fmt.Errorf("failed to parse version %s: %w", selectedEntry.Version, err)
	}

	// Generate release notes from history entry
	releaseNotes, err := template.RenderReleaseNotes([]history.Entry{selectedEntry})
	if err != nil {
		return fmt.Errorf("failed to generate release notes: %w", err)
	}

	// Create release publisher
	publisher := github.NewReleasePublisher(cwd, cfg)

	// Publish release
	ctx := context.Background()
	if err := publisher.PublishRelease(ctx, opts.Package, version, selectedEntry.Tag, releaseNotes, opts.Draft); err != nil {
		return err
	}

	// Report success
	releaseURL := fmt.Sprintf("https://github.com/%s/%s/releases/tag/%s", cfg.GitHub.Owner, cfg.GitHub.Repo, selectedEntry.Tag)

	if opts.JSON {
		// JSON output
		jsonData := map[string]interface{}{
			"success": true,
			"package": opts.Package,
			"version": version.String(),
			"tag":     selectedEntry.Tag,
			"url":     releaseURL,
		}
		return PrintJSON(os.Stdout, jsonData)
	}

	if !opts.Quiet {
		fmt.Printf("✓ Release published successfully\n")
		fmt.Printf("  Package: %s\n", opts.Package)
		fmt.Printf("  Version: %s\n", version)
		fmt.Printf("  Tag: %s\n", selectedEntry.Tag)
		fmt.Printf("  URL: %s\n", releaseURL)
	}

	return nil
}
