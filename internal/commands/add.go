package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/consignment"
	"github.com/NatoNathan/shipyard/internal/errors"
	"github.com/NatoNathan/shipyard/internal/git"
	"github.com/NatoNathan/shipyard/internal/metadata"
	"github.com/NatoNathan/shipyard/internal/prompt"
	"github.com/NatoNathan/shipyard/internal/ui"
	"github.com/NatoNathan/shipyard/pkg/types"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// AddOptions holds the options for the add command
type AddOptions struct {
	Packages  []string
	Type      string
	Summary   string
	Metadata  map[string]string
	Timestamp time.Time // For testing
}

// runAdd executes the add command logic
func runAdd(projectPath string, options AddOptions) error {
	// Verify we're in a git repository
	isGitRepo, err := git.IsRepository(projectPath)
	if err != nil || !isGitRepo {
		return errors.NewGitError("not a git repository", nil)
	}

	// Load configuration
	cfg, err := config.LoadFromDir(projectPath)
	if err != nil {
		return errors.NewConfigError("failed to load configuration", err)
	}

	// Validate packages exist
	if err := validatePackages(cfg, options.Packages); err != nil {
		return err
	}

	// Validate change type
	if err := validateChangeType(options.Type); err != nil {
		return err
	}

	// Validate summary
	if strings.TrimSpace(options.Summary) == "" {
		return errors.NewValidationError("summary", "summary cannot be empty")
	}

	// Validate metadata against config if metadata validation is configured
	if err := metadata.ValidateMetadata(cfg, options.Metadata); err != nil {
		return err
	}

	// Convert and parse metadata
	metadataMap, err := convertMetadata(cfg, options.Metadata)
	if err != nil {
		return fmt.Errorf("failed to convert metadata: %w", err)
	}

	// Generate consignment ID
	var timestamp time.Time
	if options.Timestamp.IsZero() {
		timestamp = time.Now().UTC()
	} else {
		timestamp = options.Timestamp
	}

	id, err := consignment.GenerateID(timestamp)
	if err != nil {
		return fmt.Errorf("failed to generate consignment ID: %w", err)
	}

	// Create consignment
	cons := &consignment.Consignment{
		ID:         id,
		Timestamp:  timestamp,
		Packages:   options.Packages,
		ChangeType: types.ChangeType(options.Type),
		Summary:    options.Summary,
		Metadata:   metadataMap,
	}

	// Get consignments directory from config
	consignmentsPath := cfg.Consignments.Path
	if consignmentsPath == "" {
		consignmentsPath = ".shipyard/consignments"
	}
	consignmentsDir := filepath.Join(projectPath, consignmentsPath)

	// Write consignment file
	if err := consignment.WriteConsignment(cons, consignmentsDir); err != nil {
		return fmt.Errorf("failed to write consignment: %w", err)
	}

	// Success message with styled output
	filename := fmt.Sprintf("%s.md", id)
	fmt.Println()
	fmt.Println(ui.SuccessMessage(fmt.Sprintf("Created consignment: %s", filename)))
	fmt.Println()
	fmt.Println(ui.KeyValue("Packages", strings.Join(options.Packages, ", ")))
	fmt.Println(ui.KeyValue("Type", options.Type))
	fmt.Println(ui.KeyValue("Summary", truncateSummary(options.Summary, 60)))
	fmt.Println()

	return nil
}

// validatePackages checks that all package names exist in the configuration
func validatePackages(cfg *config.Config, packages []string) error {
	if len(packages) == 0 {
		return errors.NewValidationError("packages", "at least one package is required")
	}

	// Build map of valid package names
	validPackages := make(map[string]bool)
	for _, pkg := range cfg.Packages {
		validPackages[pkg.Name] = true
	}

	// Check each package
	var invalidPackages []string
	for _, pkg := range packages {
		if !validPackages[pkg] {
			invalidPackages = append(invalidPackages, pkg)
		}
	}

	if len(invalidPackages) > 0 {
		// List available packages
		var availablePackages []string
		for _, pkg := range cfg.Packages {
			availablePackages = append(availablePackages, pkg.Name)
		}

		return fmt.Errorf("invalid package reference: %s\n\nAvailable packages:\n  - %s",
			strings.Join(invalidPackages, ", "),
			strings.Join(availablePackages, "\n  - "))
	}

	return nil
}

// validateChangeType checks that the change type is valid
func validateChangeType(changeType string) error {
	validTypes := []string{"patch", "minor", "major"}
	for _, valid := range validTypes {
		if changeType == valid {
			return nil
		}
	}

	return errors.NewValidationError("changeType",
		fmt.Sprintf("invalid change type: %s (valid: %s)", changeType, strings.Join(validTypes, ", ")))
}

