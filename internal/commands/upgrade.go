package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/NatoNathan/shipyard/internal/errors"
	"github.com/NatoNathan/shipyard/internal/logger"
	"github.com/NatoNathan/shipyard/internal/prompt"
	"github.com/NatoNathan/shipyard/internal/ui"
	"github.com/NatoNathan/shipyard/internal/upgrade"
	"github.com/spf13/cobra"
)

// UpgradeOptions contains options for the upgrade command
type UpgradeOptions struct {
	Yes     bool   // Skip confirmation
	Version string // Specific version (default: latest)
	Force   bool   // Upgrade even if on latest
	DryRun  bool   // Show plan without executing
}

// VersionInfo contains build version information
type VersionInfo struct {
	Version string
	Commit  string
	Date    string
}

// NewUpgradeCommand creates the upgrade command
func NewUpgradeCommand(versionInfo VersionInfo) *cobra.Command {
	opts := &UpgradeOptions{}

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade shipyard to the latest version",
		Long: `Upgrade shipyard to the latest version.

This command automatically detects how shipyard was installed (Homebrew, npm, Go install,
or script install) and uses the appropriate upgrade method.

Examples:
  # Upgrade with confirmation prompt
  shipyard upgrade

  # Upgrade without confirmation
  shipyard upgrade --yes

  # Show what would happen without upgrading
  shipyard upgrade --dry-run

  # Force reinstall of current version
  shipyard upgrade --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpgrade(cmd.Context(), opts, versionInfo)
		},
	}

	cmd.Flags().BoolVarP(&opts.Yes, "yes", "y", false, "Skip confirmation prompt")
	cmd.Flags().StringVar(&opts.Version, "version", "", "Upgrade to specific version (default: latest)")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "Force upgrade even if already on latest version")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Show upgrade plan without executing")

	return cmd
}

func runUpgrade(ctx context.Context, opts *UpgradeOptions, versionInfo VersionInfo) error {
	log := logger.Get()

	// Step 1: Detect installation
	var installInfo *upgrade.InstallInfo
	var detectErr error

	spinner := ui.NewSpinner("Checking installation...")
	spinner.Action(func() {
		installInfo, detectErr = upgrade.DetectInstallation(versionInfo.Version, versionInfo.Commit, versionInfo.Date)
	})

	if err := spinner.Run(); err != nil {
		return err
	}

	if detectErr != nil {
		return errors.NewUpgradeError("failed to detect installation", detectErr)
	}

	// Step 2: Check if we can upgrade
	if !installInfo.CanUpgrade {
		fmt.Println(ui.ErrorMessage(fmt.Sprintf("Cannot upgrade: %s", installInfo.Reason)))
		if installInfo.Method == upgrade.InstallMethodDocker {
			fmt.Println("\nTo upgrade Docker installations:")
			fmt.Println("  docker pull natonathan/shipyard:latest")
		} else if installInfo.Method == upgrade.InstallMethodUnknown {
			fmt.Println("\nManual upgrade instructions:")
			fmt.Println("  Visit https://github.com/NatoNathan/shipyard/releases/latest")
		}
		return nil
	}

	// Step 3: Fetch latest release
	var release *upgrade.ReleaseInfo
	var fetchErr error

	client := upgrade.NewGitHubClient()
	spinner = ui.NewSpinner("Checking for updates...")
	spinner.Action(func() {
		release, fetchErr = client.GetLatestRelease(ctx, "NatoNathan", "shipyard")
	})

	if err := spinner.Run(); err != nil {
		return err
	}

	if fetchErr != nil {
		return errors.NewNetworkError("failed to fetch latest release", fetchErr)
	}

	// Step 4: Compare versions
	isNewer, err := upgrade.IsNewer(versionInfo.Version, release.TagName)
	if err != nil {
		log.Warn("Failed to compare versions: %v", err)
		// Continue anyway, let user decide
	}

	if !isNewer && !opts.Force {
		fmt.Println(ui.SuccessMessage(fmt.Sprintf("Already on latest version %s", versionInfo.Version)))
		return nil
	}

	// Step 5: Display information
	fmt.Println()
	fmt.Println(ui.KeyValue("Current Version", versionInfo.Version))
	fmt.Println(ui.KeyValue("Latest Version", release.TagName))
	fmt.Println(ui.KeyValue("Installation Method", installInfo.Method.String()))
	fmt.Println(ui.KeyValue("Binary Path", installInfo.BinaryPath))

	if release.Body != "" {
		fmt.Println()
		fmt.Println(ui.Section("Release Notes"))
		releaseNotes := release.Body
		if len(releaseNotes) > 500 {
			releaseNotes = releaseNotes[:497] + "..."
		}
		fmt.Println(releaseNotes)
	}

	fmt.Println()

	// Dry run exit
	if opts.DryRun {
		upgrader, err := upgrade.NewUpgrader(installInfo, log)
		if err != nil {
			return errors.NewUpgradeError("failed to create upgrader", err)
		}

		fmt.Println(ui.Section("Upgrade Command"))
		fmt.Println(upgrader.GetUpgradeCommand())
		return nil
	}

	// Step 6: Confirmation prompt
	if !opts.Yes {
		confirmed, err := prompt.PromptConfirm(
			fmt.Sprintf("Upgrade to %s?", release.TagName),
			true,
		)
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Println(ui.InfoMessage("Upgrade cancelled"))
			return nil
		}
	}

	// Step 7: Execute upgrade
	upgrader, err := upgrade.NewUpgrader(installInfo, log)
	if err != nil {
		return errors.NewUpgradeError("failed to create upgrader", err)
	}

	var upgradeErr error

	spinner = ui.NewSpinner(fmt.Sprintf("Upgrading to %s...", release.TagName))
	spinner.Action(func() {
		// Use a timeout context for the upgrade
		upgradeCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()

		upgradeErr = upgrader.Upgrade(upgradeCtx, release)
	})

	if err := spinner.Run(); err != nil {
		return err
	}

	if upgradeErr != nil {
		fmt.Println(ui.ErrorMessage(fmt.Sprintf("Upgrade failed: %v", upgradeErr)))
		fmt.Println("\nYou can try upgrading manually:")
		fmt.Printf("  %s\n", upgrader.GetUpgradeCommand())
		return errors.NewUpgradeError("upgrade failed", upgradeErr)
	}

	// Step 8: Success message
	fmt.Println()
	fmt.Println(ui.SuccessMessage(fmt.Sprintf("Successfully upgraded to %s", release.TagName)))
	fmt.Println(ui.InfoMessage("Run 'shipyard --version' to verify the new version"))

	return nil
}

// truncateString truncates a string to the specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return strings.TrimSpace(s[:maxLen-3]) + "..."
}
