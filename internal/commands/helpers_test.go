package commands

import (
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
