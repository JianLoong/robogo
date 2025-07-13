package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/JianLoong/robogo/internal/actions"
	"github.com/JianLoong/robogo/internal/output"
	"github.com/JianLoong/robogo/internal/parser"
)

// TestExecutionService provides a clean, decoupled interface for test execution
// This replaces the tightly coupled TestRunner with proper separation of concerns
// Implements TestExecutor interface
type TestExecutionService struct {
	context     ExecutionContext
	stepService StepExecutor
}

// NewTestExecutionService creates a new test execution service
func NewTestExecutionService(executor *actions.ActionExecutor) TestExecutor {
	execContext := NewExecutionContext(executor)
	stepService := NewStepExecutionService(execContext)

	return &TestExecutionService{
		context:     execContext,
		stepService: stepService,
	}
}

// NewTestExecutionServiceWithContext creates a service with a custom execution context
func NewTestExecutionServiceWithContext(context ExecutionContext) TestExecutor {
	stepService := NewStepExecutionService(context)

	return &TestExecutionService{
		context:     context,
		stepService: stepService,
	}
}

// ExecuteTestCase executes a single test case with proper lifecycle management
func (tes *TestExecutionService) ExecuteTestCase(ctx context.Context, testCase *parser.TestCase, silent bool) (*parser.TestResult, error) {
	startTime := time.Now()

	result := &parser.TestResult{
		TestCase:    testCase,
		Status:      "FAILED",
		Duration:    0,
		StepResults: []parser.StepResult{},
	}

	// Initialize test case context
	if err := tes.initializeTestCase(testCase); err != nil {
		return result, fmt.Errorf("failed to initialize test case: %w", err)
	}

	// Start output capture with real-time display
	if outputCapture, ok := tes.context.Output().(*OutputCapture); ok {
		outputCapture.StartWithRealTimeDisplay(!silent)
	} else {
		tes.context.Output().StartCapture()
	}

	// Print test case info
	if !silent {
		output.PrintTestCaseStart(testCase.Name)
		if testCase.Description != "" {
			output.PrintTestCaseDescription(testCase.Description)
		}
		output.PrintTestCaseSteps(len(testCase.Steps))
	}

	// Execute steps
	stepResults, err := tes.executeSteps(ctx, testCase, silent)
	result.StepResults = stepResults

	// Calculate final status
	result.Duration = time.Since(startTime)
	result.Status = tes.calculateTestStatus(stepResults, err)

	// Capture output
	result.CapturedOutput = tes.context.Output().StopCapture()

	// Calculate step statistics
	tes.calculateStepStatistics(result)

	return result, err
}

// ExecuteTestSuite executes a test suite with proper setup/teardown
func (tes *TestExecutionService) ExecuteTestSuite(ctx context.Context, testSuite *parser.TestSuite, filePath string, silent bool) (*parser.TestSuiteResult, error) {
	startTime := time.Now()

	result := &parser.TestSuiteResult{
		TestSuite:   testSuite,
		Duration:    0,
		CaseResults: []parser.TestCaseResult{},
	}

	// Execute setup if present
	if err := tes.executeSetup(ctx, testSuite, silent); err != nil {
		result.SetupStatus = fmt.Sprintf("FAILED: %v", err)
		return result, fmt.Errorf("test suite setup failed: %w", err)
	}
	result.SetupStatus = "PASSED"

	// Execute test cases
	for _, testCaseRef := range testSuite.TestCases {
		caseResult, err := tes.executeTestCaseFromPath(ctx, testCaseRef.File, silent)
		result.CaseResults = append(result.CaseResults, caseResult)

		if err != nil && !silent {
			fmt.Printf("Test case %s failed: %v\n", testCaseRef.File, err)
		}
	}

	// Execute teardown
	if err := tes.executeTeardown(ctx, testSuite, silent); err != nil {
		result.TeardownStatus = fmt.Sprintf("FAILED: %v", err)
		if !silent {
			fmt.Printf("Test suite teardown failed: %v\n", err)
		}
	} else {
		result.TeardownStatus = "PASSED"
	}

	// Calculate final results
	result.Duration = time.Since(startTime)
	tes.calculateSuiteStatistics(result)

	return result, nil
}

// Private helper methods

func (tes *TestExecutionService) initializeTestCase(testCase *parser.TestCase) error {
	// Load templates if present
	if len(testCase.Templates) > 0 {
		templateNames := make([]string, 0, len(testCase.Templates))
		for name := range testCase.Templates {
			templateNames = append(templateNames, name)
		}

		if !tes.isQuiet() {
			output.PrintTemplatesLoaded(len(testCase.Templates), fmt.Sprintf("%v", templateNames))
		}
	}

	// Initialize variables using the VariableManager's proper initialization
	// This ensures correct multi-pass substitution and cross-substitution between secrets and variables
	if execCtx, ok := tes.context.(*DefaultExecutionContext); ok {
		execCtx.variableManager.InitializeVariables(testCase)
	} else {
		// Fallback for other ExecutionContext implementations
		if testCase.Variables.Regular != nil {
			// Load regular variables
			for key, value := range testCase.Variables.Regular {
				if err := tes.context.Variables().Set(key, value); err != nil {
					return fmt.Errorf("failed to set variable %s: %w", key, err)
				}
			}
		}

		// Load secrets
		if testCase.Variables.Secrets != nil {
			if err := tes.context.Variables().LoadSecrets(testCase.Variables.Secrets); err != nil {
				return fmt.Errorf("failed to load secrets: %w", err)
			}
		}
	}

	return nil
}

