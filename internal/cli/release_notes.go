package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/logger"
	"github.com/NatoNathan/shipyard/pkg/changelog"
	"github.com/NatoNathan/shipyard/pkg/semver"
	"github.com/NatoNathan/shipyard/pkg/shipment"
	"github.com/NatoNathan/shipyard/pkg/templates"
	"github.com/spf13/cobra"
)

var ReleaseNotesCmd = &cobra.Command{
	Use:   "release-notes [version]",
	Short: "Get release notes for a specific version",
	Long:  `Get release notes for a specific version from shipment history.`,
	Example: `
  shipyard release-notes 1.2.3          # Get release notes for version 1.2.3
  shipyard release-notes v1.2.3         # Get release notes for version v1.2.3 (v prefix is optional)
  shipyard release-notes 1.2.3 -p myapp # Get release notes for version 1.2.3 of package myapp (monorepo)
  shipyard release-notes 1.2.3 --raw    # Get raw markdown output instead of rendered
  shipyard release-notes 1.2.3 -t simple # Use a different template for rendering`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Load project configuration
		projectConfig, err := config.LoadProjectConfig()
		if err != nil {
			logger.Error("Failed to load project configuration", "error", err)
			fmt.Printf("Error: Unable to load project configuration: %v\n", err)
			fmt.Println("Please run 'shipyard init' to initialize your project.")
			os.Exit(1)
		}

		// Parse the requested version
		versionStr := args[0]
		// Remove 'v' prefix if present
		versionStr = strings.TrimPrefix(versionStr, "v")

		requestedVersion, err := semver.Parse(versionStr)
		if err != nil {
			logger.Error("Invalid version format", "version", versionStr, "error", err)
			fmt.Printf("Error: Invalid version format '%s': %v\n", versionStr, err)
			fmt.Println("Version should be in format like '1.2.3' or 'v1.2.3'")
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

		// Load shipment history
		shipmentHistory := shipment.NewShipmentHistory(projectConfig)
		history, err := shipmentHistory.LoadHistory()
		if err != nil {
			logger.Error("Failed to load shipment history", "error", err)
			fmt.Printf("Error: Unable to load shipment history: %v\n", err)
			os.Exit(1)
		}

		if len(history) == 0 {
			logger.Error("No shipment history found")
			fmt.Printf("Error: No shipment history found\n")
			fmt.Println("Use 'shipyard version' to create some releases first.")
			os.Exit(1)
		}

		// Find the shipment for the requested version
		var matchingShipment *shipment.Shipment
		var matchingPackageName string

		for _, ship := range history {
			// Check all packages in the shipment
			for packageName, version := range ship.Versions {
				// If package filter is specified, only check that package
				if packageFilter != "" && packageName != packageFilter {
					continue
				}

				// For single repo, use the package name from config if no filter
				if projectConfig.Type == config.RepositoryTypeSingleRepo && packageFilter == "" {
					packageName = projectConfig.Package.Name
				}

				if version.Equals(requestedVersion) {
					matchingShipment = ship
					matchingPackageName = packageName
					break
				}
			}
			if matchingShipment != nil {
				break
			}
		}

		if matchingShipment == nil {
			logger.Error("Version not found in shipment history", "version", requestedVersion.String(), "package", packageFilter)
			if packageFilter != "" {
				fmt.Printf("Error: Version %s not found for package '%s' in shipment history\n", requestedVersion.String(), packageFilter)
			} else {
				fmt.Printf("Error: Version %s not found in shipment history\n", requestedVersion.String())
			}
			fmt.Println("Available versions:")

			// Show available versions
			if packageFilter != "" {
				fmt.Printf("  Package '%s':\n", packageFilter)
				for _, ship := range history {
					if version, exists := ship.Versions[packageFilter]; exists {
						fmt.Printf("    - %s (shipped: %s)\n", version.String(), ship.Date.Format("2006-01-02"))
					}
				}
			} else {
				// Show all versions for all packages
				versionMap := make(map[string][]string)
				for _, ship := range history {
					for pkgName, version := range ship.Versions {
						versionMap[pkgName] = append(versionMap[pkgName], fmt.Sprintf("%s (shipped: %s)", version.String(), ship.Date.Format("2006-01-02")))
					}
				}

				for pkgName, versions := range versionMap {
					fmt.Printf("  Package '%s':\n", pkgName)
					for _, v := range versions {
						fmt.Printf("    - %s\n", v)
					}
				}
			}
			os.Exit(1)
		}

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

		// Create template engine for generating release notes
		templateEngine := templates.NewTemplateEngine(projectConfig)

		// Convert the shipment to a changelog entry for the specific version
		releaseNotesContent, err := generateReleaseNotesForShipment(matchingShipment, matchingPackageName, requestedVersion, templateEngine, projectConfig)
		if err != nil {
			logger.Error("Failed to generate release notes", "error", err)
			fmt.Printf("Error: Unable to generate release notes: %v\n", err)
			os.Exit(1)
		}

		// Output the release notes
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
				fmt.Printf("📄 Release Notes for %s %s\n", matchingPackageName, requestedVersion.String())
				fmt.Printf("========================================\n\n")
				fmt.Print(rendered)
			}
		}

		logger.Info("Release notes command completed successfully",
			"version", requestedVersion.String(),
			"package", matchingPackageName,
			"shipment_id", matchingShipment.ID,
			"shipment_date", matchingShipment.Date.Format("2006-01-02"),
		)
	},
}

