package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/logger"
	"github.com/NatoNathan/shipyard/pkg/changelog"
	"github.com/NatoNathan/shipyard/pkg/consignment"
	"github.com/spf13/cobra"
)

var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show consignment status and version information",
	Long:  "Display current consignments, what new versions would be, and optionally output release notes for packages.",
	Run: func(cmd *cobra.Command, args []string) {
		// Load project configuration
		projectConfig, err := config.LoadProjectConfig()
		if err != nil {
			logger.Error("Failed to load project configuration", "error", err)
			fmt.Printf("Error: Unable to load project configuration: %v\n", err)
			fmt.Println("Please run 'shipyard init' to initialize your project.")
			os.Exit(1)
		}

		// Create consignment manager
		manager := consignment.NewManager(projectConfig)

		// Get all consignments
		consignments, err := manager.GetConsignmens()
		if err != nil {
			logger.Error("Failed to get consignments", "error", err)
			fmt.Printf("Error: Unable to read consignments: %v\n", err)
			os.Exit(1)
		}

		// Check for package filter (monorepo only)
		packageFilter, _ := cmd.Flags().GetString("package")
		if packageFilter != "" {
			if projectConfig.Type != config.RepositoryTypeMonorepo {
				logger.Error("Package filter can only be used with monorepo projects")
				fmt.Printf("Error: --package flag can only be used with monorepo projects\n")
				os.Exit(1)
			}

			// Validate package exists
			packageExists := false
			for _, pkg := range projectConfig.Packages {
				if pkg.Name == packageFilter {
					packageExists = true
					break
				}
			}

			if !packageExists {
				logger.Error("Package not found", "package", packageFilter)
				fmt.Printf("Error: Package '%s' not found in project configuration\n", packageFilter)
				os.Exit(1)
			}
		}

		// Show project info
		fmt.Printf("📊 Shipyard Project Status\n")
		fmt.Printf("==========================\n\n")

		fmt.Printf("📁 Project Type: %s\n", projectConfig.Type)
		if projectConfig.Type == config.RepositoryTypeMonorepo {
			fmt.Printf("📦 Packages: %d\n", len(projectConfig.Packages))
		} else {
			fmt.Printf("📦 Package: %s\n", projectConfig.Package.Name)
		}
		fmt.Printf("📝 Changelog Template: %s\n", projectConfig.Changelog.Template)
		fmt.Printf("\n")

		// Show consignments
		fmt.Printf("📋 Consignments\n")
		fmt.Printf("===============\n")

		if len(consignments) == 0 {
			fmt.Println("No consignments found.")
			fmt.Println("Create some consignments with 'shipyard add' first.")
			return
		}

		fmt.Printf("Total consignments: %d\n\n", len(consignments))

		// Filter consignments if package filter is specified
		filteredConsignments := consignments
		if packageFilter != "" {
			filteredConsignments = make([]*consignment.Consignment, 0)
			for _, c := range consignments {
				if _, exists := c.Packages[packageFilter]; exists {
					filteredConsignments = append(filteredConsignments, c)
				}
			}
		}

		// Show consignments summary
		for _, c := range filteredConsignments {
			fmt.Printf("• %s (Created: %s)\n", c.ID, c.Created.Format(time.RFC3339))
			fmt.Printf("  Summary: %s\n", c.Summary)
			fmt.Printf("  Packages: ")
			first := true
			for pkgName, changeType := range c.Packages {
				if packageFilter != "" && pkgName != packageFilter {
					continue
				}
				if !first {
					fmt.Printf(", ")
				}
				fmt.Printf("%s (%s)", pkgName, changeType)
				first = false
			}
			fmt.Printf("\n\n")
		}

		// Calculate and show version information
		fmt.Printf("🔖 Version Information\n")
		fmt.Printf("=====================\n")

		versions, err := manager.CalculateAllVersions()
		if err != nil {
			logger.Error("Failed to calculate versions", "error", err)
			fmt.Printf("Error: Unable to calculate versions: %v\n", err)
			os.Exit(1)
		}

		if packageFilter != "" {
			if version, exists := versions[packageFilter]; exists {
				fmt.Printf("Next version for %s: %s\n", packageFilter, version.String())
			} else {
				fmt.Printf("No version changes for %s\n", packageFilter)
			}
		} else {
			fmt.Printf("Next versions:\n")
			for pkgName, version := range versions {
				fmt.Printf("  %s: %s\n", pkgName, version.String())
			}
		}

		// Check if release notes should be generated
		releaseNotes, _ := cmd.Flags().GetBool("release-notes")
		if releaseNotes {
			fmt.Printf("\n📄 Release Notes\n")
			fmt.Printf("================\n\n")

			// Check for template override
			templateOverride, _ := cmd.Flags().GetString("template")
			if templateOverride != "" {
				// Validate the template
				if err := changelog.ValidateTemplate(templateOverride); err != nil {
					logger.Error("Invalid template specified", "template", templateOverride, "error", err)
					fmt.Printf("Error: %v\n", err)
					os.Exit(1)
				}
				projectConfig.Changelog.Template = templateOverride
			}

			// Create changelog generator
			generator, err := changelog.NewGenerator(projectConfig)
			if err != nil {
				logger.Error("Failed to create changelog generator", "error", err)
				fmt.Printf("Error: Unable to create changelog generator: %v\n", err)
				os.Exit(1)
			}

			// Generate release notes
			var releaseNotesContent string
			if packageFilter != "" {
				// Generate release notes for specific package
				version, exists := versions[packageFilter]
				if !exists {
					fmt.Printf("No version changes for package '%s'\n", packageFilter)
					return
				}

				releaseNotesContent, err = generator.GenerateChangelogForPackage(packageFilter, consignments, version)
				if err != nil {
					logger.Error("Failed to generate release notes for package", "package", packageFilter, "error", err)
					fmt.Printf("Error: Unable to generate release notes for package '%s': %v\n", packageFilter, err)
					os.Exit(1)
				}
			} else {
				// Generate release notes for all packages
				releaseNotesContent, err = generator.GenerateChangelog(consignments, versions)
				if err != nil {
					logger.Error("Failed to generate release notes", "error", err)
					fmt.Printf("Error: Unable to generate release notes: %v\n", err)
					os.Exit(1)
				}
			}

			// Check if raw output is requested
			raw, _ := cmd.Flags().GetBool("raw")
			if raw {
				fmt.Print(releaseNotesContent)
			} else {
				// Render markdown
				rendered, err := renderMarkdown(releaseNotesContent)
				if err != nil {
					logger.Error("Failed to render markdown", "error", err)
					fmt.Printf("Warning: Failed to render markdown, showing raw content: %v\n", err)
					fmt.Print(releaseNotesContent)
				} else {
					fmt.Print(rendered)
				}
			}
		}

		fmt.Printf("\n💡 Next steps:\n")
		fmt.Printf("   - Run 'shipyard version --preview' to see the changelog\n")
		fmt.Printf("   - Run 'shipyard version --dry-run' to see changelog and version info\n")
		fmt.Printf("   - Run 'shipyard version' to generate changelog and apply versions\n")

		logger.Info("Status command completed successfully",
			"consignments", len(consignments),
			"packages", len(versions),
		)
	},
}

func init() {
	// Add flags for status command
	StatusCmd.Flags().StringP("package", "p", "", "Show status for a specific package only (monorepo only)")
	StatusCmd.Flags().Bool("release-notes", false, "Generate and display release notes")
	StatusCmd.Flags().Bool("raw", false, "Show raw markdown instead of rendered output (use with --release-notes)")
	StatusCmd.Flags().StringP("template", "t", "", "Override the changelog template (use with --release-notes)")
}
