package consignment

import (
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
// Returns error if any deletion fails
func DeleteConsignments(paths []string) error {
	// Early return for empty list
	if len(paths) == 0 {
		return nil
	}

	// Delete each file
	for _, path := range paths {
		if err := DeleteConsignment(path); err != nil {
			return err
		}
	}

	return nil
}

// DeleteConsignmentByID deletes a consignment file by its ID
// Returns error if file doesn't exist or deletion fails
func DeleteConsignmentByID(consignmentsDir, id string) error {
	path := filepath.Join(consignmentsDir, id+".md")
	return DeleteConsignment(path)
}

// DeleteConsignmentsByIDs deletes multiple consignment files by their IDs
// Returns error if any deletion fails
func DeleteConsignmentsByIDs(consignmentsDir string, ids []string) error {
	// Early return for empty list
	if len(ids) == 0 {
		return nil
	}

	// Build paths and delete
	paths := make([]string, len(ids))
	for i, id := range ids {
		paths[i] = filepath.Join(consignmentsDir, id+".md")
	}

	return DeleteConsignments(paths)
}