// generateReleaseNotesForShipment generates release notes for a specific shipment and version
func generateReleaseNotesForShipment(ship *shipment.Shipment, packageName string, version *semver.Version, templateEngine *templates.TemplateEngine, projectConfig *config.ProjectConfig) (string, error) {
	// Create a changelog entry for this specific version
	entry := templates.ChangelogEntry{
		Version:      version.String(),
		Date:         ship.Date.Format("2006-01-02"),
		DateTime:     ship.Date.Format("2006-01-02 15:04:05"),
		ShipmentDate: ship.Date.Format("2006-01-02"),
		Changes:      make(map[string][]templates.ChangelogChange),
		PackageName:  packageName,
		ShipmentID:   ship.ID,
	}

	// Process consignments for this package
	for _, c := range ship.Consignments {
		// Check if this consignment affects the requested package
		if changeType, exists := c.Packages[packageName]; exists {
			// Map change type to changelog section
			section := mapChangeTypeToSection(changeType, projectConfig)

			change := templates.ChangelogChange{
				Summary:       c.Summary,
				ChangeType:    changeType,
				Section:       section,
				PackageName:   packageName,
				ConsignmentID: c.ID,
			}

			entry.Changes[section] = append(entry.Changes[section], change)
		}
	}

	// Generate release notes using the template engine
	return templateEngine.RenderChangelogTemplate([]templates.ChangelogEntry{entry}, projectConfig.Changelog.Template)
}

// mapChangeTypeToSection maps a change type to a changelog section based on project configuration
func mapChangeTypeToSection(changeType string, projectConfig *config.ProjectConfig) string {
	// Look up the change type in the configuration
	for _, ct := range projectConfig.ChangeTypes {
		if ct.Name == changeType {
			// Use the configured section if available, otherwise use display name or name
			if ct.Section != "" {
				return ct.Section
			}
			if ct.DisplayName != "" {
				return ct.DisplayName
			}
			return ct.Name
		}
	}

	// Fallback for unknown change types
	return "Changed"
}

func init() {
	// Add flags for release-notes command
	ReleaseNotesCmd.Flags().StringP("package", "p", "", "Get release notes for a specific package only (monorepo only)")
	ReleaseNotesCmd.Flags().Bool("raw", false, "Show raw markdown instead of rendered output")
	ReleaseNotesCmd.Flags().StringP("template", "t", "", "Override the changelog template")
}
