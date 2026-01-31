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

func TestStageFiles(t *testing.T) {
	tests := []struct {
		name      string
		files     []string
		wantErr   bool
		errMsg    string
		setupFunc func(t *testing.T, repoPath string) []string
	}{
		{
			name:    "success with multiple files",
			wantErr: false,
			setupFunc: func(t *testing.T, repoPath string) []string {
				file1 := filepath.Join(repoPath, "file1.txt")
				file2 := filepath.Join(repoPath, "file2.txt")
				require.NoError(t, os.WriteFile(file1, []byte("content1"), 0644))
				require.NoError(t, os.WriteFile(file2, []byte("content2"), 0644))
				return []string{file1, file2}
			},
		},
		{
			name:    "success with single file",
			wantErr: false,
			setupFunc: func(t *testing.T, repoPath string) []string {
				file := filepath.Join(repoPath, "single.txt")
				require.NoError(t, os.WriteFile(file, []byte("content"), 0644))
				return []string{file}
			},
		},
		{
			name:    "success with empty file list",
			wantErr: false,
			setupFunc: func(t *testing.T, repoPath string) []string {
				return []string{}
			},
		},
		{
			name:    "success with relative paths",
			wantErr: false,
			setupFunc: func(t *testing.T, repoPath string) []string {
				file := filepath.Join(repoPath, "relative.txt")
				require.NoError(t, os.WriteFile(file, []byte("content"), 0644))
				// Return just the filename (relative path)
				return []string{"relative.txt"}
			},
		},
		{
			name:    "error with invalid repository",
			wantErr: true,
			errMsg:  "failed to open repository",
			setupFunc: func(t *testing.T, repoPath string) []string {
				return []string{"file.txt"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir, err := os.MkdirTemp("", "shipyard-stage-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir)

			var repoPath string
			if tt.name == "error with invalid repository" {
				// Don't initialize git repo for this test
				repoPath = tmpDir
			} else {
				// Initialize git repository
				repo, err := gogit.PlainInit(tmpDir, false)
				require.NoError(t, err)

				// Create initial commit
				worktree, err := repo.Worktree()
				require.NoError(t, err)

				initFile := filepath.Join(tmpDir, "init.txt")
				require.NoError(t, os.WriteFile(initFile, []byte("init"), 0644))
				_, err = worktree.Add("init.txt")
				require.NoError(t, err)

				_, err = worktree.Commit("Initial commit", &gogit.CommitOptions{
					Author: &object.Signature{
						Name:  "Test",
						Email: "test@example.com",
					},
				})
				require.NoError(t, err)

				repoPath = tmpDir
			}

			// Setup test files
			files := tt.setupFunc(t, repoPath)

			// Execute
			err = StageFiles(repoPath, files)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)

				// Verify files are staged (except for empty list case)
				if len(files) > 0 {
					repo, err := gogit.PlainOpen(repoPath)
					require.NoError(t, err)

					worktree, err := repo.Worktree()
					require.NoError(t, err)

					status, err := worktree.Status()
					require.NoError(t, err)

					for _, file := range files {
						// Handle both absolute and relative paths
						relPath := file
						if filepath.IsAbs(file) {
							relPath, _ = filepath.Rel(repoPath, file)
						}
						fileStatus := status.File(relPath)
						assert.NotEqual(t, gogit.Untracked, fileStatus.Staging, "File %s should be staged", relPath)
					}
				}
			}
		})
	}
}

func TestStageFiles_NonExistentFile(t *testing.T) {
	// Create temp directory with git repo
	tmpDir, err := os.MkdirTemp("", "shipyard-stage-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	repo, err := gogit.PlainInit(tmpDir, false)
	require.NoError(t, err)

	// Create initial commit
	worktree, err := repo.Worktree()
	require.NoError(t, err)

	initFile := filepath.Join(tmpDir, "init.txt")
	require.NoError(t, os.WriteFile(initFile, []byte("init"), 0644))
	_, err = worktree.Add("init.txt")
	require.NoError(t, err)

	_, err = worktree.Commit("Initial commit", &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
		},
	})
	require.NoError(t, err)

	// Try to stage non-existent file
	nonExistentFile := filepath.Join(tmpDir, "does-not-exist.txt")
	err = StageFiles(tmpDir, []string{nonExistentFile})

	// go-git should handle this gracefully or return an error
	// The behavior depends on the version, so we just check it doesn't panic
	assert.NotNil(t, err, "Should return error for non-existent file")
}
