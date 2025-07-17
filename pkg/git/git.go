// Package git provides functionality for Git operations in Shipyard projects.
package git

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/NatoNathan/shipyard/pkg/config"
	"github.com/NatoNathan/shipyard/pkg/semver"
	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// GitClient provides git operations for shipyard using go-git
type GitClient struct {
	repo *git.Repository
}

// NewGitClient creates a new git client
func NewGitClient(workingDir string) (*GitClient, error) {
	if workingDir == "" {
		workingDir = "."
	}

	repo, err := git.PlainOpen(workingDir)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	return &GitClient{
		repo: repo,
	}, nil
}

// IsGitRepository checks if the current directory is a git repository
func IsGitRepository(workingDir string) bool {
	if workingDir == "" {
		workingDir = "."
	}
	_, err := git.PlainOpen(workingDir)
	return err == nil
}

// GetCurrentBranch returns the current git branch
func (g *GitClient) GetCurrentBranch() (string, error) {
	ref, err := g.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	if ref.Name().IsBranch() {
		return ref.Name().Short(), nil
	}

	return "", fmt.Errorf("HEAD is not pointing to a branch")
}

// GetCurrentCommit returns the current commit hash
func (g *GitClient) GetCurrentCommit() (string, error) {
	ref, err := g.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	return ref.Hash().String(), nil
}

// HasUncommittedChanges checks if there are uncommitted changes
func (g *GitClient) HasUncommittedChanges() (bool, error) {
	worktree, err := g.repo.Worktree()
	if err != nil {
		return false, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return false, fmt.Errorf("failed to get status: %w", err)
	}

	return !status.IsClean(), nil
}

// AddFile adds a file to git staging
func (g *GitClient) AddFile(filepath string) error {
	worktree, err := g.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	_, err = worktree.Add(filepath)
	if err != nil {
		return fmt.Errorf("failed to add file %s: %w", filepath, err)
	}

	return nil
}

// AddFiles adds multiple files to git staging
func (g *GitClient) AddFiles(filepaths []string) error {
	worktree, err := g.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	for _, filepath := range filepaths {
		_, err = worktree.Add(filepath)
		if err != nil {
			return fmt.Errorf("failed to add file %s: %w", filepath, err)
		}
	}

	return nil
}

// Commit creates a git commit with the given message
func (g *GitClient) Commit(message string) (string, error) {
	worktree, err := g.repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	// Get git config for author info
	cfg, err := g.repo.Config()
	if err != nil {
		return "", fmt.Errorf("failed to get git config: %w", err)
	}

	// Use configured user info or defaults
	authorName := "Shipyard"
	authorEmail := "shipyard@localhost"

	if cfg.User.Name != "" {
		authorName = cfg.User.Name
	}
	if cfg.User.Email != "" {
		authorEmail = cfg.User.Email
	}

	commitHash, err := worktree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  authorName,
			Email: authorEmail,
			When:  time.Now(),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to commit: %w", err)
	}

	return commitHash.String(), nil
}

// CreateTag creates a git tag
func (g *GitClient) CreateTag(tagName, message string) error {
	head, err := g.repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Get git config for tagger info
	cfg, err := g.repo.Config()
	if err != nil {
		return fmt.Errorf("failed to get git config: %w", err)
	}

	// Use configured user info or defaults
	taggerName := "Shipyard"
	taggerEmail := "shipyard@localhost"

	if cfg.User.Name != "" {
		taggerName = cfg.User.Name
	}
	if cfg.User.Email != "" {
		taggerEmail = cfg.User.Email
	}

	_, err = g.repo.CreateTag(tagName, head.Hash(), &git.CreateTagOptions{
		Tagger: &object.Signature{
			Name:  taggerName,
			Email: taggerEmail,
			When:  time.Now(),
		},
		Message: message,
	})
	if err != nil {
		return fmt.Errorf("failed to create tag %s: %w", tagName, err)
	}

	return nil
}

// TagExists checks if a git tag exists
func (g *GitClient) TagExists(tagName string) (bool, error) {
	tags, err := g.repo.Tags()
	if err != nil {
		return false, fmt.Errorf("failed to get tags: %w", err)
	}

	exists := false
	err = tags.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().Short() == tagName {
			exists = true
		}
		return nil
	})
	if err != nil {
		return false, fmt.Errorf("failed to iterate tags: %w", err)
	}

	return exists, nil
}

