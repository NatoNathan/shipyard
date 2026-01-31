package git

import (
	"os"
	"path/filepath"
	"testing"

	gogit "github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// initGitRepo properly initializes a git repository for testing
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	_, err := gogit.PlainInit(dir, false)
	require.NoError(t, err, "Failed to initialize git repository")
}

// TestIsRepository_ValidRepository tests detection of valid git repositories
func TestIsRepository_ValidRepository(t *testing.T) {
	// Create temporary directory and initialize git repo
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Should detect as repository
	isRepo, err := IsRepository(tempDir)
	assert.NoError(t, err, "IsRepository should not return error")
	assert.True(t, isRepo, "Should detect valid git repository")
}

// TestIsRepository_NotARepository tests detection of non-git directories
func TestIsRepository_NotARepository(t *testing.T) {
	// Create temporary directory without .git
	tempDir := t.TempDir()

	// Should not detect as repository
	isRepo, err := IsRepository(tempDir)
	assert.NoError(t, err, "IsRepository should not return error")
	assert.False(t, isRepo, "Should not detect non-git directory as repository")
}

// TestIsRepository_NonexistentPath tests handling of nonexistent paths
func TestIsRepository_NonexistentPath(t *testing.T) {
	// Use a path that doesn't exist
	nonexistentPath := filepath.Join(os.TempDir(), "nonexistent-directory-12345")

	// Should return error
	isRepo, err := IsRepository(nonexistentPath)
	assert.Error(t, err, "IsRepository should return error for nonexistent path")
	assert.False(t, isRepo, "Should return false for nonexistent path")
}

// TestIsRepository_GitSubdirectory tests detection in subdirectories of git repo
func TestIsRepository_GitSubdirectory(t *testing.T) {
	// Create temporary directory and initialize git repo
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Create subdirectory
	subDir := filepath.Join(tempDir, "subdir", "nested")
	require.NoError(t, os.MkdirAll(subDir, 0755), "Failed to create subdirectory")

	// Should detect as repository even from subdirectory
	isRepo, err := IsRepository(subDir)
	assert.NoError(t, err, "IsRepository should not return error")
	assert.True(t, isRepo, "Should detect git repository from subdirectory")
}
