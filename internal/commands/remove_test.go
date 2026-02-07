package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/NatoNathan/shipyard/internal/config"
)

func setupRemoveTestProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".shipyard", "consignments"), 0755))

	cfg := config.Config{
		Packages: []config.Package{
			{Name: "my-pkg", Path: ".", Ecosystem: "go"},
		},
	}
	cfgData, err := yaml.Marshal(&cfg)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".shipyard", "shipyard.yaml"), cfgData, 0644))

	// Create two consignments
	c1 := `---
id: 20240101-120000-aaa111
timestamp: 2024-01-01T12:00:00Z
packages:
  - my-pkg
changeType: patch
---

Fix a bug
`
	c2 := `---
id: 20240102-120000-bbb222
timestamp: 2024-01-02T12:00:00Z
packages:
  - my-pkg
changeType: minor
---

Add a feature
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".shipyard", "consignments", "20240101-120000-aaa111.md"), []byte(c1), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".shipyard", "consignments", "20240102-120000-bbb222.md"), []byte(c2), 0644))

	return dir
}

func TestRemove_ByID(t *testing.T) {
	dir := setupRemoveTestProject(t)

	opts := &RemoveCommandOptions{
		IDs:   []string{"20240101-120000-aaa111"},
		Quiet: true,
	}
	err := runRemoveWithDir(dir, opts)
	assert.NoError(t, err)

	// First consignment should be gone
	_, err = os.Stat(filepath.Join(dir, ".shipyard", "consignments", "20240101-120000-aaa111.md"))
	assert.True(t, os.IsNotExist(err))

	// Second should still exist
	_, err = os.Stat(filepath.Join(dir, ".shipyard", "consignments", "20240102-120000-bbb222.md"))
	assert.NoError(t, err)
}

func TestRemove_All(t *testing.T) {
	dir := setupRemoveTestProject(t)

	opts := &RemoveCommandOptions{
		All:   true,
		Quiet: true,
	}
	err := runRemoveWithDir(dir, opts)
	assert.NoError(t, err)

	// Both consignments should be gone
	_, err = os.Stat(filepath.Join(dir, ".shipyard", "consignments", "20240101-120000-aaa111.md"))
	assert.True(t, os.IsNotExist(err))

	_, err = os.Stat(filepath.Join(dir, ".shipyard", "consignments", "20240102-120000-bbb222.md"))
	assert.True(t, os.IsNotExist(err))
}

func TestRemove_NoFlags(t *testing.T) {
	dir := setupRemoveTestProject(t)

	opts := &RemoveCommandOptions{}
	err := runRemoveWithDir(dir, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "specify --id or --all")
}

func TestRemove_NotFound(t *testing.T) {
	dir := setupRemoveTestProject(t)

	opts := &RemoveCommandOptions{
		IDs: []string{"nonexistent-id"},
	}
	err := runRemoveWithDir(dir, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "consignment not found")
}

func TestRemove_AllEmpty(t *testing.T) {
	dir := setupRemoveTestProject(t)

	// Remove all files first
	os.RemoveAll(filepath.Join(dir, ".shipyard", "consignments"))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".shipyard", "consignments"), 0755))

	opts := &RemoveCommandOptions{
		All:   true,
		Quiet: true,
	}
	err := runRemoveWithDir(dir, opts)
	assert.NoError(t, err)
}
