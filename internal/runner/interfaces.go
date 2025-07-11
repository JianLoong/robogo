package runner

import (
	"context"
	"time"

	"github.com/JianLoong/robogo/internal/actions"
	"github.com/JianLoong/robogo/internal/parser"
)

// TestExecutor defines the interface for test case execution
type TestExecutor interface {
	ExecuteTestCase(ctx context.Context, testCase *parser.TestCase, silent bool) (*parser.TestResult, error)
	ExecuteTestSuite(ctx context.Context, testSuite *parser.TestSuite, filePath string, silent bool) (*parser.TestSuiteResult, error)
	GetContext() ExecutionContext
	GetExecutor() *actions.ActionExecutor
	ShouldSkipTestCase(testCase *parser.TestCase, context string) SkipInfo
	Cleanup() error
}

// StepExecutor defines the interface for step execution
type StepExecutor interface {
	ExecuteStep(ctx context.Context, step parser.Step, silent bool) (*parser.StepResult, error)
	ExecuteSteps(ctx context.Context, steps []parser.Step, silent bool) ([]parser.StepResult, error)
	ExecuteStepsParallel(ctx context.Context, steps []parser.Step, config *parser.ParallelConfig, silent bool) ([]parser.StepResult, error)
}

// TestSuiteExecutor defines the interface for test suite execution
type TestSuiteExecutor interface {
	RunTestSuite(ctx context.Context, testSuite *parser.TestSuite, suiteFilePath string, printSummary bool) (*parser.TestSuiteResult, error)
}

// VariableManager interface for variable management operations
type VariableManagerInterface interface {
	InitializeVariables(testCase *parser.TestCase)
	SetVariable(name string, value interface{})
	GetVariable(name string) (interface{}, bool)
	SubstituteVariables(args []interface{}) []interface{}
	SubstituteString(s string) string
	resolveDotNotation(varName string) (interface{}, bool)
	substituteStringForDisplay(s string) string

	// Secret management methods
	MaskSensitiveOutput(output string) string
	IsSecretMasked(secretName string) bool
	GetSecretInfo(secretName string) (source string, masked bool, exists bool)
	ListSecrets() []string
}

// OutputManager interface for output capture and management
type OutputManager interface {
	StartCapture()
	StopCapture() string
	Write(data []byte) (int, error)
	Capture() ([]byte, error)
}

// RetryPolicy interface for retry logic
type RetryPolicy interface {
	ShouldRetry(step parser.Step, attempt int, err error) bool
	GetRetryDelay(attempt int) time.Duration
	ExecuteWithRetry(ctx context.Context, step parser.Step, executor ActionExecutor, silent bool) (interface{}, error)
}

// ContextProvider interface for providing execution context
type ContextProvider interface {
	GetExecutionContext() ExecutionContext
	CreateContext(executor *actions.ActionExecutor) ExecutionContext
	WithContext(context ExecutionContext) ContextProvider
}

// ServiceFactory interface for creating service instances
type ServiceFactory interface {
	CreateTestExecutor(executor *actions.ActionExecutor) TestExecutor
	CreateStepExecutor(context ExecutionContext) StepExecutor
	CreateTestSuiteExecutor(runner TestExecutor) TestSuiteExecutor
	CreateVariableManager() VariableManagerInterface
	CreateOutputManager() OutputManager
	CreateRetryPolicy() RetryPolicy
}

// ConfigManager interface for configuration management
type ConfigManager interface {
	GetParallelConfig() *parser.ParallelConfig
	GetRetryConfig() *parser.RetryConfig
	MergeParallelConfig(config *parser.ParallelConfig) *parser.ParallelConfig
	ValidateConfig() error
}

// ValidationEngine interface for data validation
type ValidationEngine interface {
	ValidateTestCase(testCase *parser.TestCase) []ValidationError
	ValidateTestSuite(testSuite *parser.TestSuite) []ValidationError
	ValidateStep(step parser.Step) []ValidationError
}

// ValidationError represents a validation error
type ValidationError struct {
	Type        string
	Message     string
	Field       string
	Value       interface{}
	Suggestions []string
}
