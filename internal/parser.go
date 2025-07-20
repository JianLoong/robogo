package internal

import (
	"fmt"
	"os"

	"github.com/JianLoong/robogo/internal/types"
	"gopkg.in/yaml.v3"
)

// Simple parser - no complex validation, just parse YAML
func ParseTestFile(filename string) (*types.TestCase, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	var testCase types.TestCase
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
