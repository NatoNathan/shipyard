package upgrade

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallMethod_String(t *testing.T) {
	tests := []struct {
		method   InstallMethod
		expected string
	}{
		{InstallMethodHomebrew, "Homebrew"},
		{InstallMethodNPM, "npm"},
		{InstallMethodGo, "Go install"},
		{InstallMethodScript, "Script install"},
		{InstallMethodDocker, "Docker"},
		{InstallMethodUnknown, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.method.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsHomebrew(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "homebrew on macOS",
			path:     "/opt/homebrew/bin/shipyard",
			expected: true,
		},
		{
			name:     "homebrew on Linux",
			path:     "/home/linuxbrew/.linuxbrew/bin/shipyard",
			expected: true,
		},
		{
			name:     "homebrew alt path",
			path:     "/usr/local/homebrew/bin/shipyard",
			expected: true,
		},
		{
			name:     "not homebrew",
			path:     "/usr/local/bin/shipyard",
			expected: false,
		},
		{
			name:     "go install path",
			path:     "/home/user/go/bin/shipyard",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isHomebrew(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsGoInstall(t *testing.T) {
	// Save original env
	origGOBIN := os.Getenv("GOBIN")
	origGOPATH := os.Getenv("GOPATH")
	defer func() {
		os.Setenv("GOBIN", origGOBIN)
		os.Setenv("GOPATH", origGOPATH)
	}()

	t.Run("GOBIN path", func(t *testing.T) {
		os.Setenv("GOBIN", "/custom/gobin")
		result := isGoInstall("/custom/gobin/shipyard")
		assert.True(t, result)
	})

	t.Run("GOPATH bin path", func(t *testing.T) {
		os.Setenv("GOBIN", "")
		os.Setenv("GOPATH", "/home/user/gopath")
		result := isGoInstall("/home/user/gopath/bin/shipyard")
		assert.True(t, result)
	})

	t.Run("default GOPATH", func(t *testing.T) {
		os.Setenv("GOBIN", "")
		os.Setenv("GOPATH", "")
		home, err := os.UserHomeDir()
		require.NoError(t, err)

		defaultPath := filepath.Join(home, "go", "bin", "shipyard")
		result := isGoInstall(defaultPath)
		assert.True(t, result)
	})

	t.Run("not go install", func(t *testing.T) {
		os.Setenv("GOBIN", "")
		os.Setenv("GOPATH", "")
		result := isGoInstall("/usr/local/bin/shipyard")
		assert.False(t, result)
	})
}

func TestCanWriteFile(t *testing.T) {
	t.Run("writable file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test")

		// Create a writable file
		err := os.WriteFile(testFile, []byte("test"), 0644)
		require.NoError(t, err)

		result := canWriteFile(testFile)
		assert.True(t, result)
	})

	t.Run("readonly file", func(t *testing.T) {
		// Skip on root user (e.g., in Docker containers) - root can write to any file
		if runtime.GOOS != "windows" && os.Geteuid() == 0 {
			t.Skip("skipping readonly test when running as root")
		}

		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "readonly")

		// Create a readonly file
		err := os.WriteFile(testFile, []byte("test"), 0444)
		require.NoError(t, err)

		result := canWriteFile(testFile)
		assert.False(t, result)
	})

	t.Run("nonexistent file", func(t *testing.T) {
		result := canWriteFile("/nonexistent/path/file")
		assert.False(t, result)
	})
}

func TestDetectInstallation(t *testing.T) {
	version := "1.0.0"
	commit := "abc123"
	date := "2024-01-01"

	t.Run("returns InstallInfo with version data", func(t *testing.T) {
		info, err := DetectInstallation(version, commit, date)
		require.NoError(t, err)
		require.NotNil(t, info)

		assert.Equal(t, version, info.Version)
		assert.Equal(t, commit, info.Commit)
		assert.Equal(t, date, info.Date)
		assert.NotEmpty(t, info.BinaryPath)
	})

	t.Run("detects installation method", func(t *testing.T) {
		info, err := DetectInstallation(version, commit, date)
		require.NoError(t, err)

		// Should detect some method (actual method depends on test environment)
		assert.NotEqual(t, InstallMethod(0), info.Method)
	})

	t.Run("sets CanUpgrade based on permissions", func(t *testing.T) {
		info, err := DetectInstallation(version, commit, date)
		require.NoError(t, err)

		// CanUpgrade should be set based on method and permissions
		if info.Method == InstallMethodDocker {
			assert.False(t, info.CanUpgrade)
			assert.Contains(t, info.Reason, "Docker")
		} else if info.Method == InstallMethodUnknown {
			// Unknown method may or may not be upgradeable
			if !info.CanUpgrade {
				assert.NotEmpty(t, info.Reason)
			}
		}
	})
}

func TestIsDocker(t *testing.T) {
	t.Run("not in docker by default", func(t *testing.T) {
		// This test assumes we're not running in Docker
		// In a real Docker environment, this would be true
		result := isDocker()

		// We can't assert a specific value since it depends on environment
		// Just verify it returns a boolean without panicking
		assert.IsType(t, false, result)
	})
}

func TestIsNPM(t *testing.T) {
	t.Run("returns boolean without error", func(t *testing.T) {
		// This test just verifies the function doesn't panic
		// Actual result depends on whether npm is installed and has shipyard-cli
		result := isNPM()
		assert.IsType(t, false, result)
	})
}
