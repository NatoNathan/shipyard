package git

import (
	"os"
	"path/filepath"
	"testing"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateCommit_Success tests creating a commit with staged changes
func TestCreateCommit_Success(t *testing.T) {
	// Setup: Create temp git repo with staged changes
	tempDir := t.TempDir()
	repo, err := gogit.PlainInit(tempDir, false)
	require.NoError(t, err)

	// Create and stage a file
	testFile := filepath.Join(tempDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test content"), 0644))

	worktree, err := repo.Worktree()
	require.NoError(t, err)
	_, err = worktree.Add("test.txt")
	require.NoError(t, err)

	// Test: Create commit with message
	message := "Initial commit"
	err = CreateCommit(tempDir, message)

	// Verify: Commit was created successfully
	require.NoError(t, err)

	// Verify commit exists with correct message
	ref, err := repo.Head()
	require.NoError(t, err)
	commit, err := repo.CommitObject(ref.Hash())
	require.NoError(t, err)
	assert.Equal(t, message, commit.Message)
}

// TestCreateCommit_MultilineMessage tests commit with multiline message
func TestCreateCommit_MultilineMessage(t *testing.T) {
	// Setup: Create temp git repo with staged changes
	tempDir := t.TempDir()
	repo, err := gogit.PlainInit(tempDir, false)
	require.NoError(t, err)

	// Create and stage a file
	testFile := filepath.Join(tempDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test content"), 0644))

	worktree, err := repo.Worktree()
	require.NoError(t, err)
	_, err = worktree.Add("test.txt")
	require.NoError(t, err)

	// Test: Create commit with multiline message
	message := `feat: Add new feature

This is a detailed description of the feature.
It spans multiple lines.

Co-Authored-By: Test User <test@example.com>`

	err = CreateCommit(tempDir, message)

	// Verify: Commit was created with full multiline message
	require.NoError(t, err)

	ref, err := repo.Head()
	require.NoError(t, err)
	commit, err := repo.CommitObject(ref.Hash())
	require.NoError(t, err)
	assert.Equal(t, message, commit.Message)
}

// TestCreateCommit_NoChanges tests creating commit when nothing is staged
func TestCreateCommit_NoChanges(t *testing.T) {
	// Setup: Create temp git repo with no staged changes
	tempDir := t.TempDir()
	_, err := gogit.PlainInit(tempDir, false)
	require.NoError(t, err)

	// Test: Attempt to create commit with no changes
	err = CreateCommit(tempDir, "Empty commit")

	// Verify: Error is returned
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no changes")
}

// TestCreateCommit_InvalidRepo tests committing to invalid repository
func TestCreateCommit_InvalidRepo(t *testing.T) {
	// Setup: Use non-git directory
	tempDir := t.TempDir()

	// Test: Attempt to create commit in non-repo
	err := CreateCommit(tempDir, "Test commit")

	// Verify: Error is returned
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open repository")
}

// TestCreateCommit_EmptyMessage tests creating commit with empty message
func TestCreateCommit_EmptyMessage(t *testing.T) {
	// Setup: Create temp git repo with staged changes
	tempDir := t.TempDir()
	repo, err := gogit.PlainInit(tempDir, false)
	require.NoError(t, err)

	// Create and stage a file
	testFile := filepath.Join(tempDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test content"), 0644))

	worktree, err := repo.Worktree()
	require.NoError(t, err)
	_, err = worktree.Add("test.txt")
	require.NoError(t, err)

	// Test: Attempt to create commit with empty message
	err = CreateCommit(tempDir, "")

	// Verify: Error is returned
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

// TestCreateCommit_MultipleFiles tests committing multiple staged files
func TestCreateCommit_MultipleFiles(t *testing.T) {
	// Setup: Create temp git repo with multiple staged files
	tempDir := t.TempDir()
	repo, err := gogit.PlainInit(tempDir, false)
	require.NoError(t, err)

	worktree, err := repo.Worktree()
	require.NoError(t, err)

	// Create and stage multiple files
	files := []string{"file1.txt", "file2.txt", "file3.txt"}
	for _, file := range files {
		path := filepath.Join(tempDir, file)
		require.NoError(t, os.WriteFile(path, []byte("content"), 0644))
		_, err = worktree.Add(file)
		require.NoError(t, err)
	}

	// Test: Create commit
	message := "Add multiple files"
	err = CreateCommit(tempDir, message)

	// Verify: Commit was created successfully
	require.NoError(t, err)

	ref, err := repo.Head()
	require.NoError(t, err)
	commit, err := repo.CommitObject(ref.Hash())
	require.NoError(t, err)
	assert.Equal(t, message, commit.Message)

	// Verify all files are in the commit
	tree, err := commit.Tree()
	require.NoError(t, err)
	for _, file := range files {
		_, err := tree.File(file)
		assert.NoError(t, err, "File %s should be in commit", file)
	}
}

// TestCreateCommit_AuthorInfo tests commit includes author information
func TestCreateCommit_AuthorInfo(t *testing.T) {
	// Setup: Create temp git repo with staged changes
	tempDir := t.TempDir()
	repo, err := gogit.PlainInit(tempDir, false)
	require.NoError(t, err)

	// Create and stage a file
	testFile := filepath.Join(tempDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test content"), 0644))

	worktree, err := repo.Worktree()
	require.NoError(t, err)
	_, err = worktree.Add("test.txt")
	require.NoError(t, err)

	// Test: Create commit
	err = CreateCommit(tempDir, "Test commit")
	require.NoError(t, err)

	// Verify: Commit has author information
	ref, err := repo.Head()
	require.NoError(t, err)
	commit, err := repo.CommitObject(ref.Hash())
	require.NoError(t, err)

	assert.NotEmpty(t, commit.Author.Name)
	assert.NotEmpty(t, commit.Author.Email)
	assert.False(t, commit.Author.When.IsZero())
	assert.NotEmpty(t, commit.Committer.Name)
	assert.NotEmpty(t, commit.Committer.Email)
	assert.False(t, commit.Committer.When.IsZero())
}

// TestCreateCommits_Bulk tests creating multiple commits
func TestCreateCommits_Bulk(t *testing.T) {
	// Setup: Create temp git repo
	tempDir := t.TempDir()
	repo, err := gogit.PlainInit(tempDir, false)
	require.NoError(t, err)

	worktree, err := repo.Worktree()
	require.NoError(t, err)

	// Test: Create multiple commits in sequence
	commits := []struct {
		file    string
		message string
	}{
		{"file1.txt", "Add file1"},
		{"file2.txt", "Add file2"},
		{"file3.txt", "Add file3"},
	}

	for _, c := range commits {
		// Stage new file
		path := filepath.Join(tempDir, c.file)
		require.NoError(t, os.WriteFile(path, []byte("content"), 0644))
		_, err = worktree.Add(c.file)
		require.NoError(t, err)

		// Create commit
		err = CreateCommit(tempDir, c.message)
		require.NoError(t, err)
	}

	// Verify: All commits exist in history
	ref, err := repo.Head()
	require.NoError(t, err)

	commitIter, err := repo.Log(&gogit.LogOptions{From: ref.Hash()})
	require.NoError(t, err)

	var commitMessages []string
	err = commitIter.ForEach(func(c *object.Commit) error {
		commitMessages = append(commitMessages, c.Message)
		return nil
	})
	require.NoError(t, err)

	// Commits are in reverse order (newest first)
	assert.Equal(t, "Add file3", commitMessages[0])
	assert.Equal(t, "Add file2", commitMessages[1])
	assert.Equal(t, "Add file1", commitMessages[2])
	assert.Len(t, commitMessages, 3)
}
