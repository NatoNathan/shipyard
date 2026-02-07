package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/consignment"
	"github.com/NatoNathan/shipyard/internal/ui"
	"github.com/spf13/cobra"
)

// RemoveCommandOptions holds options for the remove command
type RemoveCommandOptions struct {
	IDs   []string
	All   bool
	JSON  bool
	Quiet bool
}

// RemoveOutput is the JSON output structure for remove command
type RemoveOutput struct {
	Removed []string `json:"removed"`
	Count   int      `json:"count"`
}

// NewRemoveCommand creates the remove command
func NewRemoveCommand() *cobra.Command {
	opts := &RemoveCommandOptions{}

	cmd := &cobra.Command{
		Use:                   "remove {--id id... | --all}",
		DisableFlagsInUseLine: true,
		Aliases:               []string{"rm", "delete"},
		Short:   "Jettison cargo from the manifest",
		Long: `Remove one or more pending consignments from the manifest.

Use --id to remove specific consignments by ID, or --all to remove all pending consignments.`,
		Example: `  # Remove specific consignment(s)
  shipyard remove --id 20240101-120000-abc123

  # Remove all pending consignments
  shipyard remove --all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			globalFlags := GetGlobalFlags(cmd)
			opts.JSON = globalFlags.JSON
			opts.Quiet = globalFlags.Quiet
			return runRemove(opts)
		},
	}

	cmd.Flags().StringSliceVar(&opts.IDs, "id", nil, "Consignment ID(s) to remove")
	cmd.Flags().BoolVar(&opts.All, "all", false, "Remove all pending consignments")

	return cmd
}

func runRemove(opts *RemoveCommandOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	return runRemoveWithDir(cwd, opts)
}

func runRemoveWithDir(projectPath string, opts *RemoveCommandOptions) error {
	if !opts.All && len(opts.IDs) == 0 {
		return fmt.Errorf("specify --id or --all to remove consignments")
	}

	// Load configuration to get consignments path
	cfg, err := config.LoadFromDir(projectPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	consignmentsPath := cfg.Consignments.Path
	if consignmentsPath == "" {
		consignmentsPath = ".shipyard/consignments"
	}
	consignmentsDir := filepath.Join(projectPath, consignmentsPath)

	var removedIDs []string

	if opts.All {
		// Read all consignments
		allConsignments, err := consignment.ReadAllConsignments(consignmentsDir)
		if err != nil {
			return fmt.Errorf("failed to read consignments: %w", err)
		}

		if len(allConsignments) == 0 {
			if !opts.Quiet && !opts.JSON {
				fmt.Println("No pending consignments to remove")
			}
			if opts.JSON {
				return PrintJSON(os.Stdout, RemoveOutput{Removed: []string{}, Count: 0})
			}
			return nil
		}

		for _, c := range allConsignments {
			filePath := filepath.Join(consignmentsDir, c.ID+".md")
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("failed to remove consignment %s: %w", c.ID, err)
			}
			removedIDs = append(removedIDs, c.ID)
		}
	} else {
		for _, id := range opts.IDs {
			filePath := filepath.Join(consignmentsDir, id+".md")
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				return fmt.Errorf("consignment not found: %s", id)
			}
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("failed to remove consignment %s: %w", id, err)
			}
			removedIDs = append(removedIDs, id)
		}
	}

	if opts.JSON {
		return PrintJSON(os.Stdout, RemoveOutput{Removed: removedIDs, Count: len(removedIDs)})
	}

	if !opts.Quiet {
		fmt.Println()
		fmt.Println(ui.SuccessMessage(fmt.Sprintf("Removed %d consignment(s)", len(removedIDs))))
		for _, id := range removedIDs {
			fmt.Printf("  - %s\n", id)
		}
		fmt.Println()
	}

	return nil
}
