package ecosystem

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReplaceTopLevelJSONVersion(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		newVersion string
		expected   string
		wantErr    bool
	}{
		{
			name:       "simple version field",
			input:      "{\n  \"name\": \"pkg\",\n  \"version\": \"1.0.0\"\n}",
			newVersion: "2.0.0",
			expected:   "{\n  \"name\": \"pkg\",\n  \"version\": \"2.0.0\"\n}",
		},
		{
			name: "nested version field is not touched",
			input: `{
  "name": "pkg",
  "version": "1.0.0",
  "config": {
    "version": "0.5.0"
  },
  "dependencies": {
    "some-lib": "1.0.0"
  }
}`,
			newVersion: "2.0.0",
			expected: `{
  "name": "pkg",
  "version": "2.0.0",
  "config": {
    "version": "0.5.0"
  },
  "dependencies": {
    "some-lib": "1.0.0"
  }
}`,
		},
		{
			name: "nested version before top-level version",
			input: `{
  "scripts": {
    "version": "echo versioning"
  },
  "version": "1.0.0"
}`,
			newVersion: "1.1.0",
			expected: `{
  "scripts": {
    "version": "echo versioning"
  },
  "version": "1.1.0"
}`,
		},
		{
			name:       "string value equal to version is not treated as key",
			input:      "{\n  \"name\": \"version\",\n  \"version\": \"1.0.0\"\n}",
			newVersion: "3.0.0",
			expected:   "{\n  \"name\": \"version\",\n  \"version\": \"3.0.0\"\n}",
		},
		{
			name: "version key inside array of objects is not touched",
			input: `{
  "entries": [
    {"version": "9.9.9"}
  ],
  "version": "1.0.0"
}`,
			newVersion: "1.0.1",
			expected: `{
  "entries": [
    {"version": "9.9.9"}
  ],
  "version": "1.0.1"
}`,
		},
		{
			name: "jsonc comments preserved",
			input: `{
  // package version
  "version": "1.0.0", /* keep me */
  "name": "pkg"
}`,
			newVersion: "2.0.0",
			expected: `{
  // package version
  "version": "2.0.0", /* keep me */
  "name": "pkg"
}`,
		},
		{
			name: "comment containing version key is ignored",
			input: `{
  // "version": "0.0.0"
  "version": "1.0.0"
}`,
			newVersion: "2.0.0",
			expected: `{
  // "version": "0.0.0"
  "version": "2.0.0"
}`,
		},
		{
			name:       "escaped quotes in preceding values",
			input:      "{\n  \"description\": \"say \\\"hi\\\"\",\n  \"version\": \"1.0.0\"\n}",
			newVersion: "1.2.3",
			expected:   "{\n  \"description\": \"say \\\"hi\\\"\",\n  \"version\": \"1.2.3\"\n}",
		},
		{
			name:       "missing version field",
			input:      "{\n  \"name\": \"pkg\"\n}",
			newVersion: "1.0.0",
			wantErr:    true,
		},
		{
			name:       "only nested version field",
			input:      "{\n  \"config\": {\"version\": \"1.0.0\"}\n}",
			newVersion: "2.0.0",
			wantErr:    true,
		},
		{
			name:       "non-string top-level version",
			input:      "{\n  \"version\": 1\n}",
			newVersion: "2.0.0",
			wantErr:    true,
		},
		{
			name:       "unterminated version string",
			input:      "{\n  \"version\": \"1.0.0",
			newVersion: "2.0.0",
			wantErr:    true,
		},
		{
			name:       "unterminated version string ending in backslash",
			input:      "{\n  \"version\": \"1.0.0\\",
			newVersion: "2.0.0",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := replaceTopLevelJSONVersion([]byte(tt.input), tt.newVersion)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(result))
		})
	}
}
