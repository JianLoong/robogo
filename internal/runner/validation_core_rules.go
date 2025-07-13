package runner

import (
	"fmt"

	"github.com/JianLoong/robogo/internal/parser"
)

// RequiredFieldRule validates that required fields are present
type RequiredFieldRule struct {
	requiredFields map[string][]string // entity type -> required fields
}

func NewRequiredFieldRule() ValidationRule {
	return &RequiredFieldRule{
		requiredFields: map[string][]string{
			"testcase":  {"name", "steps"},
			"testsuite": {"name", "testcases"},
			"step":      {"name", "action"},
		},
	}
}

func (r *RequiredFieldRule) Name() string {
	return "required_fields"
}

func (r *RequiredFieldRule) Description() string {
	return "Validates that all required fields are present and non-empty"
}

func (r *RequiredFieldRule) Category() ValidationCategory {
	return CategorySyntax
}

func (r *RequiredFieldRule) Severity() ValidationSeverity {
	return SeverityError
}

func (r *RequiredFieldRule) ShouldApply(context ValidationContext) bool {
	return true
}

func (r *RequiredFieldRule) Validate(context ValidationContext) []ValidationError {
	var errors []ValidationError
	
	if testCase := context.GetTestCase(); testCase != nil {
		errors = append(errors, r.validateRequiredFields("testcase", map[string]interface{}{
			"name":  testCase.Name,
			"steps": testCase.Steps,
		})...)
		
		// Validate each step
		for i, step := range testCase.Steps {
			stepErrors := r.validateRequiredFields("step", map[string]interface{}{
				"name":   step.Name,
				"action": step.Action,
			})
			
			// Add step context to errors
			for _, err := range stepErrors {
				err.Location.Step = i + 1
				err.Field = fmt.Sprintf("steps[%d].%s", i, err.Field)
				errors = append(errors, err)
			}
		}
	}
	
	if testSuite := context.GetTestSuite(); testSuite != nil {
		errors = append(errors, r.validateRequiredFields("testsuite", map[string]interface{}{
			"name":      testSuite.Name,
			"testcases": testSuite.TestCases,
		})...)
	}
	
	return errors
}

func (r *RequiredFieldRule) validateRequiredFields(entityType string, fields map[string]interface{}) []ValidationError {
	var errors []ValidationError
	
	requiredFields, exists := r.requiredFields[entityType]
	if !exists {
		return errors
	}
	
	for _, fieldName := range requiredFields {
		value, exists := fields[fieldName]
		if !exists || r.isEmpty(value) {
			errors = append(errors, ValidationError{
				Type:     "missing_required_field",
				Category: CategorySyntax,
				Severity: SeverityError,
				Message:  fmt.Sprintf("Required field '%s' is missing or empty", fieldName),
				Field:    fieldName,
				Rule:     r.Name(),
				Code:     "REQ_FIELD_MISSING",
				Suggestions: []string{
					fmt.Sprintf("Add the required field '%s'", fieldName),
					fmt.Sprintf("Ensure '%s' has a non-empty value", fieldName),
				},
			})
		}
	}
	
	return errors
}

func (r *RequiredFieldRule) isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}
	
	switch v := value.(type) {
	case string:
		return v == ""
	case []interface{}:
		return len(v) == 0
	default:
		return false
	}
}

// ActionValidationRule validates that actions exist and have correct parameters
type ActionValidationRule struct {
	availableActions map[string]ActionMetadata
}

func NewActionValidationRule(actions map[string]ActionMetadata) ValidationRule {
	return &ActionValidationRule{
		availableActions: actions,
	}
}

func (r *ActionValidationRule) Name() string {
	return "action_validation"
}

func (r *ActionValidationRule) Description() string {
	return "Validates that actions exist and have correct parameters"
}

func (r *ActionValidationRule) Category() ValidationCategory {
	return CategoryAction
}

func (r *ActionValidationRule) Severity() ValidationSeverity {
	return SeverityError
}

func (r *ActionValidationRule) ShouldApply(context ValidationContext) bool {
	return context.GetTestCase() != nil
}

