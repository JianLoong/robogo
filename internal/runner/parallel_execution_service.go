package runner

import (
	"context"
	"fmt"
	"sync"

	"github.com/JianLoong/robogo/internal/parser"
)

// ParallelExecutionService handles parallel execution of steps
type ParallelExecutionService struct {
	stepExecutor StepExecutor
}

// NewParallelExecutionService creates a new parallel execution service
func NewParallelExecutionService(stepExecutor StepExecutor) *ParallelExecutionService {
	return &ParallelExecutionService{
		stepExecutor: stepExecutor,
	}
}

// ExecuteParallel executes steps in parallel based on configuration
func (pes *ParallelExecutionService) ExecuteParallel(ctx context.Context, steps []parser.Step, config *parser.ParallelConfig, silent bool) ([]parser.StepResult, error) {
	if config == nil || !config.Enabled {
		// Fall back to sequential execution
		return pes.stepExecutor.ExecuteSteps(ctx, steps, silent)
	}

	// Analyze dependencies and group steps
	stepGroups := pes.analyzeStepDependencies(steps, config)
	
	var allResults []parser.StepResult
	
	// Execute step groups in dependency order
	for groupIndex, group := range stepGroups {
		if !silent {
			fmt.Printf("Executing step group %d with %d steps\n", groupIndex+1, len(group))
		}
		
		groupResults, err := pes.executeStepGroupParallel(ctx, group, config.MaxConcurrency, silent)
		if err != nil {
			return append(allResults, groupResults...), err
		}
		
		allResults = append(allResults, groupResults...)
		
		// Check for failures in this group before proceeding
		for _, result := range groupResults {
			if result.Status == "FAILED" {
				return allResults, fmt.Errorf("step group %d failed: %s", groupIndex+1, result.Error)
			}
		}
	}
	
	return allResults, nil
}

// analyzeStepDependencies groups steps based on their dependencies
func (pes *ParallelExecutionService) analyzeStepDependencies(steps []parser.Step, config *parser.ParallelConfig) [][]parser.Step {
	// Simple dependency analysis - steps that don't depend on each other can run in parallel
	var groups [][]parser.Step
	var currentGroup []parser.Step
	usedVariables := make(map[string]bool)
	
	for _, step := range steps {
		// Check if this step depends on variables from previous steps
		stepDependencies := pes.getStepDependencies(step)
		hasUnmetDependency := false
		
		for _, dep := range stepDependencies {
			if !usedVariables[dep] {
				hasUnmetDependency = true
				break
			}
		}
		
		// If step has unmet dependencies or is not safe for parallel execution, start new group
		if hasUnmetDependency || !pes.isStepSafeForParallelExecution(step) {
			if len(currentGroup) > 0 {
				groups = append(groups, currentGroup)
				currentGroup = []parser.Step{}
			}
		}
		
		currentGroup = append(currentGroup, step)
		
		// Mark variables produced by this step as available
		if step.Result != "" {
			usedVariables[step.Result] = true
		}
	}
	
	// Add the last group
	if len(currentGroup) > 0 {
		groups = append(groups, currentGroup)
	}
	
	return groups
}

// executeStepGroupParallel executes a group of steps in parallel
func (pes *ParallelExecutionService) executeStepGroupParallel(ctx context.Context, steps []parser.Step, maxConcurrency int, silent bool) ([]parser.StepResult, error) {
	if len(steps) == 1 {
		// Single step - execute directly
		result, err := pes.stepExecutor.ExecuteStep(ctx, steps[0], silent)
		if result != nil {
			return []parser.StepResult{*result}, err
		}
		return []parser.StepResult{}, err
	}
	
	// Limit concurrency
	if maxConcurrency <= 0 {
		maxConcurrency = len(steps)
	}
	if maxConcurrency > len(steps) {
		maxConcurrency = len(steps)
	}
	
	semaphore := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	results := make([]parser.StepResult, len(steps))
	errors := make([]error, len(steps))
	
	// Execute steps in parallel
	for i, step := range steps {
		wg.Add(1)
		go func(index int, s parser.Step) {
			defer wg.Done()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			result, err := pes.stepExecutor.ExecuteStep(ctx, s, silent)
			if result != nil {
				results[index] = *result
			}
			errors[index] = err
		}(i, step)
	}
	
	wg.Wait()
	
	// Check for errors
	for i, err := range errors {
		if err != nil {
			return results, fmt.Errorf("step %s failed: %w", steps[i].Name, err)
		}
	}
	
	return results, nil
}

// getStepDependencies extracts variable dependencies from a step
func (pes *ParallelExecutionService) getStepDependencies(step parser.Step) []string {
	var dependencies []string
	
	// Extract variable references from arguments
	for _, arg := range step.Args {
		if str, ok := arg.(string); ok {
			deps := extractVariableReferences(str)
			dependencies = append(dependencies, deps...)
		}
	}
	
	// Extract from conditional statements
	if step.If != nil {
		deps := extractVariableReferences(step.If.Condition)
		dependencies = append(dependencies, deps...)
	}
	
	if step.For != nil {
		deps := extractVariableReferences(step.For.Condition)
		dependencies = append(dependencies, deps...)
	}
	
	if step.While != nil {
		deps := extractVariableReferences(step.While.Condition)
		dependencies = append(dependencies, deps...)
	}
	
	return dependencies
}

// isStepSafeForParallelExecution determines if a step can be safely executed in parallel
func (pes *ParallelExecutionService) isStepSafeForParallelExecution(step parser.Step) bool {
	// Steps that are generally not safe for parallel execution
	unsafeActions := map[string]bool{
		"template": true, // May modify shared templates
		"log":      false, // Logging is generally safe but order matters for readability
	}
	
	if unsafe, exists := unsafeActions[step.Action]; exists {
		return !unsafe
	}
	
	// Control flow steps are not safe for parallel execution
	if step.If != nil || step.For != nil || step.While != nil {
		return false
	}
	
	// Steps with side effects (like database modifications) may not be safe
	// This is a simplified check - in practice, you'd want more sophisticated analysis
	sideEffectActions := map[string]bool{
		"postgres": true,
		"kafka":    true,
		"rabbitmq": true,
	}
	
	if hasSideEffects := sideEffectActions[step.Action]; hasSideEffects {
		return false
	}
	
	return true
}