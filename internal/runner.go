package internal

import (
	"fmt"
	"os"
	"time"

	"github.com/JianLoong/robogo/internal/actions"
	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/execution"
	"github.com/JianLoong/robogo/internal/types"
)

// TestRunner executes a test case and manages variables and control flow.
type TestRunner struct {
	variables      *common.Variables
	strategyRouter *execution.ExecutionStrategyRouter
}

// NewTestRunner creates a new TestRunner with direct strategy router.
func NewTestRunner() *TestRunner {
	variables := common.NewVariables()
	actionRegistry := actions.NewActionRegistry()
	conditionEvaluator := execution.NewBasicConditionEvaluator(variables)
	
	// Create strategy router and register strategies directly
	router := execution.NewExecutionStrategyRouter()
	router.RegisterStrategy(execution.NewConditionalExecutionStrategy(conditionEvaluator, router))
	router.RegisterStrategy(execution.NewRetryExecutionStrategy(variables, actionRegistry))
	router.RegisterStrategy(execution.NewNestedStepsExecutionStrategy(router))
	router.RegisterStrategy(execution.NewBasicExecutionStrategy(variables, actionRegistry))
	
	return &TestRunner{
		variables:      variables,
		strategyRouter: router,
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

	// 1. Run setup phase
	setupResults, setupSkipped := r.runSetupPhase(testCase.Setup)
	result.SetupSteps = setupResults
	
	// If setup failed critically, skip the main test
	if setupSkipped {
		result.Status = "SKIPPED"
		result.Duration = time.Since(start)
		fmt.Printf("\n[SETUP] Test skipped due to critical setup failure\n")
		return result, nil
	}

	// 2. Run main test steps
	testFailed := false
	for i, step := range testCase.Steps {
		stepResult := r.strategyRouter.Execute(step, i+1, nil)
		var stepResults []types.StepResult
		if stepResult != nil {
			stepResults = append(stepResults, *stepResult)
		}
		result.Steps = append(result.Steps, stepResults...)

		if r.anyStepFailedOrErrored(stepResults) {
			result.Status = r.aggregateStatus(stepResults)
			result.ErrorInfo = r.getFirstErrorInfo(stepResults)
			testFailed = true
			
			// Check if step has continue flag
			if !step.Continue {
				break
			}
			
			fmt.Printf("⚠️  Step failed but continuing due to continue flag: %s\n", step.Name)
		}
	}

	// 3. Always run teardown phase (regardless of test outcome)
	teardownResults := r.runTeardownPhase(testCase.Teardown, testFailed)
	result.TeardownSteps = teardownResults

	result.Duration = time.Since(start)
	return result, nil
}

// printTestHeader prints the test case header information.
func (r *TestRunner) printTestHeader(testCase *types.TestCase) {
	fmt.Printf("Running test case: %s\n", testCase.Name)
	if testCase.Description != "" {
		fmt.Printf("Description: %s\n", testCase.Description)
	}
	setupCount := len(testCase.Setup)
	teardownCount := len(testCase.Teardown)
	fmt.Printf("Setup: %d, Steps: %d, Teardown: %d\n\n", setupCount, len(testCase.Steps), teardownCount)
	os.Stdout.Sync()
}

// runSetupPhase executes setup steps, returns (results, shouldSkipTest)
func (r *TestRunner) runSetupPhase(setupSteps []types.Step) ([]types.StepResult, bool) {
	if len(setupSteps) == 0 {
		return nil, false
	}

	fmt.Printf("[SETUP] Running %d setup steps...\n", len(setupSteps))
	
	var results []types.StepResult
	
	for i, step := range setupSteps {
		stepResult := r.strategyRouter.Execute(step, i+1, nil)
		var stepResults []types.StepResult
		if stepResult != nil {
			stepResults = append(stepResults, *stepResult)
		}
		results = append(results, stepResults...)

		// Check for critical failures that should skip the test
		if r.anyStepFailedOrErrored(stepResults) {
			fmt.Printf("[SETUP] ⚠️  Setup step failed: %s\n", step.Name)
			
			// For now, treat all setup failures as warnings, not critical
			// In the future, we could add a "critical: true" flag to setup steps
			fmt.Printf("[SETUP] ⚠️  Continuing with test despite setup failure...\n")
		}
	}
	
	fmt.Printf("[SETUP] ✓ Setup phase completed\n\n")
	return results, false
}

// runTeardownPhase executes teardown steps, always runs regardless of test outcome
func (r *TestRunner) runTeardownPhase(teardownSteps []types.Step, testFailed bool) []types.StepResult {
	if len(teardownSteps) == 0 {
		return nil
	}

	fmt.Printf("\n[TEARDOWN] Running %d teardown steps...\n", len(teardownSteps))
	
	var results []types.StepResult
	
	for i, step := range teardownSteps {
		stepResult := r.strategyRouter.Execute(step, i+1, nil)
		var stepResults []types.StepResult
		if stepResult != nil {
			stepResults = append(stepResults, *stepResult)
		}
		results = append(results, stepResults...)

		// Log teardown failures but don't affect test outcome
		if r.anyStepFailedOrErrored(stepResults) {
			fmt.Printf("[TEARDOWN] ⚠️  Teardown step failed: %s\n", step.Name)
			fmt.Printf("[TEARDOWN] ⚠️  Error: %s\n", r.getErrorMessage(stepResults))
		}
	}
	
	fmt.Printf("[TEARDOWN] ✓ Teardown phase completed\n")
	return results
}

// getErrorMessage extracts error message from step results
func (r *TestRunner) getErrorMessage(stepResults []types.StepResult) string {
	for _, sr := range stepResults {
		if sr.Result.ErrorInfo != nil {
			return sr.Result.ErrorInfo.Message
		}
		if sr.Result.FailureInfo != nil {
			return sr.Result.FailureInfo.Message
		}
	}
	
	return "Unknown error"
}

// anyStepFailedOrErrored returns true if any step failed or errored.
func (r *TestRunner) anyStepFailedOrErrored(stepResults []types.StepResult) bool {
	for _, sr := range stepResults {
		switch sr.Result.Status {
		case types.ActionStatusError, types.ActionStatusFailed:
			return true
		}
	}
	return false
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

// getFirstErrorInfo extracts the first ErrorInfo from step results.
// If a step has FailureInfo but no ErrorInfo, converts FailureInfo to ErrorInfo.
func (r *TestRunner) getFirstErrorInfo(stepResults []types.StepResult) *types.ErrorInfo {
	for _, sr := range stepResults {
		switch sr.Result.Status {
		case types.ActionStatusError, types.ActionStatusFailed:
			// First check for ErrorInfo
			if sr.Result.ErrorInfo != nil {
				return sr.Result.ErrorInfo
			}
			// If no ErrorInfo but has FailureInfo, convert it
			if sr.Result.FailureInfo != nil {
				return &types.ErrorInfo{
					Category:  types.ErrorCategory(sr.Result.FailureInfo.Category),
					Code:      sr.Result.FailureInfo.Code,
					Message:   sr.Result.FailureInfo.Message,
					Timestamp: sr.Result.FailureInfo.Timestamp,
				}
			}
		}
	}
	return nil
}
