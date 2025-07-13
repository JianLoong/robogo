package runner

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/JianLoong/robogo/internal/parser"
)

// RuleBasedValidationEngine implements ValidationEngine with pluggable rules
type RuleBasedValidationEngine struct {
	mu              sync.RWMutex
	rules           map[string]ValidationRule
	fieldValidators map[string]FieldValidator
	config          ValidationConfig
}

// ValidationConfig configures the validation engine behavior
type ValidationConfig struct {
	EnableSecurityValidation    bool
	EnablePerformanceValidation bool
	EnableBestPracticeValidation bool
	StrictMode                  bool
	MaxErrors                   int
	FailOnWarnings              bool
}

// NewRuleBasedValidationEngine creates a new enhanced validation engine
func NewRuleBasedValidationEngine(config ValidationConfig) ValidationEngine {
	engine := &RuleBasedValidationEngine{
		rules:           make(map[string]ValidationRule),
		fieldValidators: make(map[string]FieldValidator),
		config:          config,
	}
	
	// Register default rules
	engine.registerDefaultRules()
	
	return engine
}

// NewDefaultValidationEngine creates a validation engine with default configuration
func NewDefaultValidationEngine() ValidationEngine {
	config := ValidationConfig{
		EnableSecurityValidation:     true,
		EnablePerformanceValidation:  true,
		EnableBestPracticeValidation: true,
		StrictMode:                   false,
		MaxErrors:                    100,
		FailOnWarnings:               false,
	}
	
	return NewRuleBasedValidationEngine(config)
}

func (ve *RuleBasedValidationEngine) registerDefaultRules() {
	// Core validation rules
	ve.RegisterRule(NewRequiredFieldRule())
	ve.RegisterRule(NewActionValidationRule(ve.getBuiltInActionMetadata()))
	
	// Security validation
	if ve.config.EnableSecurityValidation {
		ve.RegisterRule(NewSecurityValidationRule())
	}
	
	// Dependency validation
	ve.RegisterRule(NewDependencyValidationRule())
	
	// Performance validation
	if ve.config.EnablePerformanceValidation {
		ve.RegisterRule(NewPerformanceValidationRule())
	}
	
	// Best practice validation
	if ve.config.EnableBestPracticeValidation {
		ve.RegisterRule(NewBestPracticeValidationRule())
	}
}

// ValidateTestCase validates a test case using all applicable rules
func (ve *RuleBasedValidationEngine) ValidateTestCase(testCase *parser.TestCase) ValidationReport {
	startTime := time.Now()
	
	context := NewValidationContext().
		WithTestCase(testCase).
		WithPhase(PhasePreExecution).
		WithAvailableActions(ve.getBuiltInActions())
	
	report := ValidationReport{
		Valid:          true,
		Errors:         make([]ValidationError, 0),
		Warnings:       make([]ValidationError, 0),
		Suggestions:    make([]ValidationSuggestion, 0),
		Timestamp:      startTime,
		ValidationTime: 0,
	}
	
	ve.mu.RLock()
	rules := make([]ValidationRule, 0, len(ve.rules))
	for _, rule := range ve.rules {
		if rule.ShouldApply(context) {
			rules = append(rules, rule)
		}
	}
	ve.mu.RUnlock()
	
	// Apply rules
	for _, rule := range rules {
		errors := rule.Validate(context)
		
		for _, err := range errors {
			err.Rule = rule.Name()
			
			switch err.Severity {
			case SeverityError, SeverityCritical:
				report.Errors = append(report.Errors, err)
				report.Valid = false
			case SeverityWarning:
				report.Warnings = append(report.Warnings, err)
				if ve.config.FailOnWarnings {
					report.Valid = false
				}
			case SeverityInfo:
				// Convert info to suggestions
				suggestion := ValidationSuggestion{
					Type:       "improvement",
					Message:    err.Message,
					Action:     "consider",
					AutoFix:    false,
					Confidence: 0.8,
					Context:    err.Context,
				}
				report.Suggestions = append(report.Suggestions, suggestion)
			}
		}
		
		// Stop if max errors reached
		if ve.config.MaxErrors > 0 && len(report.Errors) >= ve.config.MaxErrors {
			break
		}
	}
	
	// Calculate statistics
	report.Statistics = ve.calculateStatistics(rules, report)
	report.ValidationTime = time.Since(startTime)
	
	return report
}

