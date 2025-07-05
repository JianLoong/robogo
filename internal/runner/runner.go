package runner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JianLoong/robogo/internal/actions"
	"github.com/JianLoong/robogo/internal/parser"
)

// TestRunner runs test cases
type TestRunner struct {
	variables     map[string]interface{} // Store variables for the test case
	secretManager *actions.SecretManager // Secret manager for handling secrets
	tdmManager    *actions.TDMManager    // TDM manager for test data management
}

// NewTestRunner creates a new test runner
func NewTestRunner() *TestRunner {
	return &TestRunner{
		variables:     make(map[string]interface{}),
		secretManager: actions.NewSecretManager(),
		tdmManager:    actions.NewTDMManager(),
	}
}

// RunTestFiles runs multiple test cases in parallel
func RunTestFiles(paths []string, silent bool) ([]*parser.TestResult, error) {
	return RunTestFilesWithConfig(paths, silent, nil)
}

// RunTestFilesWithConfig runs multiple test cases with parallelism configuration
func RunTestFilesWithConfig(paths []string, silent bool, parallelConfig *parser.ParallelConfig) ([]*parser.TestResult, error) {
	var files []string
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("failed to stat path %s: %w", path, err)
		}

		if info.IsDir() {
			filepath.Walk(path, func(p string, i os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !i.IsDir() && (strings.HasSuffix(p, ".robogo") || strings.HasSuffix(p, ".yml") || strings.HasSuffix(p, ".yaml")) {
					files = append(files, p)
				}
				return nil
			})
		} else {
			files = append(files, path)
		}
	}

	// Merge and validate parallelism configuration
	config := parser.MergeParallelConfig(parallelConfig)
	if err := parser.ValidateParallelConfig(config); err != nil {
		return nil, fmt.Errorf("invalid parallelism configuration: %w", err)
	}

	// If parallelism is disabled, run sequentially
	if !config.Enabled || !config.TestCases {
		var results []*parser.TestResult
		for _, file := range files {
			result, err := RunTestFile(file, silent)
			if err != nil {
				fmt.Fprintf(os.Stderr, "DEBUG RunTestFilesWithConfig: Creating error result for %s\n", file)
				result = &parser.TestResult{
					TestCase:     &parser.TestCase{Name: file},
					Status:       "FAILED",
					ErrorMessage: err.Error(),
				}
			}
			fmt.Fprintf(os.Stderr, "DEBUG RunTestFilesWithConfig: Sequential result for %s, TestCase address: %p, Steps length: %d\n",
				file, result.TestCase, len(result.TestCase.Steps))
			results = append(results, result)
		}
		return results, nil
	}

	// Run in parallel with concurrency control
	var wg sync.WaitGroup
	resultsChan := make(chan *parser.TestResult, len(files))
	semaphore := make(chan struct{}, config.MaxConcurrency)

	if !silent {
		fmt.Printf("🚀 Running %d test files in parallel (max concurrency: %d)\n", len(files), config.MaxConcurrency)
	}

	for _, file := range files {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result, err := RunTestFile(file, silent)
			if err != nil {
				fmt.Fprintf(os.Stderr, "DEBUG RunTestFilesWithConfig: Creating parallel error result for %s\n", file)
				// In case of a fatal error before the test can even run properly,
				// create a dummy result to report the failure.
				result = &parser.TestResult{
					TestCase:     &parser.TestCase{Name: file},
					Status:       "FAILED",
					ErrorMessage: err.Error(),
				}
			}
			fmt.Fprintf(os.Stderr, "DEBUG RunTestFilesWithConfig: Parallel result for %s, TestCase address: %p, Steps length: %d\n",
				file, result.TestCase, len(result.TestCase.Steps))
			resultsChan <- result
		}(file)
	}

	wg.Wait()
	close(resultsChan)

	var results []*parser.TestResult
	for result := range resultsChan {
		results = append(results, result)
	}

	return results, nil
}

// RunTestFile runs a test case from a file
func RunTestFile(filename string, silent bool) (*parser.TestResult, error) {
	// Parse the test case
	testCase, err := parser.ParseTestFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to parse test file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "DEBUG RunTestFile: After parsing, testCase.Steps length: %d\n", len(testCase.Steps))
	fmt.Fprintf(os.Stderr, "DEBUG RunTestFile: testCase address: %p\n", testCase)

	// Run the test case
	result, err := RunTestCase(testCase, silent)
	if err != nil {
		// RunTestCase returns an error when the test fails, but we still want to return the result
		// The error just indicates test failure, not a fatal error
		fmt.Fprintf(os.Stderr, "DEBUG RunTestFile: RunTestCase returned error (test failure): %v\n", err)
		fmt.Fprintf(os.Stderr, "DEBUG RunTestFile: But returning the result anyway, TestCase address: %p, Steps length: %d\n",
			result.TestCase, len(result.TestCase.Steps))
		return result, nil // Return the result, not the error
	}
	return result, nil
}

