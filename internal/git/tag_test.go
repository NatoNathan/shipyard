package git

import (
	"os"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateAnnotatedTag_Success tests creating an annotated tag
func TestCreateAnnotatedTag_Success(t *testing.T) {
	// Setup: Create temp git repo with a commit
	tempDir := t.TempDir()
	repo, err := gogit.PlainInit(tempDir, false)
	require.NoError(t, err)

	// Create initial commit
	worktree, err := repo.Worktree()
	require.NoError(t, err)

	// Create a file and commit it
	testFile := tempDir + "/test.txt"
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))
	_, err = worktree.Add("test.txt")
	require.NoError(t, err)

	commit, err := worktree.Commit("Initial commit", &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)

	// Test: Create annotated tag
	tagName := "v1.0.0"
	message := "Release v1.0.0"

	err = CreateAnnotatedTag(tempDir, tagName, message)

	// Verify: Tag was created successfully
	require.NoError(t, err)

	// Verify tag exists and points to correct commit
	tag, err := repo.Tag(tagName)
	require.NoError(t, err)
	assert.NotNil(t, tag)

	// Get tag object and verify it's annotated
	tagObj, err := repo.TagObject(tag.Hash())
	require.NoError(t, err)
	assert.Equal(t, message+"\n", tagObj.Message) // git adds newline
	assert.Equal(t, commit, tagObj.Target)
}

// TestCreateAnnotatedTag_DuplicateTag tests error when tag already exists
func TestCreateAnnotatedTag_DuplicateTag(t *testing.T) {
	// Setup: Create temp git repo with a commit
	tempDir := t.TempDir()
	repo, err := gogit.PlainInit(tempDir, false)
	require.NoError(t, err)

	// Create initial commit
	worktree, err := repo.Worktree()
	require.NoError(t, err)

	testFile := tempDir + "/test.txt"
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))
	_, err = worktree.Add("test.txt")
	require.NoError(t, err)

	_, err = worktree.Commit("Initial commit", &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)

	// Create first tag
	tagName := "v1.0.0"
	message := "Release v1.0.0"
	err = CreateAnnotatedTag(tempDir, tagName, message)
	require.NoError(t, err)

	// Test: Try to create duplicate tag
	err = CreateAnnotatedTag(tempDir, tagName, message)

	// Verify: Should return error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tag already exists")
}

// TestCreateAnnotatedTag_InvalidRepo tests error when repo path is invalid
func TestCreateAnnotatedTag_InvalidRepo(t *testing.T) {
	// Test: Try to create tag in non-existent repo
	err := CreateAnnotatedTag("/nonexistent/path", "v1.0.0", "message")

	// Verify: Should return error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open repository")
}

// TestCreateAnnotatedTag_NotARepository tests error when path is not a git repo
func TestCreateAnnotatedTag_NotARepository(t *testing.T) {
	// Setup: Create temp directory without git init
	tempDir := t.TempDir()

	// Test: Try to create tag in non-git directory
	err := CreateAnnotatedTag(tempDir, "v1.0.0", "message")

	// Verify: Should return error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open repository")
}

// TestCreateMultipleTags tests creating multiple tags in sequence
func TestCreateMultipleTags(t *testing.T) {
	// Setup: Create temp git repo with a commit
	tempDir := t.TempDir()
	repo, err := gogit.PlainInit(tempDir, false)
	require.NoError(t, err)

	// Create initial commit
	worktree, err := repo.Worktree()
	require.NoError(t, err)

	testFile := tempDir + "/test.txt"
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))
	_, err = worktree.Add("test.txt")
	require.NoError(t, err)

	_, err = worktree.Commit("Initial commit", &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)

	// Test: Create multiple tags
	tags := []struct {
		name    string
		message string
	}{
		{"core/v1.0.0", "Release core v1.0.0"},
		{"api/v2.0.0", "Release api v2.0.0"},
		{"release-20260130-143000", "Release: core v1.0.0, api v2.0.0"},
	}

	for _, tag := range tags {
		err = CreateAnnotatedTag(tempDir, tag.name, tag.message)
		require.NoError(t, err, "Failed to create tag %s", tag.name)
	}

	// Verify: All tags exist
	for _, tag := range tags {
		ref, err := repo.Tag(tag.name)
		require.NoError(t, err, "Tag %s should exist", tag.name)
		assert.NotNil(t, ref)

		// Verify it's an annotated tag
		tagObj, err := repo.TagObject(ref.Hash())
		require.NoError(t, err, "Tag %s should be annotated", tag.name)
		assert.Equal(t, tag.message+"\n", tagObj.Message)
	}
}

