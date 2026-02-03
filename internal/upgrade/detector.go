package upgrade

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DetectInstallation determines how shipyard was installed
func DetectInstallation(version, commit, date string) (*InstallInfo, error) {
	info := &InstallInfo{
		Method:     InstallMethodUnknown,
		Version:    version,
		Commit:     commit,
		Date:       date,
		CanUpgrade: false,
	}

	// Get binary path
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks
	realPath, err := filepath.EvalSymlinks(exePath)
	if err != nil {
		realPath = exePath
	}
	info.BinaryPath = realPath

	// Check if running in Docker
	if isDocker() {
		info.Method = InstallMethodDocker
		info.Reason = "Cannot upgrade inside Docker container. Pull a new image instead."
		return info, nil
	}

	// Check for Homebrew
	if isHomebrew(realPath) {
		info.Method = InstallMethodHomebrew
		info.CanUpgrade = true
		return info, nil
	}

	// Check for npm
	if isNPM() {
		info.Method = InstallMethodNPM
		info.CanUpgrade = true
		return info, nil
	}

	// Check for Go install
	if isGoInstall(realPath) {
		info.Method = InstallMethodGo
		info.CanUpgrade = true
		return info, nil
	}

	// Default to script/manual install
	info.Method = InstallMethodScript

	// Check if we have write permissions
	if canWriteFile(realPath) {
		info.CanUpgrade = true
	} else {
		info.Reason = fmt.Sprintf("No write permission for %s. Try running with sudo or move the binary to a writable location.", realPath)
	}

	return info, nil
}

// isDocker checks if running inside a Docker container
func isDocker() bool {
	// Check for .dockerenv file
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// Check cgroup
	data, err := os.ReadFile("/proc/1/cgroup")
	if err == nil && strings.Contains(string(data), "docker") {
		return true
	}

	return false
}

// isHomebrew checks if the binary is in a Homebrew path
func isHomebrew(path string) bool {
	homebrewPaths := []string{
		"/homebrew/",
		"/linuxbrew/",
		"/opt/homebrew/",
	}

	for _, prefix := range homebrewPaths {
		if strings.Contains(path, prefix) {
			return true
		}
	}

	return false
}

// isNPM checks if shipyard was installed via npm
func isNPM() bool {
	cmd := exec.Command("npm", "list", "-g", "shipyard-cli", "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}

	var result struct {
		Dependencies map[string]interface{} `json:"dependencies"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return false
	}

	_, exists := result.Dependencies["shipyard-cli"]
	return exists
}

// isGoInstall checks if the binary is in a Go path
func isGoInstall(path string) bool {
	// Check GOBIN
	if gobin := os.Getenv("GOBIN"); gobin != "" {
		if strings.HasPrefix(path, gobin) {
			return true
		}
	}

	// Check GOPATH/bin
	if gopath := os.Getenv("GOPATH"); gopath != "" {
		goBin := filepath.Join(gopath, "bin")
		if strings.HasPrefix(path, goBin) {
			return true
		}
	}

	// Check default GOPATH
	if home, err := os.UserHomeDir(); err == nil {
		defaultGoBin := filepath.Join(home, "go", "bin")
		if strings.HasPrefix(path, defaultGoBin) {
			return true
		}
	}

	return false
}

// canWriteFile checks if we have write permission to a file
func canWriteFile(path string) bool {
	// Try to open for writing
	file, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return false
	}
	file.Close()
	return true
}