// ValidateTestSuite validates a test suite
func (ve *RuleBasedValidationEngine) ValidateTestSuite(testSuite *parser.TestSuite) ValidationReport {
	startTime := time.Now()
	
	context := NewValidationContext().
		WithTestSuite(testSuite).
		WithPhase(PhasePreExecution)
	
	report := ValidationReport{
		Valid:          true,
		Errors:         make([]ValidationError, 0),
		Warnings:       make([]ValidationError, 0),
		Suggestions:    make([]ValidationSuggestion, 0),
		Timestamp:      startTime,
		ValidationTime: 0,
	}
	
	ve.mu.RLock()
	rules := make([]ValidationRule, 0, len(ve.rules))
	for _, rule := range ve.rules {
		if rule.ShouldApply(context) {
			rules = append(rules, rule)
		}
	}
	ve.mu.RUnlock()
	
	// Apply rules
	for _, rule := range rules {
		errors := rule.Validate(context)
		
		for _, err := range errors {
			err.Rule = rule.Name()
			
			switch err.Severity {
			case SeverityError, SeverityCritical:
				report.Errors = append(report.Errors, err)
				report.Valid = false
			case SeverityWarning:
				report.Warnings = append(report.Warnings, err)
				if ve.config.FailOnWarnings {
					report.Valid = false
				}
			}
		}
		
		if ve.config.MaxErrors > 0 && len(report.Errors) >= ve.config.MaxErrors {
			break
		}
	}
	
	report.Statistics = ve.calculateStatistics(rules, report)
	report.ValidationTime = time.Since(startTime)
	
	return report
}

// ValidateStep validates a single step
func (ve *RuleBasedValidationEngine) ValidateStep(step parser.Step) ValidationReport {
	startTime := time.Now()
	
	context := NewValidationContext().
		WithCurrentStep(&step).
		WithPhase(PhasePreExecution)
	
	report := ValidationReport{
		Valid:          true,
		Errors:         make([]ValidationError, 0),
		Warnings:       make([]ValidationError, 0),
		Suggestions:    make([]ValidationSuggestion, 0),
		Timestamp:      startTime,
		ValidationTime: 0,
	}
	
	ve.mu.RLock()
	rules := make([]ValidationRule, 0, len(ve.rules))
	for _, rule := range ve.rules {
		if rule.ShouldApply(context) {
			rules = append(rules, rule)
		}
	}
	ve.mu.RUnlock()
	
	// Apply rules
	for _, rule := range rules {
		errors := rule.Validate(context)
		
		for _, err := range errors {
			err.Rule = rule.Name()
			
			switch err.Severity {
			case SeverityError, SeverityCritical:
				report.Errors = append(report.Errors, err)
				report.Valid = false
			case SeverityWarning:
				report.Warnings = append(report.Warnings, err)
			}
		}
	}
	
	report.Statistics = ve.calculateStatistics(rules, report)
	report.ValidationTime = time.Since(startTime)
	
	return report
}

// RegisterRule adds a validation rule
func (ve *RuleBasedValidationEngine) RegisterRule(rule ValidationRule) {
	ve.mu.Lock()
	defer ve.mu.Unlock()
	
	ve.rules[rule.Name()] = rule
}

// UnregisterRule removes a validation rule
func (ve *RuleBasedValidationEngine) UnregisterRule(ruleName string) {
	ve.mu.Lock()
	defer ve.mu.Unlock()
	
	delete(ve.rules, ruleName)
}

// ListRules returns names of all registered rules
func (ve *RuleBasedValidationEngine) ListRules() []string {
	ve.mu.RLock()
	defer ve.mu.RUnlock()
	
	names := make([]string, 0, len(ve.rules))
	for name := range ve.rules {
		names = append(names, name)
	}
	
	return names
}

// RegisterValidator adds a field validator
func (ve *RuleBasedValidationEngine) RegisterValidator(validator FieldValidator) {
	ve.mu.Lock()
	defer ve.mu.Unlock()
	
	ve.fieldValidators[validator.Name()] = validator
}

// GetValidator retrieves a field validator
func (ve *RuleBasedValidationEngine) GetValidator(fieldType string) FieldValidator {
	ve.mu.RLock()
	defer ve.mu.RUnlock()
	
	return ve.fieldValidators[fieldType]
}

