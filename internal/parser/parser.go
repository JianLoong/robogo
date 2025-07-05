package parser

import (
	"fmt"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseTestFile parses a YAML test case file
func ParseTestFile(filename string) (*TestCase, error) {
	// Check file extension
	if !isValidTestFile(filename) {
		return nil, fmt.Errorf("unsupported file extension. Use .yaml, .yml, or .robogo")
	}

	// Read file
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	// Parse YAML
	var testCase TestCase
	if err := yaml.Unmarshal(data, &testCase); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate test case
	if err := validateTestCase(&testCase); err != nil {
		return nil, fmt.Errorf("invalid test case: %w", err)
	}

	return &testCase, nil
}

// isValidTestFile checks if the file has a valid extension
func isValidTestFile(filename string) bool {
	ext := strings.ToLower(filename)
	return strings.HasSuffix(ext, ".yaml") ||
		strings.HasSuffix(ext, ".yml") ||
		strings.HasSuffix(ext, ".robogo")
}

// validateTestCase validates a test case
func validateTestCase(tc *TestCase) error {
	if tc.Name == "" {
		return fmt.Errorf("test case name is required")
	}

	if len(tc.Steps) == 0 {
		return fmt.Errorf("test case must have at least one step")
	}

	for i, step := range tc.Steps {
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
