package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/NatoNathan/shipyard/internal/fileutil"
	"github.com/NatoNathan/shipyard/internal/git"
	"github.com/go-git/go-git/v5/plumbing"
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
				continue
			}
			if err := os.Chmod(snapshot.path, snapshot.mode); err != nil {
				rollbackErr = joinRollbackError(rollbackErr, fmt.Errorf("failed to restore mode for %s: %w", snapshot.path, err))
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

// applyTransaction tracks filesystem and git changes made while applying
// versions so they can be rolled back together on failure. Rollback order is
// tags, then commit, then files.
type applyTransaction struct {
	projectPath     string
	files           *fileTransaction
	originalHead    plumbing.Hash
	originalHeadSet bool
	commitCreated   bool
	createdTags     []string
}

func newApplyTransaction(projectPath string) *applyTransaction {
	return &applyTransaction{
		projectPath:  projectPath,
		files:        newFileTransaction(),
		originalHead: plumbing.ZeroHash,
	}
}

func (tx *applyTransaction) Backup(path string) error {
	return tx.files.Backup(path)
}

// CaptureHead records the current git HEAD so a later commit can be undone.
// Must be called before CreateCommit.
func (tx *applyTransaction) CaptureHead() error {
	head, err := git.HeadHash(tx.projectPath)
	if err != nil {
		return fmt.Errorf("failed to capture git HEAD before version changes: %w", err)
	}
	tx.originalHead = head
	tx.originalHeadSet = true
	return nil
}

func (tx *applyTransaction) MarkCommitCreated() {
	tx.commitCreated = true
}

func (tx *applyTransaction) MarkTagCreated(name string) {
	tx.createdTags = append(tx.createdTags, name)
}

// RollbackOnError undoes recorded changes when err is non-nil, appending any
// rollback failures to the returned error. Intended for use in a defer with a
// named error return:
//
//	defer func() { err = tx.RollbackOnError(err) }()
func (tx *applyTransaction) RollbackOnError(err error) error {
	if err == nil {
		return nil
	}
	if len(tx.createdTags) > 0 {
		if rollbackErr := git.DeleteTags(tx.projectPath, tx.createdTags); rollbackErr != nil {
			err = fmt.Errorf("%w; additionally failed to delete created tags: %v", err, rollbackErr)
		}
	}
	// Reset whenever HEAD was captured: even if the commit was never created,
	// files may have been staged, and a mixed reset unstages them without
	// touching the worktree.
	if tx.originalHeadSet {
		if rollbackErr := git.ResetMixed(tx.projectPath, tx.originalHead); rollbackErr != nil {
			err = fmt.Errorf("%w; additionally failed to roll back git commit: %v", err, rollbackErr)
		}
	}
	if rollbackErr := tx.files.Rollback(); rollbackErr != nil {
		err = fmt.Errorf("%w; additionally failed to roll back filesystem changes: %v", err, rollbackErr)
	}
	return err
}
