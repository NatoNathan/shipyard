package template

import (
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

// TemplateParser handles parsing go templates with Sprig functions
type TemplateParser struct {
	funcMap            template.FuncMap
	options            map[string]string
	allowEnvAccess     bool
	unsafeSprigFuncMap template.FuncMap
}

// NewTemplateParser creates a new template parser with safe Sprig functions
func NewTemplateParser() *TemplateParser {
	parser := &TemplateParser{
		options:            make(map[string]string),
		unsafeSprigFuncMap: sprig.TxtFuncMap(),
	}

	// Initialize with safe Sprig functions
	parser.funcMap = getSafeSprigFunctions()

	return parser
}

// Parse parses a template string and returns a compiled template
func (p *TemplateParser) Parse(name, content string) (*template.Template, error) {
	// Create new template with name
	tmpl := template.New(name)

	// Add function map
	tmpl = tmpl.Funcs(p.funcMap)

	// Apply options
	for key, value := range p.options {
		tmpl = tmpl.Option(fmt.Sprintf("%s=%s", key, value))
	}

	// Parse the template content
	parsed, err := tmpl.Parse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return parsed, nil
}

// AddFunction adds a custom function to the function map.
func (p *TemplateParser) AddFunction(name string, fn interface{}) {
	if isBlockedTemplateFunction(name) && (!p.allowEnvAccess || !isEnvironmentTemplateFunction(name)) {
		return
	}
	p.funcMap[name] = fn
}

// EnableEnvironmentAccess enables Sprig environment lookup functions for trusted templates.
func (p *TemplateParser) EnableEnvironmentAccess() {
	p.allowEnvAccess = true
	for _, name := range environmentTemplateFunctions {
		if fn, ok := p.unsafeSprigFuncMap[name]; ok {
			p.funcMap[name] = fn
		}
	}
}

// SetOption sets a template option (e.g., "missingkey=error")
func (p *TemplateParser) SetOption(key, value string) {
	p.options[key] = value
}

var environmentTemplateFunctions = []string{
	"env",       // Environment variable access
	"expandenv", // Environment variable expansion
}

var blockedTemplateFunctions = []string{
	"env",           // Environment variable access
	"expandenv",     // Environment variable expansion
	"getHostByName", // Network/DNS access
}

// getSafeSprigFunctions returns Sprig functions with dangerous functions removed
func getSafeSprigFunctions() template.FuncMap {
	// Get all text-safe Sprig functions (excludes HTML-specific functions)
	funcMap := sprig.TxtFuncMap()

	for _, fnName := range blockedTemplateFunctions {
		delete(funcMap, fnName)
	}

	// Add custom helper functions
	addCustomFunctions(funcMap)

	return funcMap
}

func isBlockedTemplateFunction(name string) bool {
	for _, fnName := range blockedTemplateFunctions {
		if name == fnName {
			return true
		}
	}
	return false
}

func isEnvironmentTemplateFunction(name string) bool {
	for _, fnName := range environmentTemplateFunctions {
		if name == fnName {
			return true
		}
	}
	return false
}

// addCustomFunctions adds shipyard-specific template functions
func addCustomFunctions(funcMap template.FuncMap) {
	// has: Check if a slice contains a value
	funcMap["has"] = func(slice []string, value string) bool {
		for _, item := range slice {
			if item == value {
				return true
			}
		}
		return false
	}

	// keys: Get map keys
	funcMap["keys"] = func(m map[string]interface{}) []string {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		return keys
	}

	// values: Get map values
	funcMap["values"] = func(m map[string]interface{}) []interface{} {
		values := make([]interface{}, 0, len(m))
		for _, v := range m {
			values = append(values, v)
		}
		return values
	}
}

// ParseWithFunctions parses a template with custom functions
func ParseWithFunctions(name, content string, funcMap template.FuncMap) (*template.Template, error) {
	parser := NewTemplateParser()

	// Add custom functions
	for fnName, fn := range funcMap {
		parser.AddFunction(fnName, fn)
	}

	return parser.Parse(name, content)
}

// MustParse parses a template and panics on error (useful for built-in templates)
func MustParse(name, content string) *template.Template {
	parser := NewTemplateParser()
	tmpl, err := parser.Parse(name, content)
	if err != nil {
		panic(fmt.Sprintf("failed to parse template %s: %v", name, err))
	}
	return tmpl
}
