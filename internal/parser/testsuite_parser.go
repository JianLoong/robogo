package parser

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ParseTestSuite parses a test suite YAML file
func ParseTestSuite(filePath string) (*TestSuite, error) {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read test suite file '%s': %w", filePath, err)
	}

	// Parse YAML
	var testSuite TestSuite
	err = yaml.Unmarshal(data, &testSuite)
	if err != nil {
		return nil, fmt.Errorf("failed to parse test suite YAML '%s': %w", filePath, err)
	}

	// Validate the test suite
	err = validateTestSuite(&testSuite, filePath)
	if err != nil {
		return nil, fmt.Errorf("test suite validation failed '%s': %w", filePath, err)
	}

	return &testSuite, nil
}

// validateTestSuite validates a test suite configuration
func validateTestSuite(testSuite *TestSuite, filePath string) error {
	// Check required fields
	if testSuite.Name == "" {
		return fmt.Errorf("test suite name is required")
	}

	if len(testSuite.TestCases) == 0 {
		return fmt.Errorf("test suite must contain at least one test case")
	}

	// Validate test case files exist
	suiteDir := filepath.Dir(filePath)
	for i, testCase := range testSuite.TestCases {
		if testCase.File == "" {
			return fmt.Errorf("test case %d: file path is required", i+1)
		}

		// Resolve relative paths relative to the suite file
		resolvedPath := testCase.File
		if !filepath.IsAbs(testCase.File) {
			resolvedPath = filepath.Join(suiteDir, testCase.File)
		}

		// Check if file exists
		if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
			return fmt.Errorf("test case %d: file not found '%s'", i+1, resolvedPath)
		}
	}

	return nil
}

// LoadTestCases loads all test cases referenced in a test suite
func LoadTestCases(testSuite *TestSuite, suiteFilePath string) ([]*TestCase, error) {
	var testCases []*TestCase
	suiteDir := filepath.Dir(suiteFilePath)

	for i, testCaseRef := range testSuite.TestCases {
		// Resolve file path
		resolvedPath := testCaseRef.File
		if !filepath.IsAbs(testCaseRef.File) {
			resolvedPath = filepath.Join(suiteDir, testCaseRef.File)
		}

		// Parse the test case
		testCase, err := ParseTestFile(resolvedPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load test case %d '%s': %w", i+1, resolvedPath, err)
		}

		// Merge variables if provided
		if testCaseRef.Variables != nil {
			testCase = mergeVariables(testCase, testCaseRef.Variables)
		}

		testCases = append(testCases, testCase)
	}

	return testCases, nil
}

// mergeVariables merges test case variables with overrides
func mergeVariables(testCase *TestCase, overrides *Variables) *TestCase {
	// Create a copy of the test case
	merged := *testCase

	// Initialize variables if not present
	if merged.Variables.Regular == nil {
		merged.Variables.Regular = make(map[string]interface{})
	}

	// Merge regular variables
	if overrides.Regular != nil {
		for k, v := range overrides.Regular {
			merged.Variables.Regular[k] = v
		}
	}

	// Merge secrets
	if overrides.Secrets != nil {
		if merged.Variables.Secrets == nil {
			merged.Variables.Secrets = make(map[string]Secret)
		}
		for k, v := range overrides.Secrets {
			merged.Variables.Secrets[k] = v
		}
	}

	return &merged
}
