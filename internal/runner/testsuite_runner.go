package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/JianLoong/robogo/internal/actions"
	"github.com/JianLoong/robogo/internal/parser"
	"github.com/JianLoong/robogo/internal/util"
)

// TestSuiteRunner runs test suites
type TestSuiteRunner struct {
	runner       *TestRunner
	setupResults map[string]interface{} // Store setup step results
}

// NewTestSuiteRunner creates a new test suite runner
func NewTestSuiteRunner(runner *TestRunner) *TestSuiteRunner {
	return &TestSuiteRunner{
		runner:       runner,
		setupResults: make(map[string]interface{}),
	}
}

// RunTestSuite executes a test suite
func (tsr *TestSuiteRunner) RunTestSuite(testSuite *parser.TestSuite, suiteFilePath string, printSummary bool) (*parser.TestSuiteResult, error) {
	startTime := time.Now()
	result := &parser.TestSuiteResult{
		TestSuite: testSuite,
		Status:    "running",
	}

	defer func() {
		if printSummary {
			tsr.printSuiteSummary(result)
			// Ensure all output is flushed
			os.Stdout.Sync()
		}
	}()

	fmt.Printf("Starting test suite: %s\n", testSuite.Name)
	if testSuite.Description != "" {
		fmt.Printf("Description: %s\n", testSuite.Description)
	}

	// Load test cases first
	testCases, err := parser.LoadTestCases(testSuite, suiteFilePath)
	if err != nil {
		result.Status = "failed"
		result.ErrorMessage = fmt.Sprintf("Failed to load test cases: %v", err)
		result.Duration = time.Since(startTime)
		return result, err
	}

	result.TotalCases = len(testCases)

	// Run setup if present
	if len(testSuite.Setup) > 0 {
		fmt.Printf("Running suite setup (%d steps)...\n", len(testSuite.Setup))
		setupResult, err := tsr.runSetup(testSuite.Setup, testSuite.Variables)
		if err != nil {
			result.Status = "skipped"
			result.ErrorMessage = fmt.Sprintf("Setup failed: %v", err)
			result.Duration = time.Since(startTime)
			result.SetupStatus = "FAILED"

			// Mark all test cases as skipped due to setup failure (without running them)
			result.SkippedCases = len(testCases)
			result.CaseResults = make([]parser.TestCaseResult, len(testCases))
			for i, testCase := range testCases {
				result.CaseResults[i] = parser.TestCaseResult{
					TestCase: testCase,
					Status:   "skipped",
					Error:    fmt.Sprintf("Skipped due to setup failure: %v", err),
				}
			}

			// Skip teardown since no tests ran
			// tsr.printSuiteSummary(result) // This line is now handled by the defer block
			return result, nil // Return nil error since this is expected behavior
		}
		result.SetupStatus = setupResult.Status
	}

	// Run test cases (only if setup succeeded)
	if result.Status != "skipped" {
		fmt.Printf("Running %d test cases...\n", len(testCases))
		if testSuite.Parallel {
			result.CaseResults = tsr.runTestCasesParallel(testCases, testSuite.Variables, testSuite.FailFast)
		} else {
			result.CaseResults = tsr.runTestCasesSequential(testCases, testSuite.Variables, testSuite.FailFast)
		}
	} else {
		fmt.Printf("Skipping %d test cases due to setup failure\n", len(testCases))
	}

	// Calculate summary
	result.PassedCases = 0
	result.FailedCases = 0
	result.SkippedCases = 0
	result.TotalSteps = 0
	result.PassedSteps = 0
	result.FailedSteps = 0
	result.SkippedSteps = 0

	// If test cases were skipped due to setup failure, all are skipped
	if result.Status == "skipped" {
		result.SkippedCases = len(testCases)
	} else {
		// Calculate based on actual results and aggregate step results
		for _, caseResult := range result.CaseResults {
			switch caseResult.Status {
			case "passed":
				result.PassedCases++
			case "failed":
				result.FailedCases++
			case "skipped":
				result.SkippedCases++
			}

			// Aggregate step results from each test case
			if caseResult.Result != nil {
				result.TotalSteps += caseResult.Result.TotalSteps
				result.PassedSteps += caseResult.Result.PassedSteps
				result.FailedSteps += caseResult.Result.FailedSteps
				result.SkippedSteps += caseResult.Result.SkippedSteps
			}
		}
	}

	// TotalCases represents the total number of test cases in the suite
	// It's already set to len(testCases) at the beginning

	// Run teardown if present (only if tests were actually run)
	if len(testSuite.Teardown) > 0 && result.Status != "skipped" {
		fmt.Printf("Running suite teardown (%d steps)...\n", len(testSuite.Teardown))
		teardownResult, err := tsr.runTeardown(testSuite.Teardown, testSuite.Variables)
		if err != nil {
			fmt.Printf("Teardown failed: %v\n", err)
		}
		result.TeardownStatus = teardownResult.Status
	} else if result.Status == "skipped" {
		fmt.Printf("Skipping teardown due to setup failure\n")
		result.TeardownStatus = "SKIPPED"
	}

	// Set final status
	if result.FailedCases == 0 && result.SkippedCases == 0 {
		result.Status = "passed"
	} else if result.FailedCases > 0 {
		result.Status = "failed"
	} else {
		result.Status = "skipped"
	}

	result.Duration = time.Since(startTime)

	// Print summary only if requested
	// if printSummary { // This line is now handled by the defer block
	// 	tsr.printSuiteSummary(result)
	// }

	return result, nil
}