// convertMetadata converts string map to interface map with proper type parsing
func convertMetadata(cfg *config.Config, input map[string]string) (map[string]interface{}, error) {
	if input == nil {
		return nil, nil
	}

	fieldDefs := make(map[string]config.MetadataField)
	if cfg.Metadata.Fields != nil {
		for _, field := range cfg.Metadata.Fields {
			fieldDefs[field.Name] = field
		}
	}

	result := make(map[string]interface{}, len(input))
	for k, v := range input {
		if field, exists := fieldDefs[k]; exists {
			parsed, err := metadata.ParseMetadataValue(field, v)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %s: %w", k, err)
			}
			result[k] = parsed
		} else {
			result[k] = v
		}
	}

	return result, nil
}

// truncateSummary truncates a summary to the specified length
func truncateSummary(summary string, maxLen int) string {
	// Get first line only
	lines := strings.Split(summary, "\n")
	firstLine := strings.TrimSpace(lines[0])

	// Remove markdown heading markers
	firstLine = strings.TrimPrefix(firstLine, "# ")
	firstLine = strings.TrimPrefix(firstLine, "## ")
	firstLine = strings.TrimPrefix(firstLine, "### ")

	if len(firstLine) <= maxLen {
		return firstLine
	}
	return firstLine[:maxLen-3] + "..."
}

// promptForMetadata prompts for metadata fields interactively using huh
func promptForMetadata(cfg *config.Config, existingMetadata map[string]string) (map[string]string, error) {
	if len(cfg.Metadata.Fields) == 0 {
		return existingMetadata, nil
	}

	result := make(map[string]string)

	// Copy existing metadata
	for k, v := range existingMetadata {
		result[k] = v
	}

	// Prompt for each field
	for _, field := range cfg.Metadata.Fields {
		// Skip if already provided via flag
		if _, exists := existingMetadata[field.Name]; exists {
			continue
		}

		// Create appropriate field prompt based on type
		var value string
		var err error

		switch {
		case len(field.AllowedValues) > 0:
			value, err = promptForSelect(field)
		case field.Type == "int" || field.Type == "integer":
			value, err = promptForInt(field)
		default:
			value, err = promptForString(field)
		}

		if err != nil {
			return nil, err
		}

		if value != "" {
			result[field.Name] = value
		}
	}

	return result, nil
}

// promptForSelect creates a select prompt for enum fields
func promptForSelect(field config.MetadataField) (string, error) {
	if field.Default == "" && !field.Required {
		// Add empty option for optional fields
		field.AllowedValues = append([]string{""}, field.AllowedValues...)
	}

	var value string
	title := field.Name
	if field.Description != "" {
		title = fmt.Sprintf("%s (%s)", field.Name, field.Description)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(title).
				Options(huh.NewOptions(field.AllowedValues...)...).
				Value(&value),
		),
	)

	if err := form.Run(); err != nil {
		return "", err
	}

	return value, nil
}

// promptForInt creates an input prompt for integer fields with validation
func promptForInt(field config.MetadataField) (string, error) {
	var value string
	title := field.Name
	description := field.Description

	// Add range info to description
	if field.Min != nil || field.Max != nil {
		rangeInfo := ""
		if field.Min != nil && field.Max != nil {
			rangeInfo = fmt.Sprintf(" (range: %d-%d)", *field.Min, *field.Max)
		} else if field.Min != nil {
			rangeInfo = fmt.Sprintf(" (min: %d)", *field.Min)
		} else {
			rangeInfo = fmt.Sprintf(" (max: %d)", *field.Max)
		}
		if description != "" {
			description += rangeInfo
		} else {
			description = strings.TrimPrefix(rangeInfo, " ")
		}
	}

	input := huh.NewInput().
		Title(title).
		Value(&value).
		Validate(func(s string) error {
			// Allow empty for optional fields
			if s == "" && !field.Required {
				return nil
			}

			// Parse integer
			intVal, err := strconv.Atoi(strings.TrimSpace(s))
			if err != nil {
				return fmt.Errorf("invalid integer")
			}

			// Validate range
			if field.Min != nil && intVal < *field.Min {
				return fmt.Errorf("below minimum %d", *field.Min)
			}
			if field.Max != nil && intVal > *field.Max {
				return fmt.Errorf("above maximum %d", *field.Max)
			}

			return nil
		})

	if description != "" {
		input = input.Description(description)
	}
	if field.Default != "" {
		input = input.Placeholder(field.Default)
	}

	form := huh.NewForm(huh.NewGroup(input))
	if err := form.Run(); err != nil {
		return "", err
	}

	return value, nil
}

