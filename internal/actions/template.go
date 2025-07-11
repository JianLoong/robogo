package actions

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"text/template"
	
	"github.com/JianLoong/robogo/internal/util"
)

// TemplateAction renders templates using ActionContext manager or from files
//
// Parameters:
//   - template_path: Path to the template file OR template name from context
//   - data: Data object to pass to the template
//   - options: Additional options (optional)
//   - silent: Whether to suppress output
//
// Returns: Rendered template as string
//
// Example usage:
//   - ["templates/mt103.tmpl", {"transaction_id": "123", "amount": "100.00"}]
//   - ["inline_template_name", {"field": "value"}] (uses template from context)
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
//   - Templates can be loaded from files or defined in test case templates section
//   - Uses Go's text/template package
//   - Supports all Go template functions
//   - Case-sensitive field names
func TemplateAction(ctx context.Context, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	if len(args) < 2 {
		return "", util.NewArgumentCountError("template", 2, len(args))
	}

	templatePathOrName, ok := args[0].(string)
	if !ok {
		return "", util.NewArgumentTypeError("template", 0, "string", args[0])
	}

	// Get data object and convert to map
	data := args[1]
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return "", util.NewArgumentTypeError("template", 1, "map", args[1])
	}

	// Get template content from manager or file
	templateContent, err := getTemplateContent(ctx, templatePathOrName)
	if err != nil {
		return "", err
	}

	// Parse the template
	tmpl, err := template.New(templatePathOrName).Parse(templateContent)
	if err != nil {
		return "", util.NewErrorBuilder(util.ErrorTypeTemplate, "failed to parse template").
			WithAction("template").
			WithCause(err).
			WithArguments(args).
			WithOptions(options).
			WithDetails(map[string]interface{}{
				"template_source": templatePathOrName,
			}).
			Build()
	}

	// Execute the template
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, dataMap)
	if err != nil {
		return "", util.NewErrorBuilder(util.ErrorTypeTemplate, "failed to execute template").
			WithAction("template").
			WithCause(err).
			WithArguments(args).
			WithOptions(options).
			WithDetails(map[string]interface{}{
				"template_source": templatePathOrName,
			}).
			Build()
	}

	result := buf.String()

	if !silent {
		fmt.Printf("Rendered template from '%s' (%d characters)\n", templatePathOrName, len(result))
	}

	return result, nil
}

// getTemplateContent retrieves template content from ActionContext manager or file
func getTemplateContent(ctx context.Context, templatePathOrName string) (string, error) {
	// First try to get from ActionContext template manager
	actionCtx := GetActionContext(ctx)
	if actionCtx.TemplateManager != nil {
		if templateContent, exists := actionCtx.TemplateManager.GetTemplate(templatePathOrName); exists {
			return templateContent, nil
		}
	}

	// Fallback to reading from file
	templateContent, err := os.ReadFile(templatePathOrName)
	if err != nil {
		return "", util.NewErrorBuilder(util.ErrorTypeFileSystem, "template not found in context and failed to read as file").
			WithAction("template").
			WithCause(err).
			WithDetails(map[string]interface{}{
				"template_source": templatePathOrName,
				"available_templates": func() []string {
					if actionCtx.TemplateManager != nil {
						return actionCtx.TemplateManager.ListTemplates()
					}
					return []string{}
				}(),
			}).
			Build()
	}

	return string(templateContent), nil
}