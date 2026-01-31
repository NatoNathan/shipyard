package consignment

import (
	"crypto/rand"
	"fmt"
	"time"
)

// GenerateID generates a unique consignment ID with format: YYYYMMDD-HHMMSS-random6
// This is the main ID generation function that should be used for creating new consignments
func GenerateID(timestamp time.Time) (string, error) {
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
