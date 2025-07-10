package runner

import (
	"fmt"
	"strings"

	"github.com/JianLoong/robogo/internal/parser"
)

// DefaultValidationEngine implements ValidationEngine interface
// Provides comprehensive validation for test cases, suites, and steps
type DefaultValidationEngine struct {
	// Configuration could be added here
}

// NewValidationEngine creates a new validation engine
func NewValidationEngine() ValidationEngine {
	return &DefaultValidationEngine{}
}

// ValidateTestCase validates a test case and returns validation errors
func (ve *DefaultValidationEngine) ValidateTestCase(testCase *parser.TestCase) []ValidationError {
	var errors []ValidationError

	// Validate test case name
	if testCase.Name == "" {
		errors = append(errors, ValidationError{
			Type:    "required_field",
			Message: "Test case name is required",
			Field:   "name",
			Value:   testCase.Name,
			Suggestions: []string{
				"Add a descriptive name for your test case",
				"Example: 'Test user login functionality'",
			},
		})
	}

	// Validate steps
	if len(testCase.Steps) == 0 {
		errors = append(errors, ValidationError{
			Type:    "required_field",
			Message: "Test case must have at least one step",
			Field:   "steps",
			Value:   len(testCase.Steps),
			Suggestions: []string{
				"Add at least one step to your test case",
				"Example: Add a step with action 'log' to start",
			},
		})
	}

	// Validate individual steps
	for i, step := range testCase.Steps {
		stepErrors := ve.ValidateStep(step)
		for _, err := range stepErrors {
			err.Field = fmt.Sprintf("steps[%d].%s", i, err.Field)
			errors = append(errors, err)
		}
	}

	// Validate variables
	if testCase.Variables.Regular != nil {
		for key := range testCase.Variables.Regular {
			if key == "" {
				errors = append(errors, ValidationError{
					Type:    "invalid_value",
					Message: "Variable key cannot be empty",
					Field:   "variables.vars",
					Value:   key,
					Suggestions: []string{
						"Use descriptive variable names",
						"Avoid empty or whitespace-only keys",
					},
				})
			}
		}
	}

	return errors
}

// ValidateTestSuite validates a test suite and returns validation errors
func (ve *DefaultValidationEngine) ValidateTestSuite(testSuite *parser.TestSuite) []ValidationError {
	var errors []ValidationError

	// Validate suite name
	if testSuite.Name == "" {
		errors = append(errors, ValidationError{
			Type:    "required_field",
			Message: "Test suite name is required",
			Field:   "name",
			Value:   testSuite.Name,
			Suggestions: []string{
				"Add a descriptive name for your test suite",
				"Example: 'User Authentication Test Suite'",
			},
		})
	}

	// Validate test cases
	if len(testSuite.TestCases) == 0 {
		errors = append(errors, ValidationError{
			Type:    "required_field",
			Message: "Test suite must reference at least one test case",
			Field:   "testcases",
			Value:   len(testSuite.TestCases),
			Suggestions: []string{
				"Add test case file references to your suite",
				"Example: - file: 'test-login.robogo'",
			},
		})
	}

	// Validate test case references
	for i, testCaseRef := range testSuite.TestCases {
		if testCaseRef.File == "" {
			errors = append(errors, ValidationError{
				Type:    "required_field",
				Message: "Test case file reference cannot be empty",
				Field:   fmt.Sprintf("testcases[%d].file", i),
				Value:   testCaseRef.File,
				Suggestions: []string{
					"Provide a valid file path to the test case",
					"Example: 'tests/test-api.robogo'",
				},
			})
		}
	}

	// Validate setup steps
	for i, step := range testSuite.Setup {
		stepErrors := ve.ValidateStep(step)
		for _, err := range stepErrors {
			err.Field = fmt.Sprintf("setup[%d].%s", i, err.Field)
			errors = append(errors, err)
		}
	}

	// Validate teardown steps
	for i, step := range testSuite.Teardown {
		stepErrors := ve.ValidateStep(step)
		for _, err := range stepErrors {
			err.Field = fmt.Sprintf("teardown[%d].%s", i, err.Field)
			errors = append(errors, err)
		}
	}

	return errors
}

