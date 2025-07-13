package runner

import (
	"fmt"
	"strings"
	"time"

	"github.com/JianLoong/robogo/internal/parser"
)

// Enhanced validation framework with rule-based architecture

// ValidationEngine interface for comprehensive validation
type ValidationEngine interface {
	ValidateTestCase(testCase *parser.TestCase) ValidationReport
	ValidateTestSuite(testSuite *parser.TestSuite) ValidationReport
	ValidateStep(step parser.Step) ValidationReport
	
	// Rule management
	RegisterRule(rule ValidationRule)
	UnregisterRule(ruleName string)
	ListRules() []string
	
	// Validator management
	RegisterValidator(validator FieldValidator)
	GetValidator(fieldType string) FieldValidator
}

// ValidationRule interface for pluggable validation rules
type ValidationRule interface {
	Name() string
	Description() string
	Category() ValidationCategory
	Severity() ValidationSeverity
	Validate(context ValidationContext) []ValidationError
	ShouldApply(context ValidationContext) bool
}

// ValidationContext provides context information for validation rules
type ValidationContext interface {
	// Test case context
	GetTestCase() *parser.TestCase
	GetTestSuite() *parser.TestSuite
	
	// Step context
	GetStep(index int) *parser.Step
	GetCurrentStep() *parser.Step
	GetStepIndex() int
	
	// Variable and dependency analysis
	GetVariable(name string) (interface{}, bool)
	GetAvailableActions() []string
	HasCircularDependency(steps []parser.Step) bool
	GetStepDependencies(step parser.Step) []string
	
	// Cross-field validation support
	GetFieldValue(path string) (interface{}, bool)
	IsFieldRequired(path string) bool
	
	// Metadata
	GetValidationPhase() ValidationPhase
	GetContextData() map[string]interface{}
}

// FieldValidator interface for type-specific field validation
type FieldValidator interface {
	Name() string
	SupportedTypes() []string
	Validate(value interface{}, constraints map[string]interface{}) []ValidationError
}

// ValidationReport contains comprehensive validation results
type ValidationReport struct {
	Valid          bool                      `json:"valid"`
	Errors         []ValidationError         `json:"errors"`
	Warnings       []ValidationError         `json:"warnings"`
	Suggestions    []ValidationSuggestion    `json:"suggestions"`
	Statistics     ValidationStatistics      `json:"statistics"`
	Timestamp      time.Time                `json:"timestamp"`
	ValidationTime time.Duration            `json:"validation_time"`
}

