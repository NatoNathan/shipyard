package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/consignment"
	shipyarderrors "github.com/NatoNathan/shipyard/internal/errors"
	"github.com/NatoNathan/shipyard/internal/git"
	"github.com/NatoNathan/shipyard/internal/graph"
	"github.com/NatoNathan/shipyard/internal/prerelease"
	"github.com/NatoNathan/shipyard/internal/template"
	"github.com/NatoNathan/shipyard/internal/ui"
	"github.com/NatoNathan/shipyard/internal/version"
	"github.com/NatoNathan/shipyard/pkg/semver"
	"github.com/spf13/cobra"
)

// PromoteCommandOptions holds options for the promote command
type PromoteCommandOptions struct {
	Preview  bool
	NoCommit bool
	NoTag    bool
	Packages []string
	Verbose  bool
	JSON     bool
	Quiet    bool
}

// PromoteOutput is the JSON output structure for promote command
type PromoteOutput struct {
	Packages []PromotePackageOutput `json:"packages"`
}

// PromotePackageOutput represents a single package in promote JSON output
type PromotePackageOutput struct {
	Name       string `json:"name"`
	OldVersion string `json:"oldVersion"`
	NewVersion string `json:"newVersion"`
	OldStage   string `json:"oldStage"`
	NewStage   string `json:"newStage"`
	Counter    int    `json:"counter"`
	Tag        string `json:"tag,omitempty"`
}