// RunTestCase runs a test case and returns the result
func RunTestCase(testCase *parser.TestCase, silent bool) (*parser.TestResult, error) {
	fmt.Fprintf(os.Stderr, "DEBUG RunTestCase: At start, testCase.Steps length: %d\n", len(testCase.Steps))
	fmt.Fprintf(os.Stderr, "DEBUG RunTestCase: testCase address: %p\n", testCase)

	tr := NewTestRunner()
	tr.initializeVariables(testCase)
	tr.initializeTDM(testCase)

	// Create action executor
	executor := actions.NewActionExecutor()

	result := &parser.TestResult{
		TestCase:    testCase,
		Status:      "PASSED",
		StepResults: make([]parser.StepResult, 0),
		DataResults: &parser.DataResults{
			Validations: make([]parser.ValidationResult, 0),
			DataSets:    make(map[string]parser.DataSetInfo),
		},
	}

	fmt.Fprintf(os.Stderr, "DEBUG RunTestCase: After creating result, result.TestCase.Steps length: %d\n", len(result.TestCase.Steps))
	fmt.Fprintf(os.Stderr, "DEBUG RunTestCase: result.TestCase address: %p\n", result.TestCase)

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

	if !silent {
		fmt.Printf("🚀 Running test case: %s\n", testCase.Name)
		if testCase.Description != "" {
			fmt.Printf("📋 Description: %s\n", testCase.Description)
		}
		fmt.Printf("📝 Steps: %d\n\n", len(testCase.Steps))
	}

	// Execute TDM setup if configured
	if testCase.DataManagement != nil && len(testCase.DataManagement.Setup) > 0 {
		if !silent {
			fmt.Printf("🔧 Executing TDM setup...\n")
		}
		tr.executeSteps(testCase.DataManagement.Setup, executor, nil, silent, &result.StepResults, "TDM Setup: ", testCase)
		result.DataResults.SetupStatus = "COMPLETED"
	}

	// Debug output before executing steps
	fmt.Println("DEBUG RUNNER: About to execute steps...")
	var err error
	err = tr.executeStepsWithConfig(testCase.Steps, executor, nil, silent, &result.StepResults, "", testCase, testCase.Parallel)
	fmt.Printf("DEBUG RUNNER: executeStepsWithConfig returned. err=%v\n", err)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DEBUG RunTestCase: Error after step execution: %v\n", err)
	}

	// Execute TDM teardown if configured
	if testCase.DataManagement != nil && len(testCase.DataManagement.Teardown) > 0 {
		if !silent {
			fmt.Printf("🧹 Executing TDM teardown...\n")
		}
		tr.executeSteps(testCase.DataManagement.Teardown, executor, nil, silent, &result.StepResults, "TDM Teardown: ", testCase)
		result.DataResults.TeardownStatus = "COMPLETED"
	}

	result.Duration = time.Since(startTime)
	for _, sr := range result.StepResults {
		if sr.Status == "FAILED" {
			result.FailedSteps++
		} else {
			result.PassedSteps++
		}
	}
	if result.FailedSteps > 0 {
		result.Status = "FAILED"
		// Only set ErrorMessage if a non-continue-on-failure step failed
		for _, sr := range result.StepResults {
			if sr.Status == "FAILED" && !sr.Step.ContinueOnFailure {
				if sr.Error != "" {
					result.ErrorMessage = sr.Error
				} else {
					result.ErrorMessage = "Test failed due to step failure."
				}
				break
			}
		}
	}

	if !silent {
		fmt.Printf("\n🏁 Test completed in %v\n", result.Duration)
		fmt.Printf("\n📊 Test Results:\n")
		fmt.Printf("✅ Status: %s\n", result.Status)
		fmt.Printf("⏱️  Duration: %v\n", result.Duration)
		fmt.Printf("📝 Steps: %d total, %d passed, %d failed\n", len(result.StepResults), result.PassedSteps, result.FailedSteps)
	}

	// Only return error if a non-continue-on-failure step failed
	if result.Status == "FAILED" && result.ErrorMessage != "" {
		fmt.Fprintf(os.Stderr, "DEBUG RunTestCase: StepResults length before return (FAILED): %d\n", len(result.StepResults))
		fmt.Fprintf(os.Stderr, "DEBUG RunTestCase: result.TestCase.Steps length before return (FAILED): %d\n", len(result.TestCase.Steps))
		for i, sr := range result.StepResults {
			if i >= 5 {
				break
			}
			fmt.Fprintf(os.Stderr, "DEBUG RunTestCase: Step %d name: %s\n", i+1, sr.Step.Name)
		}
		return result, fmt.Errorf(result.ErrorMessage)
	}
	fmt.Fprintf(os.Stderr, "DEBUG RunTestCase: StepResults length before return: %d\n", len(result.StepResults))
	fmt.Fprintf(os.Stderr, "DEBUG RunTestCase: result.TestCase.Steps length before return: %d\n", len(result.TestCase.Steps))
	for i, sr := range result.StepResults {
		if i >= 5 {
			break
		}
		fmt.Fprintf(os.Stderr, "DEBUG RunTestCase: Step %d name: %s\n", i+1, sr.Step.Name)
	}
	return result, nil
}

