# Robogo Architecture Documentation

## Overview

Robogo uses a modern, interface-driven architecture that emphasizes decoupling, testability, and maintainability. The architecture has been completely refactored to eliminate global state, remove tight coupling, and provide clean separation of concerns through dependency injection and service-oriented design.

## Core Architecture Principles

### 1. Interface-Driven Design
- All major components depend on interfaces, not concrete implementations
- Clear contracts between components ensure API stability
- Easy to mock and test individual components

### 2. Dependency Injection
- Services receive their dependencies through constructors
- No global state or singleton patterns
- Context-aware execution with proper resource management

### 3. Service-Oriented Architecture
- Functionality is organized into focused services
- Each service has a single responsibility
- Services communicate through well-defined interfaces

### 4. Factory Pattern
- Centralized service creation with proper dependency wiring
- Consistent configuration across all services
- Easy to extend with new service types

## Core Components

### Service Interfaces

#### TestExecutor
Primary interface for test case and suite execution.

```go
type TestExecutor interface {
    ExecuteTestCase(ctx context.Context, testCase *parser.TestCase, silent bool) (*parser.TestResult, error)
    ExecuteTestSuite(ctx context.Context, testSuite *parser.TestSuite, filePath string, silent bool) (*parser.TestSuiteResult, error)
    GetContext() ExecutionContext
    GetExecutor() *actions.ActionExecutor
    ShouldSkipTestCase(testCase *parser.TestCase, context string) SkipInfo
    Cleanup() error
}
```

**Implementation:** `TestExecutionService`
- Manages complete test case lifecycle
- Handles TDM setup/teardown
- Provides step execution coordination

#### StepExecutor
Interface for individual step execution with parallel support.

```go
type StepExecutor interface {
    ExecuteStep(ctx context.Context, step parser.Step, silent bool) (*parser.StepResult, error)
    ExecuteSteps(ctx context.Context, steps []parser.Step, silent bool) ([]parser.StepResult, error)
    ExecuteStepsParallel(ctx context.Context, steps []parser.Step, config *parser.ParallelConfig, silent bool) ([]parser.StepResult, error)
}
```

**Implementation:** `StepExecutionService`
- Handles control flow (if/for/while statements)
- Manages parallel step execution
- Provides step dependency analysis

#### ExecutionContext
Central context for managing execution dependencies.

```go
type ExecutionContext interface {
    Variables() VariableStore
    Secrets() SecretStore
    Output() OutputHandler
    Retry() RetryHandler
    TestData() TDMHandler
    Actions() ActionExecutor
    Cleanup() error
}
```

**Implementation:** `DefaultExecutionContext`
- Dependency injection container
- Resource lifecycle management
- Provides adapters for existing components

### Supporting Interfaces

#### VariableManagerInterface
```go
type VariableManagerInterface interface {
    InitializeVariables(testCase *parser.TestCase)
    SetVariable(name string, value interface{})
    GetVariable(name string) (interface{}, bool)
    SubstituteVariables(args []interface{}) []interface{}
    SubstituteString(s string) string
    resolveDotNotation(varName string) (interface{}, bool)
    substituteStringForDisplay(s string) string
}
```

#### OutputManager
```go
type OutputManager interface {
    StartCapture()
    StopCapture() string
    Write(data []byte) (int, error)
    Capture() ([]byte, error)
}
```

#### RetryPolicy
```go
type RetryPolicy interface {
    ShouldRetry(step parser.Step, attempt int, err error) bool
    GetRetryDelay(attempt int) time.Duration
    ExecuteWithRetry(ctx context.Context, step parser.Step, executor ActionExecutor, silent bool) (interface{}, error)
}
```

### Factory Services

#### ServiceFactory
Centralized service creation with proper dependency injection.

```go
type ServiceFactory interface {
    CreateTestExecutor(executor *actions.ActionExecutor) TestExecutor
    CreateStepExecutor(context ExecutionContext) StepExecutor
    CreateTestSuiteExecutor(runner TestExecutor) TestSuiteExecutor
    CreateVariableManager() VariableManagerInterface
    CreateOutputManager() OutputManager
    CreateRetryPolicy() RetryPolicy
}
```

**Usage Example:**
```go
factory := runner.NewServiceFactory()
executor := actions.NewActionExecutor()
testExecutor := factory.CreateTestExecutor(executor)
suiteExecutor := factory.CreateTestSuiteExecutor(testExecutor)
```

## Architecture Layers

### 1. Interface Layer
- Defines contracts and abstractions
- Located in `internal/runner/interfaces.go`
- No implementation details

### 2. Service Layer
- Implements business logic
- Uses dependency injection
- Examples: `TestExecutionService`, `StepExecutionService`

### 3. Adapter Layer
- Bridges new interfaces with existing components
- Provides backward compatibility
- Examples: `variableStoreAdapter`, `outputHandlerAdapter`

### 4. Factory Layer
- Creates and wires services
- Manages dependencies
- Examples: `DefaultServiceFactory`, `ContextProviderImpl`

## Component Relationships

