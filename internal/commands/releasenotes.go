package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/NatoNathan/shipyard/internal/history"
	"github.com/NatoNathan/shipyard/internal/template"
	"github.com/spf13/cobra"
)

// ReleaseNotesOptions holds options for the release-notes command
type ReleaseNotesOptions struct {
	Package string
	Output  string
	Version string
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
			return runReleaseNotes(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Package, "package", "p", "", "Filter by package name")
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output file (default: stdout)")
	cmd.Flags().StringVar(&opts.Version, "version", "", "Generate notes for specific version")

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

	// Filter by package if specified
	if opts.Package != "" {
		entries = history.FilterByPackage(entries, opts.Package)
	}

	// Filter by version if specified
	if opts.Version != "" {
		entries = history.FilterByVersion(entries, opts.Version)
	}

	// Generate release notes using template
	notes, err := template.RenderReleaseNotes(entries)
	if err != nil {
		return fmt.Errorf("failed to render release notes: %w", err)
	}

	// Output
	if opts.Output != "" {
		return os.WriteFile(opts.Output, []byte(notes), 0644)
	}
	fmt.Print(notes)
	return nil
}