// TestCreateAnnotatedTags_Bulk tests creating multiple tags with one call
func TestCreateAnnotatedTags_Bulk(t *testing.T) {
	// Setup: Create temp git repo with a commit
	tempDir := t.TempDir()
	repo, err := gogit.PlainInit(tempDir, false)
	require.NoError(t, err)

	// Create initial commit
	worktree, err := repo.Worktree()
	require.NoError(t, err)

	testFile := tempDir + "/test.txt"
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))
	_, err = worktree.Add("test.txt")
	require.NoError(t, err)

	_, err = worktree.Commit("Initial commit", &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)

	// Test: Create multiple tags at once
	tags := map[string]string{
		"core/v1.0.0":              "Release core v1.0.0",
		"api/v2.0.0":               "Release api v2.0.0",
		"release-20260130-143000": "Release: core v1.0.0, api v2.0.0",
	}

	err = CreateAnnotatedTags(tempDir, tags)
	require.NoError(t, err)

	// Verify: All tags exist
	for name, message := range tags {
		ref, err := repo.Tag(name)
		require.NoError(t, err, "Tag %s should exist", name)
		assert.NotNil(t, ref)

		// Verify it's an annotated tag
		tagObj, err := repo.TagObject(ref.Hash())
		require.NoError(t, err, "Tag %s should be annotated", name)
		assert.Equal(t, message+"\n", tagObj.Message)
	}
}

// TestCreateAnnotatedTags_PartialFailure tests behavior when some tags fail
func TestCreateAnnotatedTags_PartialFailure(t *testing.T) {
	// Setup: Create temp git repo with a commit
	tempDir := t.TempDir()
	repo, err := gogit.PlainInit(tempDir, false)
	require.NoError(t, err)

	// Create initial commit
	worktree, err := repo.Worktree()
	require.NoError(t, err)

	testFile := tempDir + "/test.txt"
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))
	_, err = worktree.Add("test.txt")
	require.NoError(t, err)

	_, err = worktree.Commit("Initial commit", &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)

	// Create first tag manually
	err = CreateAnnotatedTag(tempDir, "v1.0.0", "First release")
	require.NoError(t, err)

	// Test: Try to create tags including duplicate
	tags := map[string]string{
		"v1.0.0": "First release",  // Duplicate
		"v2.0.0": "Second release", // New
	}

	err = CreateAnnotatedTags(tempDir, tags)

	// Verify: Should return error about duplicate
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "v1.0.0")

	// Verify: Second tag should NOT be created (all-or-nothing)
	_, err = repo.Tag("v2.0.0")
	assert.Error(t, err) // Tag should not exist
}

// TestCreateLightweightTag_Success tests creating a lightweight tag
func TestCreateLightweightTag_Success(t *testing.T) {
	// Setup: Create temp git repo with a commit
	tempDir := t.TempDir()
	repo, err := gogit.PlainInit(tempDir, false)
	require.NoError(t, err)

	// Create initial commit
	worktree, err := repo.Worktree()
	require.NoError(t, err)

	testFile := tempDir + "/test.txt"
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))
	_, err = worktree.Add("test.txt")
	require.NoError(t, err)

	commit, err := worktree.Commit("Initial commit", &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)

	// Test: Create lightweight tag
	tagName := "v1.0.0"

	err = CreateLightweightTag(tempDir, tagName)

	// Verify: Tag was created successfully
	require.NoError(t, err)

	// Verify tag exists and points to correct commit
	tag, err := repo.Tag(tagName)
	require.NoError(t, err)
	assert.NotNil(t, tag)

	// Verify tag is lightweight (not annotated)
	_, err = repo.TagObject(tag.Hash())
	assert.Error(t, err) // Should fail because lightweight tags have no tag object

	// Verify tag points directly to commit
	assert.Equal(t, commit, tag.Hash())
}

