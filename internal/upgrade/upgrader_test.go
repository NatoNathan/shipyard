package upgrade

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/NatoNathan/shipyard/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUpgrader(t *testing.T) {
	log := logger.New(os.Stdout, logger.LevelInfo, false)

	tests := []struct {
		name        string
		method      InstallMethod
		expectError bool
		upgraderType string
	}{
		{
			name:         "creates HomebrewUpgrader",
			method:       InstallMethodHomebrew,
			expectError:  false,
			upgraderType: "*upgrade.HomebrewUpgrader",
		},
		{
			name:         "creates NPMUpgrader",
			method:       InstallMethodNPM,
			expectError:  false,
			upgraderType: "*upgrade.NPMUpgrader",
		},
		{
			name:         "creates GoUpgrader",
			method:       InstallMethodGo,
			expectError:  false,
			upgraderType: "*upgrade.GoUpgrader",
		},
		{
			name:         "creates ScriptUpgrader",
			method:       InstallMethodScript,
			expectError:  false,
			upgraderType: "*upgrade.ScriptUpgrader",
		},
		{
			name:        "rejects Docker method",
			method:      InstallMethodDocker,
			expectError: true,
		},
		{
			name:        "rejects Unknown method",
			method:      InstallMethodUnknown,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &InstallInfo{
				Method:     tt.method,
				BinaryPath: "/test/path",
			}

			upgrader, err := NewUpgrader(info, log)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, upgrader)
			} else {
				require.NoError(t, err)
				require.NotNil(t, upgrader)
				assert.Contains(t, fmt.Sprintf("%T", upgrader), tt.upgraderType)
			}
		})
	}
}

func TestHomebrewUpgrader_GetUpgradeCommand(t *testing.T) {
	log := logger.New(os.Stdout, logger.LevelInfo, true)
	upgrader := &HomebrewUpgrader{log: log}

	cmd := upgrader.GetUpgradeCommand()
	assert.Equal(t, "brew upgrade natonathan/tap/shipyard", cmd)
}

func TestNPMUpgrader_GetUpgradeCommand(t *testing.T) {
	log := logger.New(os.Stdout, logger.LevelInfo, true)
	upgrader := &NPMUpgrader{log: log}

	cmd := upgrader.GetUpgradeCommand()
	assert.Equal(t, "npm update -g shipyard-cli", cmd)
}

func TestGoUpgrader_GetUpgradeCommand(t *testing.T) {
	log := logger.New(os.Stdout, logger.LevelInfo, true)
	upgrader := &GoUpgrader{log: log}

	cmd := upgrader.GetUpgradeCommand()
	assert.Equal(t, "go install github.com/NatoNathan/shipyard/cmd/shipyard@latest", cmd)
}

func TestScriptUpgrader_GetUpgradeCommand(t *testing.T) {
	log := logger.New(os.Stdout, logger.LevelInfo, true)
	upgrader := &ScriptUpgrader{
		binaryPath: "/usr/local/bin/shipyard",
		log:        log,
	}

	cmd := upgrader.GetUpgradeCommand()
	assert.Equal(t, "shipyard upgrade", cmd)
}

func TestScriptUpgrader_VerifyChecksum(t *testing.T) {
	log := logger.New(os.Stdout, logger.LevelInfo, true)
	upgrader := &ScriptUpgrader{log: log}

	testData := []byte("test data")
	hash := sha256.Sum256(testData)
	correctChecksum := hex.EncodeToString(hash[:])

	t.Run("valid checksum", func(t *testing.T) {
		checksums := fmt.Sprintf("%s  test.tar.gz\n", correctChecksum)
		err := upgrader.verifyChecksum(testData, "test.tar.gz", []byte(checksums))
		assert.NoError(t, err)
	})

	t.Run("invalid checksum", func(t *testing.T) {
		checksums := "0000000000000000000000000000000000000000000000000000000000000000  test.tar.gz\n"
		err := upgrader.verifyChecksum(testData, "test.tar.gz", []byte(checksums))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "checksum mismatch")
	})

	t.Run("checksum not found", func(t *testing.T) {
		checksums := fmt.Sprintf("%s  other.tar.gz\n", correctChecksum)
		err := upgrader.verifyChecksum(testData, "test.tar.gz", []byte(checksums))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "checksum not found")
	})

	t.Run("malformed checksums file", func(t *testing.T) {
		checksums := "invalid format\n"
		err := upgrader.verifyChecksum(testData, "test.tar.gz", []byte(checksums))
		assert.Error(t, err)
	})
}