// executeSteps executes a slice of steps, collecting StepResults recursively
func (tr *TestRunner) executeSteps(steps []parser.Step, executor *actions.ActionExecutor, parentLoop *parser.LoopBlock, silent bool, stepResults *[]parser.StepResult, context string, testCase *parser.TestCase) error {
	return tr.executeStepsWithConfig(steps, executor, parentLoop, silent, stepResults, context, testCase, nil)
}

// executeStepsWithConfig executes steps with parallelism configuration
func (tr *TestRunner) executeStepsWithConfig(steps []parser.Step, executor *actions.ActionExecutor, parentLoop *parser.LoopBlock, silent bool, stepResults *[]parser.StepResult, context string, testCase *parser.TestCase, parallelConfig *parser.ParallelConfig) error {
	fmt.Fprintf(os.Stderr, "DEBUG executeStepsWithConfig: Starting with %d steps\n", len(steps))

	// Check if parallel step execution is enabled
	config := parser.MergeParallelConfig(parallelConfig)
	if config.Enabled && config.Steps && len(steps) > 1 {
		fmt.Fprintln(os.Stderr, "DEBUG executeStepsWithConfig: Using parallel execution")
		return tr.executeStepsParallel(steps, executor, parentLoop, silent, stepResults, context, testCase, config)
	}

	// Execute steps sequentially (original behavior)
	fmt.Fprintln(os.Stderr, "DEBUG executeStepsWithConfig: Using sequential execution")
	for idx, step := range steps {
		fmt.Fprintf(os.Stderr, "DEBUG executeStepsWithConfig: Processing step %d: %s\n", idx+1, step.Action)

		stepContext := context
		if parentLoop != nil {
			iteration := tr.variables["iteration"]
			stepContext = context + fmt.Sprintf("Iteration[%v]: ", iteration)
		}

		if step.If != nil {
			fmt.Fprintf(os.Stderr, "DEBUG executeStepsWithConfig: Step %d is an if statement\n", idx+1)
			if err := tr.executeIfStatement(step.If, executor, silent, stepResults, stepContext+step.Name+"/If: ", testCase); err != nil {
				return err
			}
			continue
		}
		if step.For != nil {
			fmt.Fprintf(os.Stderr, "DEBUG executeStepsWithConfig: Step %d is a for loop\n", idx+1)
			if err := tr.executeForLoop(step.For, executor, silent, stepResults, stepContext+step.Name+"/For: ", testCase); err != nil {
				return err
			}
			continue
		}
		if step.While != nil {
			fmt.Fprintf(os.Stderr, "DEBUG executeStepsWithConfig: Step %d is a while loop\n", idx+1)
			if err := tr.executeWhileLoop(step.While, executor, silent, stepResults, stepContext+step.Name+"/While: ", testCase); err != nil {
				return err
			}
			continue
		}

		fmt.Fprintf(os.Stderr, "DEBUG executeStepsWithConfig: Step %d is a regular action: %s\n", idx+1, step.Action)
		stepStart := time.Now()
		stepLabel := step.Name
		if stepLabel == "" {
			stepLabel = step.Action
		}
		if !silent {
			fmt.Printf("Step %d: %s\n", len(*stepResults)+1, stepLabel)
		}

		substitutedArgs := tr.substituteVariables(step.Args)
		output, err := tr.executeStepWithRetry(step, substitutedArgs, executor, silent)
		stepDuration := time.Since(stepStart)

		// Get verbosity level for this step
		verbosityLevel := parser.GetVerbosityLevel(&step, testCase)

		// Mask secrets in output for display
		maskedOutput := tr.secretManager.MaskSecretsInString(output)

		// Format verbose output if enabled
		verboseOutput := parser.FormatVerboseOutput(verbosityLevel, step.Action, substitutedArgs, maskedOutput, stepDuration.String())

		stepResult := parser.StepResult{
			Step:      step,
			Status:    "PASSED",
			Duration:  stepDuration,
			Output:    maskedOutput,
			Timestamp: time.Now(),
		}

		// Handle expect_error property
		if step.ExpectError != nil {
			expectErr := tr.validateExpectedError(step.ExpectError, err, output, silent)
			if expectErr != nil {
				stepResult.Status = "FAILED"
				stepResult.Error = expectErr.Error()
				if !silent {
					fmt.Printf("❌ Step %d failed: %s\n", len(*stepResults)+1, expectErr.Error())
				}
			} else {
				if !silent {
					fmt.Printf("✅ Error expectation passed\n")
				}
			}
		} else if err != nil {
			// Normal error handling (no expect_error)
			stepResult.Status = "FAILED"
			stepResult.Error = err.Error()
			if !silent {
				fmt.Printf("❌ Step %d failed: %s\n", len(*stepResults)+1, err.Error())
			}
		} else {
			if !silent {
				// Display verbose output if enabled
				if verboseOutput != "" {
					fmt.Print(verboseOutput)
				} else {
					// For log actions, mask the message before displaying
					if step.Action == "log" && len(step.Args) > 0 {
						message := fmt.Sprintf("%v", substitutedArgs[0])
						maskedMessage := tr.secretManager.MaskSecretsInString(message)
						fmt.Printf("📝 %s\n", maskedMessage)
					} else {
						fmt.Printf("✅ Step %d completed in %v\n", len(*stepResults)+1, stepDuration)
					}
				}
			}
		}

		// Add context to step name for reporting
		if step.Name != "" {
			stepResult.Step.Name = stepContext + step.Name
		} else if stepContext != "" {
			stepResult.Step.Name = stepContext + fmt.Sprintf("Step%d", idx+1)
		}

		fmt.Fprintf(os.Stderr, "DEBUG executeStepsWithConfig: Adding step result for step %d\n", idx+1)
		*stepResults = append(*stepResults, stepResult)

		if err != nil {
			if step.ContinueOnFailure {
				// Log and continue to next step
				if !silent {
					fmt.Printf("⚠️  Step '%s' failed but continuing due to continue_on_failure\n", stepResult.Step.Name)
				}
				continue
			} else {
				return fmt.Errorf("step '%s' failed: %w", stepResult.Step.Name, err)
			}
		}

		// Store result in variable if specified (store actual value, not masked)
		if step.Result != "" {
			tr.variables[step.Result] = output // Store actual output
			if !silent {
				// Display masked version
				maskedValue := tr.secretManager.MaskSecretsInString(fmt.Sprintf("%v", output))
				fmt.Printf("💾 Stored result in variable: %s = %s\n", step.Result, maskedValue)
			}
		}
	}
	fmt.Fprintf(os.Stderr, "DEBUG executeStepsWithConfig: Finished processing all %d steps\n", len(steps))
	return nil
}

