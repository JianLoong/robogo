package parser

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseTestFile parses a test case from a file
func ParseTestFile(filename string) (*TestCase, error) {
	// Check file extension
	if !isValidTestFile(filename) {
		return nil, fmt.Errorf("unsupported file extension. Use .yaml, .yml, or .robogo")
	}

	// Read file content
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read test file: %w", err)
	}

	// Parse YAML
	var testCase TestCase
	if err := yaml.Unmarshal(content, &testCase); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Debug output
	fmt.Printf("DEBUG PARSER: Parsed test case '%s' with %d steps\n", testCase.Name, len(testCase.Steps))

	// Validate test case
	if err := validateTestCase(&testCase); err != nil {
		return nil, fmt.Errorf("test case validation failed: %w", err)
	}

	return &testCase, nil
}

// isValidTestFile checks if the file has a valid extension
func isValidTestFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".yaml" || ext == ".yml" || ext == ".robogo"
}

// validateTestCase validates the test case structure
func validateTestCase(testCase *TestCase) error {
	if testCase.Name == "" {
		return fmt.Errorf("test case must have a name")
	}

	if len(testCase.Steps) == 0 {
		return fmt.Errorf("test case must have at least one step")
	}

	// Validate each step
	for i, step := range testCase.Steps {
		// Check if step has any control flow or regular action
		hasControlFlow := step.If != nil || step.For != nil || step.While != nil
		hasAction := step.Action != ""

		if !hasControlFlow && !hasAction {
			return fmt.Errorf("step %d: must have either an action or control flow (if/for/while)", i+1)
		}

		if hasControlFlow && hasAction {
			return fmt.Errorf("step %d: cannot have both action and control flow", i+1)
		}
	}

	return nil
}
