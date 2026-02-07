package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/NatoNathan/shipyard/internal/commands"
	shipyarderrors "github.com/NatoNathan/shipyard/internal/errors"
	"github.com/NatoNathan/shipyard/internal/logger"
	"github.com/NatoNathan/shipyard/internal/ui"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "shipyard",
		Short: "Chart your project's version journey",
		Long: `Navigate your versioning voyage with Shipyard. Manage cargo (changes) across
your fleet (packages), chart courses to new version ports, and maintain detailed
ship's logs of your journey.`,
		Version: version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Configure logger based on flags
			quiet, _ := cmd.Flags().GetBool("quiet")
			verbose, _ := cmd.Flags().GetBool("verbose")

			log := logger.Get()
			log.SetQuiet(quiet)

			if verbose {
				log.SetLevel(logger.LevelDebug)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	rootCmd.SetVersionTemplate(fmt.Sprintf("shipyard version %s (commit: %s, built: %s)\n", version, commit, date))
	rootCmd.SetHelpFunc(ui.HelpFunc)
	rootCmd.SilenceUsage = true

	// Global flags
	rootCmd.PersistentFlags().BoolP("json", "j", false, "output in JSON format")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "suppress non-error output")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")

	// Create version info for commands that need it
	versionInfo := commands.VersionInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	}

	// Add subcommands
	rootCmd.AddCommand(commands.NewInitCommand())
	rootCmd.AddCommand(commands.NewAddCommand())
	rootCmd.AddCommand(commands.NewVersionCommand())
	rootCmd.AddCommand(commands.NewStatusCommand())
	rootCmd.AddCommand(commands.NewReleaseNotesCommand())
	rootCmd.AddCommand(commands.NewReleaseCommand())
	rootCmd.AddCommand(commands.NewCompletionCommand())
	rootCmd.AddCommand(commands.NewUpgradeCommand(versionInfo))
	rootCmd.AddCommand(commands.NewRemoveCommand())
	rootCmd.AddCommand(commands.NewValidateCommand())

	configCmd := &cobra.Command{Use: "config {show}", Aliases: []string{"cfg"}, Short: "Review the ship's standing orders"}
	configCmd.AddCommand(commands.NewConfigShowCommand())
	rootCmd.AddCommand(configCmd)

	if err := rootCmd.Execute(); err != nil {
		var exitErr *shipyarderrors.ExitCodeError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.Code)
		}
		os.Exit(1)
	}
}