// executeStepsParallel executes independent steps in parallel
func (tr *TestRunner) executeStepsParallel(steps []parser.Step, executor *actions.ActionExecutor, parentLoop *parser.LoopBlock, silent bool, stepResults *[]parser.StepResult, context string, testCase *parser.TestCase, config *parser.ParallelConfig) error {
	// Group steps into parallel and sequential groups
	stepGroups := parser.GroupIndependentSteps(steps)

	if !silent {
		fmt.Printf("📊 Executing %d step groups (parallel execution enabled)\n", len(stepGroups))
	}

	for groupIdx, group := range stepGroups {
		if len(group) == 1 {
			// Single step - execute sequentially
			step := group[0]
			stepResult, err := tr.executeSingleStep(step, executor, parentLoop, silent, stepResults, context, testCase, groupIdx)
			if err != nil {
				return err
			}
			*stepResults = append(*stepResults, *stepResult)
		} else {
			// Multiple independent steps - execute in parallel
			if !silent {
				fmt.Printf("⚡ Executing %d steps in parallel (group %d)\n", len(group), groupIdx+1)
			}

			if err := tr.executeStepGroupParallel(group, executor, parentLoop, silent, stepResults, context, testCase, config, groupIdx); err != nil {
				return err
			}
		}
	}

	return nil
}

// executeStepGroupParallel executes a group of independent steps in parallel
func (tr *TestRunner) executeStepGroupParallel(steps []parser.Step, executor *actions.ActionExecutor, parentLoop *parser.LoopBlock, silent bool, stepResults *[]parser.StepResult, context string, testCase *parser.TestCase, config *parser.ParallelConfig, groupIdx int) error {
	var wg sync.WaitGroup
	resultsChan := make(chan *parser.StepResult, len(steps))
	errorsChan := make(chan error, len(steps))
	semaphore := make(chan struct{}, config.MaxConcurrency)

	// Execute steps in parallel
	for i, step := range steps {
		wg.Add(1)
		go func(stepIdx int, step parser.Step) {
			defer wg.Done()

			// Acquire semaphore for concurrency control
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			stepResult, err := tr.executeSingleStep(step, executor, parentLoop, silent, stepResults, context, testCase, groupIdx)
			if err != nil {
				errorsChan <- fmt.Errorf("step %d failed: %w", stepIdx+1, err)
				return
			}
			resultsChan <- stepResult
		}(i, step)
	}

	// Wait for all steps to complete
	wg.Wait()
	close(resultsChan)
	close(errorsChan)

	// Check for errors
	select {
	case err := <-errorsChan:
		return err
	default:
		// No errors, collect results
		for result := range resultsChan {
			*stepResults = append(*stepResults, *result)
		}
	}

	return nil
}

