package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/logger"
	"github.com/NatoNathan/shipyard/pkg/changelog"
	"github.com/NatoNathan/shipyard/pkg/consignment"
	"github.com/NatoNathan/shipyard/pkg/git"
	"github.com/NatoNathan/shipyard/pkg/semver"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Generate changelogs and apply version updates",
	Long:  `Generate changelogs and apply version updates to package manifests.`,
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

		// Check if regenerate flag is set
		regenerate, _ := cmd.Flags().GetBool("regenerate")

		// Calculate versions for packages that have consignments (if any)
		var versions map[string]*semver.Version
		if len(consignments) > 0 && !regenerate {
			versions, err = manager.CalculateAllVersions()
			if err != nil {
				logger.Error("Failed to calculate versions", "error", err)
				fmt.Printf("Error: Unable to calculate versions: %v\n", err)
				os.Exit(1)
			}
		} else {
			// No current consignments or regenerate mode, just use empty versions map
			versions = make(map[string]*semver.Version)
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

		// Generate changelog from shipment history (primary) or current consignments (fallback)
		var changelogContent string
		var packageChangelogs map[string]string // For monorepo separate changelogs
		hasCurrentConsignments := len(consignments) > 0 && !regenerate

		if regenerate {
			// Regenerate mode: only use shipment history
			if packageFilter != "" {
				changelogContent, err = generator.GenerateChangelogFromHistoryForPackage(packageFilter)
				if err != nil {
					logger.Error("Failed to generate changelog from history for package", "package", packageFilter, "error", err)
					fmt.Printf("Error: Unable to generate changelog from history for package '%s': %v\n", packageFilter, err)
					fmt.Println("No shipment history found. Use 'shipyard version' to ship some consignments first.")
					os.Exit(1)
				}
			} else {
				// For monorepo without package filter, generate separate changelogs for each package
				if projectConfig.Type == config.RepositoryTypeMonorepo {
					packageChangelogs, err = generator.GenerateChangelogsFromHistoryForPackages()
					if err != nil {
						logger.Error("Failed to generate package changelogs from history", "error", err)
						fmt.Printf("Error: Unable to generate package changelogs from history: %v\n", err)
						fmt.Println("No shipment history found. Use 'shipyard version' to ship some consignments first.")
						os.Exit(1)
					}
				} else {
					changelogContent, err = generator.GenerateChangelogFromHistory()
					if err != nil {
						logger.Error("Failed to generate changelog from history", "error", err)
						fmt.Printf("Error: Unable to generate changelog from history: %v\n", err)
						fmt.Println("No shipment history found. Use 'shipyard version' to ship some consignments first.")
						os.Exit(1)
					}
				}
			}
		} else {
			// Normal mode: prefer history, fallback to current consignments
			if packageFilter != "" {
				// Generate changelog for specific package from history first
				changelogContent, err = generator.GenerateChangelogFromHistoryForPackage(packageFilter)
				if err != nil {
					// If no history exists and we have current consignments, use them
					if hasCurrentConsignments {
						logger.Info("No shipment history found for package, using current consignments", "package", packageFilter)
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
						logger.Error("No shipment history or current consignments found for package", "package", packageFilter)
						fmt.Printf("Error: No shipment history found for package '%s' and no current consignments to process\n", packageFilter)
						fmt.Println("Create some consignments with 'shipyard add' first, or ensure shipment history exists.")
						os.Exit(1)
					}
				}
			} else {
				// For monorepo without package filter, generate separate changelogs for each package
				if projectConfig.Type == config.RepositoryTypeMonorepo {
					// Try to generate from history first
					packageChangelogs, err = generator.GenerateChangelogsFromHistoryForPackages()
					if err != nil {
						// If no history exists and we have current consignments, use them
						if hasCurrentConsignments {
							logger.Info("No shipment history found, using current consignments for package changelogs")
							packageChangelogs, err = generator.GenerateChangelogsForPackages(consignments, versions)
							if err != nil {
								logger.Error("Failed to generate package changelogs", "error", err)
								fmt.Printf("Error: Unable to generate package changelogs: %v\n", err)
								os.Exit(1)
							}
						} else {
							logger.Error("No shipment history or current consignments found for monorepo")
							fmt.Printf("Error: No shipment history found for monorepo and no current consignments to process\n")
							fmt.Println("Create some consignments with 'shipyard add' first, or ensure shipment history exists.")
							os.Exit(1)
						}
					}
				} else {
					// Generate changelog for all packages from history first
					changelogContent, err = generator.GenerateChangelogFromHistory()
					if err != nil {
						// If no history exists and we have current consignments, use them
						if hasCurrentConsignments {
							logger.Info("No shipment history found, using current consignments")
							changelogContent, err = generator.GenerateChangelog(consignments, versions)
							if err != nil {
								logger.Error("Failed to generate changelog", "error", err)
								fmt.Printf("Error: Unable to generate changelog: %v\n", err)
								os.Exit(1)
							}
						} else {
							logger.Error("No shipment history or current consignments found")
							fmt.Printf("Error: No shipment history found and no current consignments to process\n")
							fmt.Println("Create some consignments with 'shipyard add' first, or ensure shipment history exists.")
							os.Exit(1)
						}
					}
				}
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
				if packageChangelogs != nil {
					// For monorepo with separate package changelogs
					fmt.Println("📄 Generated Package Changelogs (Raw):")
					fmt.Println("======================================")
					fmt.Println()
					for packageName, changelog := range packageChangelogs {
						fmt.Printf("--- Package: %s ---\n", packageName)
						fmt.Print(changelog)
						fmt.Println()
					}
				} else {
					// For single repo or package-filtered changelog
					fmt.Println("📄 Generated Changelog (Raw):")
					fmt.Println("=============================")
					fmt.Println()
					fmt.Print(changelogContent)
				}
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
				fmt.Printf("📦 Consignments to be processed: %d\n", len(consignments))
				fmt.Printf("📝 Template: %s\n", projectConfig.Changelog.Template)
				if packageFilter != "" {
					fmt.Printf("🎯 Package filter: %s\n", packageFilter)
				}
				if packageChangelogs != nil {
					fmt.Printf("📚 Generated %d package changelogs\n", len(packageChangelogs))
				}
				fmt.Println()
				fmt.Println("💡 To apply these changes, run 'shipyard version' without --dry-run")
			} else {
				// Preview mode - pretty print just the new changelog
				if packageChangelogs != nil {
					// For monorepo with separate package changelogs
					fmt.Println("📄 Generated Package Changelogs Preview:")
					fmt.Println("========================================")
					fmt.Println()
					for packageName, changelog := range packageChangelogs {
						fmt.Printf("--- Package: %s ---\n", packageName)
						rendered, err := renderMarkdown(changelog)
						if err != nil {
							logger.Error("Failed to render markdown for package", "package", packageName, "error", err)
							fmt.Printf("Warning: Failed to render markdown for package %s, showing raw content: %v\n", packageName, err)
							fmt.Print(changelog)
						} else {
							fmt.Print(rendered)
						}
						fmt.Println("=====================================")
						fmt.Println()
					}
					fmt.Printf("📚 Generated %d package changelogs\n", len(packageChangelogs))
				} else {
					// For single repo or package-filtered changelog
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
			var confirmationMessage string
			if regenerate {
				confirmationMessage = "This will regenerate the changelog from shipment history. Are you sure?"
			} else if hasCurrentConsignments {
				confirmationMessage = fmt.Sprintf("This will generate changelog and update package versions, clearing %d consignments. Are you sure?", len(consignments))
			} else {
				confirmationMessage = "This will regenerate the changelog from shipment history. Are you sure?"
			}

			form := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title("Generate Changelog?").
						Description(confirmationMessage).
						Value(&confirm),
				),
			)

			if err := form.Run(); err != nil {
				logger.Error("Failed to get user confirmation", "error", err)
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			if !confirm {
				fmt.Println("Changelog generation cancelled.")
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

		// Write changelog to file(s)
		if packageChangelogs != nil {
			// For monorepo, write separate changelog files for each package
			fmt.Printf("\n📝 Writing package changelogs...\n")

			usePackagePaths := projectConfig.ShouldUsePackagePaths()
			defaultFilename := projectConfig.GetChangelogOutputPath()

			for packageName, changelog := range packageChangelogs {
				var packageOutputPath string

				if usePackagePaths {
					// Use package-specific paths
					pkg := projectConfig.GetPackageByName(packageName)
					if pkg != nil {
						packageOutputPath = pkg.GetChangelogPath(defaultFilename)
					} else {
						// Fallback if package not found
						packageOutputPath = filepath.Join(packageName, defaultFilename)
					}
				} else {
					// Use root directory with package suffix
					if outputPath == "" {
						outputPath = defaultFilename
					}
					packageOutputPath = fmt.Sprintf("%s.%s.md", strings.TrimSuffix(outputPath, ".md"), packageName)
				}

				// Ensure directory exists
				dir := filepath.Dir(packageOutputPath)
				if dir != "." {
					err := os.MkdirAll(dir, 0755)
					if err != nil {
						logger.Error("Failed to create directory for package changelog", "package", packageName, "dir", dir, "error", err)
						fmt.Printf("Error: Failed to create directory %s for package %s: %v\n", dir, packageName, err)
						os.Exit(1)
					}
				}

				err := os.WriteFile(packageOutputPath, []byte(changelog), 0644)
				if err != nil {
					logger.Error("Failed to write package changelog", "package", packageName, "path", packageOutputPath, "error", err)
					fmt.Printf("Error: Failed to write changelog for package %s to %s: %v\n", packageName, packageOutputPath, err)
					os.Exit(1)
				}
				fmt.Printf("   ✓ %s: %s\n", packageName, packageOutputPath)
			}
		} else {
			// For single repo or package-filtered changelog
			if outputPath == "" {
				outputPath = projectConfig.GetChangelogOutputPath()
			}
			fmt.Printf("\n📝 Writing changelog to %s...\n", outputPath)

			// Ensure directory exists
			dir := filepath.Dir(outputPath)
			if dir != "." {
				err := os.MkdirAll(dir, 0755)
				if err != nil {
					logger.Error("Failed to create directory for changelog", "dir", dir, "error", err)
					fmt.Printf("Error: Failed to create directory %s: %v\n", dir, err)
					os.Exit(1)
				}
			}

			err := os.WriteFile(outputPath, []byte(changelogContent), 0644)
			if err != nil {
				logger.Error("Failed to write changelog", "path", outputPath, "error", err)
				fmt.Printf("Error: Failed to write changelog to %s: %v\n", outputPath, err)
				os.Exit(1)
			}
			fmt.Printf("   ✓ Changelog written successfully\n")
		}

		// Apply version updates and record shipment history BEFORE writing changelog
		var appliedVersions map[string]*semver.Version
		if hasCurrentConsignments && !regenerate {
			fmt.Printf("\n🔄 Recording shipment history...\n")

			// Initialize git operations
			gitOps := git.NewGitOperations(projectConfig)

			// Check for git flags
			gitPush, _ := cmd.Flags().GetBool("git-push")
			gitTag, _ := cmd.Flags().GetBool("git-tag")

			// If git-push is enabled, enable tag as well
			if gitPush {
				gitTag = true
			}

			// Create git tags if requested and git is available
			var gitTags map[string]string
			if gitTag && gitOps.IsAvailable() {
				// Calculate versions first for tagging
				tempVersions, err := manager.CalculateAllVersions()
				if err != nil {
					logger.Error("Failed to calculate versions for git tags", "error", err)
					fmt.Printf("Error: Unable to calculate versions for git tags: %v\n", err)
					os.Exit(1)
				}

				// Get consignment summaries for commit message
				var consignmentSummaries []string
				for _, c := range consignments {
					consignmentSummaries = append(consignmentSummaries, c.Summary)
				}

				// Create git tags
				commitMessage := gitOps.CreateShipmentCommitMessage(tempVersions, consignmentSummaries)
				gitTags, err = gitOps.CreateShipmentTags(tempVersions, commitMessage)
				if err != nil {
					logger.Error("Failed to create git tags", "error", err)
					fmt.Printf("Warning: Failed to create git tags: %v\n", err)
					// Continue without git tags
					gitTags = nil
				} else {
					fmt.Printf("✅ Git tags created!\n")
				}
			}

			// Record shipment history with git tags BEFORE generating changelog
			templateName := projectConfig.Changelog.Template
			appliedVersions, err = manager.RecordShipmentHistoryWithTags(templateName, gitTags)
			if err != nil {
				logger.Error("Failed to record shipment history", "error", err)
				fmt.Printf("Error: Unable to record shipment history: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("✅ Shipment history recorded!\n")

			// Now regenerate changelog from history (which includes current consignments)
			fmt.Printf("\n🔄 Regenerating changelog from updated history...\n")
			if packageFilter != "" {
				changelogContent, err = generator.GenerateChangelogFromHistoryForPackage(packageFilter)
				if err != nil {
					logger.Error("Failed to generate changelog from history for package", "package", packageFilter, "error", err)
					fmt.Printf("Error: Unable to generate changelog from history for package '%s': %v\n", packageFilter, err)
					os.Exit(1)
				}
			} else {
				changelogContent, err = generator.GenerateChangelogFromHistory()
				if err != nil {
					logger.Error("Failed to generate changelog from history", "error", err)
					fmt.Printf("Error: Unable to generate changelog from history: %v\n", err)
					os.Exit(1)
				}
			}

			fmt.Printf("✅ Changelog regenerated from updated history!\n")
		} else if regenerate {
			fmt.Printf("\n� Changelog regenerated from shipment history (no version updates applied)\n")
			appliedVersions = make(map[string]*semver.Version)
		} else {
			fmt.Printf("\n📋 No current consignments to apply - changelog regenerated from shipment history\n")
			appliedVersions = make(map[string]*semver.Version)
		}

		// Write changelog to file
		if err := writeChangelog(outputPath, changelogContent); err != nil {
			logger.Error("Failed to write changelog", "error", err)
			fmt.Printf("Error: Unable to write changelog: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Changelog generated successfully!\n")
		fmt.Printf("📁 File: %s\n", outputPath)

		// Apply version updates and clear consignments AFTER changelog is written
		if hasCurrentConsignments && !regenerate {
			fmt.Printf("\n🔄 Applying version updates...\n")

			// Apply version updates and clear consignments
			err = manager.ApplyVersionUpdatesAndClearConsignments(appliedVersions)
			if err != nil {
				logger.Error("Failed to apply version updates", "error", err)
				fmt.Printf("Error: Unable to apply version updates: %v\n", err)
				os.Exit(1)
			}

			// Success message for version updates
			fmt.Printf("\n✅ Version updates applied successfully!\n")
			fmt.Printf("📦 Packages updated:\n")
			for pkgName, version := range appliedVersions {
				fmt.Printf("   - %s: %s\n", pkgName, version.String())
			}
			fmt.Printf("\n🧹 Consignments cleared: %d\n", len(consignments))

			// Perform git commit and push if requested
			gitCommit, _ := cmd.Flags().GetBool("git-commit")
			gitPush, _ := cmd.Flags().GetBool("git-push")

			// If git-push is enabled, enable commit as well
			if gitPush {
				gitCommit = true
			}

			if gitCommit || gitPush {
				gitOps := git.NewGitOperations(projectConfig)
				if gitOps.IsAvailable() {
					// Get consignment summaries for commit message
					var consignmentSummaries []string
					for _, c := range consignments {
						consignmentSummaries = append(consignmentSummaries, c.Summary)
					}

					// Determine output path for git operations
					changelogGitPath := outputPath
					if changelogGitPath == "" {
						changelogGitPath = "CHANGELOG.md"
					}

					err = gitOps.PerformShipmentGitOperations(appliedVersions, changelogGitPath, consignmentSummaries, gitCommit, gitPush)
					if err != nil {
						logger.Error("Failed to perform git operations", "error", err)
						fmt.Printf("Warning: Failed to perform git operations: %v\n", err)
						// Continue without git operations
					}
				} else {
					fmt.Printf("⚠️  Git repository not found - skipping git operations\n")
				}
			}
		}

		fmt.Printf("📝 Template: %s\n", projectConfig.Changelog.Template)
		if packageFilter != "" {
			fmt.Printf("🎯 Package filter: %s\n", packageFilter)
		}

		// Show next steps based on what was done
		gitCommit, _ := cmd.Flags().GetBool("git-commit")
		gitPush, _ := cmd.Flags().GetBool("git-push")
		gitTag, _ := cmd.Flags().GetBool("git-tag")

		if gitPush {
			gitCommit = true
			gitTag = true
		}

		fmt.Printf("\n💡 Next steps:\n")
		fmt.Printf("   - Review the generated changelog\n")
		if hasCurrentConsignments && !regenerate {
			fmt.Printf("   - Review updated package manifests\n")
		}

		// Only show git-related next steps if git operations weren't performed
		if !gitCommit {
			fmt.Printf("   - Commit all changes to your repository\n")
		}
		if hasCurrentConsignments && !regenerate && !gitTag {
			fmt.Printf("   - Create and push git tags for the new versions\n")
		}

		logger.Info("Changelog generation completed successfully",
			"output", outputPath,
			"template", projectConfig.Changelog.Template,
			"packages_updated", len(appliedVersions),
			"consignments_cleared", len(consignments),
			"regenerate_mode", regenerate,
			"from_history", regenerate || !hasCurrentConsignments,
		)
	},
}

func init() {
	// Add flags for version command
	VersionCmd.Flags().Bool("dry-run", false, "Show changelog and version info without applying changes")
	VersionCmd.Flags().Bool("preview", false, "Pretty print just the new changelog without applying changes")
	VersionCmd.Flags().Bool("regenerate", false, "Regenerate changelog from shipment history without applying version updates")
	VersionCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompts")
	VersionCmd.Flags().StringP("package", "p", "", "Generate changelog and version for a specific package only")
	VersionCmd.Flags().StringP("output", "o", "", "Output file path for changelog (default: CHANGELOG.md)")
	VersionCmd.Flags().StringP("template", "t", "", "Override the changelog template")

	// Git-related flags
	VersionCmd.Flags().Bool("git-tag", false, "Create git tags for released versions")
	VersionCmd.Flags().Bool("git-commit", false, "Automatically commit changelog and version changes")
	VersionCmd.Flags().Bool("git-push", false, "Automatically push commits and tags to remote (implies --git-commit and --git-tag)")
}
