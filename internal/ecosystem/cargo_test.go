package ecosystem

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/NatoNathan/shipyard/pkg/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCargoEcosystem_ReadVersion tests reading version from Cargo.toml
func TestCargoEcosystem_ReadVersion(t *testing.T) {
	t.Run("reads version from Cargo.toml", func(t *testing.T) {
		// Setup: Create Cargo.toml
		tempDir := t.TempDir()
		cargoPath := filepath.Join(tempDir, "Cargo.toml")
		cargoContent := `[package]
name = "my-rust-project"
version = "1.2.3"
edition = "2021"
authors = ["John Doe <john@example.com>"]

[dependencies]
serde = "1.0"
`
		require.NoError(t, os.WriteFile(cargoPath, []byte(cargoContent), 0644))

		// Test: Read version
		cargo := NewCargoEcosystem(tempDir)
		version, err := cargo.ReadVersion()

		// Verify: Version read correctly
		require.NoError(t, err)
		assert.Equal(t, "1.2.3", version.String())
	})

	t.Run("returns error for missing Cargo.toml", func(t *testing.T) {
		// Setup: Empty directory
		tempDir := t.TempDir()

		// Test: Read version
		cargo := NewCargoEcosystem(tempDir)
		_, err := cargo.ReadVersion()

		// Verify: Error returned
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Cargo.toml")
	})

	t.Run("returns error for invalid TOML", func(t *testing.T) {
		// Setup: Create invalid Cargo.toml
		tempDir := t.TempDir()
		cargoPath := filepath.Join(tempDir, "Cargo.toml")
		require.NoError(t, os.WriteFile(cargoPath, []byte("[invalid toml"), 0644))

		// Test: Read version
		cargo := NewCargoEcosystem(tempDir)
		_, err := cargo.ReadVersion()

		// Verify: Error returned
		assert.Error(t, err)
	})

	t.Run("returns error for missing version field", func(t *testing.T) {
		// Setup: Create Cargo.toml without version
		tempDir := t.TempDir()
		cargoPath := filepath.Join(tempDir, "Cargo.toml")
		cargoContent := `[package]
name = "my-project"
`
		require.NoError(t, os.WriteFile(cargoPath, []byte(cargoContent), 0644))

		// Test: Read version
		cargo := NewCargoEcosystem(tempDir)
		_, err := cargo.ReadVersion()

		// Verify: Error returned
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "version")
	})
}

// TestCargoEcosystem_UpdateVersion tests updating version in Cargo.toml
func TestCargoEcosystem_UpdateVersion(t *testing.T) {
	t.Run("updates version in Cargo.toml", func(t *testing.T) {
		// Setup: Create Cargo.toml
		tempDir := t.TempDir()
		cargoPath := filepath.Join(tempDir, "Cargo.toml")
		cargoContent := `[package]
name = "my-rust-project"
version = "1.2.3"
edition = "2021"

[dependencies]
serde = "1.0"
`
		require.NoError(t, os.WriteFile(cargoPath, []byte(cargoContent), 0644))

		// Test: Update version
		cargo := NewCargoEcosystem(tempDir)
		newVersion := semver.MustParse("2.0.0")
		err := cargo.UpdateVersion(newVersion)

		// Verify: Version updated
		require.NoError(t, err)

		// Read back and verify
		updatedVersion, err := cargo.ReadVersion()
		require.NoError(t, err)
		assert.Equal(t, "2.0.0", updatedVersion.String())

		// Verify other fields preserved
		content, _ := os.ReadFile(cargoPath)
		contentStr := string(content)
		assert.Contains(t, contentStr, "name")
		assert.Contains(t, contentStr, "my-rust-project")
		assert.Contains(t, contentStr, "edition")
		assert.Contains(t, contentStr, "dependencies")
	})

	t.Run("returns error for missing Cargo.toml", func(t *testing.T) {
		// Setup: Empty directory
		tempDir := t.TempDir()

		// Test: Update version
		cargo := NewCargoEcosystem(tempDir)
		newVersion := semver.MustParse("2.0.0")
		err := cargo.UpdateVersion(newVersion)

		// Verify: Error returned
		assert.Error(t, err)
	})
}

// TestCargoEcosystem_GetVersionFiles tests getting version file paths
func TestCargoEcosystem_GetVersionFiles(t *testing.T) {
	t.Run("returns Cargo.toml path", func(t *testing.T) {
		// Setup: Create Cargo.toml
		tempDir := t.TempDir()
		cargoPath := filepath.Join(tempDir, "Cargo.toml")
		require.NoError(t, os.WriteFile(cargoPath, []byte("[package]\nversion = \"1.0.0\"\n"), 0644))

		// Test: Get version files
		cargo := NewCargoEcosystem(tempDir)
		files := cargo.GetVersionFiles()

		// Verify: Cargo.toml returned
		require.Len(t, files, 1)
		assert.Equal(t, cargoPath, files[0])
	})

	t.Run("returns empty for missing Cargo.toml", func(t *testing.T) {
		// Setup: Empty directory
		tempDir := t.TempDir()

		// Test: Get version files
		cargo := NewCargoEcosystem(tempDir)
		files := cargo.GetVersionFiles()

		// Verify: Empty slice returned
		assert.Len(t, files, 0)
	})
}

// TestDetectCargoEcosystem tests Cargo/Rust project detection
func TestDetectCargoEcosystem(t *testing.T) {
	t.Run("detects Cargo.toml", func(t *testing.T) {
		// Setup: Create Cargo.toml
		tempDir := t.TempDir()
		cargoPath := filepath.Join(tempDir, "Cargo.toml")
		require.NoError(t, os.WriteFile(cargoPath, []byte("[package]\nname = \"test\"\n"), 0644))

		// Test: Detect Cargo
		detected := DetectCargoEcosystem(tempDir)

		// Verify: Detected
		assert.True(t, detected)
	})

	t.Run("detects src directory with .rs files", func(t *testing.T) {
		// Setup: Create src directory with Rust files
		tempDir := t.TempDir()
		srcDir := filepath.Join(tempDir, "src")
		require.NoError(t, os.MkdirAll(srcDir, 0755))
		mainPath := filepath.Join(srcDir, "main.rs")
		require.NoError(t, os.WriteFile(mainPath, []byte("fn main() {}\n"), 0644))

		// Test: Detect Cargo
		detected := DetectCargoEcosystem(tempDir)

		// Verify: Detected
		assert.True(t, detected)
	})

	t.Run("returns false for no Rust files", func(t *testing.T) {
		// Setup: Empty directory
		tempDir := t.TempDir()

		// Test: Detect Cargo
		detected := DetectCargoEcosystem(tempDir)

		// Verify: Not detected
		assert.False(t, detected)
	})
}