func (ve *RuleBasedValidationEngine) getBuiltInActions() []string {
	return []string{
		"http", "assert", "log", "variable", "postgres", "spanner",
		"kafka", "rabbitmq", "template", "sleep", "get_time", "get_random",
		"concat", "length", "skip", "bytes_to_string", "json_extract",
		"if", "for", "while", "control",
	}
}

func (ve *RuleBasedValidationEngine) getBuiltInActionMetadata() map[string]ActionMetadata {
	metadata := make(map[string]ActionMetadata)
	
	actions := ve.getBuiltInActions()
	for _, action := range actions {
		metadata[action] = ActionMetadata{
			Name:        action,
			Description: fmt.Sprintf("%s action", action),
			Parameters:  []ParameterMetadata{}, // Basic metadata for now
		}
	}
	
	// Add specific metadata for key actions
	metadata["http"] = ActionMetadata{
		Name:        "http",
		Description: "HTTP request action",
		Parameters: []ParameterMetadata{
			{Name: "method", Type: "string", Required: true},
			{Name: "url", Type: "string", Required: true},
		},
	}
	
	metadata["assert"] = ActionMetadata{
		Name:        "assert",
		Description: "Assertion action",
		Parameters: []ParameterMetadata{
			{Name: "actual", Type: "any", Required: true},
			{Name: "operator", Type: "string", Required: true},
			{Name: "expected", Type: "any", Required: true},
		},
	}
	
	return metadata
}

func (ve *RuleBasedValidationEngine) calculateStatistics(rules []ValidationRule, report ValidationReport) ValidationStatistics {
	stats := ValidationStatistics{
		TotalRules:      len(ve.rules),
		AppliedRules:    len(rules),
		ErrorCount:      len(report.Errors),
		WarningCount:    len(report.Warnings),
		SuggestionCount: len(report.Suggestions),
	}
	
	// Count by severity and category
	for _, err := range report.Errors {
		switch err.Severity {
		case SeverityCritical:
			stats.CriticalErrors++
		}
		
		switch err.Category {
		case CategorySecurity:
			stats.SecurityIssues++
		case CategoryPerformance:
			stats.PerformanceIssues++
		}
	}
	
	return stats
}

// DefaultValidationContext implements ValidationContext
type DefaultValidationContext struct {
	testCase      *parser.TestCase
	testSuite     *parser.TestSuite
	currentStep   *parser.Step
	stepIndex     int
	phase         ValidationPhase
	contextData   map[string]interface{}
	variables     map[string]interface{}
	availableActions []string
}

// NewValidationContext creates a new validation context
func NewValidationContext() *DefaultValidationContext {
	return &DefaultValidationContext{
		contextData:      make(map[string]interface{}),
		variables:        make(map[string]interface{}),
		availableActions: []string{"http", "assert", "log", "variable", "postgres", "kafka", "rabbitmq"},
	}
}

func (ctx *DefaultValidationContext) WithTestCase(testCase *parser.TestCase) *DefaultValidationContext {
	ctx.testCase = testCase
	if testCase != nil && testCase.Variables.Regular != nil {
		for k, v := range testCase.Variables.Regular {
			ctx.variables[k] = v
		}
	}
	return ctx
}

func (ctx *DefaultValidationContext) WithTestSuite(testSuite *parser.TestSuite) *DefaultValidationContext {
	ctx.testSuite = testSuite
	return ctx
}

func (ctx *DefaultValidationContext) WithCurrentStep(step *parser.Step) *DefaultValidationContext {
	ctx.currentStep = step
	return ctx
}

func (ctx *DefaultValidationContext) WithPhase(phase ValidationPhase) *DefaultValidationContext {
	ctx.phase = phase
	return ctx
}

func (ctx *DefaultValidationContext) WithAvailableActions(actions []string) *DefaultValidationContext {
	ctx.availableActions = actions
	return ctx
}

func (ctx *DefaultValidationContext) GetTestCase() *parser.TestCase {
	return ctx.testCase
}

func (ctx *DefaultValidationContext) GetTestSuite() *parser.TestSuite {
	return ctx.testSuite
}

func (ctx *DefaultValidationContext) GetStep(index int) *parser.Step {
	if ctx.testCase == nil || index < 0 || index >= len(ctx.testCase.Steps) {
		return nil
	}
	return &ctx.testCase.Steps[index]
}

func (ctx *DefaultValidationContext) GetCurrentStep() *parser.Step {
	return ctx.currentStep
}

