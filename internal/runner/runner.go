package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/JianLoong/robogo/internal/actions"
	"github.com/JianLoong/robogo/internal/output"
	"github.com/JianLoong/robogo/internal/parser"
	"github.com/JianLoong/robogo/internal/util"
)

// TestRunner runs test cases - DEPRECATED: Use TestExecutionService instead
// This struct is kept for backward compatibility with existing code
// Implements SkipEvaluator interface
type TestRunner struct {
	variableManager VariableManagerInterface // Variable management
	outputCapture   OutputManager           // Output capture
	retryManager    RetryPolicy             // Retry logic
	secretManager   *actions.SecretManager  // Secret manager for handling secrets
	tdmManager      *actions.TDMManager     // TDM manager for test data management
	executor        *actions.ActionExecutor // Action executor (injected)
}

// NewTestRunner creates a new test runner - DEPRECATED: Use NewTestExecutionService instead
func NewTestRunner(executor *actions.ActionExecutor) SkipEvaluator {
	return &TestRunner{
		variableManager: NewVariableManager(),
		outputCapture:   NewOutputCapture(),
		retryManager:    NewRetryManager(),
		secretManager:   actions.NewSecretManager(),
		tdmManager:      actions.NewTDMManager(),
		executor:        executor,
	}
}

// RunTestFiles runs multiple test cases in parallel
func RunTestFiles(ctx context.Context, paths []string, silent bool, executor *actions.ActionExecutor) ([]*parser.TestResult, error) {
	return RunTestFilesWithConfig(ctx, paths, silent, nil, executor)
}

// RunTestFilesWithConfig runs multiple test cases with parallelism configuration
func RunTestFilesWithConfig(ctx context.Context, paths []string, silent bool, parallelConfig *parser.ParallelConfig, executor *actions.ActionExecutor) ([]*parser.TestResult, error) {
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
			result, err := RunTestFile(ctx, file, silent, executor)
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

			result, err := RunTestFile(ctx, file, silent, executor)
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
		output.PrintParallelFiles(len(files), config.MaxConcurrency)
	}

	return results, nil
}

// RunTestFile runs a test case from a file
func RunTestFile(ctx context.Context, filename string, silent bool, executor *actions.ActionExecutor) (*parser.TestResult, error) {
	// Parse the test case
	testCase, err := parser.ParseTestFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to parse test file: %w", err)
	}

	// Run the test case
	result, err := RunTestCase(ctx, testCase, silent, executor)
	if err != nil {
		// RunTestCase returns an error when the test fails, but we still want to return the result
		// The error just indicates test failure, not a fatal error
		return result, nil // Return the result, not the error
	}
	return result, nil
}

// RunTestCase runs a test case and returns the result
func RunTestCase(ctx context.Context, testCase *parser.TestCase, silent bool, executor *actions.ActionExecutor) (*parser.TestResult, error) {
	// Create test execution service with proper dependency injection
	testService := NewTestExecutionService(executor)
	defer testService.Cleanup()

	// Execute the test case using the new service
	result, err := testService.ExecuteTestCase(ctx, testCase, silent)
	if err != nil {
		// Return the result even on error - the error indicates test failure, not fatal error
		return result, nil
	}
	return result, nil
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
			output.PrintWarning("Failed to load data sets: %v", err)
		}
	}
	if len(testCase.Environments) > 0 {
		if err := tr.tdmManager.LoadEnvironments(testCase.Environments); err != nil {
			output.PrintWarning("Failed to load environments: %v", err)
		}
	}
	if testCase.DataManagement.Environment != "" {
		if err := tr.tdmManager.SetEnvironment(testCase.DataManagement.Environment); err != nil {
			output.PrintWarning("Failed to set environment '%s': %v", testCase.DataManagement.Environment, err)
		}
	}
	if len(testCase.DataManagement.Validation) > 0 {
		validationResults := tr.tdmManager.ValidateData(testCase.DataManagement.Validation)
		for _, result := range validationResults {
			if result.Status == parser.StatusFailed {
				output.PrintDataValidationFailure(result.Name, result.Message)
			} else if result.Status == "WARNING" {
				output.PrintDataValidationWarning(result.Name, result.Message)
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
	return tr.variableManager.SubstituteString(s)
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