// runSetup executes suite setup steps and captures results
func (tsr *TestSuiteRunner) runSetup(steps []parser.Step, variables *parser.Variables) (*parser.TestResult, error) {
	// Create a temporary test case for setup
	setupTestCase := &parser.TestCase{
		Name:  "Suite Setup",
		Steps: steps,
	}
	if variables != nil {
		setupTestCase.Variables = *variables
	}

	// Run setup and capture results
	result, err := RunTestCase(setupTestCase, false)
	if err != nil {
		return result, err
	}

	// Extract result variables from setup steps
	tsr.extractSetupResults(result.StepResults)

	return result, nil
}

// runTeardown executes suite teardown steps with access to setup results
func (tsr *TestSuiteRunner) runTeardown(steps []parser.Step, variables *parser.Variables) (*parser.TestResult, error) {
	// Create a temporary test case for teardown
	teardownTestCase := &parser.TestCase{
		Name:  "Suite Teardown",
		Steps: steps,
	}

	// Merge setup results and suite variables for teardown
	mergedTestCase := tsr.mergeSuiteVariables(teardownTestCase, variables)

	return RunTestCase(mergedTestCase, false)
}

// runTestCasesSequential runs test cases one after another
func (tsr *TestSuiteRunner) runTestCasesSequential(testCases []*parser.TestCase, suiteVariables *parser.Variables, failFast bool) []parser.TestCaseResult {
	var results []parser.TestCaseResult

	failed := false
	failReason := ""

	for i, testCase := range testCases {
		// If fail-fast and a previous test failed, skip the rest
		if failFast && failed {
			skipReason := "Skipped due to fail-fast after failure: " + failReason
			fmt.Printf("\nTest case %d/%d: %s skipped: %s\n", i+1, len(testCases), testCase.Name, skipReason)
			caseResult := parser.TestCaseResult{
				TestCase: testCase,
				Status:   "skipped",
				Error:    skipReason,
			}
			results = append(results, caseResult)
			continue
		}
		// Check for skip at the test case level using unified logic
		skipInfo := tsr.runner.ShouldSkipTestCase(testCase, fmt.Sprintf("test case %d/%d", i+1, len(testCases)))
		if skipInfo.ShouldSkip {
			PrintSkipMessage(fmt.Sprintf("Test case %d/%d", i+1, len(testCases)), testCase.Name, skipInfo.Reason, false)
			caseResult := CreateSkipTestCaseResult(testCase, skipInfo.Reason)
			results = append(results, caseResult)
			continue
		}

		fmt.Printf("\nRunning test case %d/%d: %s\n", i+1, len(testCases), testCase.Name)

		// Merge suite variables with test case variables
		mergedTestCase := tsr.mergeSuiteVariables(testCase, suiteVariables)

		// Log what variables are being passed
		fmt.Println("   Variables:")
		fmt.Printf("      %d regular, %d secrets\n", len(mergedTestCase.Variables.Regular), len(mergedTestCase.Variables.Secrets))
		fmt.Println("   Final merged variables for this test case:")
		for k, v := range mergedTestCase.Variables.Regular {
			fmt.Printf("      %s: %v\n", k, v)
		}

		testResult, err := RunTestCase(mergedTestCase, false)
		duration := time.Since(time.Now())

		caseResult := parser.TestCaseResult{
			TestCase: mergedTestCase,
			Result:   testResult,
			Duration: duration,
		}

		if err != nil {
			// Check if this is a skip error - this should be treated as a skipped test case
			if actions.IsSkipError(err) || (testResult != nil && strings.ToLower(testResult.Status) == "skipped") {
				caseResult.Status = "skipped"
				caseResult.Error = testResult.ErrorMessage
				fmt.Printf("Test case skipped: %s\n", testCase.Name)
			} else {
				caseResult.Status = "failed"
				caseResult.Error = util.FormatRobogoError(err)
				fmt.Printf("Test case failed: %v\n", err)
			}
			// After running the test case, print all step statuses and errors
			if caseResult.Status == "failed" {
				failed = true
				failReason = caseResult.Error
			}
		} else {
			// Convert status to lowercase for consistency
			caseResult.Status = strings.ToLower(testResult.Status)
			if caseResult.Status == "passed" {
				fmt.Printf("Test case passed in %v\n", duration)
			} else if caseResult.Status == "skipped" {
				fmt.Printf("Test case skipped in %v\n", duration)
				if testResult.ErrorMessage != "" {
					fmt.Printf("   Reason: %s\n", testResult.ErrorMessage)
				}
			} else {
				fmt.Printf("Test case failed in %v\n", duration)
				if testResult.ErrorMessage != "" {
					fmt.Printf("   Error: %s\n", testResult.ErrorMessage)
				}
				// After running the test case, print all step statuses and errors
			}
		}

		// Check for failures (both error and failed status) for fail-fast
		if caseResult.Status == "failed" {
			failed = true
			failReason = caseResult.Error
		}

		results = append(results, caseResult)
	}

	return results
}

