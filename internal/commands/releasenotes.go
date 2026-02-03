package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/history"
	"github.com/NatoNathan/shipyard/internal/template"
	"github.com/spf13/cobra"
)

// ReleaseNotesOptions holds options for the release-notes command
type ReleaseNotesOptions struct {
	Package        string
	Output         string
	Version        string
	AllVersions    bool
	MetadataFilter []string
	Template       string
	JSON           bool     // Output in JSON format
	Quiet          bool     // Suppress output
}

// NewReleaseNotesCommand creates the release-notes command
func NewReleaseNotesCommand() *cobra.Command {
	opts := &ReleaseNotesOptions{}

	cmd := &cobra.Command{
		Use:   "release-notes",
		Short: "Tell the tale of your voyage",
		Long: `Recount the journey from the captain's log. Transforms version history into
tales of ports visited and cargo delivered. Filter by vessel or destination,
write to parchment (file) or speak aloud (stdout).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Extract global flags
			globalFlags := GetGlobalFlags(cmd)
			opts.JSON = globalFlags.JSON
			opts.Quiet = globalFlags.Quiet
			return runReleaseNotes(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Package, "package", "p", "", "Filter by package name (required for multi-package repos)")
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output file (default: stdout)")
	cmd.Flags().StringVar(&opts.Version, "version", "", "Generate notes for specific version")
	cmd.Flags().BoolVar(&opts.AllVersions, "all-versions", false, "Show complete history instead of just latest version")
	cmd.Flags().StringArrayVar(&opts.MetadataFilter, "filter", []string{}, "Filter by custom metadata (format: key=value, can be repeated)")
	cmd.Flags().StringVar(&opts.Template, "template", "", "Template to use (path or builtin name)")

	// Register package name completion
	RegisterPackageCompletions(cmd, "package")

	return cmd
}

// runReleaseNotes executes the release notes generation
func runReleaseNotes(opts *ReleaseNotesOptions) error {
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

	// Read history
	historyPath := filepath.Join(cwd, ".shipyard", "history.json")
	entries, err := history.ReadHistory(historyPath)
	if err != nil {
		return fmt.Errorf("failed to read history: %w", err)
	}

	// Check if history is empty
	if len(entries) == 0 {
		output := "No releases found in history\n"
		if opts.Output != "" {
			return os.WriteFile(opts.Output, []byte(output), 0644)
		}
		fmt.Print(output)
		return nil
	}

	// Require --package for multi-package repos
	if len(cfg.Packages) > 1 && opts.Package == "" {
		return fmt.Errorf("--package is required for multi-package repositories")
	}

	// Auto-detect package for single-package repos
	if len(cfg.Packages) == 1 && opts.Package == "" {
		opts.Package = cfg.Packages[0].Name
	}

	// Filter by package
	if opts.Package != "" {
		entries = history.FilterByPackage(entries, opts.Package)
	}

	// Filter by custom metadata (validate against config)
	for _, filter := range opts.MetadataFilter {
		parts := strings.SplitN(filter, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid filter format: %s (expected key=value)", filter)
		}
		key, value := parts[0], parts[1]

		// Validate filter key exists in config and value is valid
		if err := validateMetadataFilter(cfg, key, value); err != nil {
			return err
		}

		entries = history.FilterConsignmentsByMetadata(entries, key, value)
	}

	// Filter by version or default to latest
	if opts.Version != "" {
		entries = history.FilterByVersion(entries, opts.Version)
	} else if !opts.AllVersions {
		// Default: show only latest version
		entries = history.SortByTimestamp(entries, true) // newest first
		if len(entries) > 0 {
			entries = entries[:1] // keep only latest
		}
	}

	// Select template
	var templateType string
	if opts.Template != "" {
		// User specified template
		templateType = opts.Template
	} else if opts.AllVersions {
		// Auto-select changelog template for multi-version
		templateType = "changelog"
	} else {
		// Default to release-notes template
		templateType = "release-notes"
	}

	// Generate release notes using selected template
	// Output based on format
	if opts.JSON {
		// JSON output with structured data
		jsonData := map[string]interface{}{
			"package": opts.Package,
			"entries": entries,
		}
		if opts.Output != "" {
			// Write JSON to file
			file, err := os.Create(opts.Output)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer func() { _ = file.Close() }()
			return PrintJSON(file, jsonData)
		}
		return PrintJSON(os.Stdout, jsonData)
	}

	// Render release notes using template
	notes, err := template.RenderReleaseNotesWithTemplate(entries, templateType)
	if err != nil {
		return fmt.Errorf("failed to render release notes: %w", err)
	}

	// Output rendered notes
	if opts.Output != "" {
		if !opts.Quiet {
			fmt.Printf("Release notes written to %s\n", opts.Output)
		}
		return os.WriteFile(opts.Output, []byte(notes), 0644)
	}

	if !opts.Quiet {
		fmt.Print(notes)
	}
	return nil
}

// validateMetadataFilter checks if metadata key/value are valid per config
func validateMetadataFilter(cfg *config.Config, key, value string) error {
	// Find the metadata field definition
	var metadataField *config.MetadataField
	for i := range cfg.Metadata.Fields {
		if cfg.Metadata.Fields[i].Name == key {
			metadataField = &cfg.Metadata.Fields[i]
			break
		}
	}

	if metadataField == nil {
		// Open metadata - allowed but warn user
		fmt.Fprintf(os.Stderr, "Warning: %s is not defined in config (open metadata)\n", key)
		return nil
	}

	// Check if type is filterable
	if metadataField.Type != "string" && metadataField.Type != "" {
		return fmt.Errorf("cannot filter on %s: type %s is not filterable (only string type supported)",
			key, metadataField.Type)
	}

	if !metadataField.Required {
		// Warning: filtering on optional metadata may return unexpected results
		fmt.Fprintf(os.Stderr, "Warning: %s is optional metadata\n", key)
	}

	// For fields with allowedValues, validate value is in the list
	if len(metadataField.AllowedValues) > 0 {
		found := false
		for _, allowed := range metadataField.AllowedValues {
			if allowed == value {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("invalid value %s for %s (allowed: %v)",
				value, key, metadataField.AllowedValues)
		}
	}

	return nil
}