// executeSingleStep executes a single step and returns the result
func (tr *TestRunner) executeSingleStep(step parser.Step, executor *actions.ActionExecutor, parentLoop *parser.LoopBlock, silent bool, stepResults *[]parser.StepResult, context string, testCase *parser.TestCase, groupIdx int) (*parser.StepResult, error) {
	stepContext := context
	if parentLoop != nil {
		iteration := tr.variables["iteration"]
		stepContext = context + fmt.Sprintf("Iteration[%v]: ", iteration)
	}

	if step.If != nil {
		if err := tr.executeIfStatement(step.If, executor, silent, stepResults, stepContext+step.Name+"/If: ", testCase); err != nil {
			return nil, err
		}
		return &parser.StepResult{Step: step, Status: "PASSED"}, nil
	}
	if step.For != nil {
		if err := tr.executeForLoop(step.For, executor, silent, stepResults, stepContext+step.Name+"/For: ", testCase); err != nil {
			return nil, err
		}
		return &parser.StepResult{Step: step, Status: "PASSED"}, nil
	}
	if step.While != nil {
		if err := tr.executeWhileLoop(step.While, executor, silent, stepResults, stepContext+step.Name+"/While: ", testCase); err != nil {
			return nil, err
		}
		return &parser.StepResult{Step: step, Status: "PASSED"}, nil
	}

	stepStart := time.Now()
	stepLabel := step.Name
	if stepLabel == "" {
		stepLabel = step.Action
	}
	if !silent {
		fmt.Printf("Step %d: %s\n", len(*stepResults)+1, stepLabel)
	}

	substitutedArgs := tr.substituteVariables(step.Args)
	output, err := tr.executeStepWithRetry(step, substitutedArgs, executor, silent)
	stepDuration := time.Since(stepStart)

	// Get verbosity level for this step
	verbosityLevel := parser.GetVerbosityLevel(&step, testCase)

	// Mask secrets in output for display
	maskedOutput := tr.secretManager.MaskSecretsInString(output)

	// Format verbose output if enabled
	verboseOutput := parser.FormatVerboseOutput(verbosityLevel, step.Action, substitutedArgs, maskedOutput, stepDuration.String())

	stepResult := parser.StepResult{
		Step:      step,
		Status:    "PASSED",
		Duration:  stepDuration,
		Output:    maskedOutput,
		Timestamp: time.Now(),
	}

	// Handle expect_error property
	if step.ExpectError != nil {
		expectErr := tr.validateExpectedError(step.ExpectError, err, output, silent)
		if expectErr != nil {
			stepResult.Status = "FAILED"
			stepResult.Error = expectErr.Error()
			if !silent {
				fmt.Printf("❌ Step %d failed: %s\n", len(*stepResults)+1, expectErr.Error())
			}
		} else {
			if !silent {
				fmt.Printf("✅ Error expectation passed\n")
			}
		}
	} else if err != nil {
		// Normal error handling (no expect_error)
		stepResult.Status = "FAILED"
		stepResult.Error = err.Error()
		if !silent {
			fmt.Printf("❌ Step %d failed: %s\n", len(*stepResults)+1, err.Error())
		}
	} else {
		if !silent {
			// Display verbose output if enabled
			if verboseOutput != "" {
				fmt.Print(verboseOutput)
			} else {
				// For log actions, mask the message before displaying
				if step.Action == "log" && len(step.Args) > 0 {
					message := fmt.Sprintf("%v", substitutedArgs[0])
					maskedMessage := tr.secretManager.MaskSecretsInString(message)
					fmt.Printf("📝 %s\n", maskedMessage)
				} else {
					fmt.Printf("✅ Step %d completed in %v\n", len(*stepResults)+1, stepDuration)
				}
			}
		}
	}

	// Add context to step name for reporting
	if step.Name != "" {
		stepResult.Step.Name = stepContext + step.Name
	} else if stepContext != "" {
		stepResult.Step.Name = stepContext + fmt.Sprintf("Step%d", groupIdx+1)
	}

	// Store result in variable if specified (store actual value, not masked)
	if step.Result != "" {
		tr.variables[step.Result] = output // Store actual output
		if !silent {
			// Display masked version
			maskedValue := tr.secretManager.MaskSecretsInString(fmt.Sprintf("%v", output))
			fmt.Printf("💾 Stored result in variable: %s = %s\n", step.Result, maskedValue)
		}
	}

	return &stepResult, nil
}

