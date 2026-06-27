package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/NatoNathan/shipyard/internal/fileutil"
)

type fileSnapshot struct {
	path   string
	data   []byte
	mode   os.FileMode
	exists bool
}

type fileTransaction struct {
	snapshots map[string]fileSnapshot
	order     []string
}

func newFileTransaction() *fileTransaction {
	return &fileTransaction{snapshots: make(map[string]fileSnapshot)}
}

func (tx *fileTransaction) Backup(path string) error {
	if _, ok := tx.snapshots[path]; ok {
		return nil
	}

	info, err := os.Stat(filepath.Clean(path))
	if err != nil {
		if os.IsNotExist(err) {
			tx.snapshots[path] = fileSnapshot{path: path, exists: false}
			tx.order = append(tx.order, path)
			return nil
		}
		return fmt.Errorf("failed to back up %s: %w", path, err)
	}

	data, err := fileutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to back up %s: %w", path, err)
	}

	copied := append([]byte(nil), data...)
	tx.snapshots[path] = fileSnapshot{path: path, data: copied, mode: info.Mode().Perm(), exists: true}
	tx.order = append(tx.order, path)
	return nil
}

func (tx *fileTransaction) Rollback() error {
	var rollbackErr error
	for i := len(tx.order) - 1; i >= 0; i-- {
		snapshot := tx.snapshots[tx.order[i]]
		if snapshot.exists {
			if err := fileutil.MkdirAll(filepath.Dir(snapshot.path), 0755); err != nil {
				rollbackErr = joinRollbackError(rollbackErr, fmt.Errorf("failed to recreate directory for %s: %w", snapshot.path, err))
				continue
			}
			if err := fileutil.WriteFile(snapshot.path, snapshot.data, snapshot.mode); err != nil {
				rollbackErr = joinRollbackError(rollbackErr, fmt.Errorf("failed to restore %s: %w", snapshot.path, err))
			}
			continue
		}

		if err := os.Remove(snapshot.path); err != nil && !os.IsNotExist(err) {
			rollbackErr = joinRollbackError(rollbackErr, fmt.Errorf("failed to remove new file %s: %w", snapshot.path, err))
		}
	}
	return rollbackErr
}

func joinRollbackError(existing error, next error) error {
	if existing == nil {
		return next
	}
	return fmt.Errorf("%v; %w", existing, next)
}