// runTestCasesParallel runs test cases in parallel
func (tsr *TestSuiteRunner) runTestCasesParallel(testCases []*parser.TestCase, suiteVariables *parser.Variables, failFast bool) []parser.TestCaseResult {
	var results []parser.TestCaseResult
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Create channels for results
	resultChan := make(chan parser.TestCaseResult, len(testCases))

	// Note: fail-fast is not supported in parallel mode. All test cases start at once.
	// The following check is intentionally omitted in parallel execution:
	// if caseResult.Status == "failed" { ... }

	// Start goroutines for each test case
	for i, testCase := range testCases {
		wg.Add(1)
		go func(index int, tc *parser.TestCase) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Panic in test case %s: %v\n", tc.Name, r)
				}
			}()
			defer wg.Done()

			var caseResult parser.TestCaseResult

			// Check for skip at the test case level using unified logic
			skipInfo := tsr.runner.ShouldSkipTestCase(tc, fmt.Sprintf("test case %d/%d", index+1, len(testCases)))
			if skipInfo.ShouldSkip {
				PrintSkipMessage(fmt.Sprintf("Test case %d/%d", index+1, len(testCases)), tc.Name, skipInfo.Reason, false)
				caseResult = CreateSkipTestCaseResult(tc, skipInfo.Reason)
				select {
				case resultChan <- caseResult:
				case <-time.After(5 * time.Second):
					fmt.Printf("Timeout sending skip result for test case: %s\n", tc.Name)
				}
				return
			}

			fmt.Printf("\nRunning test case %d/%d: %s\n", index+1, len(testCases), tc.Name)

			// Merge suite variables with test case variables
			mergedTestCase := tsr.mergeSuiteVariables(tc, suiteVariables)

			// Log what variables are being passed
			fmt.Println("   Variables:")
			fmt.Printf("      %d regular, %d secrets\n", len(mergedTestCase.Variables.Regular), len(mergedTestCase.Variables.Secrets))
			fmt.Println("   Final merged variables for this test case:")
			for k, v := range mergedTestCase.Variables.Regular {
				fmt.Printf("      %s: %v\n", k, v)
			}

			// Use silent=true for parallel execution to avoid stdout capture deadlocks
			testResult, err := RunTestCase(mergedTestCase, true)
			duration := time.Since(time.Now())

			caseResult = parser.TestCaseResult{
				TestCase: mergedTestCase,
				Result:   testResult,
				Duration: duration,
			}

			if err != nil {
				// Check if this is a skip error - this should be treated as a skipped test case
				if actions.IsSkipError(err) || (testResult != nil && strings.ToLower(testResult.Status) == "skipped") {
					caseResult.Status = "skipped"
					caseResult.Error = testResult.ErrorMessage
					fmt.Printf("Test case skipped: %s\n", tc.Name)
				} else {
					caseResult.Status = "failed"
					caseResult.Error = util.FormatRobogoError(err)
					fmt.Printf("Test case failed: %v\n", err)
				}
			} else {
				// Convert status to lowercase for consistency
				caseResult.Status = strings.ToLower(testResult.Status)
				if caseResult.Status == "passed" {
					fmt.Printf("Test case passed in %v\n", duration)
				} else if caseResult.Status == "skipped" {
					fmt.Printf("Test case skipped in %v\n", duration)
					if testResult.ErrorMessage != "" {
						fmt.Printf("   Reason: %s\n", testResult.ErrorMessage)
					}
				} else {
					fmt.Printf("Test case failed in %v\n", duration)
					if testResult.ErrorMessage != "" {
						fmt.Printf("   Error: %s\n", testResult.ErrorMessage)
					}
				}
			}

			// Send result with timeout protection
			select {
			case resultChan <- caseResult:
			case <-time.After(5 * time.Second):
				fmt.Printf("Timeout sending result for test case: %s\n", tc.Name)
			}
		}(i, testCase)
	}

	// Collect results with timeout protection
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results - this will block until all goroutines are done or timeout
	expectedResults := len(testCases)
	for {
		select {
		case result, ok := <-resultChan:
			if !ok {
				// Channel closed, all results collected
				return results
			}
			mu.Lock()
			results = append(results, result)
			currentCount := len(results)
			mu.Unlock()

			// If we have all expected results, return immediately
			if currentCount >= expectedResults {
				return results
			}
		case <-ctx.Done():
			fmt.Printf("Timeout waiting for parallel test execution to complete (got %d/%d results)\n", len(results), expectedResults)
			// Return partial results and let the test suite continue
			return results
		}
	}
}