func TestScriptUpgrader_AtomicReplace(t *testing.T) {
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "shipyard")

	log := logger.New(os.Stdout, logger.LevelInfo, true)
	upgrader := &ScriptUpgrader{
		binaryPath: binaryPath,
		log:        log,
	}

	t.Run("replaces binary atomically", func(t *testing.T) {
		// Create original binary
		originalContent := []byte("original binary")
		err := os.WriteFile(binaryPath, originalContent, 0755)
		require.NoError(t, err)

		// Replace with new binary
		newContent := []byte("new binary")
		err = upgrader.atomicReplace(newContent)
		require.NoError(t, err)

		// Verify new content
		actualContent, err := os.ReadFile(binaryPath)
		require.NoError(t, err)
		assert.Equal(t, newContent, actualContent)

		// Verify permissions
		info, err := os.Stat(binaryPath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0755), info.Mode().Perm())
	})

	t.Run("removes backup on success", func(t *testing.T) {
		err := os.WriteFile(binaryPath, []byte("test"), 0755)
		require.NoError(t, err)

		err = upgrader.atomicReplace([]byte("new"))
		require.NoError(t, err)

		backupPath := binaryPath + ".old"
		_, err = os.Stat(backupPath)
		assert.True(t, os.IsNotExist(err), "backup should be removed")
	})
}

func TestScriptUpgrader_Upgrade(t *testing.T) {
	// Create a mock HTTP server
	platform := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
	tarballName := fmt.Sprintf("shipyard_v1.0.0_%s.tar.gz", platform)

	// Create a minimal tar.gz for testing
	tmpDir := t.TempDir()
	testBinaryPath := filepath.Join(tmpDir, "shipyard")
	testBinaryContent := []byte("test binary content")
	err := os.WriteFile(testBinaryPath, testBinaryContent, 0755)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/checksums.txt" {
			// Calculate checksum of test binary
			hash := sha256.Sum256(testBinaryContent)
			checksum := hex.EncodeToString(hash[:])
			fmt.Fprintf(w, "%s  %s\n", checksum, tarballName)
		} else if r.URL.Path == "/"+tarballName {
			w.Write(testBinaryContent) // Simplified - just write the binary
		}
	}))
	defer server.Close()

	log := logger.New(os.Stdout, logger.LevelInfo, true)

	t.Run("handles missing asset", func(t *testing.T) {
		binaryPath := filepath.Join(t.TempDir(), "shipyard")
		upgrader := &ScriptUpgrader{
			binaryPath: binaryPath,
			log:        log,
		}

		release := &ReleaseInfo{
			TagName: "v1.0.0",
			Assets:  []ReleaseAsset{}, // No assets
		}

		err := upgrader.Upgrade(context.Background(), release)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no release asset found")
	})

	t.Run("handles missing checksums", func(t *testing.T) {
		binaryPath := filepath.Join(t.TempDir(), "shipyard")
		upgrader := &ScriptUpgrader{
			binaryPath: binaryPath,
			log:        log,
		}

		release := &ReleaseInfo{
			TagName: "v1.0.0",
			Assets: []ReleaseAsset{
				{
					Name:        tarballName,
					DownloadURL: server.URL + "/" + tarballName,
				},
				// No checksums.txt
			},
		}

		err := upgrader.Upgrade(context.Background(), release)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no checksums.txt found")
	})
}

func TestScriptUpgrader_DownloadFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/test" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test content"))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	log := logger.New(os.Stdout, logger.LevelInfo, true)
	upgrader := &ScriptUpgrader{log: log}

	t.Run("downloads file successfully", func(t *testing.T) {
		data, err := upgrader.downloadFile(context.Background(), server.URL+"/test")
		require.NoError(t, err)
		assert.Equal(t, []byte("test content"), data)
	})

	t.Run("handles 404", func(t *testing.T) {
		data, err := upgrader.downloadFile(context.Background(), server.URL+"/notfound")
		assert.Error(t, err)
		assert.Nil(t, data)
		assert.Contains(t, err.Error(), "404")
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		data, err := upgrader.downloadFile(ctx, server.URL+"/test")
		assert.Error(t, err)
		assert.Nil(t, data)
	})
}
