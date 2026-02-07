package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"time"

	"github.com/NatoNathan/shipyard/internal/changelog"
	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/consignment"
	"github.com/NatoNathan/shipyard/internal/git"
	"github.com/NatoNathan/shipyard/internal/graph"
	"github.com/NatoNathan/shipyard/internal/history"
	"github.com/NatoNathan/shipyard/internal/prerelease"
	"github.com/NatoNathan/shipyard/internal/template"
	"github.com/NatoNathan/shipyard/internal/ui"
	"github.com/NatoNathan/shipyard/internal/version"
	"github.com/NatoNathan/shipyard/pkg/semver"
	"github.com/spf13/cobra"
)

// VersionCommandOptions holds options for the version command
type VersionCommandOptions struct {
	Preview  bool     // --preview: Show changes without applying
	NoCommit bool     // --no-commit: Skip git commit
	NoTag    bool     // --no-tag: Skip git tag creation
	Packages []string // --package: Filter to specific packages
	Verbose  bool     // --verbose: Show detailed output
}

// NewVersionCommand creates the version command
func NewVersionCommand() *cobra.Command {
	opts := &VersionCommandOptions{}

	cmd := &cobra.Command{
		Use:                   "version [command] [-p package]... [--preview] [--no-commit] [--no-tag]",
		DisableFlagsInUseLine: true,
		Aliases:               []string{"bump", "sail"},
		Short:   "Sail to the next port",
		Long: `Set sail with your cargo and reach the next version port. Navigates the fleet
through calculated routes, updates ship's logs, plants harbor markers (tags),
and archives the voyage in history.

The voyage: Load pending cargo → Chart course with dependency-aware navigation →
Update fleet coordinates → Record in ship's logs → Mark harbors with buoys →
Archive journey in captain's log.`,
		Example: `  # Set sail for all vessels
  shipyard version

  # Preview the route without sailing
  shipyard version --preview

  # Sail specific vessels only
  shipyard version --package core --package api

  # Navigate but don't record the voyage
  shipyard version --no-commit

  # Sail and record, but don't plant harbor markers
  shipyard version --no-tag`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVersion(opts)
		},
	}

	// Add flags
	cmd.Flags().BoolVar(&opts.Preview, "preview", false, "Show changes without applying them")
	cmd.Flags().BoolVar(&opts.NoCommit, "no-commit", false, "Skip creating git commit")
	cmd.Flags().BoolVar(&opts.NoTag, "no-tag", false, "Skip creating git tags")
	cmd.Flags().StringSliceVarP(&opts.Packages, "package", "p", []string{}, "Filter to specific packages (can be specified multiple times)")
	cmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false, "Show detailed output")

	// Register package name completion
	RegisterPackageCompletions(cmd, "package")

	// Register subcommands
	cmd.AddCommand(NewPrereleaseCommand())
	cmd.AddCommand(NewPromoteCommand())
	cmd.AddCommand(NewSnapshotCommand())

	return cmd
}

// runVersion executes the version command logic in the current directory
func runVersion(opts *VersionCommandOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	return runVersionWithDir(cwd, opts)
}