func (ctx *DefaultValidationContext) GetStepIndex() int {
	return ctx.stepIndex
}

func (ctx *DefaultValidationContext) GetVariable(name string) (interface{}, bool) {
	value, exists := ctx.variables[name]
	return value, exists
}

func (ctx *DefaultValidationContext) GetAvailableActions() []string {
	return ctx.availableActions
}

func (ctx *DefaultValidationContext) HasCircularDependency(steps []parser.Step) bool {
	// Build dependency graph
	dependencies := make(map[string][]string)
	variables := make(map[string]int) // variable -> step index that defines it
	
	for i, step := range steps {
		if step.Result != "" {
			variables[step.Result] = i
		}
		
		stepDeps := ctx.GetStepDependencies(step)
		if len(stepDeps) > 0 {
			dependencies[fmt.Sprintf("step_%d", i)] = stepDeps
		}
	}
	
	// Check for circular dependencies using DFS
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	
	var hasCycle func(string) bool
	hasCycle = func(node string) bool {
		visited[node] = true
		recStack[node] = true
		
		for _, dep := range dependencies[node] {
			depKey := fmt.Sprintf("var_%s", dep)
			if !visited[depKey] {
				if hasCycle(depKey) {
					return true
				}
			} else if recStack[depKey] {
				return true
			}
		}
		
		recStack[node] = false
		return false
	}
	
	for node := range dependencies {
		if !visited[node] {
			if hasCycle(node) {
				return true
			}
		}
	}
	
	return false
}

func (ctx *DefaultValidationContext) GetStepDependencies(step parser.Step) []string {
	var dependencies []string
	
	// Extract variable references from arguments
	for _, arg := range step.Args {
		if str, ok := arg.(string); ok {
			deps := extractVariableReferences(str)
			dependencies = append(dependencies, deps...)
		}
	}
	
	// Extract from conditional statements
	if step.If != nil {
		deps := extractVariableReferences(step.If.Condition)
		dependencies = append(dependencies, deps...)
	}
	
	if step.For != nil {
		deps := extractVariableReferences(step.For.Condition)
		dependencies = append(dependencies, deps...)
	}
	
	if step.While != nil {
		deps := extractVariableReferences(step.While.Condition)
		dependencies = append(dependencies, deps...)
	}
	
	return dependencies
}

func extractVariableReferences(text string) []string {
	var variables []string
	
	// Simple extraction of ${variable} patterns
	start := 0
	for {
		startIdx := strings.Index(text[start:], "${")
		if startIdx == -1 {
			break
		}
		startIdx += start
		
		endIdx := strings.Index(text[startIdx:], "}")
		if endIdx == -1 {
			break
		}
		endIdx += startIdx
		
		varName := text[startIdx+2 : endIdx]
		if varName != "" {
			variables = append(variables, varName)
		}
		
		start = endIdx + 1
	}
	
	return variables
}

func (ctx *DefaultValidationContext) GetFieldValue(path string) (interface{}, bool) {
	// TODO: Implement field path resolution
	return nil, false
}

func (ctx *DefaultValidationContext) IsFieldRequired(path string) bool {
	// TODO: Implement field requirement checking
	return false
}

func (ctx *DefaultValidationContext) GetValidationPhase() ValidationPhase {
	return ctx.phase
}

func (ctx *DefaultValidationContext) GetContextData() map[string]interface{} {
	return ctx.contextData
}

// Additional validation rules

// PerformanceValidationRule checks for performance-related issues
type PerformanceValidationRule struct{}

func NewPerformanceValidationRule() ValidationRule {
	return &PerformanceValidationRule{}
}

func (r *PerformanceValidationRule) Name() string {
	return "performance_validation"
}

func (r *PerformanceValidationRule) Description() string {
	return "Validates performance-related configurations and patterns"
}

func (r *PerformanceValidationRule) Category() ValidationCategory {
	return CategoryPerformance
}

func (r *PerformanceValidationRule) Severity() ValidationSeverity {
	return SeverityWarning
}

func (r *PerformanceValidationRule) ShouldApply(context ValidationContext) bool {
	return context.GetTestCase() != nil
}

