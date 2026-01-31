package consignment

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"github.com/NatoNathan/shipyard/pkg/types"
)

// Consignment represents a recorded change to one or more packages
type Consignment struct {
	ID         string                 `yaml:"id"`
	Timestamp  time.Time              `yaml:"timestamp"`
	Packages   []string               `yaml:"packages"`
	ChangeType types.ChangeType       `yaml:"changeType"`
	Summary    string                 `yaml:"-"` // Stored in markdown body
	Metadata   map[string]interface{} `yaml:"metadata,omitempty"`
}

// GenerateIDFromTime generates a unique consignment ID from a timestamp
// Format: YYYYMMDD-HHMMSS-{random6}
// Deprecated: Use GenerateID(timestamp) instead
func GenerateIDFromTime(timestamp time.Time) (string, error) {
	// Format date and time components
	dateTime := timestamp.Format("20060102-150405")

	// Generate 6-character random alphanumeric string (lowercase)
	randomBytes := make([]byte, 6)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Convert to lowercase alphanumeric (a-z, 0-9)
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	for i := range randomBytes {
		randomBytes[i] = charset[int(randomBytes[i])%len(charset)]
	}

	return fmt.Sprintf("%s-%s", dateTime, string(randomBytes)), nil
}

// New creates a new Consignment with generated ID and timestamp
func New(packages []string, changeType types.ChangeType, summary string, metadata map[string]interface{}) (*Consignment, error) {
	timestamp := time.Now().UTC()
	id, err := GenerateID(timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ID: %w", err)
	}

	return &Consignment{
		ID:         id,
		Timestamp:  timestamp,
		Packages:   packages,
		ChangeType: changeType,
		Summary:    summary,
		Metadata:   metadata,
	}, nil
}

// Validate checks if the consignment is valid
func (c *Consignment) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("ID is required")
	}
	
	if c.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required")
	}
	
	if len(c.Packages) == 0 {
		return fmt.Errorf("at least one package is required")
	}
	
	if err := c.ChangeType.Validate(); err != nil {
		return fmt.Errorf("invalid change type: %w", err)
	}
	
	if strings.TrimSpace(c.Summary) == "" {
		return fmt.Errorf("summary is required")
	}
	
	return nil
}

// AffectsPackage checks if this consignment affects the specified package
func (c *Consignment) AffectsPackage(packageName string) bool {
	for _, pkg := range c.Packages {
		if pkg == packageName {
			return true
		}
	}
	return false
}
