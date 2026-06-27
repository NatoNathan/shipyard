package git

import (
	"fmt"
	"strings"
	"time"

	gogit "github.com/go-git/go-git/v5"
	gogitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// getCommitAuthor reads git config for user.name and user.email,
// falling back to defaults if not configured
func getCommitAuthor(repo *gogit.Repository) *object.Signature {
	sig := &object.Signature{
		Name:  "Shipyard",
		Email: "shipyard@local",
		When:  time.Now(),
	}

	cfg, err := repo.ConfigScoped(gogitconfig.GlobalScope)
	if err != nil || cfg.User.Name == "" {
		return sig
	}

	sig.Name = cfg.User.Name
	if cfg.User.Email != "" {
		sig.Email = cfg.User.Email
	}
	return sig
}

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
		Author: getCommitAuthor(repo),
	})
	if err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	return nil
}

// HeadHash returns the current HEAD commit hash.
func HeadHash(repoPath string) (plumbing.Hash, error) {
	repo, err := gogit.PlainOpen(repoPath)
	if err != nil {
		return plumbing.ZeroHash, fmt.Errorf("failed to open repository: %w", err)
	}

	head, err := repo.Head()
	if err != nil {
		return plumbing.ZeroHash, fmt.Errorf("failed to get HEAD: %w", err)
	}

	return head.Hash(), nil
}

// ResetHard resets the working tree and HEAD to the given commit hash.
func ResetHard(repoPath string, hash plumbing.Hash) error {
	repo, err := gogit.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	if err := worktree.Reset(&gogit.ResetOptions{
		Commit: hash,
		Mode:   gogit.HardReset,
	}); err != nil {
		return fmt.Errorf("failed to reset repository: %w", err)
	}

	return nil
}