func (r *PerformanceValidationRule) Validate(context ValidationContext) []ValidationError {
	var errors []ValidationError
	
	testCase := context.GetTestCase()
	if testCase == nil {
		return errors
	}
	
	// Check for excessive number of steps
	if len(testCase.Steps) > 50 {
		errors = append(errors, ValidationError{
			Type:     "excessive_steps",
			Category: CategoryPerformance,
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("Test case has %d steps, consider breaking into smaller test cases", len(testCase.Steps)),
			Field:    "steps",
			Value:    len(testCase.Steps),
			Rule:     r.Name(),
			Code:     "PERF_EXCESSIVE_STEPS",
			Suggestions: []string{
				"Break large test cases into smaller, focused test cases",
				"Use test suites to organize related test cases",
				"Consider if all steps are necessary for the test goal",
			},
		})
	}
	
	// Check for missing timeouts on HTTP actions
	for i, step := range testCase.Steps {
		if step.Action == "http" {
			hasTimeout := false
			if step.Options != nil {
				if _, exists := step.Options["timeout"]; exists {
					hasTimeout = true
				}
			}
			
			if !hasTimeout {
				errors = append(errors, ValidationError{
					Type:     "missing_timeout",
					Category: CategoryPerformance,
					Severity: SeverityWarning,
					Message:  "HTTP action without explicit timeout",
					Field:    fmt.Sprintf("steps[%d].options.timeout", i),
					Rule:     r.Name(),
					Code:     "PERF_MISSING_TIMEOUT",
					Suggestions: []string{
						"Add a timeout option to prevent hanging requests",
						"Example: options: {timeout: '30s'}",
					},
					Location: ValidationLocation{
						Step:  i + 1,
						Field: "options",
						Path:  fmt.Sprintf("steps[%d].options", i),
					},
				})
			}
		}
	}
	
	return errors
}

// BestPracticeValidationRule checks for best practice violations
type BestPracticeValidationRule struct{}

func NewBestPracticeValidationRule() ValidationRule {
	return &BestPracticeValidationRule{}
}

func (r *BestPracticeValidationRule) Name() string {
	return "best_practice_validation"
}

func (r *BestPracticeValidationRule) Description() string {
	return "Validates adherence to testing best practices"
}

func (r *BestPracticeValidationRule) Category() ValidationCategory {
	return CategoryBestPractice
}

func (r *BestPracticeValidationRule) Severity() ValidationSeverity {
	return SeverityInfo
}

func (r *BestPracticeValidationRule) ShouldApply(context ValidationContext) bool {
	return context.GetTestCase() != nil
}

func (r *BestPracticeValidationRule) Validate(context ValidationContext) []ValidationError {
	var errors []ValidationError
	
	testCase := context.GetTestCase()
	if testCase == nil {
		return errors
	}
	
	// Check for descriptive test case names
	if len(testCase.Name) < 10 {
		errors = append(errors, ValidationError{
			Type:     "short_test_name",
			Category: CategoryBestPractice,
			Severity: SeverityInfo,
			Message:  "Test case name is quite short, consider making it more descriptive",
			Field:    "name",
			Value:    testCase.Name,
			Rule:     r.Name(),
			Code:     "BP_SHORT_NAME",
			Suggestions: []string{
				"Use descriptive names that explain what the test validates",
				"Include the expected behavior in the test name",
				"Example: 'Should return 200 when user is authenticated'",
			},
		})
	}
	
	// Check for missing descriptions
	if testCase.Description == "" {
		errors = append(errors, ValidationError{
			Type:     "missing_description",
			Category: CategoryBestPractice,
			Severity: SeverityInfo,
			Message:  "Test case is missing a description",
			Field:    "description",
			Rule:     r.Name(),
			Code:     "BP_MISSING_DESC",
			Suggestions: []string{
				"Add a description explaining the purpose of this test",
				"Describe what scenario this test covers",
				"Include any important context or prerequisites",
			},
		})
	}
	
	// Check for steps without assertions
	hasAssertion := false
	for _, step := range testCase.Steps {
		if step.Action == "assert" {
			hasAssertion = true
			break
		}
	}
	
	if !hasAssertion {
		errors = append(errors, ValidationError{
			Type:     "no_assertions",
			Category: CategoryBestPractice,
			Severity: SeverityWarning,
			Message:  "Test case has no assertions - consider adding validation steps",
			Field:    "steps",
			Rule:     r.Name(),
			Code:     "BP_NO_ASSERTIONS",
			Suggestions: []string{
				"Add assert actions to validate expected outcomes",
				"Verify response status codes, data, or behavior",
				"Tests without assertions may not catch regressions",
			},
		})
	}
	
	return errors
}