// promptForString creates an input prompt for string fields with validation
func promptForString(field config.MetadataField) (string, error) {
	var value string
	title := field.Name
	description := field.Description

	// Add pattern info to description
	if field.Pattern != "" {
		patternInfo := fmt.Sprintf(" (pattern: %s)", field.Pattern)
		if description != "" {
			description += patternInfo
		} else {
			description = strings.TrimPrefix(patternInfo, " ")
		}
	}

	input := huh.NewInput().
		Title(title).
		Value(&value).
		Validate(func(s string) error {
			// Allow empty for optional fields
			if s == "" && !field.Required {
				return nil
			}

			// Validate using metadata package
			tempMetadata := map[string]string{field.Name: s}
			tempCfg := &config.Config{
				Metadata: config.MetadataConfig{
					Fields: []config.MetadataField{field},
				},
			}
			return metadata.ValidateMetadata(tempCfg, tempMetadata)
		})

	if description != "" {
		input = input.Description(description)
	}
	if field.Default != "" {
		input = input.Placeholder(field.Default)
	}

	form := huh.NewForm(huh.NewGroup(input))
	if err := form.Run(); err != nil {
		return "", err
	}

	return value, nil
}

// AddCmd returns the add command
func AddCmd() *cobra.Command {
	var (
		packages []string
		typeName string
		summary  string
		metadata []string
	)

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Log cargo in the ship's manifest",
		Long: `Record new cargo in your ship's manifest. Each consignment documents what's
being shipped (changes), which vessels carry it (packages), and how it affects
the voyage (patch/minor/major). Interactive mode guides you through manifest
creation, or use flags to log cargo directly.`,
		Example: `  # Interactive mode
  shipyard add

  # Non-interactive mode
  shipyard add --package core --type minor --summary "Added new feature"

  # Multiple packages
  shipyard add --package core --package api --type major --summary "Breaking change"

  # With metadata
  shipyard add --package core --type patch --summary "Fixed bug" \
    --metadata author=dev@example.com --metadata issue=JIRA-123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectPath, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}

			// Parse metadata flags
			metadataMap := make(map[string]string)
			for _, m := range metadata {
				parts := strings.SplitN(m, "=", 2)
				if len(parts) != 2 {
					return errors.NewValidationError("metadata", fmt.Sprintf("invalid metadata format: %s (expected key=value)", m))
				}
				metadataMap[parts[0]] = parts[1]
			}

			// Check if we have all required flags for non-interactive mode
			if len(packages) > 0 && typeName != "" && summary != "" {
				// Non-interactive mode
				return runAdd(projectPath, AddOptions{
					Packages: packages,
					Type:     typeName,
					Summary:  summary,
					Metadata: metadataMap,
				})
			}

			// Interactive mode: prompt for missing fields
			return runInteractiveAdd(projectPath, packages, typeName, summary, metadataMap)
		},
	}

	cmd.Flags().StringSliceVarP(&packages, "package", "p", nil, "package name(s) affected by this change")
	cmd.Flags().StringVarP(&typeName, "type", "t", "", "change type: patch, minor, or major")
	cmd.Flags().StringVarP(&summary, "summary", "s", "", "summary of the change")
	cmd.Flags().StringSliceVarP(&metadata, "metadata", "m", nil, "metadata in key=value format (can be repeated)")

	// Register package name completion
	RegisterPackageCompletions(cmd, "package")

	return cmd
}

// runInteractiveAdd runs the add command in interactive mode
func runInteractiveAdd(projectPath string, packages []string, typeName, summary string, metadata map[string]string) error {
	// Load config to get available packages
	cfg, err := config.LoadFromDir(projectPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get available package names
	availablePackages := make([]string, 0, len(cfg.Packages))
	for _, pkg := range cfg.Packages {
		availablePackages = append(availablePackages, pkg.Name)
	}

	// Prompt for packages if not provided
	if len(packages) == 0 {
		packages, err = prompt.PromptForPackages(availablePackages)
		if err != nil {
			return fmt.Errorf("failed to select packages: %w", err)
		}
	}

	// Prompt for change type if not provided
	var changeType types.ChangeType
	if typeName == "" {
		changeType, err = prompt.PromptForChangeType()
		if err != nil {
			return fmt.Errorf("failed to select change type: %w", err)
		}
	} else {
		changeType = types.ChangeType(typeName)
	}

	// Prompt for summary if not provided
	if summary == "" {
		fmt.Println()
		summary, err = prompt.PromptSummary(projectPath)
		if err != nil {
			return fmt.Errorf("failed to get summary: %w", err)
		}
	}

	// Prompt for metadata fields if configured
	metadata, err = promptForMetadata(cfg, metadata)
	if err != nil {
		return fmt.Errorf("failed to collect metadata: %w", err)
	}

	// Run the add command
	return runAdd(projectPath, AddOptions{
		Packages: packages,
		Type:     string(changeType),
		Summary:  summary,
		Metadata: metadata,
	})
}
