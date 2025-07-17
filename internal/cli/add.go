package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/logger"
	"github.com/NatoNathan/shipyard/pkg/consignment"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// TODO: add flags for fast consignment creation
var AddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new consignment",
	Long:  "Create a new consignment to describe changes made to packages. This is used to track changes for release management.",
	Run: func(cmd *cobra.Command, args []string) {
		// Load project configuration
		projectConfig, err := config.LoadProjectConfig()
		if err != nil {
			logger.Error("Failed to load project configuration", "error", err)
			fmt.Println("Error: Unable to load project configuration.")
			fmt.Println("Make sure you're in a Shipyard project directory and run 'shipyard init' if needed.")
			os.Exit(1)
		}

		// Create consignment manager
		manager := consignment.NewManager(projectConfig)

		var (
			selectedPackages []string
			changeType       string
			summary          string
		)

		// Handle package selection based on repository type
		availablePackages := manager.GetAvailablePackages()
		if len(availablePackages) == 0 {
			fmt.Println("No packages configured in this project.")
			fmt.Println("Please run 'shipyard init' to configure your packages.")
			os.Exit(1)
		}

		if projectConfig.Type == config.RepositoryTypeMonorepo {
			// Create package options for selection
			packageOptions := make([]huh.Option[string], 0, len(availablePackages))
			for _, pkg := range availablePackages {
				packageOptions = append(packageOptions, huh.NewOption(fmt.Sprintf("%s (%s)", pkg.Name, pkg.Path), pkg.Name))
			}

			// Package selection form
			packageForm := huh.NewForm(
				huh.NewGroup(
					huh.NewNote().
						Title("📦 Package Selection").
						Description("Select the packages that have been changed."),
					huh.NewMultiSelect[string]().
						Title("Which packages have changed?").
						Description("Select all packages that have been modified.").
						Options(packageOptions...).
						Value(&selectedPackages),
				),
			)

			if err := packageForm.Run(); err != nil {
				logger.Error("Failed to get package selection", "error", err)
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			if len(selectedPackages) == 0 {
				fmt.Println("No packages selected. Consignment creation cancelled.")
				return
			}
		} else {
			// Single repo - use the configured package
			selectedPackages = []string{availablePackages[0].Name}
		}

		// Change type selection form
		changeTypeOptions := make([]huh.Option[string], 0, len(projectConfig.ChangeTypes))
		for _, ct := range projectConfig.ChangeTypes {
			optionText := ct.DisplayName
			if optionText == "" {
				optionText = ct.Name
			}
			changeTypeOptions = append(changeTypeOptions, huh.NewOption(optionText, ct.Name))
		}

		changeTypeForm := huh.NewForm(
			huh.NewGroup(
				huh.NewNote().
					Title("📝 Change Type").
					Description("Select the type of change you've made."),
				huh.NewSelect[string]().
					Title("What type of change is this?").
					Description("Choose the appropriate change type.").
					Options(changeTypeOptions...).
					Value(&changeType),
			),
		)

		if err := changeTypeForm.Run(); err != nil {
			logger.Error("Failed to get change type", "error", err)
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Summary form
		summaryForm := huh.NewForm(
			huh.NewGroup(
				huh.NewNote().
					Title("📋 Change Summary").
					Description("Provide a summary of the changes made."),
				huh.NewText().
					Title("Summary").
					Description("Describe the changes made to the selected packages.").
					Placeholder("e.g. Fixed bug in user authentication flow").
					Value(&summary),
			),
		)

		if err := summaryForm.Run(); err != nil {
			logger.Error("Failed to get change summary", "error", err)
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Create consignment using the manager
		createdConsignment, err := manager.CreateConsignment(selectedPackages, changeType, summary)
		if err != nil {
			logger.Error("Failed to create consignment", "error", err)
			fmt.Printf("Error: Unable to create consignment: %v\n", err)
			os.Exit(1)
		}

		// Success message
		fmt.Printf("\n🎉 Consignment created successfully!\n")
		fmt.Printf("📄 File: %s/%s.md\n", manager.GetConsignmentDir(), createdConsignment.ID)
		fmt.Printf("📦 Packages: %s\n", strings.Join(selectedPackages, ", "))
		fmt.Printf("🔄 Type: %s\n", changeType)
		fmt.Printf("📝 Summary: %s\n", summary)
		fmt.Printf("\n💡 Next steps:\n")
		fmt.Printf("   - Review the consignment file\n")
		fmt.Printf("   - Commit the consignment to your repository\n")
		fmt.Printf("   - Run 'shipyard version' to calculate new versions\n")

		logger.Info("Consignment created successfully",
			"id", createdConsignment.ID,
			"packages", selectedPackages,
			"type", changeType,
		)
	},
}
