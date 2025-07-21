package internal

import (
	"fmt"
	"os"

	"github.com/JianLoong/robogo/internal/types"
	"gopkg.in/yaml.v3"
)

// validateSteps recursively validates steps and nested steps
func validateSteps(steps []types.Step, stepPath string) error {
	for i, step := range steps {
		currentPath := fmt.Sprintf("%sstep %d", stepPath, i+1)
		
		if step.Name == "" {
			return fmt.Errorf("%s: name is required", currentPath)
		}
		
		if step.Action == "" && len(step.Steps) == 0 {
			return fmt.Errorf("%s: either 'action' or 'steps' field is required", currentPath)
		}
		
		if step.Action != "" && len(step.Steps) > 0 {
			return fmt.Errorf("%s: cannot have both 'action' and 'steps' fields", currentPath)
		}
		
		// Recursively validate nested steps
		if len(step.Steps) > 0 {
			if err := validateSteps(step.Steps, currentPath+" -> "); err != nil {
				return err
			}
		}
	}
	return nil
}

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

	// Validate main steps
	if err := validateSteps(testCase.Steps, ""); err != nil {
		return nil, err
	}

	// Validate setup steps if present
	if len(testCase.Setup) > 0 {
		if err := validateSteps(testCase.Setup, "setup "); err != nil {
			return nil, err
		}
	}

	// Validate teardown steps if present
	if len(testCase.Teardown) > 0 {
		if err := validateSteps(testCase.Teardown, "teardown "); err != nil {
			return nil, err
		}
	}

	return &testCase, nil
}
