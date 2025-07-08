package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/JianLoong/robogo/internal/actions"
	"github.com/JianLoong/robogo/internal/parser"
	"github.com/JianLoong/robogo/internal/util"
)

// TestRunner runs test cases
type TestRunner struct {
	variableManager *VariableManager       // Variable management
	outputCapture   *OutputCapture         // Output capture
	retryManager    *RetryManager          // Retry logic
	secretManager   *actions.SecretManager // Secret manager for handling secrets
	tdmManager      *actions.TDMManager    // TDM manager for test data management
}

// NewTestRunner creates a new test runner
func NewTestRunner() *TestRunner {
	return &TestRunner{
		variableManager: NewVariableManager(),
		outputCapture:   NewOutputCapture(),
		retryManager:    NewRetryManager(),
		secretManager:   actions.NewSecretManager(),
		tdmManager:      actions.NewTDMManager(),
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
				result = &parser.TestResult{
					TestCase:     &parser.TestCase{Name: file},
					Status:       parser.StatusFailed,
					ErrorMessage: util.FormatRobogoError(err),
				}
			}
			results = append(results, result)
		}
		return results, nil
	}

	// Run in parallel with concurrency control
	var wg sync.WaitGroup
	resultsChan := make(chan *parser.TestResult, len(files))
	semaphore := make(chan struct{}, config.MaxConcurrency)

	if !silent {
		fmt.Printf("Running %d test files in parallel (max concurrency: %d)\n", len(files), config.MaxConcurrency)
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
				// In case of a fatal error before the test can even run properly,
				// create a dummy result to report the failure.
				result = &parser.TestResult{
					TestCase:     &parser.TestCase{Name: file},
					Status:       parser.StatusFailed,
					ErrorMessage: util.FormatRobogoError(err),
				}
			}
			resultsChan <- result
		}(file)
	}

	wg.Wait()
	close(resultsChan)

	var results []*parser.TestResult
	for result := range resultsChan {
		results = append(results, result)
	}

	if !silent {
		PrintParallelFiles(len(files), config.MaxConcurrency)
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

	// Run the test case
	result, err := RunTestCase(testCase, silent)
	if err != nil {
		// RunTestCase returns an error when the test fails, but we still want to return the result
		// The error just indicates test failure, not a fatal error
		return result, nil // Return the result, not the error
	}
	return result, nil
}

// RunTestCase runs a test case and returns the result
func RunTestCase(testCase *parser.TestCase, silent bool) (*parser.TestResult, error) {
	tr := NewTestRunner()

	// Create execution engine
	engine := NewExecutionEngine(tr)

	// Execute the test case using the engine
	result, err := engine.ExecuteTestCase(testCase, silent)
	fmt.Printf("[DEBUG] RunTestCase: after ExecuteTestCase for test case: %s\n", testCase.Name)
	if err != nil {
		fmt.Printf("[DEBUG] RunTestCase: returning early with error for test case: %s\n", testCase.Name)
		// RunTestCase returns an error when the test fails, but we still want to return the result
		// The error just indicates test failure, not a fatal error
		return result, nil // Return the result, not the error
	}
	fmt.Printf("[DEBUG] RunTestCase: returning normally for test case: %s\n", testCase.Name)
	return result, nil
}

// executeSteps executes a slice of steps, collecting StepResults recursively
func executeSteps(tr *TestRunner, steps []parser.Step, executor *actions.ActionExecutor, parentLoop *parser.LoopBlock, silent bool, stepResults *[]parser.StepResult, context string, testCase *parser.TestCase) error {
	return executeStepsWithConfig(tr, steps, executor, parentLoop, silent, stepResults, context, testCase, nil)
}

// executeStepsWithConfig executes steps with parallelism configuration
func executeStepsWithConfig(tr *TestRunner, steps []parser.Step, executor *actions.ActionExecutor, parentLoop *parser.LoopBlock, silent bool, stepResults *[]parser.StepResult, context string, testCase *parser.TestCase, parallelConfig *parser.ParallelConfig) error {
	// Check if parallel step execution is enabled
	config := parser.MergeParallelConfig(parallelConfig)
	if config.Enabled && config.Steps && len(steps) > 1 {
		fmt.Printf("[DEBUG] executeStepsWithConfig: running steps in parallel for test case: %s\n", testCase.Name)
		return executeStepsParallel(tr, steps, executor, parentLoop, silent, stepResults, context, testCase, config)
	}

	// Execute steps sequentially (original behavior)
	for idx, step := range steps {
		stepResult, err := executeSingleStep(tr, step, executor, parentLoop, silent, stepResults, context, testCase, idx)
		if err != nil {
			if actions.IsSkipError(err) {
				*stepResults = append(*stepResults, *stepResult)
				fmt.Printf("[DEBUG] executeStepsWithConfig: skip error, returning for test case: %s\n", testCase.Name)
				return err // propagate skip error up
			}
			if step.ContinueOnFailure {
				// Log and continue to next step
				if !silent {
					fmt.Printf("Step '%s' failed but continuing due to continue_on_failure\n", stepResult.Step.Name)
				}
				continue
			} else {
				fmt.Printf("[DEBUG] executeStepsWithConfig: step failed, returning for test case: %s\n", testCase.Name)
				return fmt.Errorf("step '%s' failed: %w", stepResult.Step.Name, err)
			}
		}
		*stepResults = append(*stepResults, *stepResult)
	}
	fmt.Printf("[DEBUG] executeStepsWithConfig: all steps executed for test case: %s\n", testCase.Name)
	return nil
}