// executeIfStatement executes an if/else block, collecting StepResults
func (tr *TestRunner) executeIfStatement(ifBlock *parser.ConditionalBlock, executor *actions.ActionExecutor, silent bool, stepResults *[]parser.StepResult, context string, testCase *parser.TestCase) error {
	condition := tr.substituteString(ifBlock.Condition)
	output, err := executor.Execute("control", []interface{}{"if", condition}, silent)
	if err != nil {
		return fmt.Errorf("failed to evaluate if condition: %w", err)
	}
	var stepsToExecute []parser.Step
	if output == "true" {
		stepsToExecute = ifBlock.Then
	} else {
		stepsToExecute = ifBlock.Else
	}
	return tr.executeSteps(stepsToExecute, executor, nil, silent, stepResults, context, testCase)
}

// executeForLoop executes a for loop, collecting StepResults
func (tr *TestRunner) executeForLoop(forBlock *parser.LoopBlock, executor *actions.ActionExecutor, silent bool, stepResults *[]parser.StepResult, context string, testCase *parser.TestCase) error {
	condition := tr.substituteString(forBlock.Condition)
	output, err := executor.Execute("control", []interface{}{"for", condition}, silent)
	if err != nil {
		return fmt.Errorf("failed to evaluate for loop condition: %w", err)
	}
	iterations, err := strconv.Atoi(output)
	if err != nil {
		return fmt.Errorf("failed to parse iteration count: %w", err)
	}
	maxIterations := forBlock.MaxIterations
	if maxIterations > 0 && iterations > maxIterations {
		iterations = maxIterations
	}
	for i := 0; i < iterations; i++ {
		tr.variables["iteration"] = i + 1
		tr.variables["index"] = i
		if err := tr.executeSteps(forBlock.Steps, executor, forBlock, silent, stepResults, context, testCase); err != nil {
			return fmt.Errorf("iteration %d failed: %w", i+1, err)
		}
	}
	return nil
}

// executeWhileLoop executes a while loop, collecting StepResults
func (tr *TestRunner) executeWhileLoop(whileBlock *parser.LoopBlock, executor *actions.ActionExecutor, silent bool, stepResults *[]parser.StepResult, context string, testCase *parser.TestCase) error {
	iteration := 0
	maxIterations := whileBlock.MaxIterations
	if maxIterations <= 0 {
		maxIterations = 1000
	}
	for {
		iteration++
		if iteration > maxIterations {
			return fmt.Errorf("while loop exceeded maximum iterations (%d)", maxIterations)
		}
		tr.variables["iteration"] = iteration
		condition := tr.substituteString(whileBlock.Condition)
		output, err := executor.Execute("control", []interface{}{"while", condition}, silent)
		if err != nil {
			return fmt.Errorf("failed to evaluate while condition: %w", err)
		}
		if output != "true" {
			break
		}
		if err := tr.executeSteps(whileBlock.Steps, executor, whileBlock, silent, stepResults, context, testCase); err != nil {
			return fmt.Errorf("while iteration %d failed: %w", iteration, err)
		}
	}
	return nil
}

// initializeVariables initializes variables from the test case
func (tr *TestRunner) initializeVariables(testCase *parser.TestCase) {
	fmt.Println("DEBUG RUNNER: Initializing variables...")
	if testCase.Variables.Regular != nil {
		for name, value := range testCase.Variables.Regular {
			tr.variables[name] = value
		}
	}
	if testCase.Variables.Secrets != nil {
		secretMap := make(map[string]interface{})
		for name, secret := range testCase.Variables.Secrets {
			secretMap[name] = map[string]interface{}{
				"value":       secret.Value,
				"file":        secret.File,
				"mask_output": secret.MaskOutput,
			}
		}
		if err := tr.secretManager.ResolveSecrets(secretMap); err != nil {
			fmt.Printf("⚠️  Warning: Failed to resolve secrets: %v\n", err)
		}
		for name := range testCase.Variables.Secrets {
			if value, exists := tr.secretManager.GetSecret(name); exists {
				tr.variables[name] = value
			}
		}
	}
	fmt.Println("DEBUG RUNNER: Finished initializing variables.")
}

