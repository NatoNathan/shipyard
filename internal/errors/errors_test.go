package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotInitializedError(t *testing.T) {
	err := NewNotInitializedError("/path/to/repo")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
	assert.Contains(t, err.Error(), "/path/to/repo")
	
	var notInitErr *NotInitializedError
	assert.True(t, errors.As(err, &notInitErr))
	assert.Equal(t, "/path/to/repo", notInitErr.Path)
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		message string
	}{
		{
			name:    "simple field error",
			field:   "packages",
			message: "cannot be empty",
		},
		{
			name:    "nested field error",
			field:   "packages[0].name",
			message: "is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewValidationError(tt.field, tt.message)
			
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.field)
			assert.Contains(t, err.Error(), tt.message)
			
			var valErr *ValidationError
			assert.True(t, errors.As(err, &valErr))
			assert.Equal(t, tt.field, valErr.Field)
			assert.Equal(t, tt.message, valErr.Message)
		})
	}
}

func TestConfigError(t *testing.T) {
	innerErr := errors.New("file not found")
	err := NewConfigError("failed to load config", innerErr)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load config")
	assert.Contains(t, err.Error(), "file not found")
	
	var cfgErr *ConfigError
	assert.True(t, errors.As(err, &cfgErr))
	assert.Equal(t, "failed to load config", cfgErr.Message)
	assert.Equal(t, innerErr, cfgErr.Cause)
}

func TestGitError(t *testing.T) {
	innerErr := errors.New("not a git repository")
	err := NewGitError("git operation failed", innerErr)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git operation failed")
	
	var gitErr *GitError
	assert.True(t, errors.As(err, &gitErr))
}

func TestConsignmentError(t *testing.T) {
	err := NewConsignmentError("20260130-143022-abc123", "invalid format")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "20260130-143022-abc123")
	assert.Contains(t, err.Error(), "invalid format")
	
	var consErr *ConsignmentError
	assert.True(t, errors.As(err, &consErr))
	assert.Equal(t, "20260130-143022-abc123", consErr.ID)
}

func TestDependencyError(t *testing.T) {
	err := NewDependencyError("circular dependency detected", []string{"pkg-a", "pkg-b", "pkg-a"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
	assert.Contains(t, err.Error(), "pkg-a")
	assert.Contains(t, err.Error(), "pkg-b")

	var depErr *DependencyError
	assert.True(t, errors.As(err, &depErr))
	assert.Equal(t, []string{"pkg-a", "pkg-b", "pkg-a"}, depErr.Cycle)
}

func TestExitCodeError(t *testing.T) {
	err := NewExitCodeError(1, "command failed")

	assert.Error(t, err)
	assert.Equal(t, "command failed", err.Error())

	var exitErr *ExitCodeError
	assert.True(t, errors.As(err, &exitErr))
	assert.Equal(t, 1, exitErr.Code)
	assert.Equal(t, "command failed", exitErr.Message)
	assert.Nil(t, exitErr.Cause)
	assert.Nil(t, exitErr.Unwrap())
}

func TestExitCodeErrorWithCause(t *testing.T) {
	cause := errors.New("underlying failure")
	err := NewExitCodeErrorWithCause(2, "command failed", cause)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "command failed")
	assert.Contains(t, err.Error(), "underlying failure")

	var exitErr *ExitCodeError
	assert.True(t, errors.As(err, &exitErr))
	assert.Equal(t, 2, exitErr.Code)
	assert.Equal(t, cause, exitErr.Cause)

	// Test Unwrap
	assert.Equal(t, cause, errors.Unwrap(err))

	// Test errors.Is through Unwrap chain
	assert.True(t, errors.Is(err, cause))
}

func TestExitCodeErrorWithNilCause(t *testing.T) {
	err := NewExitCodeErrorWithCause(1, "no cause", nil)

	assert.Error(t, err)
	assert.Equal(t, "no cause", err.Error())

	var exitErr *ExitCodeError
	assert.True(t, errors.As(err, &exitErr))
	assert.Nil(t, exitErr.Unwrap())
}