// Enhanced ValidationError with more context
type ValidationError struct {
	Type        string                 `json:"type"`
	Category    ValidationCategory     `json:"category"`
	Severity    ValidationSeverity     `json:"severity"`
	Message     string                 `json:"message"`
	Field       string                 `json:"field"`
	Value       interface{}            `json:"value,omitempty"`
	Expected    interface{}            `json:"expected,omitempty"`
	Actual      interface{}            `json:"actual,omitempty"`
	Rule        string                 `json:"rule"`
	Code        string                 `json:"code"`
	Suggestions []string               `json:"suggestions"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Location    ValidationLocation     `json:"location"`
}

// ValidationSuggestion provides actionable suggestions for improvement
type ValidationSuggestion struct {
	Type        string                 `json:"type"`
	Message     string                 `json:"message"`
	Action      string                 `json:"action"`
	AutoFix     bool                   `json:"auto_fix"`
	Confidence  float64               `json:"confidence"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// ValidationStatistics provides metrics about validation results
type ValidationStatistics struct {
	TotalRules       int `json:"total_rules"`
	AppliedRules     int `json:"applied_rules"`
	ErrorCount       int `json:"error_count"`
	WarningCount     int `json:"warning_count"`
	SuggestionCount  int `json:"suggestion_count"`
	CriticalErrors   int `json:"critical_errors"`
	SecurityIssues   int `json:"security_issues"`
	PerformanceIssues int `json:"performance_issues"`
}

// ValidationLocation pinpoints where validation issues occur
type ValidationLocation struct {
	File       string `json:"file,omitempty"`
	Line       int    `json:"line,omitempty"`
	Column     int    `json:"column,omitempty"`
	Step       int    `json:"step,omitempty"`
	Field      string `json:"field"`
	Path       string `json:"path"`
}

// Enums for validation categorization

type ValidationCategory string

const (
	CategorySyntax       ValidationCategory = "syntax"
	CategorySemantic     ValidationCategory = "semantic"
	CategorySecurity     ValidationCategory = "security"
	CategoryPerformance  ValidationCategory = "performance"
	CategoryBestPractice ValidationCategory = "best_practice"
	CategoryDependency   ValidationCategory = "dependency"
)

type ValidationSeverity int

const (
	SeverityInfo ValidationSeverity = iota
	SeverityWarning
	SeverityError
	SeverityCritical
)

func (s ValidationSeverity) String() string {
	switch s {
	case SeverityInfo:
		return "info"
	case SeverityWarning:
		return "warning"
	case SeverityError:
		return "error"
	case SeverityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

type ValidationPhase string

const (
	PhasePreExecution  ValidationPhase = "pre_execution"
	PhasePostExecution ValidationPhase = "post_execution"
	PhaseCrossField    ValidationPhase = "cross_field"
	PhaseDependency    ValidationPhase = "dependency"
)

// Default validation rule implementations

// RequiredFieldRule validates that required fields are present
type RequiredFieldRule struct {
	requiredFields map[string][]string // entity type -> required fields
}

func NewRequiredFieldRule() ValidationRule {
	return &RequiredFieldRule{
		requiredFields: map[string][]string{
			"testcase": {"name", "steps"},
			"testsuite": {"name", "testcases"},
			"step": {"name", "action"},
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
	
	for _, requiredField := range requiredFields {
		value, exists := fields[requiredField]
		if !exists || r.isEmpty(value) {
			errors = append(errors, ValidationError{
				Type:     "required_field",
				Category: CategorySyntax,
				Severity: SeverityError,
				Message:  fmt.Sprintf("Required field '%s' is missing or empty", requiredField),
				Field:    requiredField,
				Value:    value,
				Rule:     r.Name(),
				Code:     fmt.Sprintf("REQ_%s_%s", strings.ToUpper(entityType), strings.ToUpper(requiredField)),
				Suggestions: []string{
					fmt.Sprintf("Add a value for the '%s' field", requiredField),
					"Check the documentation for required field formats",
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
		return strings.TrimSpace(v) == ""
	case []interface{}:
		return len(v) == 0
	case []parser.Step:
		return len(v) == 0
	case []parser.TestCaseReference:
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
	return CategorySemantic
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
		
		metadata, exists := r.availableActions[step.Action]
		if !exists {
			// Suggest similar actions
			suggestions := r.getSimilarActions(step.Action)
			
			errors = append(errors, ValidationError{
				Type:     "unknown_action",
				Category: CategorySemantic,
				Severity: SeverityError,
				Message:  fmt.Sprintf("Unknown action '%s'", step.Action),
				Field:    fmt.Sprintf("steps[%d].action", i),
				Value:    step.Action,
				Rule:     r.Name(),
				Code:     "ACT_UNKNOWN",
				Suggestions: append([]string{
					"Check if the action name is spelled correctly",
					"Use 'robogo list' to see available actions",
				}, suggestions...),
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

func (r *ActionValidationRule) getSimilarActions(action string) []string {
	var suggestions []string
	action = strings.ToLower(action)
	
	for availableAction := range r.availableActions {
		available := strings.ToLower(availableAction)
		
		// Simple similarity check
		if strings.Contains(available, action) || strings.Contains(action, available) {
			suggestions = append(suggestions, fmt.Sprintf("Did you mean '%s'?", availableAction))
		}
	}
	
	// Limit suggestions
	if len(suggestions) > 3 {
		suggestions = suggestions[:3]
	}
	
	return suggestions
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

// ActionMetadata represents metadata about an action (simplified version)
type ActionMetadata struct {
	Name        string
	Description string
	Parameters  []ParameterMetadata
}

type ParameterMetadata struct {
	Name        string
	Type        string
	Required    bool
	Description string
}

// DependencyValidationRule validates step dependencies and prevents circular references
type DependencyValidationRule struct{}

func NewDependencyValidationRule() ValidationRule {
	return &DependencyValidationRule{}
}

func (r *DependencyValidationRule) Name() string {
	return "dependency_validation"
}

func (r *DependencyValidationRule) Description() string {
	return "Validates step dependencies and prevents circular references"
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
			Rule:     r.Name(),
			Code:     "DEP_CIRCULAR",
			Suggestions: []string{
				"Review variable assignments and references",
				"Ensure variables are defined before they are used",
				"Break circular references by using intermediate variables",
			},
		})
	}
	
	// Check for undefined variable references
	definedVars := make(map[string]bool)
	
	// Add variables from test case
	if testCase.Variables.Regular != nil {
		for varName := range testCase.Variables.Regular {
			definedVars[varName] = true
		}
	}
	
	if testCase.Variables.Secrets != nil {
		for varName := range testCase.Variables.Secrets {
			definedVars[varName] = true
		}
	}
	
	// Check each step
	for i, step := range testCase.Steps {
		dependencies := context.GetStepDependencies(step)
		
		for _, dep := range dependencies {
			// Check if it's a simple variable or dot notation
			baseVar := dep
			if dotIndex := strings.Index(dep, "."); dotIndex != -1 {
				baseVar = dep[:dotIndex]
			}
			
			if !definedVars[baseVar] {
				errors = append(errors, ValidationError{
					Type:     "undefined_variable",
					Category: CategoryDependency,
					Severity: SeverityError,
					Message:  fmt.Sprintf("Variable '%s' is used but not defined", baseVar),
					Field:    fmt.Sprintf("steps[%d]", i),
					Value:    dep,
					Rule:     r.Name(),
					Code:     "DEP_UNDEFINED_VAR",
					Suggestions: []string{
						fmt.Sprintf("Define variable '%s' in the variables section or as a step result", baseVar),
						"Check if the variable name is spelled correctly",
						"Ensure the variable is defined before this step",
					},
					Location: ValidationLocation{
						Step:  i + 1,
						Field: "args",
						Path:  fmt.Sprintf("steps[%d].args", i),
					},
				})
			}
		}
		
		// Add variables defined by this step
		if step.Result != "" {
			definedVars[step.Result] = true
		}
	}
	
	return errors
}

// SecurityValidationRule checks for security issues like exposed secrets
type SecurityValidationRule struct{}

func NewSecurityValidationRule() ValidationRule {
	return &SecurityValidationRule{}
}

func (r *SecurityValidationRule) Name() string {
	return "security_validation"
}

func (r *SecurityValidationRule) Description() string {
	return "Validates security aspects like secret exposure and credential handling"
}

func (r *SecurityValidationRule) Category() ValidationCategory {
	return CategorySecurity
}

func (r *SecurityValidationRule) Severity() ValidationSeverity {
	return SeverityCritical
}

func (r *SecurityValidationRule) ShouldApply(context ValidationContext) bool {
	return true
}

func (r *SecurityValidationRule) Validate(context ValidationContext) []ValidationError {
	var errors []ValidationError
	
	testCase := context.GetTestCase()
	if testCase == nil {
		return errors
	}
	
	// Check for hardcoded secrets in variable values
	if testCase.Variables.Regular != nil {
		for varName, value := range testCase.Variables.Regular {
			if r.looksLikeSecret(varName, value) {
				errors = append(errors, ValidationError{
					Type:     "potential_secret_exposure",
					Category: CategorySecurity,
					Severity: SeverityCritical,
					Message:  fmt.Sprintf("Variable '%s' may contain a hardcoded secret", varName),
					Field:    fmt.Sprintf("variables.vars.%s", varName),
					Value:    "***REDACTED***",
					Rule:     r.Name(),
					Code:     "SEC_HARDCODED_SECRET",
					Suggestions: []string{
						"Move sensitive values to the secrets section",
						"Use file-based secrets instead of inline values",
						"Enable output masking for this secret",
					},
				})
			}
		}
	}
	
	// Check for secrets without output masking
	if testCase.Variables.Secrets != nil {
		for secretName, secret := range testCase.Variables.Secrets {
			if !secret.MaskOutput {
				errors = append(errors, ValidationError{
					Type:     "unmasked_secret",
					Category: CategorySecurity,
					Severity: SeverityWarning,
					Message:  fmt.Sprintf("Secret '%s' does not have output masking enabled", secretName),
					Field:    fmt.Sprintf("variables.secrets.%s.mask_output", secretName),
					Value:    false,
					Expected: true,
					Rule:     r.Name(),
					Code:     "SEC_UNMASKED_SECRET",
					Suggestions: []string{
						"Set 'mask_output: true' for this secret",
						"Consider if this secret should be masked in logs",
					},
				})
			}
		}
	}
	
	return errors
}

func (r *SecurityValidationRule) looksLikeSecret(name string, value interface{}) bool {
	nameStr := strings.ToLower(name)
	secretIndicators := []string{"password", "token", "key", "secret", "credential", "auth"}
	
	for _, indicator := range secretIndicators {
		if strings.Contains(nameStr, indicator) {
			return true
		}
	}
	
	// Check value patterns (e.g., looks like a token)
	if str, ok := value.(string); ok {
		if len(str) > 20 && (strings.Contains(str, "Bearer") || 
			strings.HasPrefix(str, "sk-") || 
			strings.HasPrefix(str, "pk-")) {
			return true
		}
	}
	
	return false
}