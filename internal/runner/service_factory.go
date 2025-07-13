package runner

import (
	"github.com/JianLoong/robogo/internal/actions"
)

// DefaultServiceFactory implements ServiceFactory interface
// Provides a centralized way to create service instances with proper dependencies
type DefaultServiceFactory struct {
	// Configuration or dependencies could be stored here
}

// NewServiceFactory creates a new service factory
func NewServiceFactory() ServiceFactory {
	return &DefaultServiceFactory{}
}

// CreateTestExecutor creates a new test executor instance
func (f *DefaultServiceFactory) CreateTestExecutor(executor *actions.ActionExecutor) TestExecutor {
	return NewTestExecutionService(executor)
}

// CreateStepExecutor creates a new step executor instance
func (f *DefaultServiceFactory) CreateStepExecutor(context TestExecutionContext) StepExecutor {
	return NewStepExecutionService(context)
}

// CreateTestSuiteExecutor creates a new test suite executor instance
func (f *DefaultServiceFactory) CreateTestSuiteExecutor(runner TestExecutor) TestSuiteExecutor {
	return NewTestSuiteRunner(runner)
}


// CreateOutputManager creates a new output manager instance
func (f *DefaultServiceFactory) CreateOutputManager() OutputManager {
	return NewOutputCapture()
}

// CreateRetryPolicy creates a new retry policy instance
func (f *DefaultServiceFactory) CreateRetryPolicy() RetryPolicy {
	return NewRetryManager()
}

// ContextProviderImpl implements ContextProvider interface
type ContextProviderImpl struct {
	context TestExecutionContext
}

// NewContextProvider creates a new context provider
func NewContextProvider() ContextProvider {
	return &ContextProviderImpl{}
}

// GetExecutionContext returns the current execution context
func (cp *ContextProviderImpl) GetExecutionContext() TestExecutionContext {
	return cp.context
}

// CreateContext creates a new execution context
func (cp *ContextProviderImpl) CreateContext(executor *actions.ActionExecutor) TestExecutionContext {
	context := NewTestExecutionContext(executor)
	cp.context = context
	return context
}

// WithContext sets the execution context
func (cp *ContextProviderImpl) WithContext(context TestExecutionContext) ContextProvider {
	return &ContextProviderImpl{
		context: context,
	}
}