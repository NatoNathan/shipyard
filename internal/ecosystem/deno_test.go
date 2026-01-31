package ecosystem

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/NatoNathan/shipyard/pkg/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDenoEcosystem_ReadVersion tests reading version from deno.json
func TestDenoEcosystem_ReadVersion(t *testing.T) {
	t.Run("reads version from deno.json", func(t *testing.T) {
		// Setup: Create deno.json
		tempDir := t.TempDir()
		denoPath := filepath.Join(tempDir, "deno.json")
		denoContent := `{
  "name": "@scope/my-module",
  "version": "1.2.3",
  "exports": "./mod.ts",
  "tasks": {
    "test": "deno test"
  }
}`
		require.NoError(t, os.WriteFile(denoPath, []byte(denoContent), 0644))

		// Test: Read version
		deno := NewDenoEcosystem(tempDir)
		version, err := deno.ReadVersion()

		// Verify: Version read correctly
		require.NoError(t, err)
		assert.Equal(t, "1.2.3", version.String())
	})

	t.Run("reads version from deno.jsonc", func(t *testing.T) {
		// Setup: Create deno.jsonc with comments
		tempDir := t.TempDir()
		denoPath := filepath.Join(tempDir, "deno.jsonc")
		denoContent := `{
  // This is a comment
  "name": "@scope/my-module",
  "version": "2.0.0",  // Version comment
  "exports": "./mod.ts"
}`
		require.NoError(t, os.WriteFile(denoPath, []byte(denoContent), 0644))

		// Test: Read version
		deno := NewDenoEcosystem(tempDir)
		version, err := deno.ReadVersion()

		// Verify: Version read correctly
		require.NoError(t, err)
		assert.Equal(t, "2.0.0", version.String())
	})

	t.Run("returns error for missing deno config", func(t *testing.T) {
		// Setup: Empty directory
		tempDir := t.TempDir()

		// Test: Read version
		deno := NewDenoEcosystem(tempDir)
		_, err := deno.ReadVersion()

		// Verify: Error returned
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "deno")
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		// Setup: Create invalid deno.json
		tempDir := t.TempDir()
		denoPath := filepath.Join(tempDir, "deno.json")
		require.NoError(t, os.WriteFile(denoPath, []byte("{invalid json}"), 0644))

		// Test: Read version
		deno := NewDenoEcosystem(tempDir)
		_, err := deno.ReadVersion()

		// Verify: Error returned
		assert.Error(t, err)
	})

	t.Run("returns error for missing version field", func(t *testing.T) {
		// Setup: Create deno.json without version
		tempDir := t.TempDir()
		denoPath := filepath.Join(tempDir, "deno.json")
		denoContent := `{
  "name": "my-module",
  "exports": "./mod.ts"
}`
		require.NoError(t, os.WriteFile(denoPath, []byte(denoContent), 0644))

		// Test: Read version
		deno := NewDenoEcosystem(tempDir)
		_, err := deno.ReadVersion()

		// Verify: Error returned
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "version")
	})
}

// TestDenoEcosystem_UpdateVersion tests updating version in deno.json
func TestDenoEcosystem_UpdateVersion(t *testing.T) {
	t.Run("updates version in deno.json", func(t *testing.T) {
		// Setup: Create deno.json
		tempDir := t.TempDir()
		denoPath := filepath.Join(tempDir, "deno.json")
		denoContent := `{
  "name": "@scope/my-module",
  "version": "1.2.3",
  "exports": "./mod.ts"
}`
		require.NoError(t, os.WriteFile(denoPath, []byte(denoContent), 0644))

		// Test: Update version
		deno := NewDenoEcosystem(tempDir)
		newVersion := semver.MustParse("2.0.0")
		err := deno.UpdateVersion(newVersion)

		// Verify: Version updated
		require.NoError(t, err)

		// Read back and verify
		updatedVersion, err := deno.ReadVersion()
		require.NoError(t, err)
		assert.Equal(t, "2.0.0", updatedVersion.String())

		// Verify other fields preserved
		content, _ := os.ReadFile(denoPath)
		contentStr := string(content)
		assert.Contains(t, contentStr, "name")
		assert.Contains(t, contentStr, "@scope/my-module")
		assert.Contains(t, contentStr, "exports")
	})

	t.Run("returns error for missing deno config", func(t *testing.T) {
		// Setup: Empty directory
		tempDir := t.TempDir()

		// Test: Update version
		deno := NewDenoEcosystem(tempDir)
		newVersion := semver.MustParse("2.0.0")
		err := deno.UpdateVersion(newVersion)

		// Verify: Error returned
		assert.Error(t, err)
	})
}

