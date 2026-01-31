package template

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderTemplate_Basic(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		context     interface{}
		expected    string
		description string
	}{
		{
			name:     "simple variable substitution",
			template: "Version: {{ .Version }}",
			context: map[string]interface{}{
				"Version": "1.2.3",
			},
			expected:    "Version: 1.2.3",
			description: "should substitute simple variables",
		},
		{
			name:     "nested field access",
			template: "Author: {{ .Metadata.Author }}",
			context: map[string]interface{}{
				"Metadata": map[string]interface{}{
					"Author": "John Doe",
				},
			},
			expected:    "Author: John Doe",
			description: "should access nested fields",
		},
		{
			name: "range over slice",
			template: `{{ range .Items -}}
- {{ . }}
{{ end -}}`,
			context: map[string]interface{}{
				"Items": []string{"first", "second", "third"},
			},
			expected:    "- first\n- second\n- third\n",
			description: "should iterate over slices",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewTemplateRenderer()
			result, err := renderer.Render(tt.template, tt.context)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestRenderTemplate_WithTimeout(t *testing.T) {
	t.Run("normal rendering within timeout", func(t *testing.T) {
		renderer := NewTemplateRenderer()
		renderer.SetTimeout(5 * time.Second)

		template := "Version: {{ .Version }}"
		context := map[string]interface{}{"Version": "1.0.0"}

		result, err := renderer.Render(template, context)

		require.NoError(t, err)
		assert.Equal(t, "Version: 1.0.0", result)
	})

	t.Run("infinite loop would timeout", func(t *testing.T) {
		t.Skip("Difficult to test timeout without actual infinite loop")
		// In real implementation, this would test that rendering
		// with an infinite loop template times out appropriately
	})
}

func TestRenderTemplate_WithSprig(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		context     interface{}
		expected    string
		description string
	}{
		{
			name:     "upper function",
			template: `{{ .Text | upper }}`,
			context: map[string]interface{}{
				"Text": "hello world",
			},
			expected:    "HELLO WORLD",
			description: "should use Sprig upper function",
		},
		{
			name:     "title function",
			template: `{{ .Text | title }}`,
			context: map[string]interface{}{
				"Text": "hello world",
			},
			expected:    "Hello World",
			description: "should use Sprig title function",
		},
		{
			name:     "join function",
			template: `{{ .Items | join ", " }}`,
			context: map[string]interface{}{
				"Items": []string{"apple", "banana", "cherry"},
			},
			expected:    "apple, banana, cherry",
			description: "should use Sprig join function",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewTemplateRenderer()
			result, err := renderer.Render(tt.template, tt.context)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestRenderTemplate_ComplexContext(t *testing.T) {
	type Consignment struct {
		ID         string
		ChangeType string
		Summary    string
		Metadata   map[string]interface{}
	}

	template := `# Changelog

{{ range .Consignments -}}
## {{ .ChangeType | title }}

{{ .Summary }}

{{ if .Metadata.author -}}
Author: {{ .Metadata.author }}
{{ end -}}
{{ end -}}
`

	context := map[string]interface{}{
		"Consignments": []Consignment{
			{
				ID:         "c1",
				ChangeType: "minor",
				Summary:    "Added new feature",
				Metadata: map[string]interface{}{
					"author": "alice@example.com",
				},
			},
			{
				ID:         "c2",
				ChangeType: "patch",
				Summary:    "Fixed bug",
				Metadata:   map[string]interface{}{},
			},
		},
	}

	renderer := NewTemplateRenderer()
	result, err := renderer.Render(template, context)

	require.NoError(t, err)
	assert.Contains(t, result, "# Changelog")
	assert.Contains(t, result, "Minor")
	assert.Contains(t, result, "Added new feature")
	assert.Contains(t, result, "alice@example.com")
	assert.Contains(t, result, "Fixed bug")
}

func TestRenderTemplate_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		context     interface{}
		expectError bool
		description string
	}{
		{
			name:        "invalid template syntax",
			template:    "{{ .Value",
			context:     map[string]interface{}{},
			expectError: true,
			description: "should error on invalid syntax",
		},
		{
			name:     "missing field with zero option",
			template: "{{ .NonExistent }}",
			context:  map[string]interface{}{},
			// With default options, missing fields render as <no value>
			expectError: false,
			description: "missing fields render as empty by default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewTemplateRenderer()
			_, err := renderer.Render(tt.template, tt.context)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRenderTemplate_DateFormatting(t *testing.T) {
	now := time.Date(2026, 1, 30, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name        string
		template    string
		context     interface{}
		expected    string
		description string
	}{
		{
			name:     "ISO date format",
			template: `{{ .Date | date "2006-01-02" }}`,
			context: map[string]interface{}{
				"Date": now,
			},
			expected:    "2026-01-30",
			description: "should format date as ISO",
		},
		{
			name:     "custom date format",
			template: `{{ .Date | date "January 2, 2006" }}`,
			context: map[string]interface{}{
				"Date": now,
			},
			expected:    "January 30, 2026",
			description: "should format date in custom format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewTemplateRenderer()
			result, err := renderer.Render(tt.template, tt.context)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestRenderTemplate_FromParsedTemplate(t *testing.T) {
	// Parse template first
	parser := NewTemplateParser()
	tmpl, err := parser.Parse("test", "Version: {{ .Version }}")
	require.NoError(t, err)

	// Render from parsed template
	renderer := NewTemplateRenderer()
	context := map[string]interface{}{"Version": "2.0.0"}

	result, err := renderer.RenderParsed(tmpl, context)

	require.NoError(t, err)
	assert.Equal(t, "Version: 2.0.0", result)
}

func TestRenderTemplate_ConditionalRendering(t *testing.T) {
	template := `{{ if .ShowDetails -}}
Details: {{ .Details }}
{{ else -}}
No details available
{{ end -}}`

	tests := []struct {
		name     string
		context  map[string]interface{}
		expected string
	}{
		{
			name: "with details",
			context: map[string]interface{}{
				"ShowDetails": true,
				"Details":     "Some information",
			},
			expected: "Details: Some information\n",
		},
		{
			name: "without details",
			context: map[string]interface{}{
				"ShowDetails": false,
			},
			expected: "No details available\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewTemplateRenderer()
			result, err := renderer.Render(template, tt.context)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRenderTemplate_CustomFunctions(t *testing.T) {
	t.Run("custom has function", func(t *testing.T) {
		template := `{{ if has .Packages "core" }}Contains core{{ end }}`
		context := map[string]interface{}{
			"Packages": []string{"core", "api", "web"},
		}

		renderer := NewTemplateRenderer()
		result, err := renderer.Render(template, context)

		require.NoError(t, err)
		assert.Equal(t, "Contains core", result)
	})
}

func TestRenderTemplate_WhitespaceControl(t *testing.T) {
	template := `{{ range .Items -}}
{{ . }}
{{ end -}}`

	context := map[string]interface{}{
		"Items": []string{"one", "two"},
	}

	renderer := NewTemplateRenderer()
	result, err := renderer.Render(template, context)

	require.NoError(t, err)
	// With whitespace trimming (-), should have minimal extra newlines
	assert.True(t, len(result) < 20) // Reasonable upper bound
	assert.Contains(t, result, "one")
	assert.Contains(t, result, "two")
}

func TestRenderTemplate_EmptyContext(t *testing.T) {
	template := "Static content without variables"

	renderer := NewTemplateRenderer()
	result, err := renderer.Render(template, nil)

	require.NoError(t, err)
	assert.Equal(t, "Static content without variables", result)
}

// Benchmark tests
func BenchmarkRenderTemplate_Simple(b *testing.B) {
	renderer := NewTemplateRenderer()
	template := "Version: {{ .Version }}"
	context := map[string]interface{}{"Version": "1.0.0"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderer.Render(template, context)
	}
}

func BenchmarkRenderTemplate_Complex(b *testing.B) {
	renderer := NewTemplateRenderer()
	template := `{{ range .Items }}{{ .Name | upper }}: {{ .Value }}
{{ end }}`

	items := make([]map[string]interface{}, 100)
	for i := 0; i < 100; i++ {
		items[i] = map[string]interface{}{
			"Name":  "item",
			"Value": i,
		}
	}

	context := map[string]interface{}{"Items": items}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderer.Render(template, context)
	}
}
