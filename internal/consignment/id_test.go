package consignment

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateID_Format tests the format of generated IDs
func TestGenerateID_Format(t *testing.T) {
	timestamp := time.Date(2026, 1, 30, 14, 30, 22, 0, time.UTC)
	id, err := GenerateID(timestamp)
	require.NoError(t, err, "Should generate ID without error")

	// Verify format: YYYYMMDD-HHMMSS-random6
	pattern := regexp.MustCompile(`^\d{8}-\d{6}-[a-z0-9]{6}$`)
	assert.True(t, pattern.MatchString(id), "ID should match format YYYYMMDD-HHMMSS-random6")

	// Verify date component
	assert.True(t, id[0:8] == "20260130", "Date component should be 20260130")
	assert.True(t, id[9:15] == "143022", "Time component should be 143022")
}

// TestGenerateID_Uniqueness tests that generated IDs are unique
func TestGenerateID_Uniqueness(t *testing.T) {
	timestamp := time.Date(2026, 1, 30, 14, 30, 22, 0, time.UTC)

	// Generate multiple IDs with same timestamp
	ids := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		id, err := GenerateID(timestamp)
		require.NoError(t, err, "Should generate ID without error")

		// Check uniqueness
		if ids[id] {
			t.Fatalf("Generated duplicate ID: %s", id)
		}
		ids[id] = true
	}

	assert.Equal(t, iterations, len(ids), "Should have generated unique IDs")
}

// TestGenerateID_DifferentTimestamps tests IDs with different timestamps
func TestGenerateID_DifferentTimestamps(t *testing.T) {
	timestamp1 := time.Date(2026, 1, 30, 14, 30, 22, 0, time.UTC)
	timestamp2 := time.Date(2026, 1, 31, 9, 15, 45, 0, time.UTC)

	id1, err := GenerateID(timestamp1)
	require.NoError(t, err, "Should generate first ID without error")

	id2, err := GenerateID(timestamp2)
	require.NoError(t, err, "Should generate second ID without error")

	// Verify different timestamps produce different date/time components
	assert.NotEqual(t, id1[0:15], id2[0:15], "Date/time components should be different")

	// Both should match format
	pattern := regexp.MustCompile(`^\d{8}-\d{6}-[a-z0-9]{6}$`)
	assert.True(t, pattern.MatchString(id1), "First ID should match format")
	assert.True(t, pattern.MatchString(id2), "Second ID should match format")
}

// TestGenerateID_RandomComponent tests the random component
func TestGenerateID_RandomComponent(t *testing.T) {
	timestamp := time.Date(2026, 1, 30, 14, 30, 22, 0, time.UTC)

	id, err := GenerateID(timestamp)
	require.NoError(t, err, "Should generate ID without error")

	// Extract random component
	randomPart := id[16:22]

	// Verify it's 6 characters of lowercase alphanumeric
	pattern := regexp.MustCompile(`^[a-z0-9]{6}$`)
	assert.True(t, pattern.MatchString(randomPart), "Random component should be 6 lowercase alphanumeric characters")
}

// TestGenerateID_Consistency tests that the same timestamp produces same date/time component
func TestGenerateID_Consistency(t *testing.T) {
	timestamp := time.Date(2026, 1, 30, 14, 30, 22, 0, time.UTC)

	id1, err := GenerateID(timestamp)
	require.NoError(t, err, "Should generate first ID without error")

	id2, err := GenerateID(timestamp)
	require.NoError(t, err, "Should generate second ID without error")

	// Same timestamp should produce same date/time component
	assert.Equal(t, id1[0:15], id2[0:15], "Date/time components should be the same")

	// But random components should be different
	assert.NotEqual(t, id1[16:22], id2[16:22], "Random components should be different")
}