// initializeTDM initializes test data management
func (tr *TestRunner) initializeTDM(testCase *parser.TestCase) {
	fmt.Println("DEBUG RUNNER: Initializing TDM...")
	if testCase.DataManagement == nil {
		return
	}
	if len(testCase.DataManagement.DataSets) > 0 {
		if err := tr.tdmManager.LoadDataSets(testCase.DataManagement.DataSets); err != nil {
			fmt.Printf("⚠️  Warning: Failed to load data sets: %v\n", err)
		}
	}
	if len(testCase.Environments) > 0 {
		if err := tr.tdmManager.LoadEnvironments(testCase.Environments); err != nil {
			fmt.Printf("⚠️  Warning: Failed to load environments: %v\n", err)
		}
	}
	if testCase.DataManagement.Environment != "" {
		if err := tr.tdmManager.SetEnvironment(testCase.DataManagement.Environment); err != nil {
			fmt.Printf("⚠️  Warning: Failed to set environment '%s': %v\n", testCase.DataManagement.Environment, err)
		}
	}
	if len(testCase.DataManagement.Validation) > 0 {
		validationResults := tr.tdmManager.ValidateData(testCase.DataManagement.Validation)
		for _, result := range validationResults {
			if result.Status == "FAILED" {
				fmt.Printf("❌ Data validation failed: %s - %s\n", result.Name, result.Message)
			} else if result.Status == "WARNING" {
				fmt.Printf("⚠️  Data validation warning: %s - %s\n", result.Name, result.Message)
			}
		}
	}
	for name, value := range tr.tdmManager.GetAllVariables() {
		tr.variables[name] = value
	}
	fmt.Println("DEBUG RUNNER: Finished initializing TDM.")
}

// substituteVariables replaces ${variable} references with actual values
func (tr *TestRunner) substituteVariables(args []interface{}) []interface{} {
	substituted := make([]interface{}, len(args))

	for i, arg := range args {
		switch v := arg.(type) {
		case string:
			substituted[i] = tr.substituteString(v)
		case []interface{}:
			substituted[i] = tr.substituteVariables(v)
		case map[string]interface{}:
			substituted[i] = tr.substituteMap(v)
		default:
			substituted[i] = arg
		}
	}

	return substituted
}

// substituteString replaces ${variable} references in a string
func (tr *TestRunner) substituteString(s string) string {
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		varName := strings.TrimPrefix(strings.TrimSuffix(match, "}"), "${")
		if value, ok := tr.resolveDotNotation(varName); ok {
			return fmt.Sprintf("%v", value)
		}
		return match // Return original if variable not found
	})
}

// resolveDotNotation resolves variables with dot notation (e.g., response.status_code)
func (tr *TestRunner) resolveDotNotation(varName string) (interface{}, bool) {
	// First, check for the full key (flat variable)
	if value, exists := tr.variables[varName]; exists {
		return value, true
	}
	// Fallback to dot notation logic
	parts := strings.Split(varName, ".")
	value, exists := tr.variables[parts[0]]
	if !exists {
		return nil, false
	}
	if len(parts) == 1 {
		return value, true
	}
	// Try to parse JSON string to map if needed
	var m map[string]interface{}
	switch v := value.(type) {
	case string:
		if err := json.Unmarshal([]byte(v), &m); err == nil {
			value = m
		} else {
			return v, true // not JSON, return as is
		}
	case map[string]interface{}:
		m = v
	default:
		return value, true
	}
	// Traverse the map for each field
	for _, field := range parts[1:] {
		if m2, ok := value.(map[string]interface{}); ok {
			if v, ok := m2[field]; ok {
				value = v
				// If nested JSON string, try to parse again
				if s, isStr := v.(string); isStr && json.Valid([]byte(s)) {
					var nested map[string]interface{}
					if err := json.Unmarshal([]byte(s), &nested); err == nil {
						value = nested
					}
				}
			} else {
				return nil, false
			}
		} else {
			return nil, false
		}
	}
	return value, true
}

// substituteStringForDisplay replaces ${variable} references and masks secrets for display
func (tr *TestRunner) substituteStringForDisplay(s string) string {
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	result := re.ReplaceAllStringFunc(s, func(match string) string {
		varName := strings.TrimPrefix(strings.TrimSuffix(match, "}"), "${")
		if value, exists := tr.variables[varName]; exists {
			return fmt.Sprintf("%v", value)
		}
		return match // Return original if variable not found
	})

	// Mask secrets in the result for display only
	return tr.secretManager.MaskSecretsInString(result)
}

// substituteMap substitutes variables in map values
func (tr *TestRunner) substituteMap(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		switch val := v.(type) {
		case string:
			result[k] = tr.substituteString(val)
		case []interface{}:
			result[k] = tr.substituteVariables(val)
		case map[string]interface{}:
			result[k] = tr.substituteMap(val)
		default:
			result[k] = v
		}
	}
	return result
}

