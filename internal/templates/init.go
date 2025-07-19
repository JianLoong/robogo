package templates

import "github.com/JianLoong/robogo/internal/constants"

// InitializeErrorTemplates returns all error templates from constants package
func InitializeErrorTemplates() map[string]string {
	return constants.ErrorTemplates
}

// GetTemplateConstant returns the actual template string for a given template constant name
func GetTemplateConstant(templateName string) string {
	// Look up the template directly from the ErrorTemplates map
	templates := constants.ErrorTemplates
	if template, exists := templates[templateName]; exists {
		return template
	}
	
	// Return empty string if template not found
	return ""
}