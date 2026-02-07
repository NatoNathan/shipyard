package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/consignment"
	shipyarderrors "github.com/NatoNathan/shipyard/internal/errors"
	"github.com/NatoNathan/shipyard/internal/git"
	"github.com/NatoNathan/shipyard/internal/graph"
	"github.com/NatoNathan/shipyard/internal/template"
	"github.com/NatoNathan/shipyard/internal/ui"
	"github.com/NatoNathan/shipyard/internal/version"
	"github.com/NatoNathan/shipyard/pkg/semver"
	"github.com/spf13/cobra"
)

// SnapshotCommandOptions holds options for the snapshot command
type SnapshotCommandOptions struct {
	Preview  bool
	NoCommit bool
	NoTag    bool
	Packages []string
	Verbose  bool
	JSON     bool
	Quiet    bool
}

// SnapshotOutput is the JSON output structure for snapshot command
type SnapshotOutput struct {
	Packages []SnapshotPackageOutput `json:"packages"`
}

// SnapshotPackageOutput represents a single package in snapshot JSON output
type SnapshotPackageOutput struct {
	Name       string `json:"name"`
	OldVersion string `json:"oldVersion"`
	NewVersion string `json:"newVersion"`
	Timestamp  string `json:"timestamp"`
	Tag        string `json:"tag,omitempty"`
}

