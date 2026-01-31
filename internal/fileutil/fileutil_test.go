package fileutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	
	content := []byte("test content")
	err := AtomicWrite(filePath, content, 0644)
	require.NoError(t, err)
	
	// Verify file exists and has correct content
	data, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, content, data)
	
	// Verify file permissions
	info, err := os.Stat(filePath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0644), info.Mode().Perm())
}

func TestAtomicWrite_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	
	// Write initial content
	err := AtomicWrite(filePath, []byte("initial"), 0644)
	require.NoError(t, err)
	
	// Overwrite with new content
	newContent := []byte("updated content")
	err = AtomicWrite(filePath, newContent, 0644)
	require.NoError(t, err)
	
	// Verify new content
	data, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, newContent, data)
}

func TestEnsureDir(t *testing.T) {
	tmpDir := t.TempDir()
	
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "create single directory",
			path:    filepath.Join(tmpDir, "test"),
			wantErr: false,
		},
		{
			name:    "create nested directories",
			path:    filepath.Join(tmpDir, "a", "b", "c"),
			wantErr: false,
		},
		{
			name:    "existing directory",
			path:    tmpDir,
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := EnsureDir(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			
			// Verify directory exists
			info, err := os.Stat(tt.path)
			require.NoError(t, err)
			assert.True(t, info.IsDir())
		})
	}
}

func TestPathExists(t *testing.T) {
	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "exists.txt")
	err := os.WriteFile(existingFile, []byte("test"), 0644)
	require.NoError(t, err)
	
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "existing file",
			path: existingFile,
			want: true,
		},
		{
			name: "existing directory",
			path: tmpDir,
			want: true,
		},
		{
			name: "non-existing path",
			path: filepath.Join(tmpDir, "does-not-exist.txt"),
			want: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PathExists(tt.path)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsDir(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "file.txt")
	err := os.WriteFile(file, []byte("test"), 0644)
	require.NoError(t, err)
	
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "directory",
			path: tmpDir,
			want: true,
		},
		{
			name: "file",
			path: file,
			want: false,
		},
		{
			name: "non-existing",
			path: filepath.Join(tmpDir, "nope"),
			want: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsDir(tt.path)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestReadYAMLFile(t *testing.T) {
	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "test.yaml")
	
	content := `
name: test
value: 123
nested:
  key: value
`
	err := os.WriteFile(yamlFile, []byte(content), 0644)
	require.NoError(t, err)
	
	var result map[string]interface{}
	err = ReadYAMLFile(yamlFile, &result)
	require.NoError(t, err)
	
	assert.Equal(t, "test", result["name"])
	assert.Equal(t, 123, result["value"])
	assert.NotNil(t, result["nested"])
}

func TestWriteYAMLFile(t *testing.T) {
	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "output.yaml")
	
	data := map[string]interface{}{
		"name":  "test",
		"value": 123,
		"nested": map[string]interface{}{
			"key": "value",
		},
	}
	
	err := WriteYAMLFile(yamlFile, data, 0644)
	require.NoError(t, err)
	
	// Verify file was written
	assert.True(t, PathExists(yamlFile))
	
	// Read back and verify content
	var result map[string]interface{}
	err = ReadYAMLFile(yamlFile, &result)
	require.NoError(t, err)
	assert.Equal(t, data["name"], result["name"])
}