func (r *ActionValidationRule) Validate(context ValidationContext) []ValidationError {
	var errors []ValidationError
	
	testCase := context.GetTestCase()
	if testCase == nil {
		return errors
	}
	
	for i, step := range testCase.Steps {
		if step.Action == "" {
			continue // Handled by RequiredFieldRule
		}
		
		// Check if action exists
		metadata, exists := r.availableActions[step.Action]
		if !exists {
			errors = append(errors, ValidationError{
				Type:     "unknown_action",
				Category: CategoryAction,
				Severity: SeverityError,
				Message:  fmt.Sprintf("Unknown action '%s'", step.Action),
				Field:    fmt.Sprintf("steps[%d].action", i),
				Value:    step.Action,
				Rule:     r.Name(),
				Code:     "ACTION_UNKNOWN",
				Suggestions: []string{
					"Check the action name for typos",
					"Ensure the action is available in this environment",
					"Refer to documentation for available actions",
				},
				Location: ValidationLocation{
					Step:  i + 1,
					Field: "action",
					Path:  fmt.Sprintf("steps[%d].action", i),
				},
			})
			continue
		}
		
		// Validate action parameters
		paramErrors := r.validateActionParameters(step, metadata, i)
		errors = append(errors, paramErrors...)
	}
	
	return errors
}

func (r *ActionValidationRule) validateActionParameters(step parser.Step, metadata ActionMetadata, stepIndex int) []ValidationError {
	var errors []ValidationError
	
	// TODO: Implement parameter validation based on action metadata
	// This would validate:
	// - Required parameters are present
	// - Parameter types are correct
	// - Parameter values are within valid ranges
	// - Mutually exclusive parameters aren't both specified
	
	return errors
}

// DependencyValidationRule checks for circular dependencies and missing variables
type DependencyValidationRule struct{}

func NewDependencyValidationRule() ValidationRule {
	return &DependencyValidationRule{}
}

func (r *DependencyValidationRule) Name() string {
	return "dependency_validation"
}

func (r *DependencyValidationRule) Description() string {
	return "Validates variable dependencies and prevents circular references"
}

func (r *DependencyValidationRule) Category() ValidationCategory {
	return CategoryDependency
}

func (r *DependencyValidationRule) Severity() ValidationSeverity {
	return SeverityError
}

func (r *DependencyValidationRule) ShouldApply(context ValidationContext) bool {
	return context.GetTestCase() != nil
}

func (r *DependencyValidationRule) Validate(context ValidationContext) []ValidationError {
	var errors []ValidationError
	
	testCase := context.GetTestCase()
	if testCase == nil {
		return errors
	}
	
	// Check for circular dependencies
	if context.HasCircularDependency(testCase.Steps) {
		errors = append(errors, ValidationError{
			Type:     "circular_dependency",
			Category: CategoryDependency,
			Severity: SeverityError,
			Message:  "Circular dependency detected in step variables",
			Field:    "steps",
			Rule:     r.Name(),
			Code:     "DEP_CIRCULAR",
			Suggestions: []string{
				"Review variable dependencies between steps",
				"Ensure step results don't create circular references",
				"Consider breaking complex dependencies into simpler chains",
			},
		})
	}
	
	return errors
}

// SecurityValidationRule checks for potential security issues
type SecurityValidationRule struct{}

func NewSecurityValidationRule() ValidationRule {
	return &SecurityValidationRule{}
}

func (r *SecurityValidationRule) Name() string {
	return "security_validation"
}

func (r *SecurityValidationRule) Description() string {
	return "Validates test cases for potential security issues"
}

func (r *SecurityValidationRule) Category() ValidationCategory {
	return CategorySecurity
}

func (r *SecurityValidationRule) Severity() ValidationSeverity {
	return SeverityWarning
}

func (r *SecurityValidationRule) ShouldApply(context ValidationContext) bool {
	return true
}

func (r *SecurityValidationRule) Validate(context ValidationContext) []ValidationError {
	var errors []ValidationError
	
	// TODO: Implement security validation rules such as:
	// - Detection of hardcoded credentials
	// - Unsafe URL patterns
	// - Insecure data handling
	// - Exposure of sensitive information in logs
	
	return errors
}