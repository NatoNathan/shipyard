package upgrade

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/NatoNathan/shipyard/internal/logger"
)

// Upgrader defines the interface for upgrade strategies
type Upgrader interface {
	Upgrade(ctx context.Context, release *ReleaseInfo) error
	GetUpgradeCommand() string
}

// NewUpgrader creates an appropriate upgrader based on installation method
func NewUpgrader(info *InstallInfo, log *logger.Logger) (Upgrader, error) {
	switch info.Method {
	case InstallMethodHomebrew:
		return &HomebrewUpgrader{log: log}, nil
	case InstallMethodNPM:
		return &NPMUpgrader{log: log}, nil
	case InstallMethodGo:
		return &GoUpgrader{log: log}, nil
	case InstallMethodScript:
		return &ScriptUpgrader{
			binaryPath: info.BinaryPath,
			log:        log,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported installation method: %s", info.Method)
	}
}

// HomebrewUpgrader handles upgrades via Homebrew
type HomebrewUpgrader struct {
	log *logger.Logger
}

func (u *HomebrewUpgrader) Upgrade(ctx context.Context, release *ReleaseInfo) error {
	cmd := exec.CommandContext(ctx, "brew", "upgrade", "natonathan/tap/shipyard")
	cmd.Stdout = u.log.Writer()
	cmd.Stderr = u.log.Writer()
	return cmd.Run()
}

func (u *HomebrewUpgrader) GetUpgradeCommand() string {
	return "brew upgrade natonathan/tap/shipyard"
}

// NPMUpgrader handles upgrades via npm
type NPMUpgrader struct {
	log *logger.Logger
}

func (u *NPMUpgrader) Upgrade(ctx context.Context, release *ReleaseInfo) error {
	cmd := exec.CommandContext(ctx, "npm", "update", "-g", "shipyard-cli")
	cmd.Stdout = u.log.Writer()
	cmd.Stderr = u.log.Writer()
	return cmd.Run()
}

func (u *NPMUpgrader) GetUpgradeCommand() string {
	return "npm update -g shipyard-cli"
}

// GoUpgrader handles upgrades via go install
type GoUpgrader struct {
	log *logger.Logger
}

func (u *GoUpgrader) Upgrade(ctx context.Context, release *ReleaseInfo) error {
	version := release.TagName
	installPath := fmt.Sprintf("github.com/NatoNathan/shipyard/cmd/shipyard@%s", version)

	cmd := exec.CommandContext(ctx, "go", "install", installPath)
	cmd.Stdout = u.log.Writer()
	cmd.Stderr = u.log.Writer()
	return cmd.Run()
}

func (u *GoUpgrader) GetUpgradeCommand() string {
	return "go install github.com/NatoNathan/shipyard/cmd/shipyard@latest"
}

// ScriptUpgrader handles upgrades for script/manual installations
type ScriptUpgrader struct {
	binaryPath string
	log        *logger.Logger
}

func (u *ScriptUpgrader) Upgrade(ctx context.Context, release *ReleaseInfo) error {
	// Determine platform string
	platform := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)

	// Find the appropriate asset
	var tarballAsset *ReleaseAsset
	var checksumAsset *ReleaseAsset

	for i := range release.Assets {
		asset := &release.Assets[i]
		if strings.Contains(asset.Name, platform) && strings.HasSuffix(asset.Name, ".tar.gz") {
			tarballAsset = asset
		}
		if asset.Name == "checksums.txt" {
			checksumAsset = asset
		}
	}

	if tarballAsset == nil {
		return fmt.Errorf("no release asset found for platform %s", platform)
	}

	if checksumAsset == nil {
		return fmt.Errorf("no checksums.txt found in release")
	}

	// Download checksum file
	checksums, err := u.downloadFile(ctx, checksumAsset.DownloadURL)
	if err != nil {
		return fmt.Errorf("failed to download checksums: %w", err)
	}

	// Download tarball
	u.log.Info("Downloading %s...", tarballAsset.Name)
	tarballData, err := u.downloadFile(ctx, tarballAsset.DownloadURL)
	if err != nil {
		return fmt.Errorf("failed to download release: %w", err)
	}

	// Verify checksum
	if err := u.verifyChecksum(tarballData, tarballAsset.Name, checksums); err != nil {
		return fmt.Errorf("checksum verification failed: %w", err)
	}

	// Extract binary
	u.log.Info("Extracting binary...")
	binaryData, err := u.extractBinary(tarballData)
	if err != nil {
		return fmt.Errorf("failed to extract binary: %w", err)
	}

	// Atomic replacement
	u.log.Info("Installing binary...")
	if err := u.atomicReplace(binaryData); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	return nil
}

func (u *ScriptUpgrader) GetUpgradeCommand() string {
	return "shipyard upgrade"
}

// downloadFile downloads a file from a URL
func (u *ScriptUpgrader) downloadFile(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// verifyChecksum verifies the SHA256 checksum of the downloaded file
func (u *ScriptUpgrader) verifyChecksum(data []byte, filename string, checksums []byte) error {
	// Calculate actual checksum
	hash := sha256.Sum256(data)
	actualChecksum := hex.EncodeToString(hash[:])

	// Parse checksums file
	lines := strings.Split(string(checksums), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}
		if parts[1] == filename {
			expectedChecksum := parts[0]
			if actualChecksum != expectedChecksum {
				return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
			}
			return nil
		}
	}

	return fmt.Errorf("checksum not found for %s", filename)
}

// extractBinary extracts the binary from a tar.gz archive
func (u *ScriptUpgrader) extractBinary(tarballData []byte) ([]byte, error) {
	// Create temp file for tarball
	tmpDir, err := os.MkdirTemp("", "shipyard-upgrade-")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	tarballPath := filepath.Join(tmpDir, "shipyard.tar.gz")
	if err := os.WriteFile(tarballPath, tarballData, 0644); err != nil {
		return nil, err
	}

	// Extract tarball
	cmd := exec.Command("tar", "-xzf", tarballPath, "-C", tmpDir)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("tar extraction failed: %w", err)
	}

	// Find the binary
	binaryPath := filepath.Join(tmpDir, "shipyard")
	data, err := os.ReadFile(binaryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read extracted binary: %w", err)
	}

	return data, nil
}

// atomicReplace replaces the current binary with the new one atomically
func (u *ScriptUpgrader) atomicReplace(newBinary []byte) error {
	dir := filepath.Dir(u.binaryPath)
	name := filepath.Base(u.binaryPath)

	// Create temp file in same directory
	tmpFile, err := os.CreateTemp(dir, "."+name+"-new-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() { _ = os.Remove(tmpPath) }() // Clean up if something fails

	// Write new binary
	if _, err := tmpFile.Write(newBinary); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("failed to write new binary: %w", err)
	}
	_ = tmpFile.Close()

	// Make executable
	if err := os.Chmod(tmpPath, 0755); err != nil {
		return fmt.Errorf("failed to chmod new binary: %w", err)
	}

	// Backup old binary
	backupPath := u.binaryPath + ".old"
	if err := os.Rename(u.binaryPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup old binary: %w", err)
	}

	// Move new binary into place
	if err := os.Rename(tmpPath, u.binaryPath); err != nil {
		// Try to restore backup
		_ = os.Rename(backupPath, u.binaryPath)
		return fmt.Errorf("failed to install new binary: %w", err)
	}

	// Remove backup on success
	_ = os.Remove(backupPath)

	return nil
}
