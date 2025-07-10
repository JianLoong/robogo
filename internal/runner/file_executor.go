package runner

import (
	"context"

	"github.com/JianLoong/robogo/internal/actions"
	"github.com/JianLoong/robogo/internal/parser"
)

// DefaultFileExecutor implements TestFileExecutor interface
// Provides file-based test execution capabilities
type DefaultFileExecutor struct {
	testExecutor TestExecutor
}

// NewFileExecutor creates a new file executor
func NewFileExecutor(executor *actions.ActionExecutor) TestFileExecutor {
	return &DefaultFileExecutor{
		testExecutor: NewTestExecutionService(executor),
	}
}

// NewFileExecutorWithTestExecutor creates a file executor with existing test executor
func NewFileExecutorWithTestExecutor(testExecutor TestExecutor) TestFileExecutor {
	return &DefaultFileExecutor{
		testExecutor: testExecutor,
	}
}

// RunTestFile runs a test case from a file
func (fe *DefaultFileExecutor) RunTestFile(ctx context.Context, filename string, silent bool, executor *actions.ActionExecutor) (*parser.TestResult, error) {
	// Parse the test case
	testCase, err := parser.ParseTestFile(filename)
	if err != nil {
		return nil, err
	}

	// Execute the test case
	return fe.testExecutor.ExecuteTestCase(ctx, testCase, silent)
}

// RunTestFiles runs multiple test cases in parallel
func (fe *DefaultFileExecutor) RunTestFiles(ctx context.Context, paths []string, silent bool, executor *actions.ActionExecutor) ([]*parser.TestResult, error) {
	return fe.RunTestFilesWithConfig(ctx, paths, silent, nil, executor)
}

// RunTestFilesWithConfig runs multiple test cases with parallelism configuration
func (fe *DefaultFileExecutor) RunTestFilesWithConfig(ctx context.Context, paths []string, silent bool, parallelConfig *parser.ParallelConfig, executor *actions.ActionExecutor) ([]*parser.TestResult, error) {
	// Use the existing implementation from runner.go
	return RunTestFilesWithConfig(ctx, paths, silent, parallelConfig, executor)
}

// GetTestExecutor returns the underlying test executor
func (fe *DefaultFileExecutor) GetTestExecutor() TestExecutor {
	return fe.testExecutor
}