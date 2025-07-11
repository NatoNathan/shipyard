package cli

import (
	"fmt"
	"os"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/logger"
	"github.com/NatoNathan/shipyard/pkg/changelog"
	"github.com/NatoNathan/shipyard/pkg/consignment"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Generate changelogs and apply version updates",
	Long:  "Generate changelogs from consignments and apply version updates to package manifests.",
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

		if len(consignments) == 0 {
			fmt.Println("No consignments found.")
			fmt.Println("Create some consignments with 'shipyard add' first.")
			return
		}

		// Calculate versions for all packages
		versions, err := manager.CalculateAllVersions()
		if err != nil {
			logger.Error("Failed to calculate versions", "error", err)
			fmt.Printf("Error: Unable to calculate versions: %v\n", err)
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

		// Create changelog generator
		generator, err := changelog.NewGenerator(projectConfig)
		if err != nil {
			logger.Error("Failed to create changelog generator", "error", err)
			fmt.Printf("Error: Unable to create changelog generator: %v\n", err)
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

		// Generate changelog
		var changelogContent string
		if packageFilter != "" {
			// Generate changelog for specific package
			version, exists := versions[packageFilter]
			if !exists {
				logger.Error("No version calculated for package", "package", packageFilter)
				fmt.Printf("Error: No version calculated for package '%s'\n", packageFilter)
				os.Exit(1)
			}

			changelogContent, err = generator.GenerateChangelogForPackage(packageFilter, consignments, version)
			if err != nil {
				logger.Error("Failed to generate changelog for package", "package", packageFilter, "error", err)
				fmt.Printf("Error: Unable to generate changelog for package '%s': %v\n", packageFilter, err)
				os.Exit(1)
			}
		} else {
			// Generate changelog for all packages
			changelogContent, err = generator.GenerateChangelog(consignments, versions)
			if err != nil {
				logger.Error("Failed to generate changelog", "error", err)
				fmt.Printf("Error: Unable to generate changelog: %v\n", err)
				os.Exit(1)
			}
		}

		// Check if dry-run or preview flag is set
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		preview, _ := cmd.Flags().GetBool("preview")

		if dryRun || preview {
			// For dry-run, show both changelog and version info
			// For preview, show just the new changelog (pretty printed)
			if dryRun {
				// Raw print changelog and version info
				fmt.Println("📄 Generated Changelog (Raw):")
				fmt.Println("=============================")
				fmt.Println()
				fmt.Print(changelogContent)
				fmt.Println()
				fmt.Println("🔖 Version Information:")
				fmt.Println("======================")
				if packageFilter != "" {
					if version, exists := versions[packageFilter]; exists {
						fmt.Printf("%s: %s\n", packageFilter, version.String())
					}
				} else {
					for pkgName, version := range versions {
						fmt.Printf("%s: %s\n", pkgName, version.String())
					}
				}
				fmt.Println()
				fmt.Printf("� Consignments to be processed: %d\n", len(consignments))
				fmt.Printf("📝 Template: %s\n", projectConfig.Changelog.Template)
				if packageFilter != "" {
					fmt.Printf("🎯 Package filter: %s\n", packageFilter)
				}
				fmt.Println()
				fmt.Println("💡 To apply these changes, run 'shipyard version' without --dry-run")
			} else {
				// Preview mode - pretty print just the new changelog
				rendered, err := renderMarkdown(changelogContent)
				if err != nil {
					logger.Error("Failed to render markdown", "error", err)
					fmt.Printf("Warning: Failed to render markdown, showing raw content: %v\n", err)
					fmt.Print(changelogContent)
				} else {
					fmt.Println("📄 Generated Changelog Preview:")
					fmt.Println("===============================")
					fmt.Println()
					fmt.Print(rendered)
					fmt.Println("===============================")
				}
				fmt.Printf("📝 Template: %s\n", projectConfig.Changelog.Template)
				fmt.Printf("📦 Consignments processed: %d\n", len(consignments))
				if packageFilter != "" {
					fmt.Printf("🎯 Package filter: %s\n", packageFilter)
				}
			}
			return
		}

		// Ask for confirmation unless --yes flag is set
		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			var confirm bool
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title("Generate Changelog and Apply Version Updates?").
						Description(fmt.Sprintf("This will generate changelog, update package versions and clear %d consignments. Are you sure?", len(consignments))).
						Value(&confirm),
				),
			)

			if err := form.Run(); err != nil {
				logger.Error("Failed to get user confirmation", "error", err)
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			if !confirm {
				fmt.Println("Version update and changelog generation cancelled.")
				return
			}
		}

		// Write changelog to file first
		outputPath, _ := cmd.Flags().GetString("output")
		if outputPath == "" {
			outputPath = "CHANGELOG.md"
		}

		// Ask user for confirmation if file already exists
		if _, err := os.Stat(outputPath); err == nil {
			if !yes {
				var overwrite bool
				form := huh.NewForm(
					huh.NewGroup(
						huh.NewConfirm().
							Title("Overwrite Existing Changelog?").
							Description(fmt.Sprintf("The file %s already exists. Do you want to overwrite it?", outputPath)).
							Value(&overwrite),
					),
				)

				if err := form.Run(); err != nil {
					logger.Error("Failed to get user confirmation", "error", err)
					fmt.Printf("Error: %v\n", err)
					os.Exit(1)
				}

				if !overwrite {
					fmt.Println("Changelog generation cancelled.")
					return
				}
			}
		}

		// Write changelog to file
		if err := writeChangelog(outputPath, changelogContent); err != nil {
			logger.Error("Failed to write changelog", "error", err)
			fmt.Printf("Error: Unable to write changelog: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Changelog generated successfully!\n")
		fmt.Printf("� File: %s\n", outputPath)

		// Apply version updates
		fmt.Printf("\n�🔄 Applying version updates...\n")

		appliedVersions, err := manager.ApplyConsignments()
		if err != nil {
			logger.Error("Failed to apply consignments", "error", err)
			fmt.Printf("Error: Unable to apply version updates: %v\n", err)
			os.Exit(1)
		}

		// Success message
		fmt.Printf("\n✅ Version updates applied successfully!\n")
		fmt.Printf("📦 Packages updated:\n")
		for pkgName, version := range appliedVersions {
			fmt.Printf("   - %s: %s\n", pkgName, version.String())
		}

		fmt.Printf("\n🧹 Consignments cleared: %d\n", len(consignments))
		fmt.Printf("� Template: %s\n", projectConfig.Changelog.Template)
		if packageFilter != "" {
			fmt.Printf("🎯 Package filter: %s\n", packageFilter)
		}

		fmt.Printf("\n💡 Next steps:\n")
		fmt.Printf("   - Review the generated changelog and updated package manifests\n")
		fmt.Printf("   - Commit all changes to your repository\n")
		fmt.Printf("   - Create and push git tags for the new versions\n")

		logger.Info("Version updates and changelog generation completed successfully",
			"output", outputPath,
			"template", projectConfig.Changelog.Template,
			"packages", len(appliedVersions),
			"consignments_cleared", len(consignments),
		)
	},
}

func init() {
	// Add flags for version command
	VersionCmd.Flags().Bool("dry-run", false, "Show changelog and version info without applying changes")
	VersionCmd.Flags().Bool("preview", false, "Pretty print just the new changelog without applying changes")
	VersionCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompts")
	VersionCmd.Flags().StringP("package", "p", "", "Generate changelog and version for a specific package only")
	VersionCmd.Flags().StringP("output", "o", "", "Output file path for changelog (default: CHANGELOG.md)")
	VersionCmd.Flags().StringP("template", "t", "", "Override the changelog template")
}