// executeStepWithRetry executes a step with retry logic if configured
func (tr *TestRunner) executeStepWithRetry(step parser.Step, args []interface{}, executor *actions.ActionExecutor, silent bool) (string, error) {
	// If no retry configuration, execute normally
	if step.Retry == nil {
		return executor.Execute(step.Action, args, silent)
	}

	// Validate retry configuration
	if err := parser.ValidateRetryConfig(step.Retry); err != nil {
		return "", fmt.Errorf("invalid retry configuration: %w", err)
	}

	// Merge with defaults
	retryConfig := parser.MergeRetryConfig(step.Retry)

	// Initialize retry result
	retryResult := parser.RetryResult{
		RetryLogs: make([]string, 0),
	}

	startTime := time.Now()
	var lastError error
	var lastOutput string

	// Execute with retries
	for attempt := 1; attempt <= retryConfig.Attempts; attempt++ {
		// Execute the action
		output, err := executor.Execute(step.Action, args, silent)
		lastOutput = output

		// If successful, return immediately
		if err == nil {
			retryResult.Success = true
			retryResult.Attempts = attempt
			retryResult.TotalTime = time.Since(startTime)

			if attempt > 1 && !silent {
				fmt.Printf("✅ %s\n", parser.FormatRetrySummary(retryResult, false))
			}
			return output, nil
		}

		lastError = err

		// Check if we should retry based on conditions
		if !parser.ShouldRetry(err, output, retryConfig.Conditions) {
			break
		}

		// If this is the last attempt, don't retry
		if attempt == retryConfig.Attempts {
			break
		}

		// Calculate delay for next attempt
		delay := parser.CalculateDelay(retryConfig.Delay, attempt, retryConfig.Backoff, retryConfig.MaxDelay, retryConfig.Jitter)

		// Log retry attempt
		retryLog := parser.FormatRetryLog(attempt, retryConfig.Attempts, delay, err, false)
		retryResult.RetryLogs = append(retryResult.RetryLogs, retryLog)

		if !silent {
			fmt.Printf("%s\n", retryLog)
		}

		// Wait before next attempt
		time.Sleep(delay)
	}

	// All attempts failed
	retryResult.Success = false
	retryResult.Attempts = retryConfig.Attempts
	retryResult.TotalTime = time.Since(startTime)
	retryResult.LastError = lastError
	retryResult.LastOutput = lastOutput

	if !silent {
		fmt.Printf("❌ %s\n", parser.FormatRetrySummary(retryResult, false))
	}

	return lastOutput, lastError
}

// validateExpectedError validates that an error occurred as expected
func (tr *TestRunner) validateExpectedError(expectError interface{}, actualErr error, output string, silent bool) error {
	var errorType string
	var expectedMessage string

	// Parse expect_error configuration
	switch v := expectError.(type) {
	case string:
		// Simple string format - check if it's "any" or default to "contains"
		if v == "any" {
			errorType = "any"
			expectedMessage = ""
		} else {
			errorType = "contains"
			expectedMessage = v
		}
	case map[string]interface{}:
		// Detailed configuration
		if t, ok := v["type"].(string); ok {
			errorType = t
		} else {
			errorType = "any" // Default to any if type not specified
		}
		if msg, ok := v["message"].(string); ok {
			expectedMessage = msg
		}
	default:
		return fmt.Errorf("invalid expect_error format: must be string or object")
	}

	// If no error occurred but we expected one, that's a failure
	if actualErr == nil {
		if errorType == "any" {
			return fmt.Errorf("expected any error but action succeeded with result: '%s'", output)
		}
		return fmt.Errorf("expected error but action succeeded with result: '%s'", output)
	}

	actualErrorMsg := actualErr.Error()

	// Validate the error based on error_type
	var result bool

	switch errorType {
	case "any":
		// Any error is acceptable
		result = true
	case "contains":
		result = strings.Contains(actualErrorMsg, expectedMessage)
	case "not_contains":
		result = !strings.Contains(actualErrorMsg, expectedMessage)
	case "matches":
		// Regex pattern matching
		matched, err := regexp.MatchString(expectedMessage, actualErrorMsg)
		if err != nil {
			return fmt.Errorf("invalid regex pattern '%s': %v", expectedMessage, err)
		}
		result = matched
	case "not_matches":
		// Regex pattern not matching
		matched, err := regexp.MatchString(expectedMessage, actualErrorMsg)
		if err != nil {
			return fmt.Errorf("invalid regex pattern '%s': %v", expectedMessage, err)
		}
		result = !matched
	case "exact":
		result = actualErrorMsg == expectedMessage
	case "starts_with":
		result = strings.HasPrefix(actualErrorMsg, expectedMessage)
	case "ends_with":
		result = strings.HasSuffix(actualErrorMsg, expectedMessage)
	default:
		return fmt.Errorf("unsupported error type: %s (supported: any, contains, not_contains, matches, not_matches, exact, starts_with, ends_with)", errorType)
	}

	if !result {
		if errorType == "any" {
			return fmt.Errorf("error expectation failed: expected any error but got '%s'", actualErrorMsg)
		}
		return fmt.Errorf("error expectation failed: '%s' %s '%s'", actualErrorMsg, errorType, expectedMessage)
	}

	return nil
}