// executeStepsParallel executes independent steps in parallel
func executeStepsParallel(tr *TestRunner, steps []parser.Step, executor *actions.ActionExecutor, parentLoop *parser.LoopBlock, silent bool, stepResults *[]parser.StepResult, context string, testCase *parser.TestCase, config *parser.ParallelConfig) error {
	// Group steps into parallel and sequential groups
	stepGroups := parser.GroupIndependentSteps(steps)

	if !silent {
		PrintParallelStepGroups(len(stepGroups))
	}

	for groupIdx, group := range stepGroups {
		if len(group) == 1 {
			// Single step - execute sequentially
			step := group[0]
			stepResult, err := executeSingleStep(tr, step, executor, parentLoop, silent, stepResults, context, testCase, groupIdx)
			if err != nil {
				return err
			}
			*stepResults = append(*stepResults, *stepResult)
		} else {
			// Multiple independent steps - execute in parallel
			if !silent {
				PrintParallelSteps(len(group), groupIdx)
			}

			if err := executeStepGroupParallel(tr, group, executor, parentLoop, silent, stepResults, context, testCase, config, groupIdx); err != nil {
				return err
			}
		}
	}

	return nil
}

// executeStepGroupParallel executes a group of independent steps in parallel
func executeStepGroupParallel(tr *TestRunner, steps []parser.Step, executor *actions.ActionExecutor, parentLoop *parser.LoopBlock, silent bool, stepResults *[]parser.StepResult, context string, testCase *parser.TestCase, config *parser.ParallelConfig, groupIdx int) error {
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

			stepResult, err := executeSingleStep(tr, step, executor, parentLoop, silent, stepResults, context, testCase, groupIdx)
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

// initializeVariables initializes variables from the test case
func (tr *TestRunner) initializeVariables(testCase *parser.TestCase) {
	tr.variableManager.InitializeVariables(testCase)
}

// initializeTDM initializes test data management
func (tr *TestRunner) initializeTDM(testCase *parser.TestCase) {
	if testCase.DataManagement == nil {
		return
	}
	if len(testCase.DataManagement.DataSets) > 0 {
		if err := tr.tdmManager.LoadDataSets(testCase.DataManagement.DataSets); err != nil {
			PrintWarning("Failed to load data sets: %v", err)
		}
	}
	if len(testCase.Environments) > 0 {
		if err := tr.tdmManager.LoadEnvironments(testCase.Environments); err != nil {
			PrintWarning("Failed to load environments: %v", err)
		}
	}
	if testCase.DataManagement.Environment != "" {
		if err := tr.tdmManager.SetEnvironment(testCase.DataManagement.Environment); err != nil {
			PrintWarning("Failed to set environment '%s': %v", testCase.DataManagement.Environment, err)
		}
	}
	if len(testCase.DataManagement.Validation) > 0 {
		validationResults := tr.tdmManager.ValidateData(testCase.DataManagement.Validation)
		for _, result := range validationResults {
			if result.Status == parser.StatusFailed {
				PrintDataValidationFailure(result.Name, result.Message)
			} else if result.Status == "WARNING" {
				PrintDataValidationWarning(result.Name, result.Message)
			}
		}
	}
	for name, value := range tr.tdmManager.GetAllVariables() {
		tr.variableManager.SetVariable(name, value)
	}
}

// substituteVariables replaces ${variable} references with actual values
func (tr *TestRunner) substituteVariables(args []interface{}) []interface{} {
	return tr.variableManager.SubstituteVariables(args)
}

// substituteString replaces ${variable} references in a string
func (tr *TestRunner) substituteString(s string) string {
	return tr.variableManager.substituteString(s)
}

// resolveDotNotation resolves variables with dot notation (e.g., response.status_code)
func (tr *TestRunner) resolveDotNotation(varName string) (interface{}, bool) {
	return tr.variableManager.resolveDotNotation(varName)
}

// substituteStringForDisplay replaces ${variable} references and masks secrets for display
func (tr *TestRunner) substituteStringForDisplay(s string) string {
	result := tr.variableManager.substituteStringForDisplay(s)
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
