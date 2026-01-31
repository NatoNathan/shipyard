package git

import (
	"fmt"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// CreateAnnotatedTag creates an annotated git tag at HEAD
func CreateAnnotatedTag(repoPath, tagName, message string) error {
	// Open repository
	repo, err := gogit.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Get HEAD reference
	head, err := repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Create annotated tag
	_, err = repo.CreateTag(tagName, head.Hash(), &gogit.CreateTagOptions{
		Tagger: &object.Signature{
			Name:  "Shipyard",
			Email: "shipyard@local",
			When:  time.Now(),
		},
		Message: message,
	})
	if err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	return nil
}

// CreateAnnotatedTags creates multiple annotated git tags at HEAD
// Returns error on first failure (no tags created if any fail)
func CreateAnnotatedTags(repoPath string, tags map[string]string) error {
	// Validate all tags can be created first
	repo, err := gogit.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Check for existing tags
	for tagName := range tags {
		_, err := repo.Tag(tagName)
		if err == nil {
			return fmt.Errorf("tag already exists: %s", tagName)
		}
	}

	// Create all tags
	for tagName, message := range tags {
		if err := CreateAnnotatedTag(repoPath, tagName, message); err != nil {
			return fmt.Errorf("failed to create tag %s: %w", tagName, err)
		}
	}

	return nil
}

// CreateLightweightTag creates a lightweight git tag at HEAD
func CreateLightweightTag(repoPath, tagName string) error {
	// Open repository
	repo, err := gogit.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Get HEAD reference
	head, err := repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Check if tag already exists
	_, err = repo.Tag(tagName)
	if err == nil {
		return fmt.Errorf("tag already exists: %s", tagName)
	}

	// Create lightweight tag (no tag object, just a reference)
	_, err = repo.CreateTag(tagName, head.Hash(), nil)
	if err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	return nil
}

// CreateLightweightTags creates multiple lightweight git tags at HEAD
// Returns error on first failure (no tags created if any fail)
func CreateLightweightTags(repoPath string, tagNames []string) error {
	// Validate all tags can be created first
	repo, err := gogit.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Check for existing tags
	for _, tagName := range tagNames {
		_, err := repo.Tag(tagName)
		if err == nil {
			return fmt.Errorf("tag already exists: %s", tagName)
		}
	}

	// Create all tags
	for _, tagName := range tagNames {
		if err := CreateLightweightTag(repoPath, tagName); err != nil {
			return fmt.Errorf("failed to create tag %s: %w", tagName, err)
		}
	}

	return nil
}
