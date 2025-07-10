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


// RunTestFiles runs multiple test cases in parallel
func RunTestFiles(ctx context.Context, paths []string, silent bool, executor *actions.ActionExecutor) ([]*parser.TestResult, error) {
	return RunTestFilesWithConfig(ctx, paths, silent, nil, executor)
}

// RunTestFilesWithConfig runs multiple test cases with parallelism configuration
func RunTestFilesWithConfig(ctx context.Context, paths []string, silent bool, parallelConfig *parser.ParallelConfig, executor *actions.ActionExecutor) ([]*parser.TestResult, error) {
	return RunTestFilesWithConfigAndDebug(ctx, paths, silent, parallelConfig, executor, false)
}

// RunTestFilesWithConfigAndDebug runs multiple test cases with parallelism configuration and optional variable debugging
func RunTestFilesWithConfigAndDebug(ctx context.Context, paths []string, silent bool, parallelConfig *parser.ParallelConfig, executor *actions.ActionExecutor, variableDebug bool) ([]*parser.TestResult, error) {
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
			result, err := RunTestFileWithDebug(ctx, file, silent, executor, variableDebug)
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

			result, err := RunTestFileWithDebug(ctx, file, silent, executor, variableDebug)
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
	return RunTestFileWithDebug(ctx, filename, silent, executor, false)
}

// RunTestFileWithDebug runs a test file with optional variable debugging
func RunTestFileWithDebug(ctx context.Context, filename string, silent bool, executor *actions.ActionExecutor, variableDebug bool) (*parser.TestResult, error) {
	// Parse the test case
	testCase, err := parser.ParseTestFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to parse test file: %w", err)
	}

	// Run the test case with debugging
	result, err := RunTestCaseWithDebug(ctx, testCase, silent, executor, variableDebug)
	if err != nil {
		// RunTestCase returns an error when the test fails, but we still want to return the result
		// The error just indicates test failure, not a fatal error
		return result, nil // Return the result, not the error
	}
	return result, nil
}

// RunTestCase runs a test case and returns the result
func RunTestCase(ctx context.Context, testCase *parser.TestCase, silent bool, executor *actions.ActionExecutor) (*parser.TestResult, error) {
	return RunTestCaseWithDebug(ctx, testCase, silent, executor, false)
}

// RunTestCaseWithDebug runs a test case with optional variable debugging
func RunTestCaseWithDebug(ctx context.Context, testCase *parser.TestCase, silent bool, executor *actions.ActionExecutor, variableDebug bool) (*parser.TestResult, error) {
	// Create test execution service with proper dependency injection
	testService := NewTestExecutionService(executor)
	defer testService.Cleanup()

	// Enable variable debugging if requested
	if variableDebug {
		testService.GetContext().EnableVariableDebugging(true)
	}

	// Execute the test case using the new service
	result, err := testService.ExecuteTestCase(ctx, testCase, silent)
	if err != nil {
		// Return the result even on error - the error indicates test failure, not fatal error
		return result, nil
	}
	return result, nil
}

