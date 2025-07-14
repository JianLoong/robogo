package internal

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Simple parser - no complex validation, just parse YAML
func ParseTestFile(filename string) (*TestCase, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	var testCase TestCase
	if err := yaml.Unmarshal(data, &testCase); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Basic validation
	if testCase.Name == "" {
		return nil, fmt.Errorf("test case name is required")
	}

	if len(testCase.Steps) == 0 {
		return nil, fmt.Errorf("test case must have at least one step")
	}

	for i, step := range testCase.Steps {
		if step.Name == "" {
			return nil, fmt.Errorf("step %d: name is required", i+1)
		}
		if step.Action == "" {
			return nil, fmt.Errorf("step %d: action is required", i+1)
		}
	}

	return &testCase, nil
}

func ParseTestSuite(filename string) (*TestSuite, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	var testSuite TestSuite
	if err := yaml.Unmarshal(data, &testSuite); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Basic validation
	if testSuite.Name == "" {
		return nil, fmt.Errorf("test suite name is required")
	}

	if len(testSuite.TestCases) == 0 {
		return nil, fmt.Errorf("test suite must have at least one test case")
	}

	return &testSuite, nil
}