// runVersionWithDir executes the version command logic in a specific directory
func runVersionWithDir(projectPath string, opts *VersionCommandOptions) error {
	// Phase 1: Validation and initialization
	if opts.Preview {
		fmt.Println()
		fmt.Println(ui.InfoMessage("Preview Mode (no changes will be applied)"))
		fmt.Println()
	}

	// 1. Load configuration
	cfg, err := config.LoadFromDir(projectPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// 2. Read pending consignments
	consignmentsDir := filepath.Join(projectPath, ".shipyard", "consignments")
	var consignments []*consignment.Consignment
	if len(opts.Packages) > 0 {
		consignments, err = consignment.ReadAllConsignmentsFiltered(consignmentsDir, opts.Packages)
	} else {
		consignments, err = consignment.ReadAllConsignments(consignmentsDir)
	}
	if err != nil {
		return fmt.Errorf("failed to read consignments: %w", err)
	}

	// If no consignments, nothing to do
	if len(consignments) == 0 {
		if opts.Verbose {
			fmt.Println()
			fmt.Println(ui.InfoMessage("No pending consignments found"))
			fmt.Println()
		}
		return nil
	}

	// 3. Build dependency graph
	depGraph, err := graph.BuildGraph(cfg)
	if err != nil {
		return fmt.Errorf("failed to build dependency graph: %w", err)
	}

	// 4. Read current versions for all packages
	currentVersions, err := ReadAllCurrentVersions(projectPath, cfg)
	if err != nil {
		return err
	}

	// 5. Calculate version bumps (with propagation)
	propagator, err := version.NewPropagator(depGraph)
	if err != nil {
		return fmt.Errorf("failed to create propagator: %w", err)
	}
	versionBumps, err := propagator.Propagate(currentVersions, consignments)
	if err != nil {
		return fmt.Errorf("failed to calculate version bumps: %w", err)
	}

	// Preview mode: Show what would change and exit
	if opts.Preview {
		displayPreview(versionBumps, consignments, cfg)
		return nil
	}

	// 6. Apply version bumps to files
	for _, pkg := range cfg.Packages {
		bump, hasBump := versionBumps[pkg.Name]
		if !hasBump {
			continue
		}

		pkgPath := filepath.Join(projectPath, pkg.Path)
		handler, err := GetEcosystemHandler(pkg, pkgPath)
		if err != nil {
			return err
		}

		if err := handler.UpdateVersion(bump.NewVersion); err != nil {
			return fmt.Errorf("failed to update version for %s: %w", pkg.Name, err)
		}

		if opts.Verbose {
			fmt.Println(ui.Dimmed(fmt.Sprintf("Updated %s: %s -> %s", pkg.Name, bump.OldVersion, bump.NewVersion)))
		}
	}

	// 7. Generate tag names (needed for history entries)
	generator := changelog.NewChangelogGenerator()
	generator.SetBaseDir(projectPath)

	tagTemplateSource := "builtin:default"
	if cfg.Templates.TagName != nil && cfg.Templates.TagName.Source != "" {
		tagTemplateSource = cfg.Templates.TagName.Source
	}

	versions := make(map[string]semver.Version)
	for pkgName, bump := range versionBumps {
		versions[pkgName] = bump.NewVersion
	}

	packageTags, err := generator.GenerateAllPackageTags(consignments, versions, tagTemplateSource)
	if err != nil {
		return fmt.Errorf("failed to generate tags: %w", err)
	}

	// 8. Archive consignments to history with version context
	historyPath := filepath.Join(projectPath, ".shipyard", "history.json")

	var historyEntries []history.Entry
	for _, pkg := range cfg.Packages {
		bump, hasBump := versionBumps[pkg.Name]
		if !hasBump {
			continue
		}

		pkgConsignments := filterConsignmentsForPackage(consignments, pkg.Name)
		if len(pkgConsignments) == 0 {
			continue
		}

		historyConsignments := make([]history.Consignment, len(pkgConsignments))
		for i, c := range pkgConsignments {
			historyConsignments[i] = history.Consignment{
				ID:         c.ID,
				Summary:    c.Summary,
				ChangeType: string(c.ChangeType),
				Metadata:   c.Metadata,
			}
		}

		tagName := ""
		if tag, exists := packageTags[pkg.Name]; exists {
			tagName = tag.Name
		}

		entry := history.Entry{
			Version:      bump.NewVersion.String(),
			Package:      pkg.Name,
			Tag:          tagName,
			Timestamp:    time.Now(),
			Consignments: historyConsignments,
		}
		historyEntries = append(historyEntries, entry)
	}

	if err := history.AppendToHistory(historyPath, historyEntries); err != nil {
		return fmt.Errorf("failed to archive consignments: %w", err)
	}

	if opts.Verbose {
		fmt.Println(ui.Dimmed(fmt.Sprintf("Archived %d history entry/entries to history", len(historyEntries))))
	}

	// 9. Generate changelogs (must happen AFTER archiving so current version is in history)
	allEntries, err := history.ReadHistory(historyPath)
	if err != nil {
		return fmt.Errorf("failed to read history for changelog generation: %w", err)
	}

	for _, pkg := range cfg.Packages {
		_, hasBump := versionBumps[pkg.Name]
		if !hasBump {
			continue
		}

		pkgEntries := history.FilterByPackage(allEntries, pkg.Name)
		if len(pkgEntries) == 0 {
			continue
		}

		templateSource := "changelog"
		if cfg.Templates.Changelog.Source != "" {
			templateSource = cfg.Templates.Changelog.Source
		}

		changelogContent, err := template.RenderReleaseNotesWithTemplate(pkgEntries, templateSource)
		if err != nil {
			return fmt.Errorf("failed to generate changelog for %s: %w", pkg.Name, err)
		}

		changelogPath := filepath.Join(projectPath, pkg.Path, "CHANGELOG.md")
		if err := os.WriteFile(changelogPath, []byte(changelogContent), 0644); err != nil {
			return fmt.Errorf("failed to write changelog for %s: %w", pkg.Name, err)
		}

		if opts.Verbose {
			fmt.Println(ui.Dimmed(fmt.Sprintf("Generated changelog for %s", pkg.Name)))
		}
	}

	// 10. Delete processed consignment files
	for _, c := range consignments {
		consignmentPath := filepath.Join(consignmentsDir, c.ID+".md")
		if err := os.Remove(consignmentPath); err != nil {
			return fmt.Errorf("failed to delete consignment %s: %w", c.ID, err)
		}
	}

	if opts.Verbose {
		fmt.Println(ui.Dimmed(fmt.Sprintf("Deleted %d consignment file(s)", len(consignments))))
	}

	// 11. Git operations (commit and tag)
	changedPackages := make(map[string]bool)
	for pkgName := range versionBumps {
		changedPackages[pkgName] = true
	}
	filesToStage, err := CollectVersionFiles(projectPath, cfg, changedPackages)
	if err != nil {
		return err
	}

	if _, err := os.Stat(historyPath); err == nil {
		filesToStage = append(filesToStage, historyPath)
	}

	for _, c := range consignments {
		consignmentPath := filepath.Join(projectPath, ".shipyard", "consignments", c.ID+".md")
		filesToStage = append(filesToStage, consignmentPath)
	}

	prereleaseStatePath := filepath.Join(projectPath, ".shipyard", "prerelease.yml")
	if prerelease.Exists(prereleaseStatePath) {
		if err := prerelease.DeleteState(prereleaseStatePath); err != nil {
			return fmt.Errorf("failed to delete prerelease state: %w", err)
		}
		filesToStage = append(filesToStage, prereleaseStatePath)
		if opts.Verbose {
			fmt.Println(ui.Dimmed("Deleted .shipyard/prerelease.yml"))
		}
	}

	if !opts.NoCommit && len(filesToStage) > 0 {
		if err := git.StageFiles(projectPath, filesToStage); err != nil {
			return fmt.Errorf("failed to stage files: %w", err)
		}

		commitTemplateSource := "builtin:default"
		if cfg.Templates.CommitMessage != nil && cfg.Templates.CommitMessage.Source != "" {
			commitTemplateSource = cfg.Templates.CommitMessage.Source
		}

		changelogBumps := make(map[string]changelog.VersionBump)
		for name, bump := range versionBumps {
			changelogBumps[name] = changelog.VersionBump{
				Package:    bump.Package,
				OldVersion: bump.OldVersion,
				NewVersion: bump.NewVersion,
				ChangeType: bump.ChangeType,
			}
		}

		commitMessage, err := generator.GenerateCommitMessage(consignments, changelogBumps, commitTemplateSource)
		if err != nil {
			return fmt.Errorf("failed to generate commit message: %w", err)
		}

		if err := git.CreateCommit(projectPath, commitMessage); err != nil {
			return fmt.Errorf("failed to create commit: %w", err)
		}

		if opts.Verbose {
			fmt.Println(ui.Dimmed(fmt.Sprintf("Created commit with %d file(s)", len(filesToStage))))
		}
	}

	if !opts.NoTag && len(packageTags) > 0 {
		annotatedTags := make(map[string]string)
		lightweightTags := []string{}
		for pkgName, tag := range packageTags {
			if tag.Message != "" {
				annotatedTags[tag.Name] = tag.Message
			} else {
				lightweightTags = append(lightweightTags, tag.Name)
			}

			if opts.Verbose {
				if tag.Message != "" {
					fmt.Println(ui.Dimmed(fmt.Sprintf("Creating annotated tag for %s: %s", pkgName, tag.Name)))
				} else {
					fmt.Println(ui.Dimmed(fmt.Sprintf("Creating lightweight tag for %s: %s", pkgName, tag.Name)))
				}
			}
		}

		if len(annotatedTags) > 0 {
			if err := git.CreateAnnotatedTags(projectPath, annotatedTags); err != nil {
				return fmt.Errorf("failed to create annotated tags: %w", err)
			}
		}

		if len(lightweightTags) > 0 {
			if err := git.CreateLightweightTags(projectPath, lightweightTags); err != nil {
				return fmt.Errorf("failed to create lightweight tags: %w", err)
			}
		}

		if opts.Verbose {
			fmt.Println(ui.Dimmed(fmt.Sprintf("Created %d tag(s)", len(packageTags))))
		}
	}

	// Success summary
	fmt.Println()
	fmt.Println(ui.SuccessMessage(fmt.Sprintf("Versioned %d package(s)", len(versionBumps))))
	var summaryRows [][]string
	for _, pkg := range cfg.Packages {
		if bump, ok := versionBumps[pkg.Name]; ok {
			summaryRows = append(summaryRows, []string{
				pkg.Name,
				bump.OldVersion.String(),
				bump.NewVersion.String(),
			})
		}
	}
	fmt.Println(ui.Table([]string{"Package", "Old Version", "New Version"}, summaryRows))

	return nil
}

// filterConsignmentsForPackage returns consignments that affect the given package
func filterConsignmentsForPackage(consignments []*consignment.Consignment, packageName string) []*consignment.Consignment {
	var filtered []*consignment.Consignment
	for _, c := range consignments {
		if slices.Contains(c.Packages, packageName) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}


// displayPreview shows what changes would be made without applying them
func displayPreview(versionBumps map[string]version.VersionBump, consignments []*consignment.Consignment, cfg *config.Config) {
	// Convert version bumps to PackageChange structs for preview display
	var changes []ui.PackageChange

	previewKeys := make([]string, 0, len(versionBumps))
	for k := range versionBumps {
		previewKeys = append(previewKeys, k)
	}
	sort.Strings(previewKeys)
	for _, pkgName := range previewKeys {
		bump := versionBumps[pkgName]
		// Get consignments for this package
		pkgConsignments := filterConsignmentsForPackage(consignments, pkgName)

		// Extract change summaries
		var changeSummaries []string
		for _, c := range pkgConsignments {
			changeSummaries = append(changeSummaries, c.Summary)
		}

		changes = append(changes, ui.PackageChange{
			Name:       pkgName,
			OldVersion: bump.OldVersion,
			NewVersion: bump.NewVersion,
			ChangeType: string(bump.ChangeType),
			Changes:    changeSummaries,
		})
	}

	// Display the preview
	preview := ui.RenderPreview(changes)
	fmt.Println(preview)
	fmt.Println()
	fmt.Println(ui.InfoMessage("Run without --preview to apply these changes"))
	fmt.Println()
}
