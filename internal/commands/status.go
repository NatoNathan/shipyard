package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/consignment"
	"github.com/NatoNathan/shipyard/internal/errors"
	"github.com/NatoNathan/shipyard/internal/graph"
	"github.com/NatoNathan/shipyard/internal/ui"
	"github.com/NatoNathan/shipyard/internal/version"
	"github.com/spf13/cobra"
)

// StatusOptions holds the options for the status command
type StatusOptions struct {
	Packages []string
	Output   string
	Quiet    bool
	Verbose  bool
}

// NewStatusCommand creates the status command
func NewStatusCommand() *cobra.Command {
	opts := &StatusOptions{}

	cmd := &cobra.Command{
		Use:                   "status [-p package]... [-o {table|json}]",
		DisableFlagsInUseLine: true,
		Aliases:               []string{"ls", "list"},
		Short:   "Check cargo and chart your course",
		Long: `Review pending cargo and see which ports of call (versions) await. Shows all
loaded consignments grouped by vessel with calculated destination coordinates.`,
		Example: `  # Show pending changes
  shipyard status

  # Filter by package
  shipyard status --package core

  # Output as JSON
  shipyard status --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check for global --json flag if local --output flag wasn't explicitly set
			if !cmd.Flags().Changed("output") {
				// Check parent for global --json flag
				if parent := cmd.Parent(); parent != nil {
					if jsonFlag, err := parent.PersistentFlags().GetBool("json"); err == nil && jsonFlag {
						opts.Output = "json"
					}
				}
			}
			return runStatus(opts)
		},
	}

	cmd.Flags().StringSliceVarP(&opts.Packages, "package", "p", nil, "Filter by package name(s)")
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "table", "Output format (table, json)")
	cmd.Flags().BoolVarP(&opts.Quiet, "quiet", "q", false, "Minimal output")
	cmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false, "Verbose output with timestamps and metadata")

	// Register package name completion
	RegisterPackageCompletions(cmd, "package")

	return cmd
}

// runStatus executes the status command
func runStatus(opts *StatusOptions) error {
	// Check if shipyard is initialized
	shipyardDir := ".shipyard"
	if _, err := os.Stat(shipyardDir); os.IsNotExist(err) {
		return errors.ErrNotInitialized
	}

	// Load configuration
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	cfg, err := config.LoadFromDir(cwd)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Read all pending consignments
	consignmentsDir := filepath.Join(shipyardDir, "consignments")
	consignments, err := readAllConsignments(consignmentsDir)
	if err != nil {
		return fmt.Errorf("failed to read consignments: %w", err)
	}

	// Filter by packages if specified
	if len(opts.Packages) > 0 {
		consignments = filterConsignmentsByPackages(consignments, opts.Packages)
	}

	// Check if there are any consignments
	if len(consignments) == 0 {
		fmt.Println(ui.InfoMessage("No pending consignments"))
		return nil
	}

	// Calculate version bumps with propagation
	versionBumps, err := calculateVersionBumpsForStatus(cfg, cwd, consignments)
	if err != nil {
		return fmt.Errorf("failed to calculate version bumps: %w", err)
	}

	// Group consignments by package
	grouped := groupConsignmentsByPackage(consignments)

	// Output based on format
	switch opts.Output {
	case "json":
		return outputJSONWithBumps(grouped, versionBumps, opts)
	default:
		return outputTableWithBumps(grouped, versionBumps, opts)
	}
}

// calculateVersionBumpsForStatus calculates version bumps including propagation
func calculateVersionBumpsForStatus(cfg *config.Config, projectPath string, consignments []*consignment.Consignment) (map[string]version.VersionBump, error) {
	// Build dependency graph
	depGraph, err := graph.BuildGraph(cfg)
	if err != nil {
		return nil, err
	}

	// Read current versions
	currentVersions, err := ReadAllCurrentVersions(projectPath, cfg)
	if err != nil {
		return nil, err
	}

	// Calculate bumps with propagation
	propagator, err := version.NewPropagator(depGraph)
	if err != nil {
		return nil, err
	}

	return propagator.Propagate(currentVersions, consignments)
}

// readAllConsignments reads all consignment files from a directory
func readAllConsignments(dir string) ([]*consignment.Consignment, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var consignments []*consignment.Consignment
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		c, err := consignment.ReadConsignment(path)
		if err != nil {
			// Skip invalid consignments
			continue
		}

		consignments = append(consignments, c)
	}

	return consignments, nil
}

// filterConsignmentsByPackages filters consignments to only those affecting specified packages
func filterConsignmentsByPackages(consignments []*consignment.Consignment, packages []string) []*consignment.Consignment {
	packageSet := make(map[string]bool)
	for _, pkg := range packages {
		packageSet[pkg] = true
	}

	var filtered []*consignment.Consignment
	for _, c := range consignments {
		for _, pkg := range c.Packages {
			if packageSet[pkg] {
				filtered = append(filtered, c)
				break
			}
		}
	}

	return filtered
}

// groupConsignmentsByPackage groups consignments by package
func groupConsignmentsByPackage(consignments []*consignment.Consignment) map[string][]*consignment.Consignment {
	grouped := make(map[string][]*consignment.Consignment)

	for _, c := range consignments {
		for _, pkg := range c.Packages {
			grouped[pkg] = append(grouped[pkg], c)
		}
	}

	return grouped
}

// outputJSONWithBumps outputs status in JSON format with calculated version bumps
func outputJSONWithBumps(grouped map[string][]*consignment.Consignment, versionBumps map[string]version.VersionBump, opts *StatusOptions) error {
	// Build JSON structure
	output := make(map[string]interface{})

	// Include all packages that have bumps (direct or propagated)
	jsonKeys := make([]string, 0, len(versionBumps))
	for k := range versionBumps {
		jsonKeys = append(jsonKeys, k)
	}
	sort.Strings(jsonKeys)
	for _, pkg := range jsonKeys {
		bump := versionBumps[pkg]
		pkgData := make(map[string]interface{})

		// Get consignments for this package (may be empty for propagated bumps)
		consignments := grouped[pkg]
		pkgData["count"] = len(consignments)
		pkgData["bump"] = bump.ChangeType
		pkgData["source"] = bump.Source
		pkgData["oldVersion"] = bump.OldVersion.String()
		pkgData["newVersion"] = bump.NewVersion.String()

		// Include consignment details if verbose
		if opts.Verbose && len(consignments) > 0 {
			var details []map[string]interface{}
			for _, c := range consignments {
				details = append(details, map[string]interface{}{
					"id":       c.ID,
					"type":     c.ChangeType,
					"summary":  c.Summary,
					"metadata": c.Metadata,
				})
			}
			pkgData["consignments"] = details
		}

		output[pkg] = pkgData
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

// outputTableWithBumps outputs status in table format with calculated version bumps
func outputTableWithBumps(grouped map[string][]*consignment.Consignment, versionBumps map[string]version.VersionBump, opts *StatusOptions) error {
	tableKeys := make([]string, 0, len(versionBumps))
	for k := range versionBumps {
		tableKeys = append(tableKeys, k)
	}
	sort.Strings(tableKeys)

	if opts.Quiet {
		// Quiet mode: just package names and bump types
		for _, pkg := range tableKeys {
			bump := versionBumps[pkg]
			fmt.Printf("%s: %s\n", pkg, bump.ChangeType)
		}
		return nil
	}

	// Normal mode: render table
	fmt.Println(ui.Header("\U0001F4E6", "Pending consignments"))
	fmt.Println()

	var rows [][]string
	for _, pkg := range tableKeys {
		bump := versionBumps[pkg]
		consignments := grouped[pkg]
		rows = append(rows, []string{
			pkg,
			bump.OldVersion.String(),
			bump.NewVersion.String(),
			ui.ChangeTypeBadge(string(bump.ChangeType)),
			string(bump.Source),
			strconv.Itoa(len(consignments)),
		})
	}

	fmt.Println(ui.Table(
		[]string{"Package", "Current", "Next", "Bump", "Source", "Changes"},
		rows,
	))

	// Verbose mode: show consignment details per package
	if opts.Verbose {
		for _, pkg := range tableKeys {
			consignments := grouped[pkg]
			if len(consignments) == 0 {
				continue
			}
			fmt.Println()
			fmt.Println(ui.Section(pkg))
			for _, c := range consignments {
				fmt.Println(ui.KeyValue("ID", c.ID))
				fmt.Println(ui.KeyValue("Type", string(c.ChangeType)))
				fmt.Println(ui.KeyValue("Summary", c.Summary))
			}
		}
	}
	fmt.Println()

	return nil
}

