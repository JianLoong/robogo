package internal

import (
	"fmt"
	"os"
	"time"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/types"
)

// TestRunner executes a test case and manages variables and control flow.
type TestRunner struct {
	variables           *common.Variables
	controlFlowExecutor *ControlFlowExecutor
}

// NewTestRunner creates a new TestRunner with fresh variables.
func NewTestRunner() *TestRunner {
	variables := common.NewVariables()
	return &TestRunner{
		variables:           variables,
		controlFlowExecutor: NewControlFlowExecutor(variables),
	}
}

// RunTest executes a single test file and returns the aggregated result.
func (r *TestRunner) RunTest(filename string) (*types.TestResult, error) {
	testCase, err := ParseTestFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to parse test file: %w", err)
	}

	if testCase.Variables.Vars != nil {
		r.variables.Load(testCase.Variables.Vars)
	}

	start := time.Now()
	result := &types.TestResult{
		Name:   testCase.Name,
		Status: string(types.ActionStatusPassed),
		Steps:  make([]types.StepResult, 0, len(testCase.Steps)),
	}

	r.printTestHeader(testCase)

	for i, step := range testCase.Steps {
		stepResults, stepErr := r.controlFlowExecutor.ExecuteStepWithControlFlow(step, i+1)
		result.Steps = append(result.Steps, stepResults...)

		if r.anyStepFailedOrErrored(stepResults, stepErr) {
			result.Status = r.aggregateStatus(stepResults)
			result.Error = r.getFirstError(stepResults)
			break
		}
	}

	result.Duration = time.Since(start)
	return result, nil
}

// printTestHeader prints the test case header information.
func (r *TestRunner) printTestHeader(testCase *types.TestCase) {
	fmt.Printf("Running test case: %s\n", testCase.Name)
	if testCase.Description != "" {
		fmt.Printf("Description: %s\n", testCase.Description)
	}
	fmt.Printf("Steps: %d\n\n", len(testCase.Steps))
	os.Stdout.Sync()
}

// anyStepFailedOrErrored returns true if any step failed or errored, or if stepErr is not nil.
func (r *TestRunner) anyStepFailedOrErrored(stepResults []types.StepResult, stepErr error) bool {
	for _, sr := range stepResults {
		switch sr.Result.Status {
		case types.ActionStatusError, types.ActionStatusFailed:
			return true
		}
	}
	return stepErr != nil
}

// aggregateStatus returns the most severe status among the step results.
func (r *TestRunner) aggregateStatus(stepResults []types.StepResult) string {
	for _, sr := range stepResults {
		switch sr.Result.Status {
		case types.ActionStatusError:
			return string(types.ActionStatusError)
		case types.ActionStatusFailed:
			return string(types.ActionStatusFailed)
		}
	}
	return string(types.ActionStatusPassed)
}

// getFirstError extracts the first error from step results.
func (r *TestRunner) getFirstError(stepResults []types.StepResult) string {
	for _, sr := range stepResults {
		switch sr.Result.Status {
		case types.ActionStatusError, types.ActionStatusFailed:
			if sr.Result.Error != "" {
				return sr.Result.Error
			}
		}
	}
	return ""
}
