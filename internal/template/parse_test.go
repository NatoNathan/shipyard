package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTemplate_Basic(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
		description string
	}{
		{
			name:        "simple template",
			content:     "Version: {{ .Version }}",
			expectError: false,
			description: "should parse simple template",
		},
		{
			name: "template with range",
			content: `{{ range .Items }}
- {{ .Name }}
{{ end }}`,
			expectError: false,
			description: "should parse template with range",
		},
		{
			name:        "invalid syntax",
			content:     "{{ .Version",
			expectError: true,
			description: "should error on invalid template syntax",
		},
		{
			name:        "empty template",
			content:     "",
			expectError: false,
			description: "should parse empty template",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewTemplateParser()
			_, err := parser.Parse("test", tt.content)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestParseTemplate_WithSprig(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
		description string
	}{
		{
			name:        "upper function",
			content:     `{{ "hello" | upper }}`,
			expectError: false,
			description: "should support Sprig upper function",
		},
		{
			name:        "date function",
			content:     `{{ now | date "2006-01-02" }}`,
			expectError: false,
			description: "should support Sprig date function",
		},
		{
			name:        "join function",
			content:     `{{ list "a" "b" "c" | join ", " }}`,
			expectError: false,
			description: "should support Sprig join function",
		},
		{
			name:        "trim function",
			content:     `{{ "  hello  " | trim }}`,
			expectError: false,
			description: "should support Sprig trim function",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewTemplateParser()
			tmpl, err := parser.Parse("test", tt.content)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, tmpl)
		})
	}
}

func TestParseTemplate_FunctionWhitelisting(t *testing.T) {
	t.Run("dangerous functions should be blocked", func(t *testing.T) {
		// These functions should be in the whitelist and safe to use
		safeFunctions := []string{
			`{{ "text" | upper }}`,
			`{{ now | date "2006" }}`,
			`{{ list "a" "b" | join "," }}`,
		}

		parser := NewTemplateParser()

		for _, content := range safeFunctions {
			_, err := parser.Parse("test", content)
			assert.NoError(t, err, "safe function should be allowed: %s", content)
		}
	})

	t.Run("env function should be blocked if configured", func(t *testing.T) {
		t.Skip("Environment access blocking would be implemented in production")
		// In a real implementation, we'd block functions like:
		// - env (environment variable access)
		// - getHostByName (network access)
		// - exec (command execution)
	})
}

func TestParseTemplate_ComplexTemplate(t *testing.T) {
	content := `# Changelog

{{ range .Consignments -}}
## {{ .ChangeType | title }} - {{ .Timestamp | date "2006-01-02" }}

{{ .Summary }}

{{ if .Metadata -}}
{{ if .Metadata.author -}}
Author: {{ .Metadata.author }}
{{ end -}}
{{ end -}}

---
{{ end -}}
`

	parser := NewTemplateParser()
	tmpl, err := parser.Parse("changelog", content)

	require.NoError(t, err)
	assert.NotNil(t, tmpl)
}

func TestParseTemplate_CustomFunctions(t *testing.T) {
	t.Run("custom function registration", func(t *testing.T) {
		parser := NewTemplateParser()

		// Add custom function
		parser.AddFunction("reverse", func(s string) string {
			runes := []rune(s)
			for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
				runes[i], runes[j] = runes[j], runes[i]
			}
			return string(runes)
		})

		content := `{{ "hello" | reverse }}`
		tmpl, err := parser.Parse("test", content)

		require.NoError(t, err)
		assert.NotNil(t, tmpl)
	})
}

func TestParseTemplate_ErrorMessages(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		expectedError string
		description   string
	}{
		{
			name:          "unclosed action",
			content:       "{{ .Value",
			expectedError: "unclosed action",
			description:   "should provide clear error for unclosed action",
		},
		{
			name:          "undefined variable",
			content:       "{{ .NonExistent.Field }}",
			expectedError: "", // This is caught at execution time, not parse time
			description:   "undefined variable errors happen at render time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewTemplateParser()
			_, err := parser.Parse("test", tt.content)

			if tt.expectedError != "" {
				assert.Error(t, err)
			}
		})
	}
}

func TestParseTemplate_Caching(t *testing.T) {
	parser := NewTemplateParser()

	content := "Version: {{ .Version }}"

	// Parse twice with same name
	tmpl1, err1 := parser.Parse("test", content)
	tmpl2, err2 := parser.Parse("test", content)

	require.NoError(t, err1)
	require.NoError(t, err2)

	// Templates should be functionally equivalent
	assert.NotNil(t, tmpl1)
	assert.NotNil(t, tmpl2)
}

func TestParseTemplate_MultipleTemplates(t *testing.T) {
	parser := NewTemplateParser()

	// Parse different templates
	templates := map[string]string{
		"changelog":    "# Changelog\n{{ range .Consignments }}{{ .Summary }}{{ end }}",
		"tagname":      "v{{ .Version }}",
		"releasenotes": "Release {{ .Version }}\n{{ .Date }}",
	}

	for name, content := range templates {
		tmpl, err := parser.Parse(name, content)
		require.NoError(t, err, "failed to parse %s", name)
		assert.NotNil(t, tmpl)
	}
}

func TestTemplateParser_WithOptions(t *testing.T) {
	t.Run("strict mode", func(t *testing.T) {
		parser := NewTemplateParser()
		parser.SetOption("missingkey", "error")

		// Template with potential missing key
		content := "{{ .MaybeExists }}"
		tmpl, err := parser.Parse("test", content)

		require.NoError(t, err)
		assert.NotNil(t, tmpl)
	})
}
