package parser

import (
	"fmt"
	"runtime"
	"strings"
)

// GetDefaultParallelConfig returns default parallelism configuration
func GetDefaultParallelConfig() *ParallelConfig {
	return &ParallelConfig{
		Enabled:        false, // Disabled by default for safety
		MaxConcurrency: runtime.NumCPU(),
		TestCases:      true,  // Test case parallelism is safe
		Steps:          false, // Step parallelism needs careful implementation
		HTTPRequests:   false, // HTTP parallelism needs rate limiting
		DatabaseOps:    false, // Database parallelism needs connection pooling
		FileOperations: false, // File operations parallelism needs file locking
	}
}

// MergeParallelConfig merges a custom config with defaults
func MergeParallelConfig(custom *ParallelConfig) *ParallelConfig {
	if custom == nil {
		return GetDefaultParallelConfig()
	}

	defaults := GetDefaultParallelConfig()

	if custom.MaxConcurrency <= 0 {
		custom.MaxConcurrency = defaults.MaxConcurrency
	}

	// If enabled is true, enable safe features by default
	if custom.Enabled {
		if !custom.TestCases {
			custom.TestCases = defaults.TestCases
		}
	}

	return custom
}

// ValidateParallelConfig validates parallelism configuration
func ValidateParallelConfig(config *ParallelConfig) error {
	if config == nil {
		return nil
	}

	if config.MaxConcurrency < 1 {
		return fmt.Errorf("max_concurrency must be at least 1")
	}

	if config.MaxConcurrency > 100 {
		return fmt.Errorf("max_concurrency cannot exceed 100 for safety")
	}

	return nil
}

// IsStepIndependent determines if a step can run independently
func IsStepIndependent(step *Step) bool {
	// Steps with control flow cannot run in parallel
	if step.If != nil || step.For != nil || step.While != nil {
		return false
	}

	// Steps that store results in variables might have dependencies
	if step.Result != "" {
		return false
	}

	// Steps that expect errors might have dependencies
	if step.ExpectError != nil {
		return false
	}

	// Steps that continue on failure might have dependencies
	if step.ContinueOnFailure {
		return false
	}

	// Safe actions that can run in parallel
	safeActions := []string{
		"log",
		"sleep",
		"get_time",
		"get_random",
		"length",
		"http",      // With rate limiting
		"postgres",  // Database operations with connection pooling
	}

	action := strings.ToLower(step.Action)
	for _, safe := range safeActions {
		if action == safe {
			return true
		}
	}

	return false
}

// GetStepDependencies analyzes step dependencies
func GetStepDependencies(step *Step) []string {
	var dependencies []string

	// Check if step uses variables that might be set by other steps
	if step.Args != nil {
		for _, arg := range step.Args {
			if str, ok := arg.(string); ok {
				// Look for variable references like ${var_name}
				if strings.Contains(str, "${") {
					// Extract variable names (simplified)
					// This is a basic implementation
					start := strings.Index(str, "${")
					end := strings.Index(str, "}")
					if start != -1 && end != -1 && end > start {
						varName := str[start+2 : end]
						dependencies = append(dependencies, varName)
					}
				}
			}
		}
	}

	return dependencies
}

// CanStepsRunInParallel checks if two steps can run in parallel
func CanStepsRunInParallel(step1, step2 *Step) bool {
	// If either step is not independent, they cannot run in parallel
	if !IsStepIndependent(step1) || !IsStepIndependent(step2) {
		return false
	}

	// Check for variable dependencies
	deps1 := GetStepDependencies(step1)
	deps2 := GetStepDependencies(step2)

	// If step1 produces a variable that step2 uses, they cannot run in parallel
	if step1.Result != "" {
		for _, dep := range deps2 {
			if dep == step1.Result {
				return false
			}
		}
	}

	// If step2 produces a variable that step1 uses, they cannot run in parallel
	if step2.Result != "" {
		for _, dep := range deps1 {
			if dep == step2.Result {
				return false
			}
		}
	}

	return true
}

// GroupIndependentSteps groups steps that can run in parallel
func GroupIndependentSteps(steps []Step) [][]Step {
	var groups [][]Step
	var currentGroup []Step

	for _, step := range steps {
		// If this step is independent and can run with the current group
		if IsStepIndependent(&step) {
			canAdd := true
			for _, existingStep := range currentGroup {
				if !CanStepsRunInParallel(&step, &existingStep) {
					canAdd = false
					break
				}
			}

			if canAdd {
				currentGroup = append(currentGroup, step)
			} else {
				// Start a new group
				if len(currentGroup) > 0 {
					groups = append(groups, currentGroup)
				}
				currentGroup = []Step{step}
			}
		} else {
			// Non-independent step starts a new group
			if len(currentGroup) > 0 {
				groups = append(groups, currentGroup)
			}
			groups = append(groups, []Step{step})
			currentGroup = nil
		}
	}

	// Add the last group
	if len(currentGroup) > 0 {
		groups = append(groups, currentGroup)
	}

	return groups
}