// mergeSuiteVariables merges suite variables and setup results with test case variables
func (tsr *TestSuiteRunner) mergeSuiteVariables(testCase *parser.TestCase, suiteVariables *parser.Variables) *parser.TestCase {
	// Create a copy of the test case
	merged := *testCase

	// Initialize variables if not present
	if merged.Variables.Regular == nil {
		merged.Variables.Regular = make(map[string]interface{})
	}
	if merged.Variables.Secrets == nil {
		merged.Variables.Secrets = make(map[string]parser.Secret)
	}

	// Create a new variables map with correct precedence
	// Order: setup results (lowest) -> suite variables (medium) -> test case variables (highest)
	finalVars := make(map[string]interface{})
	finalSecrets := make(map[string]parser.Secret)

	// 1. Start with setup results (lowest precedence)
	for k, v := range tsr.setupResults {
		finalVars[k] = v
	}

	// 2. Override with suite variables (medium precedence)
	if suiteVariables != nil {
		for k, v := range suiteVariables.Regular {
			finalVars[k] = v
		}
		for k, v := range suiteVariables.Secrets {
			finalSecrets[k] = v
		}
	}

	// 3. Override with test case variables (highest precedence)
	for k, v := range merged.Variables.Regular {
		finalVars[k] = v
	}
	for k, v := range merged.Variables.Secrets {
		finalSecrets[k] = v
	}

	// Set the final merged variables
	merged.Variables.Regular = finalVars
	merged.Variables.Secrets = finalSecrets

	return &merged
}

