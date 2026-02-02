package main

import (
	"fmt"
	"os"

	"github.com/NatoNathan/shipyard/internal/commands"
	"github.com/NatoNathan/shipyard/internal/logger"
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
			cmd.Help()
		},
	}

	rootCmd.SetVersionTemplate(fmt.Sprintf("shipyard version %s (commit: %s, built: %s)\n", version, commit, date))

	// Global flags
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default is .shipyard/shipyard.yaml)")
	rootCmd.PersistentFlags().BoolP("json", "j", false, "output in JSON format")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "suppress non-error output")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")

	// Add subcommands
	rootCmd.AddCommand(commands.InitCmd())
	rootCmd.AddCommand(commands.AddCmd())
	rootCmd.AddCommand(commands.NewVersionCommand())
	rootCmd.AddCommand(commands.NewStatusCommand())
	rootCmd.AddCommand(commands.NewReleaseCommand())
	rootCmd.AddCommand(commands.NewReleaseNotesCommand())
	rootCmd.AddCommand(commands.NewCompletionCommand())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