```
┌─────────────────────────────────────────────────────────────┐
│                    Client Code (main.go)                    │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                  ServiceFactory                             │
│  ┌─────────────────────────────────────────────────────────┐│
│  │        Creates and wires all services                   ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                Service Interfaces                           │
│  ┌─────────────┐ ┌──────────────┐ ┌─────────────────────┐   │
│  │TestExecutor │ │ StepExecutor │ │ TestSuiteExecutor   │   │
│  └─────────────┘ └──────────────┘ └─────────────────────┘   │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│              Service Implementations                        │
│  ┌─────────────────┐ ┌───────────────────┐ ┌─────────────┐  │
│  │TestExecService  │ │ StepExecService   │ │OutputCapture│  │
│  └─────────────────┘ └───────────────────┘ └─────────────┘  │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                ExecutionContext                             │
│  ┌─────────────────────────────────────────────────────────┐│
│  │     Dependency injection container                      ││
│  │     Resource lifecycle management                       ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

## Validation Framework

### ValidationEngine Interface

Provides comprehensive validation for test cases, suites, and steps:

```go
type ValidationEngine interface {
    ValidateTestCase(testCase *parser.TestCase) []ValidationError
    ValidateTestSuite(testSuite *parser.TestSuite) []ValidationError
    ValidateStep(step parser.Step) []ValidationError
}
```

### ValidationError Structure

```go
type ValidationError struct {
    Type        string      // Error category
    Message     string      // Human-readable message
    Field       string      // Field that failed validation
    Value       interface{} // Invalid value
    Suggestions []string    // Actionable suggestions for fixing
}
```

### Validation Features

- **Required Field Validation**: Ensures mandatory fields are present
- **Action-Specific Validation**: Validates arguments based on action type
- **Control Flow Validation**: Prevents conflicting control statements
- **Detailed Error Messages**: Provides specific, actionable feedback
- **Suggestions**: Offers concrete steps to fix validation errors

**Example Usage:**
```go
validator := runner.NewValidationEngine()
errors := validator.ValidateTestCase(testCase)
for _, err := range errors {
    fmt.Printf("Error in %s: %s\n", err.Field, err.Message)
    for _, suggestion := range err.Suggestions {
        fmt.Printf("  Suggestion: %s\n", suggestion)
    }
}
```

## Migration Guide

### From Legacy Architecture

The architecture has been completely refactored to use interfaces and dependency injection. Here's how to migrate:

#### Before (Legacy)
```go
// Global state and tight coupling
tr := runner.NewTestRunner(executor)
engine := runner.NewExecutionEngine(tr)
result, err := engine.ExecuteTestCase(ctx, testCase, silent)
```

#### After (Interface-Based)
```go
// Interface-based with dependency injection
testExecutor := runner.NewTestExecutionService(executor)
result, err := testExecutor.ExecuteTestCase(ctx, testCase, silent)
```

### Service Creation

#### Using Factory Pattern
```go
factory := runner.NewServiceFactory()
testExecutor := factory.CreateTestExecutor(executor)
suiteExecutor := factory.CreateTestSuiteExecutor(testExecutor)
```

#### Direct Service Creation
```go
testExecutor := runner.NewTestExecutionService(executor)
stepExecutor := runner.NewStepExecutionService(context)
```

## Benefits of New Architecture

### 1. Improved Testability
- **Mock-Friendly**: All dependencies are interfaces that can be easily mocked
- **Isolated Testing**: Components can be tested in isolation
- **Dependency Injection**: Easy to provide test doubles

```go
// Example: Testing with mocks
mockStepExecutor := &MockStepExecutor{}
testService := &TestExecutionService{
    stepService: mockStepExecutor,
    context:     mockContext,
}
```

### 2. Enhanced Maintainability
- **Clear Contracts**: Interface contracts ensure API stability
- **Loose Coupling**: Changes to implementations don't affect clients
- **Single Responsibility**: Each service has a focused purpose

### 3. Flexible Configuration
- **Runtime Composition**: Services can be composed at runtime
- **Multiple Implementations**: Easy to provide alternative implementations
- **Plugin Architecture**: New services can be added without modification

### 4. Better Error Handling
- **Comprehensive Validation**: Proactive error detection
- **Detailed Error Messages**: Clear, actionable feedback
- **Error Recovery**: Structured approach to error handling

## Performance Considerations

### Service Creation
- Services are lightweight and stateless where possible
- Factory pattern enables efficient service reuse
- Context provides resource pooling and cleanup

### Memory Management
- ExecutionContext manages resource lifecycle
- Proper cleanup prevents memory leaks
- Services can be garbage collected when context is cleaned up

### Parallel Execution
- Interface-based design doesn't impact performance
- Parallel execution capabilities are preserved
- Service boundaries don't introduce overhead

## Future Extensibility

The interface-based architecture provides several extension points:

### 1. Custom Service Implementations
- Implement interfaces to provide custom behavior
- Register with factory for dependency injection
- No changes needed to existing code

### 2. Event-Driven Architecture
- EventPublisher interface enables event-driven patterns
- Services can publish events without tight coupling
- Easy to add monitoring and observability

### 3. Plugin System
- Interfaces enable plugin-like extensions
- Services can be loaded dynamically
- Configuration-driven service composition

### 4. Alternative Execution Engines
- TestExecutor interface allows multiple execution strategies
- Distributed execution can be implemented
- Cloud-native execution models possible

## Conclusion

The new interface-based architecture provides a solid foundation for Robogo's continued evolution. By emphasizing decoupling, testability, and maintainability, the architecture enables confident development and extension while preserving all existing functionality.

The comprehensive interface system, combined with proper dependency injection and service-oriented design, creates a flexible and robust platform for test automation that can adapt to future requirements and scale with growing complexity.