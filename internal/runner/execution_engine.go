package runner

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/JianLoong/robogo/internal/actions"
	"github.com/JianLoong/robogo/internal/parser"
)

// ExecutionEngine handles the core test execution logic
type ExecutionEngine struct {
	runner *TestRunner
}

// NewExecutionEngine creates a new execution engine
func NewExecutionEngine(runner *TestRunner) *ExecutionEngine {
	return &ExecutionEngine{
		runner: runner,
	}
}

// ExecuteTestCase executes a test case and returns the result
func (engine *ExecutionEngine) ExecuteTestCase(testCase *parser.TestCase, silent bool) (*parser.TestResult, error) {
	// Initialize test runner
	engine.runner.initializeVariables(testCase)
	engine.runner.initializeTDM(testCase)

	// Initialize template context if templates are defined
	if testCase.Templates != nil && len(testCase.Templates) > 0 {
		actions.SetTemplateContext(testCase.Templates)
		if !silent {
			PrintTemplatesLoaded(len(testCase.Templates), getTemplateNames(testCase.Templates))
		}
	}

	// Create action executor
	executor := actions.NewActionExecutor()

	result := &parser.TestResult{
		TestCase:    testCase,
		Status:      parser.StatusPending,
		StepResults: make([]parser.StepResult, 0),
		DataResults: &parser.DataResults{
			Validations: make([]parser.ValidationResult, 0),
			DataSets:    make(map[string]parser.DataSetInfo),
		},
	}

	startTime := time.Now()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	defer func() {
		w.Close()
		os.Stdout = oldStdout // Restore stdout
		result.CapturedOutput = <-outC
	}()

	// Print test case start information
	if !silent {
		PrintTestCaseStart(testCase.Name)
		if testCase.Description != "" {
			PrintTestCaseDescription(testCase.Description)
		}
		PrintTestCaseSteps(len(testCase.Steps))
	}

	// Execute TDM setup if configured
	if testCase.DataManagement != nil && len(testCase.DataManagement.Setup) > 0 {
		if !silent {
			PrintTDMSetup()
		}
		executeSteps(engine.runner, testCase.DataManagement.Setup, executor, nil, silent, &result.StepResults, "TDM Setup: ", testCase)
		result.DataResults.SetupStatus = "COMPLETED"
	}

	// Execute main test steps
	var err error
	err = executeStepsWithConfig(engine.runner, testCase.Steps, executor, nil, silent, &result.StepResults, "", testCase, testCase.Parallel)
	if err != nil {
		if actions.IsSkipError(err) {
			result.Status = parser.StatusSkipped
			// Set skip reason from the first skipped step
			for _, sr := range result.StepResults {
				if sr.Status == parser.StatusSkipped {
					result.ErrorMessage = sr.Error
					break
				}
			}
			// Skip TDM teardown for skipped tests
			engine.calculateTestResults(result, startTime)
			return result, nil
		}
		// Error occurred during step execution (not skip)
	}

	// Execute TDM teardown if configured
	if testCase.DataManagement != nil && len(testCase.DataManagement.Teardown) > 0 {
		if !silent {
			PrintTDMTeardown()
		}
		executeSteps(engine.runner, testCase.DataManagement.Teardown, executor, nil, silent, &result.StepResults, "TDM Teardown: ", testCase)
		result.DataResults.TeardownStatus = "COMPLETED"
	}

	// Calculate test results
	engine.calculateTestResults(result, startTime)

	// Return appropriate error based on test status
	return engine.determineReturnValue(result)
}

// calculateTestResults calculates the final test results and status
func (engine *ExecutionEngine) calculateTestResults(result *parser.TestResult, startTime time.Time) {
	result.Duration = time.Since(startTime)
	result.TotalSteps = len(result.StepResults)
	result.PassedSteps = 0
	result.FailedSteps = 0
	result.SkippedSteps = 0

	for _, sr := range result.StepResults {
		switch sr.Status {
		case parser.StatusPassed:
			result.PassedSteps++
		case parser.StatusFailed:
			result.FailedSteps++
		case parser.StatusSkipped:
			result.SkippedSteps++
		}
	}

	// Determine test case status
	if result.FailedSteps > 0 {
		result.Status = parser.StatusFailed
		// Only set ErrorMessage if a non-continue-on-failure step failed
		for _, sr := range result.StepResults {
			if sr.Status == parser.StatusFailed && !sr.Step.ContinueOnFailure {
				if sr.Error != "" {
					result.ErrorMessage = sr.Error
				} else {
					result.ErrorMessage = "Test failed due to step failure."
				}
				break
			}
		}
	} else if result.SkippedSteps == result.TotalSteps {
		// Only mark as SKIPPED if ALL steps were skipped
		result.Status = parser.StatusSkipped
		// Set error message from the first skipped step
		for _, sr := range result.StepResults {
			if sr.Status == parser.StatusSkipped {
				result.ErrorMessage = sr.Error
				break
			}
		}
	} else {
		result.Status = parser.StatusPassed
	}
}

// determineReturnValue determines what to return based on test status
func (engine *ExecutionEngine) determineReturnValue(result *parser.TestResult) (*parser.TestResult, error) {
	// Only return error if a non-continue-on-failure step failed
	if result.Status == parser.StatusFailed && result.ErrorMessage != "" {
		return result, fmt.Errorf(result.ErrorMessage)
	}
	// Don't return error for skipped tests
	if result.Status == parser.StatusSkipped {
		return result, nil
	}
	return result, nil
}
