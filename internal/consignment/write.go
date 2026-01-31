package consignment

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/NatoNathan/shipyard/internal/fileutil"
)

// WriteConsignment writes a consignment to a markdown file with atomic write
func WriteConsignment(cons *Consignment, dir string) error {
	// Ensure directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create consignments directory: %w", err)
	}

	// Serialize consignment
	content, err := Serialize(cons)
	if err != nil {
		return fmt.Errorf("failed to serialize consignment: %w", err)
	}

	// Build file path
	filename := fmt.Sprintf("%s.md", cons.ID)
	filePath := filepath.Join(dir, filename)

	// Write atomically
	if err := fileutil.AtomicWrite(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write consignment file: %w", err)
	}

	return nil
}
