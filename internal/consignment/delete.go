package consignment

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// DeleteConsignment deletes a single consignment file
// Returns error if file doesn't exist or deletion fails
func DeleteConsignment(path string) error {
	// Check if path is a directory
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to delete consignment: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("failed to delete consignment: path is a directory")
	}

	// Delete the file
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete consignment: %w", err)
	}

	return nil
}

// DeleteConsignments deletes multiple consignment files
// Collects all errors instead of failing on first error
func DeleteConsignments(paths []string) error {
	if len(paths) == 0 {
		return nil
	}

	var errs []error
	for _, path := range paths {
		if err := DeleteConsignment(path); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// DeleteConsignmentByID deletes a consignment file by its ID
// Returns error if file doesn't exist or deletion fails
func DeleteConsignmentByID(consignmentsDir, id string) error {
	path := filepath.Join(consignmentsDir, id+".md")
	return DeleteConsignment(path)
}

// DeleteConsignmentsByIDs deletes multiple consignment files by their IDs
// Collects all errors instead of failing on first error
func DeleteConsignmentsByIDs(consignmentsDir string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	var errs []error
	for _, id := range ids {
		path := filepath.Join(consignmentsDir, id+".md")
		if err := os.Remove(path); err != nil {
			errs = append(errs, fmt.Errorf("failed to delete %s: %w", id, err))
		}
	}

	return errors.Join(errs...)
}
