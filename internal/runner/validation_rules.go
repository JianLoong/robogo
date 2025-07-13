package runner

import (
	"fmt"
)

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