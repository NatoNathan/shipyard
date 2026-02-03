package commands

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// changeToDir changes to the specified directory and returns a cleanup function
func changeToDir(t *testing.T, dir string) func() {
	t.Helper()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	return func() { os.Chdir(oldDir) }
}

// captureOutput captures stdout during function execution
func captureOutput(f func()) string {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run function
	f()

	// Restore stdout
	w.Close()
	os.Stdout = old

	// Read captured output
	var buf []byte
	buf = make([]byte, 4096)
	n, _ := r.Read(buf)
	return string(buf[:n])
}

// assertJSONOutput validates that output is valid JSON and contains the expected keys
func assertJSONOutput(t *testing.T, output string, expectedKeys ...string) {
	t.Helper()

	var data map[string]interface{}
	err := json.Unmarshal([]byte(output), &data)
	require.NoError(t, err, "Output should be valid JSON")

	for _, key := range expectedKeys {
		_, exists := data[key]
		require.True(t, exists, "JSON output should contain key: %s", key)
	}
}

// assertContainsEmoji validates that output contains the specified emoji
func assertContainsEmoji(t *testing.T, output, emoji string) {
	t.Helper()
	require.Contains(t, output, emoji, "Output should contain emoji: %s", emoji)
}

// assertContainsArrow validates that output contains the → arrow character
func assertContainsArrow(t *testing.T, output string) {
	t.Helper()
	require.Contains(t, output, "→", "Output should contain → arrow character")
}
