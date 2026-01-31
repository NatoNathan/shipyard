package template

import (
	"bytes"
	"context"
	"fmt"
	"text/template"
	"time"
)

// TemplateRenderer handles rendering templates with context and timeout
type TemplateRenderer struct {
	parser  *TemplateParser
	timeout time.Duration
}

// NewTemplateRenderer creates a new template renderer
func NewTemplateRenderer() *TemplateRenderer {
	return &TemplateRenderer{
		parser:  NewTemplateParser(),
		timeout: 30 * time.Second, // Default 30 second timeout
	}
}

// SetTimeout sets the maximum time allowed for template rendering
func (r *TemplateRenderer) SetTimeout(timeout time.Duration) {
	r.timeout = timeout
}

// Render renders a template string with the given context
func (r *TemplateRenderer) Render(templateContent string, ctx interface{}) (string, error) {
	// Parse the template
	tmpl, err := r.parser.Parse("template", templateContent)
	if err != nil {
		return "", err
	}

	// Render the parsed template
	return r.RenderParsed(tmpl, ctx)
}

// RenderParsed renders a pre-parsed template with the given context
func (r *TemplateRenderer) RenderParsed(tmpl *template.Template, ctx interface{}) (string, error) {
	var buf bytes.Buffer

	// Create context with timeout
	renderCtx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	// Channel to signal completion
	done := make(chan error, 1)

	// Execute template in goroutine to allow timeout
	go func() {
		done <- tmpl.Execute(&buf, ctx)
	}()

	// Wait for completion or timeout
	select {
	case err := <-done:
		if err != nil {
			return "", fmt.Errorf("template execution failed: %w", err)
		}
		return buf.String(), nil
	case <-renderCtx.Done():
		return "", fmt.Errorf("template rendering timed out after %v", r.timeout)
	}
}

// RenderWithName renders a template with a specific name (useful for error messages)
func (r *TemplateRenderer) RenderWithName(name, templateContent string, ctx interface{}) (string, error) {
	tmpl, err := r.parser.Parse(name, templateContent)
	if err != nil {
		return "", err
	}

	return r.RenderParsed(tmpl, ctx)
}

// MustRender renders a template and panics on error (useful for testing)
func (r *TemplateRenderer) MustRender(templateContent string, ctx interface{}) string {
	result, err := r.Render(templateContent, ctx)
	if err != nil {
		panic(fmt.Sprintf("template rendering failed: %v", err))
	}
	return result
}
