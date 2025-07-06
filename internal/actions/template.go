package actions

import (
	"bytes"
	"fmt"
	"os"
	"text/template"
)

// TemplateAction renders templates from a file using Go's template engine
//
// Parameters:
//   - template_path: Path to the template file
//   - data: Data object to pass to the template
//   - options: Additional options (optional)
//   - silent: Whether to suppress output
//
// Returns: Rendered template as string
//
// Example usage:
//   - ["templates/mt103.tmpl", {"transaction_id": "123", "amount": "100.00"}]
//
// Template Syntax:
//   - Variables: {{.FieldName}}
//   - Nested fields: {{.Sender.BIC}}
//   - Conditionals: {{if .Field}}{{.Field}}{{end}}
//   - Loops: {{range .Items}}{{.Item}}{{end}}
//
// Use Cases:
//   - Payment message generation
//   - API request formatting
//   - Data transformation
//   - Report generation
//
// Notes:
//   - Templates must be defined in the test case's templates section
//   - Uses Go's text/template package
//   - Supports all Go template functions
//   - Case-sensitive field names
func TemplateAction(args []interface{}, options map[string]interface{}, silent bool) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("template action requires at least 2 arguments: template file path and data")
	}

	templatePath, ok := args[0].(string)
	if !ok {
		return "", fmt.Errorf("template file path must be a string")
	}

	// Read template file
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template file '%s': %w", templatePath, err)
	}

	// Get data object and convert to map
	data := args[1]
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("template data must be a map")
	}

	// Parse the template
	tmpl, err := template.New(templatePath).Parse(string(templateContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template '%s': %w", templatePath, err)
	}

	// Execute the template
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, dataMap)
	if err != nil {
		return "", fmt.Errorf("failed to execute template '%s': %w", templatePath, err)
	}

	result := buf.String()

	if !silent {
		fmt.Printf("📄 Rendered template from file '%s' (%d characters)\n", templatePath, len(result))
	}

	return result, nil
}

// SetTemplateContext sets the template context for the current test case
// This will be called by the test runner to provide access to templates
var templateContext map[string]string

// SetTemplateContext sets the available templates for the current test case
func SetTemplateContext(templates map[string]string) {
	templateContext = templates
}

// GetTemplateContext returns the current template context
func GetTemplateContext() map[string]string {
	return templateContext
}

// getTemplateFromContext retrieves a template from the current test context
func getTemplateFromContext(templateName string) (string, error) {
	if templateContext == nil {
		return "", fmt.Errorf("no template context available")
	}

	template, exists := templateContext[templateName]
	if !exists {
		return "", fmt.Errorf("template '%s' not found in context", templateName)
	}

	return template, nil
}
