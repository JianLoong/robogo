package runner

import (
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
	EnableSecurityValidation     bool
	EnablePerformanceValidation  bool
	EnableBestPracticeValidation bool
	StrictMode                   bool
	MaxErrors                    int
	FailOnWarnings               bool
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

// getBuiltInActionMetadata returns metadata for built-in actions
func (ve *RuleBasedValidationEngine) getBuiltInActionMetadata() map[string]ActionMetadata {
	return map[string]ActionMetadata{
		"http": {
			Name:        "http",
			Description: "Performs HTTP requests",
			Parameters: []ParameterMetadata{
				{Name: "method", Type: "string", Required: true, Description: "HTTP method"},
				{Name: "url", Type: "string", Required: true, Description: "Request URL"},
			},
		},
		"assert": {
			Name:        "assert",
			Description: "Performs assertions and validations",
			Parameters: []ParameterMetadata{
				{Name: "actual", Type: "interface{}", Required: true, Description: "Actual value"},
				{Name: "operator", Type: "string", Required: true, Description: "Comparison operator"},
				{Name: "expected", Type: "interface{}", Required: true, Description: "Expected value"},
			},
		},
		"log": {
			Name:        "log",
			Description: "Logs messages",
			Parameters: []ParameterMetadata{
				{Name: "message", Type: "string", Required: true, Description: "Message to log"},
			},
		},
		"variable": {
			Name:        "variable",
			Description: "Sets variables",
			Parameters: []ParameterMetadata{
				{Name: "name", Type: "string", Required: true, Description: "Variable name"},
				{Name: "value", Type: "interface{}", Required: true, Description: "Variable value"},
			},
		},
		"postgres": {
			Name:        "postgres",
			Description: "Executes PostgreSQL operations",
			Parameters: []ParameterMetadata{
				{Name: "query", Type: "string", Required: true, Description: "SQL query"},
			},
		},
		"kafka": {
			Name:        "kafka",
			Description: "Produces/consumes Kafka messages",
			Parameters: []ParameterMetadata{
				{Name: "topic", Type: "string", Required: true, Description: "Kafka topic"},
			},
		},
		"rabbitmq": {
			Name:        "rabbitmq",
			Description: "Produces/consumes RabbitMQ messages",
			Parameters: []ParameterMetadata{
				{Name: "queue", Type: "string", Required: true, Description: "RabbitMQ queue"},
			},
		},
	}
}

// getBuiltInActions returns a list of built-in action names
func (ve *RuleBasedValidationEngine) getBuiltInActions() []string {
	return []string{"http", "assert", "log", "variable", "postgres", "kafka", "rabbitmq", "spanner", "template", "if", "for", "while", "get_time", "get_random", "concat", "length"}
}