// TestDenoEcosystem_GetVersionFiles tests getting version file paths
func TestDenoEcosystem_GetVersionFiles(t *testing.T) {
	t.Run("returns deno.json path", func(t *testing.T) {
		// Setup: Create deno.json
		tempDir := t.TempDir()
		denoPath := filepath.Join(tempDir, "deno.json")
		require.NoError(t, os.WriteFile(denoPath, []byte(`{"version": "1.0.0"}`), 0644))

		// Test: Get version files
		deno := NewDenoEcosystem(tempDir)
		files := deno.GetVersionFiles()

		// Verify: deno.json returned
		require.Len(t, files, 1)
		assert.Equal(t, denoPath, files[0])
	})

	t.Run("returns deno.jsonc path", func(t *testing.T) {
		// Setup: Create deno.jsonc
		tempDir := t.TempDir()
		denoPath := filepath.Join(tempDir, "deno.jsonc")
		require.NoError(t, os.WriteFile(denoPath, []byte(`{"version": "1.0.0"}`), 0644))

		// Test: Get version files
		deno := NewDenoEcosystem(tempDir)
		files := deno.GetVersionFiles()

		// Verify: deno.jsonc returned
		require.Len(t, files, 1)
		assert.Equal(t, denoPath, files[0])
	})

	t.Run("returns empty for missing config", func(t *testing.T) {
		// Setup: Empty directory
		tempDir := t.TempDir()

		// Test: Get version files
		deno := NewDenoEcosystem(tempDir)
		files := deno.GetVersionFiles()

		// Verify: Empty slice returned
		assert.Len(t, files, 0)
	})
}

// TestDetectDenoEcosystem tests Deno project detection
func TestDetectDenoEcosystem(t *testing.T) {
	t.Run("detects deno.json", func(t *testing.T) {
		// Setup: Create deno.json
		tempDir := t.TempDir()
		denoPath := filepath.Join(tempDir, "deno.json")
		require.NoError(t, os.WriteFile(denoPath, []byte(`{"name": "test"}`), 0644))

		// Test: Detect Deno
		detected := DetectDenoEcosystem(tempDir)

		// Verify: Detected
		assert.True(t, detected)
	})

	t.Run("detects deno.jsonc", func(t *testing.T) {
		// Setup: Create deno.jsonc
		tempDir := t.TempDir()
		denoPath := filepath.Join(tempDir, "deno.jsonc")
		require.NoError(t, os.WriteFile(denoPath, []byte(`{"name": "test"}`), 0644))

		// Test: Detect Deno
		detected := DetectDenoEcosystem(tempDir)

		// Verify: Detected
		assert.True(t, detected)
	})

	t.Run("detects mod.ts file", func(t *testing.T) {
		// Setup: Create mod.ts
		tempDir := t.TempDir()
		modPath := filepath.Join(tempDir, "mod.ts")
		require.NoError(t, os.WriteFile(modPath, []byte("export {}\n"), 0644))

		// Test: Detect Deno
		detected := DetectDenoEcosystem(tempDir)

		// Verify: Detected
		assert.True(t, detected)
	})

	t.Run("returns false for no Deno files", func(t *testing.T) {
		// Setup: Empty directory
		tempDir := t.TempDir()

		// Test: Detect Deno
		detected := DetectDenoEcosystem(tempDir)

		// Verify: Not detected
		assert.False(t, detected)
	})
}
