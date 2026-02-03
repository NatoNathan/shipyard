package errors

import (
	"errors"
	"fmt"
	"strings"
)

// Common errors
var (
	ErrNotInitialized = errors.New("shipyard not initialized (run 'shipyard init')")
)

// NotInitializedError indicates that shipyard has not been initialized in the repository
type NotInitializedError struct {
	Path string
}

func (e *NotInitializedError) Error() string {
	return fmt.Sprintf("shipyard not initialized in %s (run 'shipyard init')", e.Path)
}

// NewNotInitializedError creates a new NotInitializedError
func NewNotInitializedError(path string) error {
	return &NotInitializedError{Path: path}
}

// ValidationError indicates a validation failure for a specific field
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s: %s", e.Field, e.Message)
}

// NewValidationError creates a new ValidationError
func NewValidationError(field, message string) error {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// ConfigError indicates a configuration-related error
type ConfigError struct {
	Message string
	Cause   error
}

func (e *ConfigError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("config error: %s: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("config error: %s", e.Message)
}

func (e *ConfigError) Unwrap() error {
	return e.Cause
}

// NewConfigError creates a new ConfigError
func NewConfigError(message string, cause error) error {
	return &ConfigError{
		Message: message,
		Cause:   cause,
	}
}

// GitError indicates a git operation error
type GitError struct {
	Message string
	Cause   error
}

func (e *GitError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("git error: %s: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("git error: %s", e.Message)
}

func (e *GitError) Unwrap() error {
	return e.Cause
}

// NewGitError creates a new GitError
func NewGitError(message string, cause error) error {
	return &GitError{
		Message: message,
		Cause:   cause,
	}
}

// ConsignmentError indicates an error related to a specific consignment
type ConsignmentError struct {
	ID      string
	Message string
}

func (e *ConsignmentError) Error() string {
	return fmt.Sprintf("consignment error [%s]: %s", e.ID, e.Message)
}

// NewConsignmentError creates a new ConsignmentError
func NewConsignmentError(id, message string) error {
	return &ConsignmentError{
		ID:      id,
		Message: message,
	}
}

// DependencyError indicates an error in the dependency graph
type DependencyError struct {
	Message string
	Cycle   []string
}

func (e *DependencyError) Error() string {
	if len(e.Cycle) > 0 {
		return fmt.Sprintf("dependency error: %s: %s", e.Message, strings.Join(e.Cycle, " -> "))
	}
	return fmt.Sprintf("dependency error: %s", e.Message)
}

// NewDependencyError creates a new DependencyError
func NewDependencyError(message string, cycle []string) error {
	return &DependencyError{
		Message: message,
		Cycle:   cycle,
	}
}

// UpgradeError indicates an error during the upgrade process
type UpgradeError struct {
	Message string
	Cause   error
}

func (e *UpgradeError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("upgrade error: %s: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("upgrade error: %s", e.Message)
}

func (e *UpgradeError) Unwrap() error {
	return e.Cause
}

// NewUpgradeError creates a new UpgradeError
func NewUpgradeError(message string, cause error) error {
	return &UpgradeError{
		Message: message,
		Cause:   cause,
	}
}

// NetworkError indicates a network-related error
type NetworkError struct {
	Message string
	Cause   error
}

func (e *NetworkError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("network error: %s: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("network error: %s", e.Message)
}

func (e *NetworkError) Unwrap() error {
	return e.Cause
}

// NewNetworkError creates a new NetworkError
func NewNetworkError(message string, cause error) error {
	return &NetworkError{
		Message: message,
		Cause:   cause,
	}
}