// printSuiteSummary prints a summary of the test suite execution
func (tsr *TestSuiteRunner) printSuiteSummary(result *parser.TestSuiteResult) {
	fmt.Printf("\n" + strings.Repeat("=", 60) + "\n")
	fmt.Printf("Test Suite Summary: %s\n", result.TestSuite.Name)
	fmt.Printf("Duration: %v\n", result.Duration)
	fmt.Printf("\n## Test Case Summary\n")
	fmt.Printf("| %-4s | %-24s | %-8s | %-10s | %-24s |\n", "#", "Name", "Status", "Duration", "Error")
	fmt.Printf("|------|--------------------------|----------|------------|--------------------------|\n")
	for i, caseResult := range result.CaseResults {
		icon := ""
		status := ""
		switch caseResult.Status {
		case "passed":
			icon = "PASSED"
			status = "PASSED"
		case "failed":
			icon = "FAILED"
			status = "FAILED"
		case "skipped":
			icon = "SKIPPED"
			status = "SKIPPED"
		}
		duration := ""
		if caseResult.Duration > 0 {
			duration = fmt.Sprintf("%.4gs", caseResult.Duration.Seconds())
		}
		error := caseResult.Error
		if len(error) > 24 {
			error = error[:21] + "..."
		}
		name := caseResult.TestCase.Name
		if len(name) > 24 {
			name = name[:21] + "..."
		}
		fmt.Printf("| %-4d | %-24s | %-8s | %-10s | %-24s |\n", i+1, name, icon+" "+status, duration, error)
	}

	// Print step results for each test case
	for _, caseResult := range result.CaseResults {
		if caseResult.Result != nil && len(caseResult.Result.StepResults) > 0 {
			title := "### Step Results for " + caseResult.TestCase.Name
			PrintStepResultsMarkdown(caseResult.Result.StepResults, title)
		}
	}

	// Print step summary table
	fmt.Printf("\n## Step Summary\n")
	fmt.Printf("| %-8s | %-8s | %-8s | %-8s |\n", "Total", "Passed", "Failed", "Skipped")
	fmt.Printf("|----------|----------|----------|----------|\n")
	fmt.Printf("| %-8d | %-8d | %-8d | %-8d |\n", result.TotalSteps, result.PassedSteps, result.FailedSteps, result.SkippedSteps)

	if result.SetupStatus != "" {
		fmt.Printf("\nSetup: %s\n", result.SetupStatus)
	}
	if result.TeardownStatus != "" {
		fmt.Printf("\nTeardown: %s\n", result.TeardownStatus)
	}

	if result.FailedCases > 0 {
		// fmt.Printf("\n❌ Failed Test Cases:\n")
		// for _, caseResult := range result.CaseResults {
		// 	if caseResult.Status == "failed" {
		// 		fmt.Printf("  - %s: %s\n", caseResult.TestCase.Name, caseResult.Error)
		// 	}
		// }
	}

	if result.SkippedCases > 0 {
		// fmt.Printf("\n⏭️  Skipped Test Cases:\n")
		// for _, caseResult := range result.CaseResults {
		// 	if caseResult.Status == "skipped" {
		// 		fmt.Printf("  - %s: %s\n", caseResult.TestCase.Name, caseResult.Error)
		// 	}
		// }
	}

	switch result.Status {
	case "passed":
		fmt.Printf("\nTest suite completed successfully!\n")
	case "failed":
		fmt.Printf("\nTest suite failed!\n")
	case "skipped":
		fmt.Printf("\nTest suite skipped due to setup failure!\n")
	}
	fmt.Printf(strings.Repeat("=", 60) + "\n")
}

// extractSetupResults extracts result variables from setup step results
func (tsr *TestSuiteRunner) extractSetupResults(stepResults []parser.StepResult) {
	for _, stepResult := range stepResults {
		if stepResult.Step.Result != "" && stepResult.Output != "" {
			// Try to parse the output as JSON to get the actual result
			var resultData map[string]interface{}
			if err := json.Unmarshal([]byte(stepResult.Output), &resultData); err == nil {
				// If it's a JSON object, store the whole thing
				tsr.setupResults[stepResult.Step.Result] = resultData
			} else {
				// If it's not JSON, store the raw output
				tsr.setupResults[stepResult.Step.Result] = stepResult.Output
			}
		}
	}
}