// TestCreateLightweightTag_DuplicateTag tests error when lightweight tag already exists
func TestCreateLightweightTag_DuplicateTag(t *testing.T) {
	// Setup: Create temp git repo with a commit
	tempDir := t.TempDir()
	repo, err := gogit.PlainInit(tempDir, false)
	require.NoError(t, err)

	// Create initial commit
	worktree, err := repo.Worktree()
	require.NoError(t, err)

	testFile := tempDir + "/test.txt"
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))
	_, err = worktree.Add("test.txt")
	require.NoError(t, err)

	_, err = worktree.Commit("Initial commit", &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)

	// Create first tag
	tagName := "v1.0.0"
	err = CreateLightweightTag(tempDir, tagName)
	require.NoError(t, err)

	// Test: Try to create duplicate tag
	err = CreateLightweightTag(tempDir, tagName)

	// Verify: Should return error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tag already exists")
}

// TestCreateLightweightTag_InvalidRepo tests error when repo path is invalid
func TestCreateLightweightTag_InvalidRepo(t *testing.T) {
	// Test: Try to create tag in non-existent repo
	err := CreateLightweightTag("/nonexistent/path", "v1.0.0")

	// Verify: Should return error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open repository")
}

// TestCreateLightweightTags_Bulk tests creating multiple lightweight tags
func TestCreateLightweightTags_Bulk(t *testing.T) {
	// Setup: Create temp git repo with a commit
	tempDir := t.TempDir()
	repo, err := gogit.PlainInit(tempDir, false)
	require.NoError(t, err)

	// Create initial commit
	worktree, err := repo.Worktree()
	require.NoError(t, err)

	testFile := tempDir + "/test.txt"
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))
	_, err = worktree.Add("test.txt")
	require.NoError(t, err)

	commit, err := worktree.Commit("Initial commit", &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)

	// Test: Create multiple lightweight tags at once
	tagNames := []string{"v1.0.0", "core/v1.0.0", "release-20260130"}

	err = CreateLightweightTags(tempDir, tagNames)
	require.NoError(t, err)

	// Verify: All tags exist and are lightweight
	for _, name := range tagNames {
		ref, err := repo.Tag(name)
		require.NoError(t, err, "Tag %s should exist", name)
		assert.NotNil(t, ref)

		// Verify tag is lightweight (not annotated)
		_, err = repo.TagObject(ref.Hash())
		assert.Error(t, err, "Tag %s should be lightweight", name)

		// Verify tag points to commit
		assert.Equal(t, commit, ref.Hash())
	}
}

// TestCreateLightweightTags_PartialFailure tests behavior when some tags fail
func TestCreateLightweightTags_PartialFailure(t *testing.T) {
	// Setup: Create temp git repo with a commit
	tempDir := t.TempDir()
	repo, err := gogit.PlainInit(tempDir, false)
	require.NoError(t, err)

	// Create initial commit
	worktree, err := repo.Worktree()
	require.NoError(t, err)

	testFile := tempDir + "/test.txt"
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))
	_, err = worktree.Add("test.txt")
	require.NoError(t, err)

	_, err = worktree.Commit("Initial commit", &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)

	// Create first tag manually
	err = CreateLightweightTag(tempDir, "v1.0.0")
	require.NoError(t, err)

	// Test: Try to create tags including duplicate
	tagNames := []string{"v1.0.0", "v2.0.0"}

	err = CreateLightweightTags(tempDir, tagNames)

	// Verify: Should return error about duplicate
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "v1.0.0")

	// Verify: Second tag should NOT be created (all-or-nothing)
	_, err = repo.Tag("v2.0.0")
	assert.Error(t, err) // Tag should not exist
}

// TestMixedTagTypes tests that lightweight and annotated tags can coexist
func TestMixedTagTypes(t *testing.T) {
	// Setup: Create temp git repo with a commit
	tempDir := t.TempDir()
	repo, err := gogit.PlainInit(tempDir, false)
	require.NoError(t, err)

	// Create initial commit
	worktree, err := repo.Worktree()
	require.NoError(t, err)

	testFile := tempDir + "/test.txt"
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))
	_, err = worktree.Add("test.txt")
	require.NoError(t, err)

	commit, err := worktree.Commit("Initial commit", &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)

	// Test: Create both lightweight and annotated tags
	err = CreateLightweightTag(tempDir, "v1.0.0")
	require.NoError(t, err)

	err = CreateAnnotatedTag(tempDir, "v1.0.0-annotated", "Release v1.0.0")
	require.NoError(t, err)

	// Verify: Lightweight tag points to commit
	lightTag, err := repo.Tag("v1.0.0")
	require.NoError(t, err)
	assert.Equal(t, commit, lightTag.Hash())
	_, err = repo.TagObject(lightTag.Hash())
	assert.Error(t, err) // No tag object

	// Verify: Annotated tag has tag object
	annotTag, err := repo.Tag("v1.0.0-annotated")
	require.NoError(t, err)
	tagObj, err := repo.TagObject(annotTag.Hash())
	require.NoError(t, err)
	assert.Equal(t, commit, tagObj.Target)
}

