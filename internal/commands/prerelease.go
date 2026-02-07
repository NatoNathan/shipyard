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

// PrereleaseCommandOptions holds options for the prerelease command
type PrereleaseCommandOptions struct {
	Preview  bool
	NoCommit bool
	NoTag    bool
	Packages []string
	Verbose  bool
	JSON     bool
	Quiet    bool
}

// PrereleaseOutput is the JSON output structure for prerelease command
type PrereleaseOutput struct {
	Packages []PrereleasePackageOutput `json:"packages"`
}

// PrereleasePackageOutput represents a single package in prerelease JSON output
type PrereleasePackageOutput struct {
	Name       string `json:"name"`
	OldVersion string `json:"oldVersion"`
	NewVersion string `json:"newVersion"`
	Stage      string `json:"stage"`
	Counter    int    `json:"counter"`
	Tag        string `json:"tag,omitempty"`
}

// NewPrereleaseCommand creates the prerelease subcommand
func NewPrereleaseCommand() *cobra.Command {
	opts := &PrereleaseCommandOptions{}

	cmd := &cobra.Command{
		Use:   "prerelease",
		Short: "Chart test waters before the main voyage",
		Long: `Create or increment a pre-release version at the current stage.
Creates pre-release versions for testing changes before creating a stable release.

The stage is determined from .shipyard/prerelease.yml state file:
  - First pre-release starts at the lowest-order stage
  - Subsequent runs increment the counter (e.g., alpha.1 → alpha.2)
  - Use 'shipyard version promote' to advance stages

Examples:
  # Create pre-release at current stage
  shipyard version prerelease

  # Preview without changes
  shipyard version prerelease --preview

  # Pre-release specific packages
  shipyard version prerelease --package core --package api
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			globalFlags := GetGlobalFlags(cmd)
			opts.JSON = globalFlags.JSON
			opts.Quiet = globalFlags.Quiet
			if globalFlags.Verbose {
				opts.Verbose = true
			}
			return runPrerelease(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Preview, "preview", false, "Show changes without applying them")
	cmd.Flags().BoolVar(&opts.NoCommit, "no-commit", false, "Skip creating git commit")
	cmd.Flags().BoolVar(&opts.NoTag, "no-tag", false, "Skip creating git tags")
	cmd.Flags().StringSliceVarP(&opts.Packages, "package", "p", []string{}, "Filter to specific packages")

	RegisterPackageCompletions(cmd, "package")

	return cmd
}

func runPrerelease(opts *PrereleaseCommandOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	return runPrereleaseWithDir(cwd, opts)
}

func runPrereleaseWithDir(projectPath string, opts *PrereleaseCommandOptions) error {
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

	// Validate pre-release stages exist
	if len(cfg.PreRelease.Stages) == 0 {
		return fmt.Errorf("no pre-release stages defined in configuration")
	}
	if err := cfg.PreRelease.Validate(); err != nil {
		return fmt.Errorf("invalid pre-release configuration: %w", err)
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

	// 3. Build graph and calculate version bumps (target versions)
	depGraph, err := graph.BuildGraph(cfg)
	if err != nil {
		return fmt.Errorf("failed to build dependency graph: %w", err)
	}

	currentVersions, err := ReadAllCurrentVersions(projectPath, cfg)
	if err != nil {
		return err
	}

	// Use base versions (without pre-release) for propagation
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

	// 4. Read state from .shipyard/prerelease.yml
	statePath := filepath.Join(projectPath, ".shipyard", "prerelease.yml")
	state, err := prerelease.ReadState(statePath)
	if err != nil {
		return fmt.Errorf("failed to read prerelease state: %w", err)
	}

	// 5. For each package with bumps, determine stage and counter
	renderer := template.NewTemplateRenderer()
	type prereleaseResult struct {
		pkg            string
		oldVersion     semver.Version
		newVersion     semver.Version
		stage          config.StageConfig
		counter        int
		tagName        string
		targetVersion  string
	}
	var results []prereleaseResult

	pkgNames := make([]string, 0, len(versionBumps))
	for k := range versionBumps {
		pkgNames = append(pkgNames, k)
	}
	sort.Strings(pkgNames)
	for _, pkgName := range pkgNames {
		bump := versionBumps[pkgName]
		targetVersion := bump.NewVersion.String()
		pkgState, hasState := state.Packages[pkgName]

		var stage config.StageConfig
		var counter int

		if !hasState {
			// No state: start at lowest-order stage, counter=1
			s, ok := cfg.PreRelease.GetLowestOrderStage()
			if !ok {
				return fmt.Errorf("no pre-release stages configured")
			}
			stage = s
			counter = 1
		} else if pkgState.TargetVersion != targetVersion {
			// Target changed: warn, reset counter
			fmt.Printf("⚠ Warning: Target version changed from %s to %s for %s (consignments modified)\n",
				pkgState.TargetVersion, targetVersion, pkgName)
			s, ok := cfg.PreRelease.GetStageByName(pkgState.Stage)
			if !ok {
				return fmt.Errorf("stage '%s' not found in configuration", pkgState.Stage)
			}
			stage = s
			counter = 1
		} else {
			// Same target: increment counter
			s, ok := cfg.PreRelease.GetStageByName(pkgState.Stage)
			if !ok {
				return fmt.Errorf("stage '%s' not found in configuration", pkgState.Stage)
			}
			stage = s
			counter = pkgState.Counter + 1
		}

		// Build pre-release version
		preReleaseID := fmt.Sprintf("%s.%d", stage.Name, counter)
		newVersion := bump.NewVersion.WithPreRelease(preReleaseID)

		// Render tag from stage's tagTemplate
		tagTemplate := stage.TagTemplate
		if tagTemplate == "" {
			tagTemplate = "v{{.Version}}-{{.Stage}}.{{.Counter}}"
		}
		tagCtx := map[string]interface{}{
			"Version": bump.NewVersion.String(),
			"Counter": counter,
			"Package": pkgName,
			"Stage":   stage.Name,
		}
		tagName, err := renderer.Render(tagTemplate, tagCtx)
		if err != nil {
			return fmt.Errorf("failed to render tag template for %s: %w", pkgName, err)
		}

		results = append(results, prereleaseResult{
			pkg:           pkgName,
			oldVersion:    currentVersions[pkgName],
			newVersion:    newVersion,
			stage:         stage,
			counter:       counter,
			tagName:       tagName,
			targetVersion: targetVersion,
		})
	}

	// Preview mode
	if opts.Preview {
		if opts.JSON {
			output := PrereleaseOutput{}
			for _, r := range results {
				output.Packages = append(output.Packages, PrereleasePackageOutput{
					Name:       r.pkg,
					OldVersion: r.oldVersion.String(),
					NewVersion: r.newVersion.String(),
					Stage:      r.stage.Name,
					Counter:    r.counter,
					Tag:        r.tagName,
				})
			}
			return PrintJSON(os.Stdout, output)
		}
		if !opts.Quiet {
			fmt.Println("📦 Preview: Pre-release version changes")
			for _, r := range results {
				fmt.Printf("  - %s: %s → %s (%s)\n", r.pkg, r.oldVersion, r.newVersion, r.stage.Name)
				fmt.Printf("    Target version: %s\n", r.targetVersion)
			}
			fmt.Println()
			fmt.Println(ui.InfoMessage("Preview mode: no changes made"))
			fmt.Println()
		}
		return nil
	}

	// 6. Update ecosystem version files
	if !opts.Quiet && !opts.JSON {
		fmt.Println("📦 Creating pre-release versions...")
	}
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
		if !opts.Quiet && !opts.JSON {
			_, hasOldState := state.Packages[r.pkg]
			if !hasOldState {
				fmt.Printf("  - %s: %s → %s (%s, first pre-release)\n", r.pkg, r.oldVersion, r.newVersion, r.stage.Name)
			} else {
				fmt.Printf("  - %s: %s → %s (%s)\n", r.pkg, r.oldVersion, r.newVersion, r.stage.Name)
			}
		}
	}

	// 7. Update state
	for _, r := range results {
		state.Packages[r.pkg] = prerelease.PackageState{
			Stage:         r.stage.Name,
			Counter:       r.counter,
			TargetVersion: r.targetVersion,
		}
	}
	if err := prerelease.WriteState(statePath, state); err != nil {
		return fmt.Errorf("failed to write prerelease state: %w", err)
	}

	// 8. Git operations
	if !opts.NoCommit {
		// Collect files to stage
		changedPackages := make(map[string]bool)
		for _, r := range results {
			changedPackages[r.pkg] = true
		}
		filesToStage, err := CollectVersionFiles(projectPath, cfg, changedPackages)
		if err != nil {
			return err
		}
		// Add state file
		filesToStage = append(filesToStage, statePath)

		if err := git.StageFiles(projectPath, filesToStage); err != nil {
			return fmt.Errorf("failed to stage files: %w", err)
		}

		// Build commit message
		commitMsg := "chore: pre-release"
		for _, r := range results {
			commitMsg += fmt.Sprintf(" %s v%s", r.pkg, r.newVersion)
		}

		if err := git.CreateCommit(projectPath, commitMsg); err != nil {
			return fmt.Errorf("failed to create commit: %w", err)
		}
		if !opts.Quiet && !opts.JSON {
			fmt.Printf("✓ Created commit: \"%s\"\n", commitMsg)
		}
	} else {
		if !opts.Quiet && !opts.JSON {
			fmt.Println("✓ Updated version files")
			fmt.Println("⊘ Skipped git commit (--no-commit)")
		}
	}

	// 9. Create tags
	if !opts.NoCommit && !opts.NoTag {
		for _, r := range results {
			if err := git.CreateLightweightTag(projectPath, r.tagName); err != nil {
				return fmt.Errorf("failed to create tag %s: %w", r.tagName, err)
			}
			if !opts.Quiet && !opts.JSON {
				fmt.Printf("✓ Created tag: %s\n", r.tagName)
			}
		}
	} else if opts.NoTag {
		if !opts.Quiet && !opts.JSON {
			fmt.Println("⊘ Skipped git tags (--no-tag)")
		}
	}

	// JSON output at end
	if opts.JSON {
		output := PrereleaseOutput{}
		for _, r := range results {
			output.Packages = append(output.Packages, PrereleasePackageOutput{
				Name:       r.pkg,
				OldVersion: r.oldVersion.String(),
				NewVersion: r.newVersion.String(),
				Stage:      r.stage.Name,
				Counter:    r.counter,
				Tag:        r.tagName,
			})
		}
		return PrintJSON(os.Stdout, output)
	}

	return nil
}
