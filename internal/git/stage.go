package git

import (
	"fmt"
	"path/filepath"

	gogit "github.com/go-git/go-git/v5"
)

// StageFiles stages multiple files in the git repository
func StageFiles(repoPath string, filePaths []string) error {
	repo, err := gogit.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	for _, filePath := range filePaths {
		// Convert to relative path from repo root
		// If filePath is already relative, use it as-is
		// If filePath is absolute, convert it to relative
		relPath := filePath
		if filepath.IsAbs(filePath) {
			var err error
			relPath, err = filepath.Rel(repoPath, filePath)
			if err != nil {
				// If we can't make it relative, use original path
				relPath = filePath
			}
		}

		_, err = worktree.Add(relPath)
		if err != nil {
			return fmt.Errorf("failed to stage %s: %w", relPath, err)
		}
	}

	return nil
}