// TestVerifyTagExists_Success tests verifying an existing tag
func TestVerifyTagExists_Success(t *testing.T) {
	// Setup: Create temp git repo with a commit and tag
	tempDir := t.TempDir()
	repo, err := gogit.PlainInit(tempDir, false)
	require.NoError(t, err)

	// Create initial commit
	worktree, err := repo.Worktree()
	require.NoError(t, err)

	testFile := tempDir + "/test.txt"
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))
	_, err = worktree.Add("test.txt")
	require.NoError(t, err)

	_, err = worktree.Commit("Initial commit", &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)

	// Create a tag
	tagName := "v1.0.0"
	err = CreateAnnotatedTag(tempDir, tagName, "Test release")
	require.NoError(t, err)

	// Test: Verify tag exists
	exists, err := VerifyTagExists(tempDir, tagName)

	// Verify: Returns true, no error
	require.NoError(t, err)
	assert.True(t, exists)
}

// TestVerifyTagExists_NotFound tests verifying a non-existent tag
func TestVerifyTagExists_NotFound(t *testing.T) {
	// Setup: Create temp git repo with commit but no tag
	tempDir := t.TempDir()
	repo, err := gogit.PlainInit(tempDir, false)
	require.NoError(t, err)

	// Create initial commit
	worktree, err := repo.Worktree()
	require.NoError(t, err)

	testFile := tempDir + "/test.txt"
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))
	_, err = worktree.Add("test.txt")
	require.NoError(t, err)

	_, err = worktree.Commit("Initial commit", &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)

	// Test: Verify non-existent tag
	exists, err := VerifyTagExists(tempDir, "v9.9.9")

	// Verify: Returns false, no error
	require.NoError(t, err)
	assert.False(t, exists)
}

// TestVerifyTagExists_InvalidRepo tests error when repo path is invalid
func TestVerifyTagExists_InvalidRepo(t *testing.T) {
	// Test: Call VerifyTagExists on non-existent path
	exists, err := VerifyTagExists("/nonexistent/path", "v1.0.0")

	// Verify: Returns false with error
	assert.Error(t, err)
	assert.False(t, exists)
	assert.Contains(t, err.Error(), "failed to open repository")
}

// TestVerifyTagPushedToRemote_InvalidRepo tests error when repo path is invalid
func TestVerifyTagPushedToRemote_InvalidRepo(t *testing.T) {
	// Test: Call VerifyTagPushedToRemote on non-existent path
	pushed, err := VerifyTagPushedToRemote("/nonexistent/path", "origin", "v1.0.0")

	// Verify: Returns false with error
	assert.Error(t, err)
	assert.False(t, pushed)
	assert.Contains(t, err.Error(), "failed to open repository")
}

// TestVerifyTagPushedToRemote_NoRemote tests error when remote doesn't exist
func TestVerifyTagPushedToRemote_NoRemote(t *testing.T) {
	// Setup: Create temp git repo with commit but no remote
	tempDir := t.TempDir()
	repo, err := gogit.PlainInit(tempDir, false)
	require.NoError(t, err)

	// Create initial commit
	worktree, err := repo.Worktree()
	require.NoError(t, err)

	testFile := tempDir + "/test.txt"
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))
	_, err = worktree.Add("test.txt")
	require.NoError(t, err)

	_, err = worktree.Commit("Initial commit", &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)

	// Test: Call VerifyTagPushedToRemote without setting up remote
	pushed, err := VerifyTagPushedToRemote(tempDir, "origin", "v1.0.0")

	// Verify: Returns false with error
	assert.Error(t, err)
	assert.False(t, pushed)
	assert.Contains(t, err.Error(), "failed to get remote")
}
