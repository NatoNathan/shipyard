package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/consignment"
	"github.com/NatoNathan/shipyard/internal/graph"
	"github.com/NatoNathan/shipyard/internal/ui"
	"github.com/spf13/cobra"
)

// ValidateOutput is the JSON output structure for validate command
type ValidateOutput struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

// NewValidateCommand creates the validate command
func NewValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration and consignments",
		Long: `Validate shipyard configuration, consignment files, and the dependency graph.

Reports any errors or warnings found during validation.

Examples:
  # Validate everything
  shipyard validate

  # Validate with JSON output
  shipyard validate --json
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			globalFlags := GetGlobalFlags(cmd)
			return runValidate(globalFlags)
		},
	}

	return cmd
}

func runValidate(flags GlobalFlags) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	return runValidateWithDir(cwd, flags)
}

func runValidateWithDir(projectPath string, flags GlobalFlags) error {
	var validationErrors []string
	var warnings []string

	// 1. Load and validate config
	cfg, err := config.LoadFromDir(projectPath)
	if err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("config load error: %s", err))
	}

	if cfg != nil {
		if err := cfg.Validate(); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("config validation: %s", err))
		}

		if err := config.ValidateDependencies(cfg); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("dependency validation: %s", err))
		}
	}

	// 2. Read consignments and check for parse errors
	if cfg != nil {
		consignmentsPath := cfg.Consignments.Path
		if consignmentsPath == "" {
			consignmentsPath = ".shipyard/consignments"
		}
		consignmentsDir := filepath.Join(projectPath, consignmentsPath)

		if _, err := os.Stat(consignmentsDir); err == nil {
			entries, err := os.ReadDir(consignmentsDir)
			if err != nil {
				validationErrors = append(validationErrors, fmt.Sprintf("consignments directory: %s", err))
			} else {
				for _, entry := range entries {
					if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
						continue
					}
					filePath := filepath.Join(consignmentsDir, entry.Name())
					_, err := consignment.ReadConsignment(filePath)
					if err != nil {
						validationErrors = append(validationErrors, fmt.Sprintf("consignment %s: %s", entry.Name(), err))
					}
				}
			}
		}

		// 3. Build dependency graph and check for cycles
		depGraph, err := graph.BuildGraph(cfg)
		if err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("dependency graph: %s", err))
		} else {
			hasCycles, cycles := graph.DetectCycles(depGraph)
			if hasCycles {
				for _, cycle := range cycles {
					warnings = append(warnings, fmt.Sprintf("dependency cycle detected: %s", strings.Join(cycle, " -> ")))
				}
			}
		}
	}

	valid := len(validationErrors) == 0

	// Output
	if flags.JSON {
		return PrintJSON(os.Stdout, ValidateOutput{
			Valid:    valid,
			Errors:   validationErrors,
			Warnings: warnings,
		})
	}

	if flags.Quiet {
		if !valid {
			return fmt.Errorf("validation failed")
		}
		return nil
	}

	if len(validationErrors) > 0 {
		fmt.Println()
		fmt.Println("Errors:")
		for _, e := range validationErrors {
			fmt.Printf("  - %s\n", e)
		}
	}

	if len(warnings) > 0 {
		fmt.Println()
		fmt.Println("Warnings:")
		for _, w := range warnings {
			fmt.Printf("  - %s\n", w)
		}
	}

	fmt.Println()
	if valid {
		fmt.Println(ui.SuccessMessage("Validation passed"))
	} else {
		fmt.Println("Validation failed")
		return fmt.Errorf("validation failed with %d error(s)", len(validationErrors))
	}

	return nil
}
