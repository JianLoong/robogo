package runner

import (
	"time"

	"github.com/JianLoong/robogo/internal/parser"
)

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
	Type       string                 `json:"type"`
	Message    string                 `json:"message"`
	Action     string                 `json:"action"`
	AutoFix    bool                   `json:"auto_fix"`
	Confidence float64                `json:"confidence"`
	Context    map[string]interface{} `json:"context,omitempty"`
}

// ValidationLocation provides specific location information
type ValidationLocation struct {
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`
	Column   int    `json:"column,omitempty"`
	Step     int    `json:"step,omitempty"`
	Field    string `json:"field,omitempty"`
	Path     string `json:"path,omitempty"`
}

// ValidationStatistics provides comprehensive validation metrics
type ValidationStatistics struct {
	TotalRules         int `json:"total_rules"`
	AppliedRules       int `json:"applied_rules"`
	ErrorCount         int `json:"error_count"`
	WarningCount       int `json:"warning_count"`
	SuggestionCount    int `json:"suggestion_count"`
	CriticalErrors     int `json:"critical_errors"`
	SecurityIssues     int `json:"security_issues"`
	PerformanceIssues  int `json:"performance_issues"`
	BestPracticeIssues int `json:"best_practice_issues"`
}

// Validation enums and constants
type ValidationCategory string

const (
	CategorySyntax       ValidationCategory = "syntax"
	CategorySecurity     ValidationCategory = "security"
	CategoryPerformance  ValidationCategory = "performance"
	CategoryBestPractice ValidationCategory = "best_practice"
	CategoryDependency   ValidationCategory = "dependency"
	CategoryData         ValidationCategory = "data"
	CategoryAction       ValidationCategory = "action"
)

type ValidationSeverity string

const (
	SeverityInfo     ValidationSeverity = "info"
	SeverityWarning  ValidationSeverity = "warning"
	SeverityError    ValidationSeverity = "error"
	SeverityCritical ValidationSeverity = "critical"
)

type ValidationPhase string

const (
	PhasePreExecution  ValidationPhase = "pre_execution"
	PhaseDuringExecution ValidationPhase = "during_execution"
	PhasePostExecution ValidationPhase = "post_execution"
)

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