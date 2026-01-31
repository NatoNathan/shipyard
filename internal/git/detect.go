package git

import (
	"fmt"
	"os"
	"path/filepath"

	gogit "github.com/go-git/go-git/v5"
)

// IsRepository checks if the given path is within a git repository
func IsRepository(path string) (bool, error) {
	// Check if path exists
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, fmt.Errorf("path does not exist: %s", path)
		}
		return false, fmt.Errorf("failed to stat path: %w", err)
	}

	// Try to open the repository (go-git will search up the directory tree)
	_, err := gogit.PlainOpenWithOptions(path, &gogit.PlainOpenOptions{
		DetectDotGit: true,
	})

	if err != nil {
		// If the error is "repository does not exist", it's not a git repo
		if err == gogit.ErrRepositoryNotExists {
			return false, nil
		}
		// Other errors should be returned
		return false, fmt.Errorf("failed to open repository: %w", err)
	}

	return true, nil
}

// FindRepositoryRoot finds the root directory of the git repository containing the given path
func FindRepositoryRoot(path string) (string, error) {
	// Try to open the repository
	repo, err := gogit.PlainOpenWithOptions(path, &gogit.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		if err == gogit.ErrRepositoryNotExists {
			return "", fmt.Errorf("not a git repository: %s", path)
		}
		return "", fmt.Errorf("failed to open repository: %w", err)
	}

	// Get the worktree
	worktree, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	// Get the absolute path to the root
	root, err := filepath.Abs(worktree.Filesystem.Root())
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	return root, nil
}
