package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/NatoNathan/shipyard/internal/changelog"
	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/consignment"
	"github.com/NatoNathan/shipyard/internal/ecosystem"
	"github.com/NatoNathan/shipyard/internal/git"
	"github.com/NatoNathan/shipyard/internal/graph"
	"github.com/NatoNathan/shipyard/internal/history"
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
		Use:   "version",
		Short: "Sail to the next port",
		Long: `Set sail with your cargo and reach the next version port. Navigates the fleet
through calculated routes, updates ship's logs, plants harbor markers (tags),
and archives the voyage in history.

The voyage: Load pending cargo → Chart course with dependency-aware navigation →
Update fleet coordinates → Record in ship's logs → Mark harbors with buoys →
Archive journey in captain's log.

Examples:
  # Set sail for all vessels
  shipyard version

  # Preview the route without sailing
  shipyard version --preview

  # Sail specific vessels only
  shipyard version --package core --package api

  # Navigate but don't record the voyage
  shipyard version --no-commit

  # Sail and record, but don't plant harbor markers
  shipyard version --no-tag
`,
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
	currentVersions := make(map[string]semver.Version)
	for _, pkg := range cfg.Packages {
		pkgPath := filepath.Join(projectPath, pkg.Path)
		var handler interface {
			ReadVersion() (semver.Version, error)
			UpdateVersion(semver.Version) error
		}

		switch pkg.Ecosystem {
		case config.EcosystemGo:
			// Check for tag-only mode via versionFiles
			if pkg.IsTagOnly() {
				handler = ecosystem.NewGoEcosystemWithOptions(pkgPath, &ecosystem.GoEcosystemOptions{TagOnly: true})
			} else {
				handler = ecosystem.NewGoEcosystem(pkgPath)
			}
		case config.EcosystemNPM:
			handler = ecosystem.NewNPMEcosystem(pkgPath)
		case config.EcosystemPython:
			handler = ecosystem.NewPythonEcosystem(pkgPath)
		case config.EcosystemHelm:
			handler = ecosystem.NewHelmEcosystem(pkgPath)
		case config.EcosystemCargo:
			handler = ecosystem.NewCargoEcosystem(pkgPath)
		case config.EcosystemDeno:
			handler = ecosystem.NewDenoEcosystem(pkgPath)
		default:
			return fmt.Errorf("unsupported ecosystem: %s", pkg.Ecosystem)
		}

		currentVer, err := handler.ReadVersion()
		if err != nil {
			return fmt.Errorf("failed to read version for %s: %w", pkg.Name, err)
		}
		currentVersions[pkg.Name] = currentVer
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

	// 6. Apply version bumps to files (unless preview)
	if !opts.Preview {
		for _, pkg := range cfg.Packages {
			bump, hasBump := versionBumps[pkg.Name]
			if !hasBump {
				continue
			}

			// Get the ecosystem handler
			pkgPath := filepath.Join(projectPath, pkg.Path)
			var handler interface {
				UpdateVersion(semver.Version) error
			}

			switch pkg.Ecosystem {
			case config.EcosystemGo:
				// Check for tag-only mode via versionFiles
				if pkg.IsTagOnly() {
					handler = ecosystem.NewGoEcosystemWithOptions(pkgPath, &ecosystem.GoEcosystemOptions{TagOnly: true})
				} else {
					handler = ecosystem.NewGoEcosystem(pkgPath)
				}
			case config.EcosystemNPM:
				handler = ecosystem.NewNPMEcosystem(pkgPath)
			case config.EcosystemPython:
				handler = ecosystem.NewPythonEcosystem(pkgPath)
			case config.EcosystemHelm:
				handler = ecosystem.NewHelmEcosystem(pkgPath)
			case config.EcosystemCargo:
				handler = ecosystem.NewCargoEcosystem(pkgPath)
			case config.EcosystemDeno:
				handler = ecosystem.NewDenoEcosystem(pkgPath)
			default:
				return fmt.Errorf("unsupported ecosystem: %s", pkg.Ecosystem)
			}

			// Update version files
			if err := handler.UpdateVersion(bump.NewVersion); err != nil {
				return fmt.Errorf("failed to update version for %s: %w", pkg.Name, err)
			}

			if opts.Verbose {
				fmt.Printf("Updated %s: %s -> %s\n", pkg.Name, bump.OldVersion, bump.NewVersion)
			}
		}
	}

	// 3. Generate tag names (needed for history entries)
	var packageTags map[string]changelog.PackageTag
	if !opts.Preview {
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

		var err error
		packageTags, err = generator.GenerateAllPackageTags(consignments, versions, tagTemplateSource)
		if err != nil {
			return fmt.Errorf("failed to generate tags: %w", err)
		}
	}

	// 4. Archive consignments to history with version context (unless preview)
	if !opts.Preview {
		historyPath := filepath.Join(projectPath, ".shipyard", "history.json")

		// Build history entries with version context
		var historyEntries []history.Entry
		for _, pkg := range cfg.Packages {
			bump, hasBump := versionBumps[pkg.Name]
			if !hasBump {
				continue
			}

			// Filter consignments for this package
			pkgConsignments := filterConsignmentsForPackage(consignments, pkg.Name)
			if len(pkgConsignments) == 0 {
				continue
			}

			// Convert to history.Consignment format
			historyConsignments := make([]history.Consignment, len(pkgConsignments))
			for i, c := range pkgConsignments {
				historyConsignments[i] = history.Consignment{
					ID:         c.ID,
					Summary:    c.Summary,
					ChangeType: string(c.ChangeType),
					Metadata:   c.Metadata,
				}
			}

			// Get tag name for this package
			tagName := ""
			if tag, exists := packageTags[pkg.Name]; exists {
				tagName = tag.Name
			}

			// Create entry with version context
			entry := history.Entry{
				Version:      bump.NewVersion.String(),
				Package:      pkg.Name,
				Tag:          tagName,
				Timestamp:    time.Now(),
				Consignments: historyConsignments,
			}
			historyEntries = append(historyEntries, entry)
		}

		// Archive with proper structure
		if err := history.AppendToHistory(historyPath, historyEntries); err != nil {
			return fmt.Errorf("failed to archive consignments: %w", err)
		}

		if opts.Verbose {
			fmt.Printf("Archived %d history entry/entries to history\n", len(historyEntries))
		}
	}

	// 5. Generate changelogs (unless preview) - Must happen AFTER archiving so current version is in history
	if !opts.Preview {
		// Read complete history to regenerate changelogs (now includes newly archived entries)
		historyPath := filepath.Join(projectPath, ".shipyard", "history.json")
		allEntries, err := history.ReadHistory(historyPath)
		if err != nil {
			return fmt.Errorf("failed to read history for changelog generation: %w", err)
		}

		for _, pkg := range cfg.Packages {
			_, hasBump := versionBumps[pkg.Name]
			if !hasBump {
				continue
			}

			// Filter history for this package
			pkgEntries := history.FilterByPackage(allEntries, pkg.Name)
			if len(pkgEntries) == 0 {
				continue
			}

			// Determine template source from config
			templateSource := "changelog" // Auto-selects builtin changelog template
			if cfg.Templates.Changelog.Source != "" {
				templateSource = cfg.Templates.Changelog.Source
			}

			// Generate full changelog from all history (multi-version)
			changelogContent, err := template.RenderReleaseNotesWithTemplate(pkgEntries, templateSource)
			if err != nil {
				return fmt.Errorf("failed to generate changelog for %s: %w", pkg.Name, err)
			}

			// Write (replace) changelog to file
			changelogPath := filepath.Join(projectPath, pkg.Path, "CHANGELOG.md")
			if err := os.WriteFile(changelogPath, []byte(changelogContent), 0644); err != nil {
				return fmt.Errorf("failed to write changelog for %s: %w", pkg.Name, err)
			}

			if opts.Verbose {
				fmt.Printf("Generated changelog for %s\n", pkg.Name)
			}
		}
	}

	// 6. Delete processed consignment files (unless preview)
	if !opts.Preview {
		for _, c := range consignments {
			consignmentPath := filepath.Join(consignmentsDir, c.ID+".md")
			if err := os.Remove(consignmentPath); err != nil {
				return fmt.Errorf("failed to delete consignment %s: %w", c.ID, err)
			}
		}

		if opts.Verbose {
			fmt.Printf("Deleted %d consignment file(s)\n", len(consignments))
		}
	}

	// 7. Create git operations (skip if preview/no-commit/no-tag)
	if !opts.Preview {
		// Collect files to stage (version files + changelogs)
		filesToStage := []string{}

		for _, pkg := range cfg.Packages {
			if _, hasBump := versionBumps[pkg.Name]; hasBump {
				pkgPath := filepath.Join(projectPath, pkg.Path)

				// Get ecosystem handler to find version files
				var handler interface{ GetVersionFiles() []string }
				switch pkg.Ecosystem {
				case config.EcosystemGo:
					if pkg.IsTagOnly() {
						handler = ecosystem.NewGoEcosystemWithOptions(pkgPath, &ecosystem.GoEcosystemOptions{TagOnly: true})
					} else {
						handler = ecosystem.NewGoEcosystem(pkgPath)
					}
				case config.EcosystemNPM:
					handler = ecosystem.NewNPMEcosystem(pkgPath)
				case config.EcosystemPython:
					handler = ecosystem.NewPythonEcosystem(pkgPath)
				case config.EcosystemHelm:
					handler = ecosystem.NewHelmEcosystem(pkgPath)
				case config.EcosystemCargo:
					handler = ecosystem.NewCargoEcosystem(pkgPath)
				case config.EcosystemDeno:
					handler = ecosystem.NewDenoEcosystem(pkgPath)
				default:
					return fmt.Errorf("unsupported ecosystem: %s", pkg.Ecosystem)
				}

				// Add version files
				if handler != nil {
					versionFiles := handler.GetVersionFiles()
					for _, vf := range versionFiles {
						filesToStage = append(filesToStage, filepath.Join(pkgPath, vf))
					}
				}

				// Add changelog if it exists
				changelogPath := filepath.Join(pkgPath, "CHANGELOG.md")
				if _, err := os.Stat(changelogPath); err == nil {
					filesToStage = append(filesToStage, changelogPath)
				}
			}
		}

		// Add history.json to staging
		historyPath := filepath.Join(projectPath, ".shipyard", "history.json")
		if _, err := os.Stat(historyPath); err == nil {
			filesToStage = append(filesToStage, historyPath)
		}

		// Add deleted consignment files to staging
		for _, c := range consignments {
			consignmentPath := filepath.Join(projectPath, ".shipyard", "consignments", c.ID+".md")
			filesToStage = append(filesToStage, consignmentPath)
		}

		// Create generator for commit message and tags
		generator := changelog.NewChangelogGenerator()
		generator.SetBaseDir(projectPath)

		// Stage and commit
		if !opts.NoCommit && len(filesToStage) > 0 {
			if err := git.StageFiles(projectPath, filesToStage); err != nil {
				return fmt.Errorf("failed to stage files: %w", err)
			}

			// Generate commit message

			commitTemplateSource := "builtin:default"
			if cfg.Templates.CommitMessage != nil && cfg.Templates.CommitMessage.Source != "" {
				commitTemplateSource = cfg.Templates.CommitMessage.Source
			}

			// Convert version.VersionBump to changelog.VersionBump
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
				fmt.Printf("Created commit with %d file(s)\n", len(filesToStage))
			}
		}

		// Create tags (using already-generated packageTags from step 3)
		if !opts.NoTag && len(packageTags) > 0 {
			// Separate annotated and lightweight tags
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
						fmt.Printf("Creating annotated tag for %s: %s\n", pkgName, tag.Name)
					} else {
						fmt.Printf("Creating lightweight tag for %s: %s\n", pkgName, tag.Name)
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
				fmt.Printf("Created %d tag(s)\n", len(packageTags))
			}
		}
	}

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

	for pkgName, bump := range versionBumps {
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
