package git

import (
	"fmt"
	"strings"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// CreateCommit creates a git commit with the given message
// Returns error if repository is invalid or no changes are staged
func CreateCommit(repoPath, message string) error {
	// Validate message
	if strings.TrimSpace(message) == "" {
		return fmt.Errorf("commit message cannot be empty")
	}

	// Open repository
	repo, err := gogit.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Get worktree
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Check if there are staged changes
	status, err := worktree.Status()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	hasChanges := false
	for _, fileStatus := range status {
		// Check if file is staged (added to index)
		if fileStatus.Staging != gogit.Unmodified {
			hasChanges = true
			break
		}
	}

	if !hasChanges {
		return fmt.Errorf("no changes staged for commit")
	}

	// Create commit
	_, err = worktree.Commit(message, &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "Shipyard",
			Email: "shipyard@local",
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	return nil
}