func (tes *TestExecutionService) executeSteps(ctx context.Context, testCase *parser.TestCase, silent bool) ([]parser.StepResult, error) {
	// Check for parallel execution configuration
	if testCase.Parallel != nil && testCase.Parallel.Steps {
		return tes.stepService.ExecuteStepsParallel(ctx, testCase.Steps, testCase.Parallel, silent)
	}

	// Sequential execution
	return tes.stepService.ExecuteSteps(ctx, testCase.Steps, silent)
}

func (tes *TestExecutionService) executeSetup(ctx context.Context, testSuite *parser.TestSuite, silent bool) error {
	if len(testSuite.Setup) == 0 {
		return nil
	}

	setupTestCase := &parser.TestCase{
		Name:  "Suite Setup",
		Steps: testSuite.Setup,
	}

	_, err := tes.ExecuteTestCase(ctx, setupTestCase, silent)
	return err
}

func (tes *TestExecutionService) executeTeardown(ctx context.Context, testSuite *parser.TestSuite, silent bool) error {
	if len(testSuite.Teardown) == 0 {
		return nil
	}

	teardownTestCase := &parser.TestCase{
		Name:  "Suite Teardown",
		Steps: testSuite.Teardown,
	}

	_, err := tes.ExecuteTestCase(ctx, teardownTestCase, silent)
	return err
}

func (tes *TestExecutionService) executeTestCaseFromPath(ctx context.Context, path string, silent bool) (parser.TestCaseResult, error) {
	// Parse the test case from file
	testCase, err := parser.ParseTestFile(path)
	if err != nil {
		return parser.TestCaseResult{
			TestCase: &parser.TestCase{Name: path},
			Status:   "FAILED",
			Error:    fmt.Sprintf("Failed to parse test case: %v", err),
		}, err
	}

	// Execute the test case
	result, err := tes.ExecuteTestCase(ctx, testCase, silent)
	if err != nil {
		return parser.TestCaseResult{
			TestCase: testCase,
			Status:   "FAILED",
			Error:    fmt.Sprintf("Failed to execute test case: %v", err),
		}, err
	}

	// Convert TestResult to TestCaseResult
	return parser.TestCaseResult{
		TestCase: testCase,
		Status:   result.Status,
		Error:    result.ErrorMessage,
	}, nil
}

func (tes *TestExecutionService) calculateTestStatus(stepResults []parser.StepResult, err error) string {
	if err != nil {
		return "FAILED"
	}

	for _, result := range stepResults {
		if result.Status == "FAILED" {
			return "FAILED"
		}
	}

	return "PASSED"
}

func (tes *TestExecutionService) calculateStepStatistics(result *parser.TestResult) {
	for _, stepResult := range result.StepResults {
		result.TotalSteps++
		switch stepResult.Status {
		case "PASSED":
			result.PassedSteps++
		case "FAILED":
			result.FailedSteps++
		case "SKIPPED":
			result.SkippedSteps++
		}
	}
}

func (tes *TestExecutionService) calculateSuiteStatistics(result *parser.TestSuiteResult) {
	for _, caseResult := range result.CaseResults {
		if caseResult.Result != nil {
			result.TotalSteps += caseResult.Result.TotalSteps
			result.PassedSteps += caseResult.Result.PassedSteps
			result.FailedSteps += caseResult.Result.FailedSteps
			result.SkippedSteps += caseResult.Result.SkippedSteps
		}
	}
}

func (tes *TestExecutionService) isQuiet() bool {
	// This could be configured via the execution context
	return false
}

// GetContext returns the execution context for advanced usage
func (tes *TestExecutionService) GetContext() ExecutionContext {
	return tes.context
}

// Cleanup performs any necessary cleanup
func (tes *TestExecutionService) Cleanup() error {
	return tes.context.Cleanup()
}

// GetExecutor returns the action executor (for interface compatibility)
func (tes *TestExecutionService) GetExecutor() *actions.ActionExecutor {
	// Extract executor from context - this is a workaround for interface compatibility
	if actionExec, ok := tes.context.Actions().(*actionExecutorAdapter); ok {
		return actionExec.executor
	}
	return nil
}

// ShouldSkipTestCase evaluates skip condition for test case (for interface compatibility)
func (tes *TestExecutionService) ShouldSkipTestCase(testCase *parser.TestCase, context string) SkipInfo {
	// Simple skip evaluation logic
	if testCase.Skip == nil {
		return SkipInfo{ShouldSkip: false}
	}

	switch v := testCase.Skip.(type) {
	case bool:
		if v {
			return SkipInfo{ShouldSkip: true, Reason: "skip condition is true"}
		}
		return SkipInfo{ShouldSkip: false}
	case string:
		substituted := tes.context.Variables().Substitute(v)
		if substituted != "" && substituted != "false" && substituted != "0" {
			return SkipInfo{ShouldSkip: true, Reason: substituted}
		}
		return SkipInfo{ShouldSkip: false}
	default:
		strValue := fmt.Sprintf("%v", v)
		if strValue != "" && strValue != "false" && strValue != "0" {
			return SkipInfo{ShouldSkip: true, Reason: strValue}
		}
		return SkipInfo{ShouldSkip: false}
	}
}