// ValidateStep validates a single step and returns validation errors
func (ve *DefaultValidationEngine) ValidateStep(step parser.Step) []ValidationError {
	var errors []ValidationError

	// Validate step name
	if step.Name == "" {
		errors = append(errors, ValidationError{
			Type:    "required_field",
			Message: "Step name is required",
			Field:   "name",
			Value:   step.Name,
			Suggestions: []string{
				"Add a descriptive name for your step",
				"Example: 'Send HTTP GET request'",
				"Step names help with debugging and reporting",
			},
		})
	}

	// Validate action
	if step.Action == "" {
		errors = append(errors, ValidationError{
			Type:    "required_field",
			Message: "Step action is required",
			Field:   "action",
			Value:   step.Action,
			Suggestions: []string{
				"Specify an action for this step",
				"Common actions: http, assert, log, variable",
				"Use 'robogo list' to see all available actions",
			},
		})
	} else {
		// Validate known actions
		validActions := []string{
			"http", "http_get", "http_post", "postgres", "spanner",
			"assert", "log", "variable", "template", "sleep",
			"get_time", "get_random", "if", "for", "while",
			"kafka", "rabbitmq", "tdm", "control",
		}
		
		isValidAction := false
		for _, validAction := range validActions {
			if step.Action == validAction {
				isValidAction = true
				break
			}
		}
		
		if !isValidAction {
			errors = append(errors, ValidationError{
				Type:    "unknown_action",
				Message: fmt.Sprintf("Unknown action '%s'", step.Action),
				Field:   "action",
				Value:   step.Action,
				Suggestions: []string{
					"Check if the action name is spelled correctly",
					"Use 'robogo list' to see available actions",
					fmt.Sprintf("Did you mean one of: %s", strings.Join(validActions[:5], ", ")),
				},
			})
		}
	}

	// Validate control flow exclusivity
	controlFlowCount := 0
	if step.If != nil {
		controlFlowCount++
	}
	if step.For != nil {
		controlFlowCount++
	}
	if step.While != nil {
		controlFlowCount++
	}
	
	if controlFlowCount > 1 {
		errors = append(errors, ValidationError{
			Type:    "invalid_configuration",
			Message: "Step cannot have multiple control flow statements (if/for/while)",
			Field:   "control_flow",
			Value:   controlFlowCount,
			Suggestions: []string{
				"Use only one control flow statement per step",
				"Split complex logic into multiple steps",
			},
		})
	}

	// Validate action-specific requirements
	errors = append(errors, ve.validateActionSpecificRequirements(step)...)

	return errors
}

// validateActionSpecificRequirements validates requirements specific to certain actions
func (ve *DefaultValidationEngine) validateActionSpecificRequirements(step parser.Step) []ValidationError {
	var errors []ValidationError

	switch step.Action {
	case "http", "http_get", "http_post":
		if len(step.Args) == 0 {
			errors = append(errors, ValidationError{
				Type:    "required_field",
				Message: "HTTP action requires at least one argument (URL)",
				Field:   "args",
				Value:   len(step.Args),
				Suggestions: []string{
					"Add URL as first argument",
					"Example: args: ['https://api.example.com/users']",
				},
			})
		}

	case "assert":
		if len(step.Args) < 3 {
			errors = append(errors, ValidationError{
				Type:    "insufficient_arguments",
				Message: "Assert action requires at least 3 arguments (actual, operator, expected)",
				Field:   "args",
				Value:   len(step.Args),
				Suggestions: []string{
					"Format: args: [actual_value, operator, expected_value]",
					"Example: args: ['${response.status}', '==', '200']",
					"Operators: ==, !=, <, >, <=, >=, contains, matches",
				},
			})
		}

	case "variable":
		if len(step.Args) < 2 {
			errors = append(errors, ValidationError{
				Type:    "insufficient_arguments",
				Message: "Variable action requires at least 2 arguments (name, value)",
				Field:   "args",
				Value:   len(step.Args),
				Suggestions: []string{
					"Format: args: [variable_name, value]",
					"Example: args: ['user_id', '12345']",
				},
			})
		}

	case "postgres":
		if len(step.Args) == 0 {
			errors = append(errors, ValidationError{
				Type:    "required_field",
				Message: "Postgres action requires SQL query as argument",
				Field:   "args",
				Value:   len(step.Args),
				Suggestions: []string{
					"Add SQL query as first argument",
					"Example: args: ['SELECT * FROM users WHERE id = $1']",
				},
			})
		}

	case "log":
		if len(step.Args) == 0 {
			errors = append(errors, ValidationError{
				Type:    "required_field",
				Message: "Log action requires message as argument",
				Field:   "args",
				Value:   len(step.Args),
				Suggestions: []string{
					"Add log message as argument",
					"Example: args: ['Test step completed']",
				},
			})
		}
	}

	return errors
}