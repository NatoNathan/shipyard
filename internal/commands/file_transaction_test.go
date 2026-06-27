package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileTransactionRollbackPreservesFileMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "script.sh")
	require.NoError(t, os.WriteFile(path, []byte("original"), 0750))

	tx := newFileTransaction()
	require.NoError(t, tx.Backup(path))
	require.NoError(t, os.WriteFile(path, []byte("changed"), 0644))
	require.NoError(t, os.Chmod(path, 0644))

	require.NoError(t, tx.Rollback())

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "original", string(content))

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0750), info.Mode().Perm())
}