// NewSnapshotCommand creates the snapshot subcommand
func NewSnapshotCommand() *cobra.Command {
	opts := &SnapshotCommandOptions{}

	cmd := &cobra.Command{
		Use:                   "snapshot [-p package]... [--preview] [--no-commit] [--no-tag]",
		DisableFlagsInUseLine: true,
		Aliases:               []string{"snap"},
		Short:   "Take a navigational reading of the current state",
		Long: `Create a timestamped snapshot pre-release version.
Snapshots are independent of the stage-based pre-release system.

Uses timestamps (YYYYMMDD-HHMMSS) for unique, chronologically ordered builds.
Does not affect .shipyard/prerelease.yml or stage tracking.`,
		Example: `  # Create snapshot
  shipyard version snapshot

  # Preview snapshot
  shipyard version snapshot --preview

  # Snapshot without git operations (for CI builds)
  shipyard version snapshot --no-commit --no-tag`,
		RunE: func(cmd *cobra.Command, args []string) error {
			globalFlags := GetGlobalFlags(cmd)
			opts.JSON = globalFlags.JSON
			opts.Quiet = globalFlags.Quiet
			if globalFlags.Verbose {
				opts.Verbose = true
			}
			return runSnapshot(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Preview, "preview", false, "Show changes without applying them")
	cmd.Flags().BoolVar(&opts.NoCommit, "no-commit", false, "Skip creating git commit")
	cmd.Flags().BoolVar(&opts.NoTag, "no-tag", false, "Skip creating git tags")
	cmd.Flags().StringSliceVarP(&opts.Packages, "package", "p", []string{}, "Filter to specific packages")

	RegisterPackageCompletions(cmd, "package")

	return cmd
}

func runSnapshot(opts *SnapshotCommandOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	return runSnapshotWithDir(cwd, opts, time.Now().UTC())
}

func runSnapshotWithDir(projectPath string, opts *SnapshotCommandOptions, now time.Time) error {
	if opts.Preview && !opts.Quiet && !opts.JSON {
		fmt.Println()
		fmt.Println(ui.InfoMessage("Preview Mode (no changes will be applied)"))
		fmt.Println()
	}

	// 1. Load configuration
	cfg, err := config.LoadFromDir(projectPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// 2. Read consignments
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

	if len(consignments) == 0 {
		return shipyarderrors.NewExitCodeError(2, "no pending consignments found")
	}

	// 3. Calculate target versions
	depGraph, err := graph.BuildGraph(cfg)
	if err != nil {
		return fmt.Errorf("failed to build dependency graph: %w", err)
	}

	currentVersions, err := ReadAllCurrentVersions(projectPath, cfg)
	if err != nil {
		return err
	}

	// Use base versions for propagation
	baseVersions := make(map[string]semver.Version)
	for k, v := range currentVersions {
		baseVersions[k] = v.BaseVersion()
	}

	propagator, err := version.NewPropagator(depGraph)
	if err != nil {
		return fmt.Errorf("failed to create propagator: %w", err)
	}
	versionBumps, err := propagator.Propagate(baseVersions, consignments)
	if err != nil {
		return fmt.Errorf("failed to calculate version bumps: %w", err)
	}

	// 4. Generate timestamp
	timestamp := now.Format("20060102-150405")

	// 5. Build snapshot versions and tags
	renderer := template.NewTemplateRenderer()

	snapshotTemplate := cfg.PreRelease.SnapshotTagTemplate
	if snapshotTemplate == "" {
		snapshotTemplate = "v{{.Version}}-snapshot.{{.Timestamp}}"
	}

	type snapshotResult struct {
		pkg           string
		oldVersion    semver.Version
		newVersion    semver.Version
		tagName       string
		targetVersion string
	}
	var results []snapshotResult

	snapshotKeys := make([]string, 0, len(versionBumps))
	for k := range versionBumps {
		snapshotKeys = append(snapshotKeys, k)
	}
	sort.Strings(snapshotKeys)
	for _, pkgName := range snapshotKeys {
		bump := versionBumps[pkgName]
		targetVersion := bump.NewVersion.String()

		// Build snapshot version
		preReleaseID := "snapshot." + timestamp
		newVersion := bump.NewVersion.WithPreRelease(preReleaseID)

		// Render tag
		tagCtx := map[string]interface{}{
			"Version":   targetVersion,
			"Timestamp": timestamp,
			"Package":   pkgName,
		}
		tagName, err := renderer.Render(snapshotTemplate, tagCtx)
		if err != nil {
			return fmt.Errorf("failed to render snapshot tag template for %s: %w", pkgName, err)
		}

		results = append(results, snapshotResult{
			pkg:           pkgName,
			oldVersion:    currentVersions[pkgName],
			newVersion:    newVersion,
			tagName:       tagName,
			targetVersion: targetVersion,
		})
	}

	// Preview mode
	if opts.Preview {
		if opts.JSON {
			output := SnapshotOutput{}
			for _, r := range results {
				output.Packages = append(output.Packages, SnapshotPackageOutput{
					Name:       r.pkg,
					OldVersion: r.oldVersion.String(),
					NewVersion: r.newVersion.String(),
					Timestamp:  timestamp,
					Tag:        r.tagName,
				})
			}
			return PrintJSON(os.Stdout, output)
		}
		if !opts.Quiet {
			fmt.Println(ui.Header("\U0001F4E6", "Preview: Snapshot version"))
			fmt.Println()
			var previewRows [][]string
			for _, r := range results {
				previewRows = append(previewRows, []string{
					r.pkg,
					r.oldVersion.String(),
					r.newVersion.String(),
					r.targetVersion,
				})
			}
			fmt.Println(ui.Table([]string{"Package", "Current", "Snapshot", "Target"}, previewRows))
			fmt.Println()
			fmt.Println(ui.InfoMessage("Preview mode: no changes made"))
			fmt.Println()
		}
		return nil
	}

	// 6. Update ecosystem version files
	for _, r := range results {
		pkg, ok := cfg.GetPackage(r.pkg)
		if !ok {
			return fmt.Errorf("package %s not found in configuration", r.pkg)
		}
		pkgPath := filepath.Join(projectPath, pkg.Path)
		handler, err := GetEcosystemHandler(pkg, pkgPath)
		if err != nil {
			return err
		}
		if err := handler.UpdateVersion(r.newVersion); err != nil {
			return fmt.Errorf("failed to update version for %s: %w", r.pkg, err)
		}
	}

	if !opts.Quiet && !opts.JSON {
		fmt.Println(ui.Header("\U0001F4E6", "Creating snapshot version"))
		fmt.Println()
		var execRows [][]string
		for _, r := range results {
			execRows = append(execRows, []string{
				r.pkg,
				r.oldVersion.String(),
				r.newVersion.String(),
			})
		}
		fmt.Println(ui.Table([]string{"Package", "Current", "Snapshot"}, execRows))
	}

	// 7. Git operations — NO state file changes for snapshots
	if !opts.NoCommit {
		changedPackages := make(map[string]bool)
		for _, r := range results {
			changedPackages[r.pkg] = true
		}
		filesToStage, err := CollectVersionFiles(projectPath, cfg, changedPackages)
		if err != nil {
			return err
		}

		if err := git.StageFiles(projectPath, filesToStage); err != nil {
			return fmt.Errorf("failed to stage files: %w", err)
		}

		commitMsg := "chore: snapshot"
		for _, r := range results {
			commitMsg += fmt.Sprintf(" %s v%s", r.pkg, r.newVersion)
		}

		if err := git.CreateCommit(projectPath, commitMsg); err != nil {
			return fmt.Errorf("failed to create commit: %w", err)
		}
		if !opts.Quiet && !opts.JSON {
			fmt.Println(ui.SuccessMessage(fmt.Sprintf("Created commit: \"%s\"", commitMsg)))
		}
	} else {
		if !opts.Quiet && !opts.JSON {
			fmt.Println(ui.SuccessMessage("Updated version files"))
			fmt.Println(ui.Dimmed("Skipped git commit (--no-commit)"))
		}
	}

	// 8. Create tags
	if !opts.NoCommit && !opts.NoTag {
		for _, r := range results {
			if err := git.CreateLightweightTag(projectPath, r.tagName); err != nil {
				return fmt.Errorf("failed to create tag %s: %w", r.tagName, err)
			}
			if !opts.Quiet && !opts.JSON {
				fmt.Println(ui.SuccessMessage(fmt.Sprintf("Created tag: %s", r.tagName)))
			}
		}
	} else if opts.NoTag {
		if !opts.Quiet && !opts.JSON {
			fmt.Println(ui.Dimmed("Skipped git tags (--no-tag)"))
		}
	}

	// JSON output at end
	if opts.JSON {
		output := SnapshotOutput{}
		for _, r := range results {
			output.Packages = append(output.Packages, SnapshotPackageOutput{
				Name:       r.pkg,
				OldVersion: r.oldVersion.String(),
				NewVersion: r.newVersion.String(),
				Timestamp:  timestamp,
				Tag:        r.tagName,
			})
		}
		return PrintJSON(os.Stdout, output)
	}

	return nil
}