// GetLatestTag returns the latest git tag
func (g *GitClient) GetLatestTag() (string, error) {
	tags, err := g.repo.Tags()
	if err != nil {
		return "", fmt.Errorf("failed to get tags: %w", err)
	}

	var latestTag string
	var latestTime time.Time

	err = tags.ForEach(func(ref *plumbing.Reference) error {
		obj, err := g.repo.TagObject(ref.Hash())
		if err != nil {
			// Not an annotated tag, try to get commit
			commit, err := g.repo.CommitObject(ref.Hash())
			if err != nil {
				return err
			}
			if commit.Author.When.After(latestTime) {
				latestTime = commit.Author.When
				latestTag = ref.Name().Short()
			}
		} else {
			// Annotated tag
			if obj.Tagger.When.After(latestTime) {
				latestTime = obj.Tagger.When
				latestTag = ref.Name().Short()
			}
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to iterate tags: %w", err)
	}

	if latestTag == "" {
		return "", fmt.Errorf("no tags found")
	}

	return latestTag, nil
}

// GetTagsForPattern returns all tags matching a pattern
func (g *GitClient) GetTagsForPattern(pattern string) ([]string, error) {
	tags, err := g.repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	var matchingTags []string
	err = tags.ForEach(func(ref *plumbing.Reference) error {
		tagName := ref.Name().Short()
		// Simple pattern matching - could be enhanced with regex
		if strings.Contains(tagName, pattern) || pattern == "*" {
			matchingTags = append(matchingTags, tagName)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to iterate tags: %w", err)
	}

	return matchingTags, nil
}

// PushTags pushes all tags to remote
func (g *GitClient) PushTags() error {
	// Get the default remote (usually "origin")
	remote, err := g.repo.Remote("origin")
	if err != nil {
		return fmt.Errorf("failed to get origin remote: %w", err)
	}

	err = remote.Push(&git.PushOptions{
		RefSpecs: []gitconfig.RefSpec{gitconfig.RefSpec("+refs/tags/*:refs/tags/*")},
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to push tags: %w", err)
	}

	return nil
}

// Push pushes current branch to remote
func (g *GitClient) Push() error {
	// Get the default remote (usually "origin")
	remote, err := g.repo.Remote("origin")
	if err != nil {
		return fmt.Errorf("failed to get origin remote: %w", err)
	}

	err = remote.Push(&git.PushOptions{})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to push: %w", err)
	}

	return nil
}

// GitOperations provides high-level git operations for shipyard
type GitOperations struct {
	client        *GitClient
	projectConfig *config.ProjectConfig
	available     bool
}

// NewGitOperations creates a new git operations manager
func NewGitOperations(projectConfig *config.ProjectConfig) *GitOperations {
	return NewGitOperationsWithDir(projectConfig, ".")
}

// NewGitOperationsWithDir creates a new git operations manager with custom directory
func NewGitOperationsWithDir(projectConfig *config.ProjectConfig, workingDir string) *GitOperations {
	client, err := NewGitClient(workingDir)
	available := err == nil

	return &GitOperations{
		client:        client,
		projectConfig: projectConfig,
		available:     available,
	}
}

// IsAvailable checks if git is available and we're in a git repository
func (g *GitOperations) IsAvailable() bool {
	return g.available
}

// GenerateTagName generates a git tag name for a package version
func (g *GitOperations) GenerateTagName(packageName string, version *semver.Version) string {
	if g.projectConfig.Type == config.RepositoryTypeMonorepo {
		return fmt.Sprintf("%s/v%s", packageName, version.String())
	}
	return fmt.Sprintf("v%s", version.String())
}

// CreateShipmentTags creates git tags for a shipment
func (g *GitOperations) CreateShipmentTags(versions map[string]*semver.Version, commitMessage string) (map[string]string, error) {
	if !g.IsAvailable() {
		return nil, fmt.Errorf("git is not available or not in a git repository")
	}

	tags := make(map[string]string)

	for packageName, version := range versions {
		tagName := g.GenerateTagName(packageName, version)

		// Check if tag already exists
		exists, err := g.client.TagExists(tagName)
		if err != nil {
			return nil, fmt.Errorf("failed to check if tag %s exists: %w", tagName, err)
		}

		if exists {
			// Tag already exists, skip it
			tags[packageName] = tagName
			continue
		}

		// Create the tag
		tagMessage := fmt.Sprintf("Release %s %s\n\n%s", packageName, version.String(), commitMessage)
		if err := g.client.CreateTag(tagName, tagMessage); err != nil {
			return nil, fmt.Errorf("failed to create tag %s: %w", tagName, err)
		}

		tags[packageName] = tagName
	}

	return tags, nil
}

// CommitShipmentChanges commits shipment-related changes
func (g *GitOperations) CommitShipmentChanges(filesToCommit []string, commitMessage string) (string, error) {
	if !g.IsAvailable() {
		return "", fmt.Errorf("git is not available or not in a git repository")
	}

	// Add files to staging
	if err := g.client.AddFiles(filesToCommit); err != nil {
		return "", fmt.Errorf("failed to stage files: %w", err)
	}

	// Commit the changes
	commitHash, err := g.client.Commit(commitMessage)
	if err != nil {
		return "", fmt.Errorf("failed to commit changes: %w", err)
	}

	return commitHash, nil
}

// GetFilesToCommit returns the list of files that should be committed for a shipment
func (g *GitOperations) GetFilesToCommit(changelogPath string, packages []string) []string {
	var files []string

	// Always include changelog
	files = append(files, changelogPath)

	// Include package manifests based on project type
	if g.projectConfig.Type == config.RepositoryTypeMonorepo {
		for _, packageName := range packages {
			pkg := g.projectConfig.GetPackageByName(packageName)
			if pkg != nil {
				// Add package-specific manifest files
				switch pkg.Ecosystem {
				case "go":
					files = append(files, filepath.Join(pkg.Path, "go.mod"))
				case "npm":
					files = append(files, filepath.Join(pkg.Path, "package.json"))
				case "helm":
					files = append(files, filepath.Join(pkg.Path, "Chart.yaml"))
				}
			}
		}
	} else {
		// Single repository - include the main manifest
		pkg := &g.projectConfig.Package
		switch pkg.Ecosystem {
		case "go":
			files = append(files, "go.mod")
		case "npm":
			files = append(files, "package.json")
		case "helm":
			files = append(files, "Chart.yaml")
		}
	}

	// Include shipment history
	files = append(files, ".shipyard/shipment-history.json")

	return files
}

// CreateShipmentCommitMessage creates a commit message for a shipment
func (g *GitOperations) CreateShipmentCommitMessage(versions map[string]*semver.Version, consignmentSummaries []string) string {
	var message strings.Builder

	if g.projectConfig.Type == config.RepositoryTypeMonorepo {
		message.WriteString("chore: release multiple packages\n\n")
		for packageName, version := range versions {
			message.WriteString(fmt.Sprintf("- %s: %s\n", packageName, version.String()))
		}
	} else {
		// Single package
		for _, version := range versions {
			message.WriteString(fmt.Sprintf("chore: release %s\n", version.String()))
			break
		}
	}

	if len(consignmentSummaries) > 0 {
		message.WriteString("\nChanges:\n")
		for _, summary := range consignmentSummaries {
			message.WriteString(fmt.Sprintf("- %s\n", summary))
		}
	}

	return message.String()
}

// PerformShipmentGitOperations performs all git operations for a shipment
func (g *GitOperations) PerformShipmentGitOperations(versions map[string]*semver.Version, changelogPath string, consignmentSummaries []string, autoCommit, autoPush bool) error {
	if !g.IsAvailable() {
		// Git not available, skip git operations
		return nil
	}

	// Get list of packages
	var packages []string
	for packageName := range versions {
		packages = append(packages, packageName)
	}

	// Create commit message
	commitMessage := g.CreateShipmentCommitMessage(versions, consignmentSummaries)

	var commitHash string
	if autoCommit {
		// Get files to commit
		filesToCommit := g.GetFilesToCommit(changelogPath, packages)

		// Commit changes
		hash, err := g.CommitShipmentChanges(filesToCommit, commitMessage)
		if err != nil {
			return fmt.Errorf("failed to commit shipment changes: %w", err)
		}
		commitHash = hash
	}

	// Create tags
	tags, err := g.CreateShipmentTags(versions, commitMessage)
	if err != nil {
		return fmt.Errorf("failed to create shipment tags: %w", err)
	}

	if autoPush {
		// Push commits and tags
		if err := g.client.Push(); err != nil {
			return fmt.Errorf("failed to push commits: %w", err)
		}

		if err := g.client.PushTags(); err != nil {
			return fmt.Errorf("failed to push tags: %w", err)
		}
	}

	// Log created tags
	fmt.Printf("\n🏷️  Git tags created:\n")
	for packageName, tagName := range tags {
		fmt.Printf("   - %s: %s\n", packageName, tagName)
	}

	if autoCommit {
		fmt.Printf("\n📝 Changes committed to git")
		if commitHash != "" {
			fmt.Printf(" (%s)", commitHash[:8])
		}
		fmt.Printf("\n")
	}

	if autoPush {
		fmt.Printf("🚀 Changes and tags pushed to remote\n")
	}

	return nil
}