// NewPromoteCommand creates the promote subcommand
func NewPromoteCommand() *cobra.Command {
	opts := &PromoteCommandOptions{}

	cmd := &cobra.Command{
		Use:                   "promote [-p package]... [--preview] [--no-commit] [--no-tag]",
		DisableFlagsInUseLine: true,
		Aliases:               []string{"advance"},
		Short:   "Advance through the harbor channel",
		Long: `Promote a pre-release to the next stage in order.
Advances pre-releases through configured stages (e.g., alpha → beta → rc).

At the highest stage, returns an error—use 'shipyard version' to promote to stable.`,
		Example: `  # Promote to next stage
  shipyard version promote

  # Preview promotion
  shipyard version promote --preview

  # Promote specific packages
  shipyard version promote --package core`,
		RunE: func(cmd *cobra.Command, args []string) error {
			globalFlags := GetGlobalFlags(cmd)
			opts.JSON = globalFlags.JSON
			opts.Quiet = globalFlags.Quiet
			if globalFlags.Verbose {
				opts.Verbose = true
			}
			return runPromote(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Preview, "preview", false, "Show changes without applying them")
	cmd.Flags().BoolVar(&opts.NoCommit, "no-commit", false, "Skip creating git commit")
	cmd.Flags().BoolVar(&opts.NoTag, "no-tag", false, "Skip creating git tags")
	cmd.Flags().StringSliceVarP(&opts.Packages, "package", "p", []string{}, "Filter to specific packages")

	RegisterPackageCompletions(cmd, "package")

	return cmd
}

func runPromote(opts *PromoteCommandOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	return runPromoteWithDir(cwd, opts)
}

func runPromoteWithDir(projectPath string, opts *PromoteCommandOptions) error {
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

	if len(cfg.PreRelease.Stages) == 0 {
		return fmt.Errorf("no pre-release stages defined in configuration")
	}
	if err := cfg.PreRelease.Validate(); err != nil {
		return fmt.Errorf("invalid pre-release configuration: %w", err)
	}

	// 2. Read state file — error exit 3 if no state exists
	statePath := filepath.Join(projectPath, ".shipyard", "prerelease.yml")
	if !prerelease.Exists(statePath) {
		return shipyarderrors.NewExitCodeError(3, "no pre-release state exists (use 'shipyard version prerelease' first)")
	}

	state, err := prerelease.ReadState(statePath)
	if err != nil {
		return fmt.Errorf("failed to read prerelease state: %w", err)
	}

	if len(state.Packages) == 0 {
		return shipyarderrors.NewExitCodeError(3, "no pre-release state exists (use 'shipyard version prerelease' first)")
	}

	// 3. Read consignments and calculate target versions
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

	depGraph, err := graph.BuildGraph(cfg)
	if err != nil {
		return fmt.Errorf("failed to build dependency graph: %w", err)
	}

	currentVersions, err := ReadAllCurrentVersions(projectPath, cfg)
	if err != nil {
		return err
	}

	propagator, err := version.NewPropagator(depGraph)
	if err != nil {
		return fmt.Errorf("failed to create propagator: %w", err)
	}

	// Use base versions for propagation (strip pre-release)
	baseVersions := make(map[string]semver.Version)
	for k, v := range currentVersions {
		baseVersions[k] = v.BaseVersion()
	}

	versionBumps, err := propagator.Propagate(baseVersions, consignments)
	if err != nil {
		return fmt.Errorf("failed to calculate version bumps: %w", err)
	}

	// 4. For each package with state, determine next stage
	renderer := template.NewTemplateRenderer()
	type promoteResult struct {
		pkg           string
		oldVersion    semver.Version
		newVersion    semver.Version
		oldStage      string
		newStage      config.StageConfig
		counter       int
		tagName       string
		targetVersion string
	}
	var results []promoteResult

	packagesToPromote := state.Packages
	if len(opts.Packages) > 0 {
		filtered := make(map[string]prerelease.PackageState)
		for _, name := range opts.Packages {
			if ps, ok := state.Packages[name]; ok {
				filtered[name] = ps
			}
		}
		packagesToPromote = filtered
	}

	promoteKeys := make([]string, 0, len(packagesToPromote))
	for k := range packagesToPromote {
		promoteKeys = append(promoteKeys, k)
	}
	sort.Strings(promoteKeys)
	for _, pkgName := range promoteKeys {
		pkgState := packagesToPromote[pkgName]
		// Check if at highest stage — error exit 2
		if cfg.PreRelease.IsHighestStage(pkgState.Stage) {
			return shipyarderrors.NewExitCodeError(2,
				fmt.Sprintf("already at highest pre-release stage '%s' for %s (use 'shipyard version' for stable release)", pkgState.Stage, pkgName))
		}

		// Find next stage
		nextStage, ok := cfg.PreRelease.GetNextStage(pkgState.Stage)
		if !ok {
			return fmt.Errorf("failed to find next stage after '%s'", pkgState.Stage)
		}

		// Determine target version
		targetVersion := pkgState.TargetVersion
		if bump, hasBump := versionBumps[pkgName]; hasBump {
			newTarget := bump.NewVersion.String()
			if newTarget != pkgState.TargetVersion {
				fmt.Println(ui.WarningMessage(fmt.Sprintf("Target version changed from %s to %s for %s (consignments modified)",
					pkgState.TargetVersion, newTarget, pkgName)))
				targetVersion = newTarget
			}
		}

		targetVer, err := semver.Parse(targetVersion)
		if err != nil {
			return fmt.Errorf("failed to parse target version for %s: %w", pkgName, err)
		}

		// Reset counter to 1 for new stage
		counter := 1
		preReleaseID := fmt.Sprintf("%s.%d", nextStage.Name, counter)
		newVersion := targetVer.WithPreRelease(preReleaseID)

		// Render tag
		tagTemplate := nextStage.TagTemplate
		if tagTemplate == "" {
			tagTemplate = "v{{.Version}}-{{.Stage}}.{{.Counter}}"
		}
		tagCtx := map[string]interface{}{
			"Version": targetVersion,
			"Counter": counter,
			"Package": pkgName,
			"Stage":   nextStage.Name,
		}
		tagName, err := renderer.Render(tagTemplate, tagCtx)
		if err != nil {
			return fmt.Errorf("failed to render tag template for %s: %w", pkgName, err)
		}

		results = append(results, promoteResult{
			pkg:           pkgName,
			oldVersion:    currentVersions[pkgName],
			newVersion:    newVersion,
			oldStage:      pkgState.Stage,
			newStage:      nextStage,
			counter:       counter,
			tagName:       tagName,
			targetVersion: targetVersion,
		})
	}

	// Preview mode
	if opts.Preview {
		if opts.JSON {
			output := PromoteOutput{}
			for _, r := range results {
				output.Packages = append(output.Packages, PromotePackageOutput{
					Name:       r.pkg,
					OldVersion: r.oldVersion.String(),
					NewVersion: r.newVersion.String(),
					OldStage:   r.oldStage,
					NewStage:   r.newStage.Name,
					Counter:    r.counter,
					Tag:        r.tagName,
				})
			}
			return PrintJSON(os.Stdout, output)
		}
		if !opts.Quiet {
			fmt.Println(ui.Header("\U0001F4E6", "Preview: Promote to next stage"))
			fmt.Println()
			var previewRows [][]string
			for _, r := range results {
				previewRows = append(previewRows, []string{
					r.pkg,
					r.oldVersion.String(),
					r.newVersion.String(),
					r.oldStage,
					r.newStage.Name,
					r.targetVersion,
				})
			}
			fmt.Println(ui.Table([]string{"Package", "Current", "Promoted", "From", "To", "Target"}, previewRows))
			fmt.Println()
			fmt.Println(ui.InfoMessage("Preview mode: no changes made"))
			fmt.Println()
		}
		return nil
	}

	// 5. Update ecosystem version files
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
		fmt.Println(ui.Header("\U0001F4E6", "Promoting to next stage"))
		fmt.Println()
		var execRows [][]string
		for _, r := range results {
			execRows = append(execRows, []string{
				r.pkg,
				r.oldVersion.String(),
				r.newVersion.String(),
				r.oldStage,
				r.newStage.Name,
			})
		}
		fmt.Println(ui.Table([]string{"Package", "Current", "Promoted", "From", "To"}, execRows))
	}

	// 6. Update state
	for _, r := range results {
		state.Packages[r.pkg] = prerelease.PackageState{
			Stage:         r.newStage.Name,
			Counter:       r.counter,
			TargetVersion: r.targetVersion,
		}
	}
	if err := prerelease.WriteState(statePath, state); err != nil {
		return fmt.Errorf("failed to write prerelease state: %w", err)
	}
	if !opts.Quiet && !opts.JSON {
		fmt.Println(ui.SuccessMessage("Updated .shipyard/prerelease.yml"))
	}

	// 7. Git operations
	if !opts.NoCommit {
		changedPackages := make(map[string]bool)
		for _, r := range results {
			changedPackages[r.pkg] = true
		}
		filesToStage, err := CollectVersionFiles(projectPath, cfg, changedPackages)
		if err != nil {
			return err
		}
		filesToStage = append(filesToStage, statePath)

		if err := git.StageFiles(projectPath, filesToStage); err != nil {
			return fmt.Errorf("failed to stage files: %w", err)
		}

		commitMsg := "chore: promote"
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
		output := PromoteOutput{}
		for _, r := range results {
			output.Packages = append(output.Packages, PromotePackageOutput{
				Name:       r.pkg,
				OldVersion: r.oldVersion.String(),
				NewVersion: r.newVersion.String(),
				OldStage:   r.oldStage,
				NewStage:   r.newStage.Name,
				Counter:    r.counter,
				Tag:        r.tagName,
			})
		}
		return PrintJSON(os.Stdout, output)
	}

	return nil
}
